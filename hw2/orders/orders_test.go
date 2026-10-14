package main

import "testing"

func TestNewOrder_InvalidID(t *testing.T) {
	_, err := NewOrder(0)
	if err == nil {
		t.Error("ожидали ошибку для ID=0")
	}
	_, err = NewOrder(-1)
	if err == nil {
		t.Error("ожидали ошибку для ID=-1")
	}
}

func TestNewOrder_MapInitialized(t *testing.T) {
	o, err := NewOrder(1)
	if err != nil {
		t.Fatal(err)
	}
	if o.Metadata == nil {
		t.Error("Metadata map не инициализирована через make")
	}
}

func TestCalculateTotal_EmptyOrder(t *testing.T) {
	o, _ := NewOrder(1)
	if sum := CalculateTotal(o); sum != 0 {
		t.Errorf("ожидали 0, получили %d", sum)
	}
}

func TestCalculateTotal_WithItems(t *testing.T) {
	o, _ := NewOrder(1)
	AddItem(o, Item{Name: "A", Price: 100})
	AddItem(o, Item{Name: "B", Price: 200})
	if sum := CalculateTotal(o); sum != 300 {
		t.Errorf("ожидали 300, получили %d", sum)
	}
}

func TestAddItem_NilOrder(t *testing.T) {
	var o *Order
	err := AddItem(o, Item{Name: "A", Price: 10})
	if err == nil {
		t.Error("ожидали ошибку для nil заказа")
	}
}

func TestAddItem_EmptyName(t *testing.T) {
	o, _ := NewOrder(1)
	err := AddItem(o, Item{Name: "", Price: 10})
	if err == nil {
		t.Error("ожидали ошибку для пустого имени")
	}
}

func TestAddItem_NegativePrice(t *testing.T) {
	o, _ := NewOrder(1)
	err := AddItem(o, Item{Name: "A", Price: -1})
	if err == nil {
		t.Error("ожидали ошибку для отрицательной цены")
	}
}

func TestCopy_DoesNotAffectOriginal(t *testing.T) {
	o1, _ := NewOrder(1)
	AddItem(o1, Item{Name: "A", Price: 100})

	o2 := &Order{
		ID:    o1.ID,
		Items: make([]Item, len(o1.Items)),
	}
	copy(o2.Items, o1.Items)

	AddItem(o2, Item{Name: "B", Price: 200})

	if len(o1.Items) != 1 {
		t.Errorf("оригинал изменился: len=%d, ожидали 1", len(o1.Items))
	}
}
