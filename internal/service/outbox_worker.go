package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type OutboxWorker struct {
	db     *pgxpool.Pool
	writer *kafka.Writer
	logger *slog.Logger
}

func NewOutboxWorker(db *pgxpool.Pool, writer *kafka.Writer, logger *slog.Logger) *OutboxWorker {
	return &OutboxWorker{db: db, writer: writer, logger: logger}
}

func (w *OutboxWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	rows, err := w.db.Query(ctx, `
		SELECT id, event_type, aggregate_id, payload 
		FROM outbox_events WHERE status = 'PENDING' LIMIT 50`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, eventType, aggregateID string
		var payload []byte
		rows.Scan(&id, &eventType, &aggregateID, &payload)

		err = w.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(aggregateID),
			Value: payload,
		})

		if err != nil {
			w.logger.Error("kafka publish fail", "id", id, "err", err)
			continue
		}

		w.db.Exec(ctx, "UPDATE outbox_events SET status = 'PROCESSED', processed_at = NOW() WHERE id = $1", id)
	}
}
