package worker

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type OrderConsumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

func NewOrderConsumer(brokers []string, topic string, groupID string, logger *slog.Logger) *OrderConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 1e6,  // 1MB
	})

	return &OrderConsumer{
		reader: reader,
		logger: logger,
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
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("failed to commit message offset", "error", err)
		}
	}
}

func (c *OrderConsumer) Close() error {
	return c.reader.Close()
}
