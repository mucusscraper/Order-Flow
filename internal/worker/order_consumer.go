package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type OrderConsumer struct {
	reader    *kafka.Reader
	dlqWriter *kafka.Writer
	logger    *slog.Logger
	db        *pgxpool.Pool
}

func NewOrderConsumer(brokers []string, topic string, groupID string, logger *slog.Logger, db *pgxpool.Pool) *OrderConsumer {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    1e6, // 1MB
		StartOffset: kafka.FirstOffset,
		MaxWait:     1 * time.Second,
		Dialer:      dialer,
	})
	dlqWriter := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic + ".dlq", // Ex: orders.created.dlq
		Balancer: &kafka.LeastBytes{},
	}
	return &OrderConsumer{
		reader:    reader,
		dlqWriter: dlqWriter,
		logger:    logger,
		db:        db,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) {
	c.logger.Info("order consumer started, waiting for messages...")
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.logger.Error("failed to fetch message from kafka", "error", err)
			if ctx.Err() != nil {
				return
			}
			continue
		}
		c.logger.Info("received event from kafka",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"key", string(msg.Key),
			"payload", string(msg.Value),
		)
		err = c.processMessageWithRetry(ctx, msg)
		if err != nil {
			c.logger.Error("failed to process message after retries/dlq", "error", err)
			// Se falhou até na DLQ, não comitamos para tentar novamente depois
			continue
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("failed to commit message offset", "error", err)
		}
	}
}

func (c *OrderConsumer) isAlreadyProcessedOrSave(ctx context.Context, eventID string) (bool, error) {
	if eventID == "" {
		return false, nil // Se por acaso não veio ID, deixa passar ou trata conforme regra
	}

	query := `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT (event_id) DO NOTHING`
	result, err := c.db.Exec(ctx, query, eventID)
	if err != nil {
		return false, err
	}

	// Se RowsAffected for 0, significa que a chave já existia na tabela (já foi processada)
	if result.RowsAffected() == 0 {
		return true, nil
	}

	return false, nil
}

func (c *OrderConsumer) processMessageWithRetry(ctx context.Context, msg kafka.Message) error {
	maxRetries := 3
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Simulamos ou executamos a lógica real aqui
		err = c.executeBusinessLogic(ctx, msg)
		if err == nil {
			return nil // Sucesso!
		}

		c.logger.Warn("failed to process message, retrying...",
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", err,
		)
		// Backoff simples de 1 segundo por tentativa
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	// Se esgotou as tentativas, envia para a DLQ
	c.logger.Error("max retries reached, sending message to DLQ", "topic", c.dlqWriter.Topic, "error", err)

	dlqErr := c.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
		Headers: append(msg.Headers, kafka.Header{
			Key:   "error-reason",
			Value: []byte(err.Error()),
		}),
	})

	if dlqErr != nil {
		c.logger.Error("CRITICAL: failed to write to DLQ", "error", dlqErr)
		return dlqErr // Se falhar na DLQ, retornamos erro para não comitar o offset
	}

	return nil
}

func (c *OrderConsumer) executeBusinessLogic(ctx context.Context, msg kafka.Message) error {
	var event struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err // Erro de JSON (vai para DLQ após retries, pois JSON malformado nunca vai passar)
	}

	processed, err := c.isAlreadyProcessedOrSave(ctx, event.ID)
	if err != nil {
		return err // Erro de banco (vai tentar de novo caso o banco volte)
	}

	if processed {
		c.logger.Info("duplicate event detected, skipping business logic", "event_id", event.ID)
		return nil
	}

	// ==========================================
	// SUA LÓGICA DE NEGÓCIO REAL VEM AQUI
	// ==========================================
	c.logger.Info("processed order event successfully", "event_id", event.ID)
	return nil
}

func (c *OrderConsumer) Close() error {
	if err := c.dlqWriter.Close(); err != nil {
		c.logger.Error("failed to close dlq writer", "error", err)
	}
	return c.reader.Close()
}
