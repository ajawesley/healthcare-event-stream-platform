package normalizer

import (
	"context"
	"strings"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type HL7Normalizer struct{}

func NewHL7Normalizer() *HL7Normalizer {
	return &HL7Normalizer{}
}

func (n *HL7Normalizer) Normalize(_ context.Context, raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	ctx := context.Background()
	log := observability.WithTrace(ctx).With(
		zap.String("component", "hl7_normalizer"),
		zap.String("event_id", env.EventID),
	)

	log.Info("hl7_normalization_start")

	msg, err := models.ParseHL7(string(raw))
	if err != nil {
		log.Error("hl7_parse_failed", zap.Error(err))
		return nil, err
	}

	msh := msg.MSH
	pid := msg.PID
	pv1 := msg.PV1

	// MSH-9: ORU^R01
	messageType := safeHL7Field(msh, 8)

	// PID-3: Patient ID
	patientID := safeHL7Field(pid, 3)

	// PID-5: Doe^John
	var lastName, firstName string
	if len(pid) > 5 {
		name := pid[5]
		parts := strings.Split(name, "^")
		if len(parts) > 0 {
			lastName = parts[0]
		}
		if len(parts) > 1 {
			firstName = parts[1]
		}
	}

	// PV1-20: Encounter ID
	encounterID := safeHL7Field(pv1, 20)

	log.Debug("hl7_parsed_fields",
		zap.String("message_type", messageType),
		zap.String("patient_id", patientID),
		zap.String("first_name", firstName),
		zap.String("last_name", lastName),
		zap.String("encounter_id", encounterID),
	)

	ne := models.NewNormalizedEvent(config.FormatHL7, raw)

	// Fields
	ne.Fields["msh.message_type"] = messageType
	ne.Fields["pid.id"] = patientID
	ne.Fields["pid.first_name"] = firstName
	ne.Fields["pid.last_name"] = lastName
	ne.Fields["pv1.encounter_id"] = encounterID

	// Metadata
	ne.Metadata["event_id"] = env.EventID
	ne.Metadata["source_system"] = env.SourceSystem

	log.Info("hl7_normalization_complete",
		zap.Any("fields", ne.Fields),
		zap.Any("metadata", ne.Metadata),
	)

	return ne, nil
}

func safeHL7Field(fields []string, idx int) string {
	if len(fields) > idx {
		return fields[idx]
	}
	return ""
}
