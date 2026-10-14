package main

import "errors"

var StatusNames = map[string]string{
	"new":        "Новый",
	"processing": "В обработке",
	"completed":  "Завершен",
}

type Item struct {
	Name  string
	Price int
}

type Order struct {
	ID       int
	Items    []Item
	Metadata map[string]string
}

func NewOrder(id int) (*Order, error) {
	if id <= 0 {
		return nil, errors.New("ID должен быть положительным")
	}
	return &Order{
		ID:       id,
		Items:    make([]Item, 0),
		Metadata: make(map[string]string),
	}, nil
}

func AddItem(o *Order, item Item) error {
	if o == nil {
		return errors.New("заказ nil")
	}
	if item.Name == "" {
		return errors.New("имя позиции пустое")
	}
	if item.Price < 0 {
		return errors.New("цена отрицательная")
	}
	o.Items = append(o.Items, item)
	return nil
}

func CalculateTotal(o *Order) int {
	if o == nil {
		return 0
	}
	sum := 0
	for _, item := range o.Items {
		sum += item.Price
	}
	return sum
}
