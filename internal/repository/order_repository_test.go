package repository

/*
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mucusscraper/Order-Flow/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresOrderRepository_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDB := SetupTestDB(t)
	defer testDB.Closer()

	ctx := context.Background()
	repo := NewPostgresOrderRepository(testDB.Pool)

	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: "cust-123",
		Items: []domain.OrderItem{
			{ProductID: "prod-1", Quantity: 2, Price: 10.50},
		},
		Total:  21.00,
		Status: "CREATED",
	}

	err := repo.Save(ctx, order)
	require.NoError(t, err)
	require.NotEmpty(t, order.ID)

	fetchedOrder, err := repo.FindByID(ctx, order.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedOrder)

	assert.Equal(t, order.ID, fetchedOrder.ID)
	assert.Equal(t, order.CustomerID, fetchedOrder.CustomerID)
	assert.Equal(t, domain.OrderStatus("CREATED"), fetchedOrder.Status)
	assert.Len(t, fetchedOrder.Items, 1)
	assert.Equal(t, "prod-1", fetchedOrder.Items[0].ProductID)
}
*/
