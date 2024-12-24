package orderbook

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
