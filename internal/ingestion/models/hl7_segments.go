package models

import (
	"context"
	"errors"
	"strings"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

var (
	ErrHL7MissingMSH = errors.New("missing MSH segment")
	ErrHL7MissingPID = errors.New("missing PID segment")
	ErrHL7MissingPV1 = errors.New("missing PV1 segment")
)

type HL7Message struct {
	MSH []string
	PID []string
	PV1 []string
}

func ParseHL7(raw string) (*HL7Message, error) {
	ctx := context.Background()

	log := observability.WithTrace(ctx)

	log = log.With(zap.String("component", "hl7_parser"))

	// Preview raw HL7 (truncate for safety)
	preview := raw
	if len(preview) > 300 {
		preview = preview[:300] + "...(truncated)"
	}
	log.Debug("hl7_parse_input", zap.String("raw_preview", preview))

	// Normalize line endings
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	lines := strings.Split(raw, "\n")
	log.Debug("hl7_parse_lines", zap.Int("count", len(lines)))

	msg := &HL7Message{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		fields := strings.Split(trimmed, "|")
		segment := fields[0]

		log.Debug("hl7_segment",
			zap.Int("index", i),
			zap.String("segment", segment),
			zap.Int("field_count", len(fields)),
		)

		switch segment {
		case "MSH":
			msg.MSH = fields
			log.Debug("hl7_segment_msh", zap.Any("fields", fields))

		case "PID":
			msg.PID = fields
			log.Debug("hl7_segment_pid", zap.Any("fields", fields))

		case "PV1":
			msg.PV1 = fields
			log.Debug("hl7_segment_pv1", zap.Any("fields", fields))
		}
	}

	// Validate required segments
	if msg.MSH == nil {
		log.Error("hl7_parse_error", zap.String("error", ErrHL7MissingMSH.Error()))
		return nil, ErrHL7MissingMSH
	}
	if msg.PID == nil {
		log.Error("hl7_parse_error", zap.String("error", ErrHL7MissingPID.Error()))
		return nil, ErrHL7MissingPID
	}
	if msg.PV1 == nil {
		log.Error("hl7_parse_error", zap.String("error", ErrHL7MissingPV1.Error()))
		return nil, ErrHL7MissingPV1
	}

	log.Info("hl7_parse_success",
		zap.Int("msh_len", len(msg.MSH)),
		zap.Int("pid_len", len(msg.PID)),
		zap.Int("pv1_len", len(msg.PV1)),
	)

	return msg, nil
}
