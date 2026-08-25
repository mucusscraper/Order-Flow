package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mucusscraper/Order-Flow/internal/domain"
	"github.com/mucusscraper/Order-Flow/internal/events"
	"github.com/mucusscraper/Order-Flow/internal/repository"
)

// CreateOrderDTO define os dados recebidos na requisição HTTP
type CreateOrderDTO struct {
	Items []OrderItemDTO `json:"items"`
}

type OrderItemDTO struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderService struct {
	dbPool    *pgxpool.Pool
	orderRepo *repository.PostgresOrderRepository
	publisher events.Publisher
}

func NewOrderService(dbPool *pgxpool.Pool, orderRepo *repository.PostgresOrderRepository, publisher events.Publisher) *OrderService {
	return &OrderService{
		dbPool:    dbPool,
		orderRepo: orderRepo,
		publisher: publisher,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error) {
	if len(dto.Items) == 0 {
		return nil, domain.ErrInvalidOrderItems
	}

	// Calcula o total e converte o DTO para o Domain Model
	var total float64
	var items []domain.OrderItem

	for _, itemDTO := range dto.Items {
		if itemDTO.Quantity <= 0 || itemDTO.Price < 0 {
			return nil, domain.ErrInvalidOrderItems
		}
		total += float64(itemDTO.Quantity) * itemDTO.Price
		items = append(items, domain.OrderItem{
			ProductID: itemDTO.ProductID,
			Quantity:  itemDTO.Quantity,
			Price:     itemDTO.Price,
		})
	}

	// Extrai o customer_id do contexto (injetado pelo middleware de Auth, se houver) ou define um padrão
	customerID := "cust-default"
	if sub, ok := ctx.Value("sub").(string); ok && sub != "" {
		customerID = sub
	}

	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		Total:      total,
		Status:     domain.OrderStatus("CREATED"), // Ou string dependendo do seu domain
		CreatedAt:  time.Now(),
		Items:      items,
	}

	// 1. Abre a transação com o pgx
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 2. Cria o pedido usando a transação externa
	err = s.orderRepo.SaveWithTx(ctx, tx, order)
	if err != nil {
		return nil, err
	}

	// 3. Monta o evento para o Outbox
	event := domain.Event{
		EventID:     uuid.New().String(),
		EventType:   "orders.created",
		AggregateID: order.ID,
		OccurredAt:  time.Now(),
		Payload:     order,
	}

	// 4. Salva o evento na tabela outbox_events na MESMA transação
	if err := s.publisher.Publish(ctx, tx, event); err != nil {
		return nil, err
	}

	// 5. Efetiva a transação no banco de dados
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.orderRepo.FindByID(ctx, id)
}

func (s *OrderService) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	// Busca o pedido primeiro para validar o estado se necessário
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Exemplo de validação de estado (ajuste conforme seu domain)
	if order.Status == "CANCELLED" || order.Status == "COMPLETED" {
		return nil, domain.ErrInvalidOrderState
	}

	err = s.orderRepo.UpdateStatus(ctx, id, domain.OrderStatus("CANCELLED"))
	if err != nil {
		return nil, err
	}

	order.Status = "CANCELLED"
	return order, nil
}
