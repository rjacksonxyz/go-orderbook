package orderbook

import (
	"fmt"
	"go-orderbook/pkg/ds/list"
	"go-orderbook/pkg/ds/rbmap"
	"go-orderbook/pkg/util"
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
	order    Order
	location int
}

type Orderbook struct {
	bids   *rbmap.Map[Price, Orders]
	asks   *rbmap.Map[Price, Orders]
	orders map[OrderId]OrderEntry
}

func (o *Orderbook) New() Orderbook {
	return Orderbook{
		bids:   rbmap.NewMap[Price, Orders](rbmap.Ascending[Price]),
		asks:   rbmap.NewMap[Price, Orders](rbmap.Descending[Price]),
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

func (o *Orderbook) MatchOrders() (Trades, error) {
	var trades Trades

	for {

		// check for empty bids or asks
		if o.bids.Empty() || o.asks.Empty() {
			break
		}

		// retrieve the best bid and ask prices, along with the
		// corresponding orders
		bidIt := o.bids.Begin()
		bidPrice, bids := bidIt.Key(), bidIt.Value()

		askIt := o.asks.Begin()
		askPrice, asks := askIt.Key(), askIt.Value()

		if bidPrice < askPrice {
			break
		}

		// While there are bids and asks, match them
		for bids.Size() > 0 && asks.Size() > 0 {
			// retrieve the first bid and ask (time priority)
			bid, _ := bids.Head()
			ask, _ := asks.Head()

			// determine the quantity to match
			quantity := util.Min(
				bid.remainingQuantity,
				ask.remainingQuantity,
			)

			// fill the orders
			err := bid.Fill(quantity)
			if err != nil {
				return trades, err
			}
			err = ask.Fill(quantity)
			if err != nil {
				return trades, err
			}

			if bid.IsFilled() {
				bids.DeleteHead()
				delete(o.orders, bid.OrderId())
			}

			if ask.IsFilled() {
				asks.DeleteHead()
				delete(o.orders, ask.OrderId())
			}

			if bids.IsEmpty() {
				o.bids.Delete(bidPrice)
			}

			if asks.IsEmpty() {
				o.asks.Delete(askPrice)
			}

			// append the trade to the list of trades
			trades = append(trades,
				Trade{
					bidTrade: TradeInfo{
						orderId:  bid.OrderId(),
						price:    bid.Price(),
						quantity: quantity,
					},
					askTrade: TradeInfo{
						orderId:  ask.OrderId(),
						price:    ask.Price(),
						quantity: quantity,
					},
				},
			)

			// handle FillAndKill orders
			if !bids.IsEmpty() {
				bid, _ = bids.Head()
				if bid.OrderType() == FillAndKill {
					bids.DeleteHead()
					delete(o.orders, bid.OrderId())
				}
			}

			if !asks.IsEmpty() {
				ask, _ = asks.Head()
				if ask.OrderType() == FillAndKill {
					asks.DeleteHead()
					delete(o.orders, ask.OrderId())
				}
			}
		}
	}
	return trades, nil
}

func (o *Orderbook) Add(order Order) (Trades, error) {
	if _, exists := o.orders[order.OrderId()]; exists {
	}

	if order.OrderType() == FillAndKill &&
		!o.CanMatch(order.Side(), order.Price()) {
	}

	var orders Orders

	// TODO: Refactor this code create zero values by deafult
	// check if price level exists and create if not, inserting the order.
	// store the Orders for the appropriate side in `orders`
	if order.Side() == Buy {
		if _, exists := o.bids.Get(order.Price()); !exists {
			o.bids.Insert(order.Price(), orders)
		}
		orders, _ = o.bids.Get(order.Price())
	} else {
		if _, exists := o.asks.Get(order.Price()); !exists {
			o.asks.Insert(order.Price(), orders)
		}
		orders, _ = o.asks.Get(order.Price())
	}

	o.orders[order.OrderId()] = OrderEntry{
		order:    order,
		location: orders.Size() - 1,
	}
	return o.MatchOrders()
}
