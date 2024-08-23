package orderbook

type Price = float64
type Size = float64

type sizeMessageId struct {
	size      Size
	messageId int
}

type OrderBook struct {
	bids    map[Price]sizeMessageId
	asks    map[Price]sizeMessageId
	bestBid Entry
	bestAsk Entry
}

type Entry struct {
	Price Price `json:"price,string"`
	Size  Size  `json:"size,string"`
}

type OrderBookSummary struct {
	BestBid Entry   `json:"best_bid"`
	BestAsk Entry   `json:"best_ask"`
	Spread  float64 `json:"spread"`
}
