package orderbook

type LevelInfo struct {
	Price    Price
	Quantity Quantity
}

type LevelsInfo []LevelInfo
