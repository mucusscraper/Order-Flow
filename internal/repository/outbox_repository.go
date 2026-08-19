package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/mucusscraper/Order-Flow/internal/domain"
)

type OutboxRepository interface {
	Save(ctx context.Context, tx pgx.Tx, event domain.Event) error
}

type PostgresOutboxRepository struct{}

func NewPostgresOutboxRepository() *PostgresOutboxRepository {
	return &PostgresOutboxRepository{}
}

func (r *PostgresOutboxRepository) Save(ctx context.Context, tx pgx.Tx, event domain.Event) error {
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO outbox_events (id, aggregate_id, event_type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, 'PENDING', $5)
	`

	_, err = tx.Exec(ctx, query, event.EventID, event.AggregateID, event.EventType, payloadBytes, event.OccurredAt)
	return err
}
