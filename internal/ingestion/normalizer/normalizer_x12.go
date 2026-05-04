package normalizer

import (
	"fmt"

	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type x12Normalizer struct{}

func NewX12Normalizer() Normalizer {
	return &x12Normalizer{}
}

func (n *x12Normalizer) Normalize(value any, meta api.Envelope) (*models.CanonicalEvent, error) {
	raw, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("expected X12 string, got %T", value)
	}

	msg, err := models.ParseX12(raw)
	if err != nil {
		return nil, err
	}

	// Patient Name: NM1*IL*Last*First
	pLast := safeX12(msg.NM1Patient, 3)
	pFirst := safeX12(msg.NM1Patient, 4)

	// Patient ID: NM1*IL*...*MemberID (field 9)
	pID := safeX12(msg.NM1Patient, 9)

	// Claim/Encounter ID: CLM01
	claimID := safeX12(msg.CLM, 1)

	// Event Type: ST01 (e.g., 837)
	eventType := safeX12(msg.ST, 1)

	// Service Date: DTP*472*D8*YYYYMMDD
	serviceDate := ""
	if len(msg.DTPService) > 3 {
		serviceDate = msg.DTPService[3]
	}

	_ = serviceDate // TODO: use service date in observation value or event timestamp

	patient := &models.CanonicalPatient{
		ID:        pID,
		FirstName: pFirst,
		LastName:  pLast,
	}

	encounter := &models.CanonicalEncounter{
		ID:   claimID,
		Type: eventType,
	}

	observation := &models.CanonicalObservation{
		Code:  eventType,
		Value: nil,
	}

	return &models.CanonicalEvent{
		EventID:      meta.EventID,
		SourceSystem: meta.SourceSystem,
		Format:       detector.FormatX12,
		Patient:      patient,
		Encounter:    encounter,
		Observation:  observation,
		RawValue:     raw,
	}, nil
}

// safeX12 returns field idx or empty string
func safeX12(fields []string, idx int) string {
	if idx >= len(fields) {
		return ""
	}
	return fields[idx]
}
