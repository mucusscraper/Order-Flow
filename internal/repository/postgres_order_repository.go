package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mucusscraper/Order-Flow/internal/domain"
)

type PostgresOrderRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrderRepository(db *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Save mantém o comportamento original caso precise ser chamado isoladamente
func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.SaveWithTx(ctx, tx, order); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SaveWithTx permite salvar o pedido usando uma transação externa (ideal para o Outbox)
func (r *PostgresOrderRepository) SaveWithTx(ctx context.Context, tx pgx.Tx, order *domain.Order) error {
	queryOrder := `INSERT INTO orders (id, customer_id, total, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.Exec(ctx, queryOrder, order.ID, order.CustomerID, order.Total, order.Status, order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	queryItem := `INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, queryItem, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}
	return nil
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	queryOrder := `SELECT id, customer_id, total, status, created_at FROM orders WHERE id = $1`
	order := &domain.Order{}
	err := r.db.QueryRow(ctx, queryOrder, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Total,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}
	queryItems := `SELECT product_id, quantity, price FROM order_items WHERE order_id = $1`
	rows, err := r.db.Query(ctx, queryItems, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()
	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	order.Items = items
	return order, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	cmdTag, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}
