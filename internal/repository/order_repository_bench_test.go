package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mucusscraper/Order-Flow/internal/domain"
)

func BenchmarkCreateOrder(b *testing.B) {
	dbPool, cleanup := SetupTestDB(&testing.T{})
	defer cleanup()

	repo := NewPostgresOrderRepository(dbPool)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		orderID := uuid.New().String()
		order := &domain.Order{
			ID:         orderID,
			CustomerID: "cust_benchmark",
			Items: []domain.OrderItem{
				{
					ProductID: "new_prod_id",
					Quantity:  1,
					Price:     75.0,
				},
			},
			Total:     75.0,
			Status:    "CREATED",
			CreatedAt: time.Now(),
		}
		repo.Save(ctx, order)
	}
}
