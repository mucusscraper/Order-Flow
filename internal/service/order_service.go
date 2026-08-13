package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mucusscraper/Order-Flow/internal/domain"
	"github.com/mucusscraper/Order-Flow/internal/repository"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

type CreateOrderDTO struct {
	CustomerID string             `json:"customer_id"`
	Items      []domain.OrderItem `json:"items"`
}

func (s *OrderService) CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error) {
	if len(dto.Items) == 0 {
		return nil, domain.ErrInvalidOrderItems
	}
	var total float64
	for _, item := range dto.Items {
		if item.Quantity <= 0 || item.Price < 0 {
			return nil, domain.ErrInvalidOrderItems
		}
		total += float64(item.Quantity) * item.Price
	}
	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: dto.CustomerID,
		Items:      dto.Items,
		Total:      total,
		Status:     domain.StatusCreated,
		CreatedAt:  time.Now(),
	}
	err := s.repo.Save(ctx, order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !order.CanTransitionTo(domain.StatusCancelled) {
		return nil, domain.ErrInvalidOrderState
	}
	if err := s.repo.UpdateStatus(ctx, id, domain.StatusCancelled); err != nil {
		return nil, err
	}
	order.Status = domain.StatusCancelled
	return order, nil
}
