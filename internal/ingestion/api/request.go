package api

import (
	"encoding/json"
	"fmt"
	"time"
)

type IngestRequest struct {
	Envelope json.RawMessage `json:"envelope"`
	Payload  json.RawMessage `json:"payload"`
}

type Envelope struct {
	EventID      string    `json:"event_id"`
	EventType    string    `json:"event_type"`
	SourceSystem string    `json:"source_system"`
	ProducedAt   time.Time `json:"produced_at"`
}

func (e *Envelope) UnmarshalJSON(data []byte) error {
	// Define an alias with ProducedAt as string
	var aux struct {
		EventID      string `json:"event_id"`
		EventType    string `json:"event_type"`
		SourceSystem string `json:"source_system"`
		ProducedAt   string `json:"produced_at"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	ts, err := time.Parse(time.RFC3339, aux.ProducedAt)
	if err != nil {
		return fmt.Errorf("invalid produced_at timestamp: %w", err)
	}

	e.EventID = aux.EventID
	e.EventType = aux.EventType
	e.SourceSystem = aux.SourceSystem
	e.ProducedAt = ts

	return nil
}
