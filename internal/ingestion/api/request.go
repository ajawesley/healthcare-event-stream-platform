package api

import "encoding/json"

type IngestRequest struct {
	Envelope json.RawMessage `json:"envelope"`
	Payload  json.RawMessage `json:"payload"`
}

type Envelope struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	ProducedAt   string `json:"produced_at"`
	SourceSystem string `json:"source_system"`
}
