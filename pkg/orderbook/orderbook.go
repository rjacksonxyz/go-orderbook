package orderbook

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
	intialQuantiy     Quantity
	remainingQuantity Quantity
}

func (o *Order) New() Order {
	return Order{}
}
