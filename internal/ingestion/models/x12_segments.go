package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/ajawes/hesp/internal/observability"
	"go.uber.org/zap"
)

var (
	ErrX12MissingISA   = fmt.Errorf("x12: missing ISA segment")
	ErrX12MissingGS    = fmt.Errorf("x12: missing GS segment")
	ErrX12MissingST    = fmt.Errorf("x12: missing ST segment")
	ErrX12MissingNM1IL = fmt.Errorf("x12: missing NM1*IL patient segment")
	ErrX12MissingCLM   = fmt.Errorf("x12: missing CLM segment")
)

type X12Message struct {
	ISA []string
	GS  []string
	ST  []string

	NM1Patient []string
	CLM        []string
	DTPService []string
}

func ParseX12(raw string) (*X12Message, error) {
	ctx := context.Background()

	log := observability.WithTrace(ctx)

	log = log.With(zap.String("component", "x12_parser"))

	// Preview raw input
	preview := raw
	if len(preview) > 300 {
		preview = preview[:300] + "...(truncated)"
	}
	log.Debug("x12_parse_input", zap.String("raw_preview", preview))

	// Normalize line endings
	raw = strings.ReplaceAll(raw, "\r\n", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\r", "")

	segments := strings.Split(raw, "~")
	log.Debug("x12_parse_segments", zap.Int("count", len(segments)))

	msg := &X12Message{}

	for i, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		fields := strings.Split(seg, "*")
		segment := fields[0]

		log.Debug("x12_segment",
			zap.Int("index", i),
			zap.String("segment", segment),
			zap.Int("field_count", len(fields)),
		)

		switch segment {
		case "ISA":
			msg.ISA = fields
			log.Debug("x12_segment_isa", zap.Any("fields", fields))

		case "GS":
			msg.GS = fields
			log.Debug("x12_segment_gs", zap.Any("fields", fields))

		case "ST":
			msg.ST = fields
			log.Debug("x12_segment_st", zap.Any("fields", fields))

		case "NM1":
			if len(fields) > 2 && fields[1] == "IL" {
				msg.NM1Patient = fields
				log.Debug("x12_segment_nm1_il", zap.Any("fields", fields))
			}

		case "CLM":
			msg.CLM = fields
			log.Debug("x12_segment_clm", zap.Any("fields", fields))

		case "DTP":
			if len(fields) > 2 && fields[1] == "472" {
				msg.DTPService = fields
				log.Debug("x12_segment_dtp_472", zap.Any("fields", fields))
			}
		}
	}

	// Required segments
	if msg.ISA == nil {
		log.Error("x12_parse_error", zap.String("error", ErrX12MissingISA.Error()))
		return nil, ErrX12MissingISA
	}
	if msg.GS == nil {
		log.Error("x12_parse_error", zap.String("error", ErrX12MissingGS.Error()))
		return nil, ErrX12MissingGS
	}
	if msg.ST == nil {
		log.Error("x12_parse_error", zap.String("error", ErrX12MissingST.Error()))
		return nil, ErrX12MissingST
	}
	if msg.NM1Patient == nil {
		log.Error("x12_parse_error", zap.String("error", ErrX12MissingNM1IL.Error()))
		return nil, ErrX12MissingNM1IL
	}
	if msg.CLM == nil {
		log.Error("x12_parse_error", zap.String("error", ErrX12MissingCLM.Error()))
		return nil, ErrX12MissingCLM
	}

	log.Info("x12_parse_success",
		zap.Int("isa_len", len(msg.ISA)),
		zap.Int("gs_len", len(msg.GS)),
		zap.Int("st_len", len(msg.ST)),
		zap.Int("nm1_il_len", len(msg.NM1Patient)),
		zap.Int("clm_len", len(msg.CLM)),
		zap.Int("dtp_len", len(msg.DTPService)),
	)

	return msg, nil
}
