package service

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
	mu         sync.Mutex
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}

	var total int64

	for _, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" || item.Quantity <= 0 || item.Price <= 0 {
			return order.Order{}, order.ErrInvalidOrder
		}

		if item.Price > (math.MaxInt64-total)/int64(item.Quantity) {
			return order.Order{}, order.ErrInvalidOrder
		}

		total += item.Price * int64(item.Quantity)
	}

	now := time.Now().UTC()

	value := order.Order{
		Items:      input.Items,
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: now},
		},
	}

	return s.repository.Create(ctx, value)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
	default:
		return order.Order{}, order.ErrInvalidStatus
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	switch value.Status {
	case order.StatusNew:
		if status != order.StatusProcessing && status != order.StatusCancelled {
			return order.Order{}, order.ErrInvalidStatusTransition
		}
	case order.StatusProcessing:
		if status != order.StatusCompleted && status != order.StatusCancelled {
			return order.Order{}, order.ErrInvalidStatusTransition
		}
	default:
		return order.Order{}, order.ErrInvalidStatusTransition
	}

	value.Status = status
	value.History = append(value.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now().UTC(),
	})

	return s.repository.Update(ctx, value)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	history := append([]order.StatusChange{}, value.History...)

	sort.SliceStable(history, func(i, j int) bool {
		return history[i].ChangedAt.Before(history[j].ChangedAt)
	})

	return history, nil
}
