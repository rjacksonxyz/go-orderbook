package orderbook

import (
	"fmt"
	"go-orderbook/pkg/ds/rbmap"
)

type OrderType int

const (
	GoodTillCancel OrderType = iota
	FillAndKill
)

type Side int

const (
	Buy Side = iota
	Sell
)

type (
	Price    int32
	Quantity uint32
	OrderId  uint64
)

type LevelInfo struct {
	Price    Price
	Quantity Quantity
}

type LevelsInfo []LevelInfo

// OrderbookLevelsInfo stores state of the bids and asks for given levels in the
// order book.
type OrderbookLevelsInfo struct {
	bids LevelsInfo
	asks LevelsInfo
}

func (o *OrderbookLevelsInfo) New(
	bids, asks LevelsInfo,
) OrderbookLevelsInfo {
	return OrderbookLevelsInfo{
		bids: bids,
		asks: asks,
	}
}

func (o *OrderbookLevelsInfo) GetBids() LevelsInfo {
	return o.bids
}

func (o *OrderbookLevelsInfo) GetAsks() LevelsInfo {
	return o.asks
}

type Order struct {
	orderType         OrderType
	orderId           OrderId
	side              Side
	price             Price
	intialQuantity    Quantity
	remainingQuantity Quantity
}

func (o *Order) New(
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
		intialQuantity:    quantity,
		remainingQuantity: quantity,
	}
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
	return o.intialQuantity
}

func (o *Order) FilledQuantity() Quantity {
	return o.intialQuantity - o.remainingQuantity
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

type Orders []*Order

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

func (o *OrderModify) ToOrder(orderType OrderType) *Order {
	return &Order{
		orderType:         orderType,
		orderId:           o.orderId,
		side:              o.side,
		price:             o.price,
		intialQuantity:    o.quantity,
		remainingQuantity: o.quantity,
	}
}

type TradeInfo struct {
	orderId  OrderId
	price    Price
	quantity Quantity
}

// A Trade represents a matching bid and ask.
type Trade struct {
	bidTrade TradeInfo
	askTrade TradeInfo
}

func (t *Trade) New(
	bidTrade, askTrade TradeInfo,
) Trade {
	return Trade{
		bidTrade: bidTrade,
		askTrade: askTrade,
	}
}

type Trades []Trade

type OrderEntry struct {
	order    *Order
	location int
}

type Orderbook struct {
	bids   *rbmap.Map[Price, Orders]
	asks   *rbmap.Map[Price, Orders]
	orders map[OrderId]OrderEntry
}

func (o *Orderbook) New() Orderbook {
	return Orderbook{
		bids: rbmap.NewMap[Price, Orders](func(a, b Price) bool {
			return a < b
		}),
		asks: rbmap.NewMap[Price, Orders](func(a, b Price) bool {
			return a > b
		}),
		orders: make(map[OrderId]OrderEntry),
	}
}

// CanMatch checks if a given order can be matched at a given price.
func (o *Orderbook) CanMatch(
	side Side,
	price Price,
) bool {
	if side == Buy {
		if o.asks.Empty() {
			return false
		}
		asks := o.asks.Begin()
		bestAsk := asks.Key()

		return price >= bestAsk
	} else {
		if o.bids.Empty() {
			return false
		}
		bids := o.bids.Begin()
		bestBid := bids.Key()
		return price <= bestBid
	}
}

//
//
