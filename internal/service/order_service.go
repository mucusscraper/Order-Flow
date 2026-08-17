package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mucusscraper/Order-Flow/internal/domain"
	"github.com/mucusscraper/Order-Flow/internal/repository"
	"github.com/redis/go-redis/v9"
)

type OrderService struct {
	repo repository.OrderRepository
	rdb  *redis.Client
}

func NewOrderService(repo repository.OrderRepository, rdb *redis.Client) *OrderService {
	return &OrderService{repo: repo, rdb: rdb}
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
	cacheKey := "order:" + id
	val, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var order domain.Order
		if json.Unmarshal([]byte(val), &order) == nil {
			return &order, nil
		}
	}
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(order)
	if err == nil {
		s.rdb.Set(ctx, cacheKey, data, 5*time.Minute)
	}
	return order, nil
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
	s.rdb.Del(ctx, "order:"+id)
	order.Status = domain.StatusCancelled
	return order, nil
}
