package compliance

import "time"

type Rule struct {
	ID             string
	EntityType     string
	EntityID       string
	RuleType       string
	ComplianceFlag bool
	ReasonCode     string
	SourceFormat   *string
	EventType      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
