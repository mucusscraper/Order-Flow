package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mucusscraper/Order-Flow/internal/domain"
)

type OutboxPublisher struct {
	repo OutboxRepository
}

func NewOutboxPublisher(repo OutboxRepository) *OutboxPublisher {
	return &OutboxPublisher{repo: repo}
}

func (p *OutboxPublisher) Publish(ctx context.Context, tx pgx.Tx, event domain.Event) error {
	return p.repo.Save(ctx, tx, event)
}
