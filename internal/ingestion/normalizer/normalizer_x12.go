package normalizer

import (
	"log/slog"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type X12Normalizer struct{}

func NewX12Normalizer() *X12Normalizer {
	return &X12Normalizer{}
}

func (n *X12Normalizer) Normalize(raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	logger := slog.Default().With(
		"component", "x12_normalizer",
		"event_id", env.EventID,
	)

	logger.Info("starting X12 normalization")

	msg, err := models.ParseX12(string(raw))
	if err != nil {
		logger.Error("X12 parse failed", "error", err)
		return nil, err
	}

	isa := msg.ISA
	gs := msg.GS
	st := msg.ST
	nm1 := msg.NM1Patient
	clm := msg.CLM
	dtp := msg.DTPService

	var patientID, firstName, lastName string
	if len(nm1) > 8 {
		lastName = safeX12Field(nm1, 3)
		firstName = safeX12Field(nm1, 4)
		patientID = safeX12Field(nm1, 9)
	}

	encounterID := safeX12Field(clm, 1)
	observationCode := safeX12Field(st, 1)

	serviceDate := ""
	if len(dtp) > 3 {
		serviceDate = dtp[3]
	}

	logger.Info("parsed X12 fields",
		"patient_id", patientID,
		"first_name", firstName,
		"last_name", lastName,
		"encounter_id", encounterID,
		"observation_code", observationCode,
		"service_date", serviceDate,
	)

	ne := models.NewNormalizedEvent(config.FormatX12, raw)

	// Fields
	ne.Fields["isa.raw"] = isa
	ne.Fields["gs.raw"] = gs
	ne.Fields["st.transaction_set"] = observationCode
	ne.Fields["nm1.patient_id"] = patientID
	ne.Fields["nm1.first_name"] = firstName
	ne.Fields["nm1.last_name"] = lastName
	ne.Fields["clm.encounter_id"] = encounterID
	ne.Fields["dtp.service_date"] = serviceDate

	// Metadata
	ne.Metadata["event_id"] = env.EventID
	ne.Metadata["source_system"] = env.SourceSystem

	logger.Info("X12 normalization complete")
	return ne, nil
}

func safeX12Field(fields []string, idx int) string {
	if len(fields) > idx {
		return fields[idx]
	}
	return ""
}
