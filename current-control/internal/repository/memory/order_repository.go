package memory

import (
	"context"
	"sync"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	mu     sync.Mutex
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func copyOrder(value order.Order) order.Order {
	value.Items = append([]order.Item{}, value.Items...)
	value.History = append([]order.StatusChange{}, value.History...)
	return value
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	value.ID = r.nextID
	r.nextID++

	r.items = append(r.items, copyOrder(value))
	return value, nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, value := range r.items {
		if value.ID == id {
			return copyOrder(value), nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]order.Order, len(r.items))
	for i, value := range r.items {
		result[i] = copyOrder(value)
	}

	return result, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.items {
		if existing.ID == value.ID {
			r.items[i] = copyOrder(value)
			return value, nil
		}
	}

	return order.Order{}, order.ErrOrderNotFound
}
