package ingestion

import (
	"fmt"
	"strings"
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

	NM1Patient []string // NM1*IL
	CLM        []string // Claim Information
	DTPService []string // DTP*472 (Service Date)
}

func ParseX12(raw string) (*X12Message, error) {
	// Normalize line endings and separators
	raw = strings.ReplaceAll(raw, "\r\n", "")
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\r", "")

	segments := strings.Split(raw, "~")

	msg := &X12Message{}

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}

		fields := strings.Split(seg, "*")
		switch fields[0] {
		case "ISA":
			msg.ISA = fields
		case "GS":
			msg.GS = fields
		case "ST":
			msg.ST = fields
		case "NM1":
			if len(fields) > 2 && fields[1] == "IL" {
				msg.NM1Patient = fields
			}
		case "CLM":
			msg.CLM = fields
		case "DTP":
			if len(fields) > 2 && fields[1] == "472" {
				msg.DTPService = fields
			}
		}
	}

	// Required segments
	if msg.ISA == nil {
		return nil, ErrX12MissingISA
	}
	if msg.GS == nil {
		return nil, ErrX12MissingGS
	}
	if msg.ST == nil {
		return nil, ErrX12MissingST
	}
	if msg.NM1Patient == nil {
		return nil, ErrX12MissingNM1IL
	}
	if msg.CLM == nil {
		return nil, ErrX12MissingCLM
	}

	return msg, nil
}
