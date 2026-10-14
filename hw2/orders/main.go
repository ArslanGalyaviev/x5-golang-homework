package main

import "fmt"

func main() {
	o, _ := NewOrder(1)

	AddItem(o, Item{Name: "Хлеб", Price: 50})
	AddItem(o, Item{Name: "Молоко", Price: 90})
	AddItem(o, Item{Name: "Сыр", Price: 300})

	fmt.Printf("Заказ #%d, сумма: %d\n", o.ID, CalculateTotal(o))
	fmt.Printf("Статус 'new': %s\n", StatusNames["new"])
}
