package orderbook

import (
	"fmt"
	"go-orderbook/pkg/ds/rbmap"
	"go-orderbook/pkg/util"
	"sync"
	"sync/atomic"
	"time"
)

type Orderbook struct {
	m        *sync.Mutex
	bids     *rbmap.Map[Price, Orders]
	asks     *rbmap.Map[Price, Orders]
	orders   map[OrderId]OrderEntry
	shutdown atomic.Bool
	cond     *sync.Cond
}

func NewOrderbook() Orderbook {
	return Orderbook{
		m:      &sync.Mutex{},
		bids:   rbmap.NewMap[Price, Orders](rbmap.Ascending[Price]),
		asks:   rbmap.NewMap[Price, Orders](rbmap.Descending[Price]),
		orders: make(map[OrderId]OrderEntry),
	}
}

func (o *Orderbook) Start() {
	o.cond = sync.NewCond(&sync.Mutex{})
	go o.PruneGoodForDayOrders()
}

func (o *Orderbook) Shutdown() {
	o.shutdown.Store(true)
	o.cond.Signal()
}

func (o *Orderbook) Size() int {
	return len(o.orders)
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

// TODO: Finish this function
func (o *Orderbook) CanFullyFill(
	side Side,
	price Price,
	quantity Quantity,
) bool {
	if !o.CanMatch(side, price) {
		return false
	}
	var _ Price
	return false
}

// MatchOrders checks the bid and asks maps and attempt to
// generate Trades from their stored Orders. If a bid is available at
// a price greater than or equal to that of the best ask, a trade is generated.
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

func (o *Orderbook) AddOrder(order Order) (Trades, error) {
	o.m.Lock()
	defer o.m.Unlock()

	if _, exists := o.orders[order.OrderId()]; exists {
		return nil, fmt.Errorf(
			"Order %d already exists",
			order.OrderId(),
		)
	}

	// Market orders are converted to GoodTillCancel with the max/worst price
	// available in the asks, ensuring execution with the best asks price once
	// `MatchOrders` is called

	if order.OrderType() == Market {
		var err error
		if order.Side() == Buy && !o.asks.Empty() {
			maxPrice, _, _ := o.asks.Last()
			err = order.ToGoodTillCancel(maxPrice)
		} else if order.Side() == Sell && !o.bids.Empty() {
			maxPrice, _, _ := o.bids.Last()
			err = order.ToGoodTillCancel(maxPrice)
		} else {
			// TODO: Improve this message
			return nil, fmt.Errorf("invalid state")
		}
		if err != nil {
			return nil, err
		}
	}

	if order.OrderType() == FillAndKill &&
		!o.CanMatch(order.Side(), order.Price()) {
		return nil, fmt.Errorf(
			"Order %d cannot be filled immediately",
			order.OrderId(),
		)
	}

	if order.OrderType() == FillOrKill {
	}

	var orders Orders

	// TODO: Refactor this code create zero values by default
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

func (o *Orderbook) CancelOrder(orderId OrderId) error {
	o.m.Lock()
	defer o.m.Unlock()
	return o.cancelOrder(orderId)
}

func (o *Orderbook) CancelOrders(orderIds OrderIds) error {
	o.m.Lock()
	defer o.m.Unlock()
	for _, id := range orderIds {
		if err := o.cancelOrder(id); err != nil {
			return err
		}
	}
	return nil
}

func (o *Orderbook) cancelOrder(orderId OrderId) error {
	if _, exists := o.orders[orderId]; !exists {
		return fmt.Errorf("Order %d does not exist", orderId)
	}

	entry := o.orders[orderId]
	order := entry.order
	location := entry.location
	delete(o.orders, orderId)

	if order.Side() == Buy {
		orders, _ := o.bids.Get(order.Price())
		orders.RemoveAt(location)
		if orders.IsEmpty() {
			o.bids.Delete(order.Price())
		}
	} else {
		orders, _ := o.asks.Get(order.Price())
		orders.RemoveAt(location)
		if orders.IsEmpty() {
			o.asks.Delete(order.Price())
		}
	}
	return nil
}

func (o *Orderbook) ModifyOrder(modify OrderModify) (Trades, error) {
	if _, exists := o.orders[modify.OrderId()]; !exists {
		return nil, fmt.Errorf("Order %d does not exist", modify.OrderId())
	}

	existingOrder := o.orders[modify.OrderId()].order
	o.CancelOrder(modify.OrderId())
	return o.AddOrder(modify.ToOrder(existingOrder.OrderType()))
}

// PruneGoodForDayOrders removes all GoodForDay orders from the orderbook at 4pm
// EST. This is a naive implementation and should be improved.
func (o *Orderbook) PruneGoodForDayOrders() error {
	endHour := 16 // 16:00 / 04:00 PM

	for {
		// get current time and convert to local time
		now := time.Now()
		location, err := time.LoadLocation("America/New_York")
		if err != nil {
			return fmt.Errorf("error loading location: %v", err)
		}
		next := now.In(location)

		if next.Hour() >= endHour {
			next = next.AddDate(0, 0, 1)
		}
		next = time.Date(
			next.Year(),
			next.Month(),
			next.Day(),
			endHour,
			0,
			0,
			0,
			next.Location(),
		)

		until := next.Sub(now) + (100 * time.Millisecond)
		if o.shutdown.Load() {
			return nil
		}

		// TODO: update this func to use a channel
		// to trigger cond.Singal() separately from
		// the timer.
		go func() {
			o.cond.L.Lock()
			defer o.cond.L.Unlock()

			time.Sleep(until)
			o.cond.Signal()
		}()

		o.cond.Wait()

		var orderIds OrderIds
		{
			o.m.Lock()
			defer o.m.Unlock()

			for id, entry := range o.orders {
				if entry.order.OrderType() == GoodForDay {
					orderIds = append(orderIds, id)
				}
			}
		}
		if err := o.CancelOrders(orderIds); err != nil {
			return fmt.Errorf("error cancelling orders: %v", err)
		}
	}
}

func (o *Orderbook) OrderInfo() OrderbookLevelsInfo {
	var (
		bidsInfo LevelsInfo
		asksInfo LevelsInfo
	)
	for bids := o.bids.Begin(); bids.Valid(); bids.Next() {
		var l LevelInfo
		var q Quantity
		l.Price = bids.Key()
		orders := bids.Value()
		it := orders.Iterator()
		for order, ok := it.Next(); ok; order, ok = it.Next() {
			q += order.remainingQuantity
		}
		l.Quantity = q
		bidsInfo = append(bidsInfo, l)
	}

	for asks := o.asks.Begin(); asks.Valid(); asks.Next() {
		var l LevelInfo
		var q Quantity
		l.Price = asks.Key()
		orders := asks.Value()
		it := orders.Iterator()
		for order, ok := it.Next(); ok; order, ok = it.Next() {
			q += order.remainingQuantity
		}
		l.Quantity = q
		asksInfo = append(asksInfo, l)
	}

	return OrderbookLevelsInfo{
		bids: bidsInfo,
		asks: asksInfo,
	}
}
