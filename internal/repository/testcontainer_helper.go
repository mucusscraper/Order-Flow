package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()
	pgCont, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("orderflow_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(1)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := pgCont.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get a connection string: %s", err)
	}
	var dbPool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		dbPool, err = pgxpool.New(ctx, connStr)
		if err == nil {
			err = dbPool.Ping(ctx)
			if err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to connect to test database: %s", err)
	}
	_, err = dbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(255) PRIMARY KEY,
			customer_id VARCHAR(255) NOT NULL,
			total NUMERIC(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS order_items (
			id SERIAL PRIMARY KEY,
			order_id VARCHAR(255) REFERENCES orders(id) ON DELETE CASCADE,
			product_id VARCHAR(255) NOT NULL,
			quantity INT NOT NULL,
			price NUMERIC(10, 2) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS outbox_messages (
			id VARCHAR(255) PRIMARY KEY,
			aggregate_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			payload TEXT NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create test schema: %s", err)
	}
	cleanup := func() {
		dbPool.Close()
		if err := pgCont.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}

	return dbPool, cleanup
}
