package repository

import (
	"context"
	"testing"
	"time"

	"github.com/mucusscraper/Order-Flow/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestPostgresOrderRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dbPool, cleanup := SetupTestDB(t)
	defer cleanup()
	repo := NewPostgresOrderRepository(dbPool)
	ctx := context.Background()

	order := &domain.Order{
		ID:         "test-order-uuid-123",
		CustomerID: "cust-test-1",
		Total:      150,
		Status:     "CREATED",
		CreatedAt:  time.Now(),
		Items: []domain.OrderItem{
			{
				ProductID: "prod-abc",
				Quantity:  2,
				Price:     75.00,
			},
		},
	}
	err := repo.Save(ctx, order)
	assert.NoError(t, err)
	fetchedOrder, err := repo.FindByID(ctx, order.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedOrder)
	assert.Equal(t, order.ID, fetchedOrder.ID)
	assert.Equal(t, order.CustomerID, fetchedOrder.CustomerID)
	assert.Equal(t, order.Total, fetchedOrder.Total)
	assert.Equal(t, order.Status, fetchedOrder.Status)
	assert.Len(t, fetchedOrder.Items, 1)
	assert.Equal(t, "prod-abc", fetchedOrder.Items[0].ProductID)
}
