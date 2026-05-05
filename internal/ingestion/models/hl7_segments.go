package models

import (
	"errors"
	"log"
	"strings"
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

	// Log raw HL7 input (truncate for safety)
	preview := raw
	if len(preview) > 300 {
		preview = preview[:300] + "...(truncated)"
	}
	log.Printf(`hl7_parse_input raw_preview="%s"`, preview)

	// Normalize line endings
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	lines := strings.Split(raw, "\n")
	log.Printf(`hl7_parse_lines count=%d`, len(lines))

	msg := &HL7Message{}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		fields := strings.Split(trimmed, "|")

		// Log each segment header
		if len(fields) > 0 {
			log.Printf(`hl7_segment index=%d segment="%s" field_count=%d`, i, fields[0], len(fields))
		}

		switch fields[0] {
		case "MSH":
			msg.MSH = fields
			log.Printf(`hl7_segment_msh fields=%v`, fields)
		case "PID":
			msg.PID = fields
			log.Printf(`hl7_segment_pid fields=%v`, fields)
		case "PV1":
			msg.PV1 = fields
			log.Printf(`hl7_segment_pv1 fields=%v`, fields)
		}
	}

	// Validate required segments
	if msg.MSH == nil {
		log.Printf(`hl7_parse_error error="missing MSH segment"`)
		return nil, ErrHL7MissingMSH
	}
	if msg.PID == nil {
		log.Printf(`hl7_parse_error error="missing PID segment"`)
		return nil, ErrHL7MissingPID
	}
	if msg.PV1 == nil {
		log.Printf(`hl7_parse_error error="missing PV1 segment"`)
		return nil, ErrHL7MissingPV1
	}

	log.Printf(`hl7_parse_success msh_len=%d pid_len=%d pv1_len=%d`,
		len(msg.MSH), len(msg.PID), len(msg.PV1))

	return msg, nil
}
