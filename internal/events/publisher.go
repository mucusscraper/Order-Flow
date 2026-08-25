package events

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mucusscraper/Order-Flow/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, tx pgx.Tx, event domain.Event) error
}
