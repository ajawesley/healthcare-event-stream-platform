package api

type IngestResponse struct {
	EventID    string `json:"event_id"`
	IngestedAt string `json:"ingested_at"`
	Format     string `json:"format"`
}
