package normalizer

import (
	"context"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/api"
	"github.com/ajawes/hesp/internal/ingestion/models"
	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

type X12Normalizer struct{}

func NewX12Normalizer() *X12Normalizer {
	return &X12Normalizer{}
}

func (n *X12Normalizer) Normalize(ctx context.Context, raw []byte, env api.Envelope) (*models.NormalizedEvent, error) {
	log := observability.WithTrace(ctx).With(
		zap.String("component", "x12_normalizer"),
		zap.String("event_id", env.EventID),
	)

	log.Info("x12_normalization_start")

	msg, err := models.ParseX12(string(raw))
	if err != nil {
		log.Error("x12_parse_failed", zap.Error(err))
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

	log.Debug("x12_parsed_fields",
		zap.String("patient_id", patientID),
		zap.String("first_name", firstName),
		zap.String("last_name", lastName),
		zap.String("encounter_id", encounterID),
		zap.String("observation_code", observationCode),
		zap.String("service_date", serviceDate),
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

	log.Info("x12_normalization_complete",
		zap.Any("fields", ne.Fields),
		zap.Any("metadata", ne.Metadata),
	)

	return ne, nil
}

func safeX12Field(fields []string, idx int) string {
	if len(fields) > idx {
		return fields[idx]
	}
	return ""
}
