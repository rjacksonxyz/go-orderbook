package orderbook

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
