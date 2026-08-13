package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mucusscraper/Order-Flow/internal/domain"
	"github.com/mucusscraper/Order-Flow/internal/repository"
)

func TestPostgresOrderRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/orderflow?sslmode=disable"
	}
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("skipping integration test, cannot connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		t.Skipf("skipping integration test, database not reachable: %v", err)
	}
	repo := repository.NewPostgresOrderRepository(dbPool)
	order := &domain.Order{
		ID:         "test-id-" + time.Now().Format("20060102150405"),
		CustomerID: "cust-test",
		Total:      150.0,
		Status:     domain.StatusCreated,
		CreatedAt:  time.Now(),
		Items: []domain.OrderItem{
			{ProductID: "prod-A", Quantity: 2, Price: 75.0},
		},
	}
	err = repo.Save(ctx, order)
	if err != nil {
		t.Fatalf("failed to save order: %v", err)
	}
	found, err := repo.FindByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to find order by id: %v", err)
	}

	if found.ID != order.ID || found.CustomerID != order.CustomerID || len(found.Items) != 1 {
		t.Errorf("found order data mismatch: got %+v", found)
	}
	err = repo.UpdateStatus(ctx, order.ID, domain.StatusCancelled)
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	updated, err := repo.FindByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("failed to find updated order: %v", err)
	}
	if updated.Status != domain.StatusCancelled {
		t.Errorf("expected status to be CANCELLED, got %s", updated.Status)
	}
}
