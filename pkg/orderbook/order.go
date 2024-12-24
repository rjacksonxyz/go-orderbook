package orderbook

import (
	"fmt"
	"go-orderbook/pkg/ds/list"
)

type OrderType int

const (
	Market OrderType = iota
	GoodTillCancel
	GoodForDay
	FillAndKill
	FillOrKill
)

type Side int

const (
	Buy Side = iota
	Sell
)

type Order struct {
	orderType         OrderType
	orderId           OrderId
	side              Side
	price             Price
	initialQuantity   Quantity
	remainingQuantity Quantity
}

type OrderEntry struct {
	order    Order
	location int
}

func NewOrder(
	orderType OrderType,
	orderId OrderId,
	side Side,
	price Price,
	quantity Quantity,
) Order {
	return Order{
		orderType:         orderType,
		orderId:           orderId,
		side:              side,
		price:             price,
		initialQuantity:   quantity,
		remainingQuantity: quantity,
	}
}

func NewMarketOrder(
	orderId OrderId,
	side Side,
	quantity Quantity,
) Order {
	return NewOrder(Market, orderId, side, 0, quantity)
}

func (o *Order) OrderId() OrderId {
	return o.orderId
}

func (o *Order) OrderType() OrderType {
	return o.orderType
}

func (o *Order) Side() Side {
	return o.side
}

func (o *Order) Price() Price {
	return o.price
}

func (o *Order) InitialQuantity() Quantity {
	return o.initialQuantity
}

func (o *Order) FilledQuantity() Quantity {
	return o.initialQuantity - o.remainingQuantity
}

func (o *Order) IsFilled() bool {
	return o.remainingQuantity == 0
}

func (o *Order) Fill(quantity Quantity) error {
	if quantity > o.remainingQuantity {
		return fmt.Errorf(
			"Order %d cannot be filled for more than it's remaining quantity",
			o.orderId,
		)
	}
	o.remainingQuantity -= quantity
	return nil
}

func (o *Order) ToGoodTillCancel(price Price) error {
	if o.OrderType() != Market {
		return fmt.Errorf(
			"Order %d cannot be converted to GoodTillCancel, must be Market",
			o.OrderId(),
		)
	}
	o.price = price
	o.orderType = GoodTillCancel
	return nil
}

type Orders struct {
	list.LinkedList[Order]
}

type OrderModify struct {
	orderId  OrderId
	side     Side
	price    Price
	quantity Quantity
}

func (o *OrderModify) New(
	orderId OrderId,
	price Price,
	side Side,
	quantity Quantity,
) OrderModify {
	return OrderModify{
		price:    price,
		side:     side,
		orderId:  orderId,
		quantity: quantity,
	}
}

func (o *OrderModify) OrderId() OrderId {
	return o.orderId
}

func (o *OrderModify) Side() Side {
	return o.side
}

func (o *OrderModify) Price() Price {
	return o.price
}

func (o *OrderModify) Quantity() Quantity {
	return o.quantity
}

func (o *OrderModify) ToOrder(orderType OrderType) Order {
	return Order{
		orderType:         orderType,
		orderId:           o.orderId,
		side:              o.side,
		price:             o.price,
		initialQuantity:   o.quantity,
		remainingQuantity: o.quantity,
	}
}
