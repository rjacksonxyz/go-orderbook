package main

import (
	"go-orderbook/pkg/orderbook"
)

func main() {
	ob := orderbook.NewOrderbook()
	order := orderbook.NewOrder(
		orderbook.GoodTillCancel,
		orderbook.OrderId(1),
		orderbook.Buy,
		orderbook.Price(100),
		orderbook.Quantity(10),
	)
	if _, err := ob.AddOrder(order); err != nil {
		panic(err)
	}
}
