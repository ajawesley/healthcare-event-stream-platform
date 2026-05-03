package ingestion

import (
	"errors"
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
	lines := strings.Split(raw, "\n")

	msg := &HL7Message{}

	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "|")
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "MSH":
			msg.MSH = fields
		case "PID":
			msg.PID = fields
		case "PV1":
			msg.PV1 = fields
		}
	}

	if msg.MSH == nil {
		return nil, ErrHL7MissingMSH
	}
	if msg.PID == nil {
		return nil, ErrHL7MissingPID
	}
	if msg.PV1 == nil {
		return nil, ErrHL7MissingPV1
	}

	return msg, nil
}
