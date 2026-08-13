package domain

import (
	"errors"
	"time"
)

type OrderStatus string

const (
	StatusCreated        OrderStatus = "CREATED"
	StatusPaymentPending OrderStatus = "PAYMENT_PENDING"
	StatusPaid           OrderStatus = "PAID"
	StatusProcessing     OrderStatus = "PROCESSING"
	StatusCompleted      OrderStatus = "COMPLETED"
	StatusCancelled      OrderStatus = "CANCELlED"
	StatusFailed         OrderStatus = "FAILED"
)

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	Total      float64     `json:"total"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
}

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidOrderState = errors.New("invalid state transition")
	ErrInvalidOrderItems = errors.New("order must have at least one item")
)

func (o *Order) CanTransitionTo(next OrderStatus) bool {
	switch o.Status {
	case StatusCreated:
		return next == StatusPaymentPending || next == StatusCancelled
	case StatusPaymentPending:
		return next == StatusPaid || next == StatusFailed || next == StatusCancelled
	case StatusPaid:
		return next == StatusProcessing || next == StatusCancelled
	case StatusProcessing:
		return next == StatusCompleted || next == StatusCancelled
	default:
		return false
	}
}
