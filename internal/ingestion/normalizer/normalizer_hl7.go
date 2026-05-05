package normalizer

import (
	"log/slog"
	"strings"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
)

type HL7Normalizer struct{}

func NewHL7Normalizer() *HL7Normalizer {
	return &HL7Normalizer{}
}

func (n *HL7Normalizer) Normalize(raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	logger := slog.Default().With(
		"component", "hl7_normalizer",
		"event_id", env.EventID,
	)

	logger.Info("starting HL7 normalization")

	msg, err := models.ParseHL7(string(raw))
	if err != nil {
		logger.Error("HL7 parse failed", "error", err)
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

	// PV1-20: Encounter ID (correct index)
	encounterID := safeHL7Field(pv1, 20)

	logger.Info("parsed HL7 fields",
		"message_type", messageType,
		"patient_id", patientID,
		"first_name", firstName,
		"last_name", lastName,
		"encounter_id", encounterID,
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

	logger.Info("HL7 normalization complete")
	return ne, nil
}

func safeHL7Field(fields []string, idx int) string {
	if len(fields) > idx {
		return fields[idx]
	}
	return ""
}
