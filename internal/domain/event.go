package domain

import (
	"time"
)

type Event struct {
	EventID     string      `json:"event_id"`
	EventType   string      `json:"event_type"`
	AggregateID string      `json:"aggregate_id"`
	OccurredAt  time.Time   `json:"occurred_at"`
	Payload     interface{} `json:"payload"`
}
