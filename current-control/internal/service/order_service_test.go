package service_test

import (
	"context"
	"math"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/require"
)

type fakeRepository struct {
	values []order.Order
}

func (r *fakeRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = int64(len(r.values) + 1)
	r.values = append(r.values, value)
	return value, nil
}

func (r *fakeRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, value := range r.values {
		if value.ID == id {
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *fakeRepository) List(_ context.Context) ([]order.Order, error) {
	return r.values, nil
}

func (r *fakeRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i, existing := range r.values {
		if existing.ID == value.ID {
			r.values[i] = value
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func TestCreate(t *testing.T) {
	cases := []struct {
		name    string
		items   []order.Item
		total   int64
		wantErr error
	}{
		{"valid order", []order.Item{
			{Name: "A", Quantity: 2, Price: 100},
			{Name: "B", Quantity: 3, Price: 50},
		}, 350, nil},
		{"empty items", nil, 0, order.ErrInvalidOrder},
		{"empty name", []order.Item{{Name: "", Quantity: 1, Price: 100}}, 0, order.ErrInvalidOrder},
		{"spaces", []order.Item{{Name: "  ", Quantity: 1, Price: 100}}, 0, order.ErrInvalidOrder},
		{"zero quantity", []order.Item{{Name: "A", Quantity: 0, Price: 100}}, 0, order.ErrInvalidOrder},
		{"negative quantity", []order.Item{{Name: "A", Quantity: -1, Price: 100}}, 0, order.ErrInvalidOrder},
		{"zero price", []order.Item{{Name: "A", Quantity: 1, Price: 0}}, 0, order.ErrInvalidOrder},
		{"negative price", []order.Item{{Name: "A", Quantity: 1, Price: -100}}, 0, order.ErrInvalidOrder},
		{"overflow", []order.Item{{Name: "A", Quantity: 2, Price: math.MaxInt64}}, 0, order.ErrInvalidOrder},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := service.NewOrderService(&fakeRepository{})

			got, err := s.Create(context.Background(), order.CreateOrderInput{
				Items: tc.items,
			})

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, int64(1), got.ID)
			require.Equal(t, tc.items, got.Items)
			require.Equal(t, tc.total, got.TotalPrice)
			require.Equal(t, order.StatusNew, got.Status)
			require.False(t, got.CreatedAt.IsZero())
			require.Len(t, got.History, 1)
			require.Equal(t, order.StatusNew, got.History[0].Status)
		})
	}
}

func TestGetAndList(t *testing.T) {
	ctx := context.Background()
	s := service.NewOrderService(&fakeRepository{})

	input := order.CreateOrderInput{
		Items: []order.Item{{Name: "A", Quantity: 1, Price: 100}},
	}

	first, err := s.Create(ctx, input)
	require.NoError(t, err)

	second, err := s.Create(ctx, input)
	require.NoError(t, err)

	got, err := s.GetByID(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, first, got)

	_, err = s.GetByID(ctx, 999)
	require.ErrorIs(t, err, order.ErrOrderNotFound)

	values, err := s.List(ctx)
	require.NoError(t, err)
	require.Equal(t, []order.Order{first, second}, values)
}

func TestUpdateStatus(t *testing.T) {
	cases := []struct {
		from    order.Status
		to      order.Status
		allowed bool
	}{
		{order.StatusNew, order.StatusProcessing, true},
		{order.StatusNew, order.StatusCancelled, true},
		{order.StatusProcessing, order.StatusCompleted, true},
		{order.StatusProcessing, order.StatusCancelled, true},
		{order.StatusNew, order.StatusCompleted, false},
		{order.StatusNew, order.StatusNew, false},
		{order.StatusProcessing, order.StatusNew, false},
		{order.StatusProcessing, order.StatusProcessing, false},
		{order.StatusCompleted, order.StatusCancelled, false},
		{order.StatusCancelled, order.StatusProcessing, false},
	}

	for _, tc := range cases {
		t.Run(string(tc.from)+"_to_"+string(tc.to), func(t *testing.T) {
			repo := &fakeRepository{
				values: []order.Order{{ID: 1, Status: tc.from}},
			}

			s := service.NewOrderService(repo)
			got, err := s.UpdateStatus(context.Background(), 1, tc.to)

			if !tc.allowed {
				require.ErrorIs(t, err, order.ErrInvalidStatusTransition)
				require.Equal(t, tc.from, repo.values[0].Status)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.to, got.Status)
			require.Equal(t, tc.to, repo.values[0].Status)
		})
	}

	s := service.NewOrderService(&fakeRepository{})

	_, err := s.UpdateStatus(context.Background(), 1, order.StatusProcessing)
	require.ErrorIs(t, err, order.ErrOrderNotFound)

	_, err = s.UpdateStatus(context.Background(), 1, "wrong")
	require.ErrorIs(t, err, order.ErrInvalidStatus)
}

func TestHistory(t *testing.T) {
	ctx := context.Background()
	s := service.NewOrderService(&fakeRepository{})

	value, err := s.Create(ctx, order.CreateOrderInput{
		Items: []order.Item{{Name: "A", Quantity: 1, Price: 100}},
	})
	require.NoError(t, err)

	_, err = s.UpdateStatus(ctx, value.ID, order.StatusProcessing)
	require.NoError(t, err)

	_, err = s.UpdateStatus(ctx, value.ID, order.StatusCompleted)
	require.NoError(t, err)

	history, err := s.History(ctx, value.ID)
	require.NoError(t, err)
	require.Len(t, history, 3)
	require.Equal(t, order.StatusNew, history[0].Status)
	require.Equal(t, order.StatusProcessing, history[1].Status)
	require.Equal(t, order.StatusCompleted, history[2].Status)

	for i := 1; i < len(history); i++ {
		require.False(t, history[i].ChangedAt.Before(history[i-1].ChangedAt))
	}

	_, err = s.History(ctx, 999)
	require.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestHistoryOrder(t *testing.T) {
	now := time.Now()

	first := order.StatusChange{
		Status:    order.StatusNew,
		ChangedAt: now.Add(-time.Minute),
	}

	second := order.StatusChange{
		Status:    order.StatusProcessing,
		ChangedAt: now,
	}

	repo := &fakeRepository{
		values: []order.Order{
			{ID: 1, History: []order.StatusChange{second, first}},
		},
	}

	s := service.NewOrderService(repo)

	history, err := s.History(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []order.StatusChange{first, second}, history)
	require.Equal(t, []order.StatusChange{second, first}, repo.values[0].History)
}
