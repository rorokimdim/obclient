package orderbook

import (
	"encoding/json"
	"fmt"
	"math"
)

type Price = float64
type Size = float64

type OrderBook struct {
	bids    map[Price]Size
	asks    map[Price]Size
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

func New() *OrderBook {
	return &OrderBook{
		bids:    map[Price]Size{},
		asks:    map[Price]Size{},
		bestBid: Entry{},
		bestAsk: Entry{Price: math.MaxFloat64},
	}
}

func (ob *OrderBook) Summarize() OrderBookSummary {
	return OrderBookSummary{
		BestAsk: ob.bestAsk,
		BestBid: ob.bestBid,
		Spread:  ob.bestAsk.Price - ob.bestBid.Price,
	}
}

func (ob *OrderBook) String() string {
	summary := ob.Summarize()
	b, err := json.Marshal(summary)
	if err != nil {
		panic(fmt.Sprintf("Could not marshall orderbook summary: %v", summary))
	}
	return string(b)
}

func (ob *OrderBook) Update(asks []Entry, bids []Entry) bool {
	prevBestAsk := ob.bestAsk
	prevBestBid := ob.bestBid
	prevSpread := computeSpread(ob)

	for _, a := range asks {
		if a.Size == 0 {
			delete(ob.asks, a.Price)
		} else {
			ob.asks[a.Price] = a.Size
		}
	}

	for _, b := range bids {
		if b.Size == 0 {
			delete(ob.bids, b.Price)
		} else {
			ob.bids[b.Price] = b.Size
		}
	}

	ob.bestAsk = computeBestAsk(ob)
	ob.bestBid = computeBestBid(ob)
	spread := computeSpread(ob)

	bestAskChanged := ob.bestAsk != prevBestAsk
	bestBidChanged := ob.bestBid != prevBestBid
	spreadChanged := spread != prevSpread

	return bestAskChanged || bestBidChanged || spreadChanged

}

func computeSpread(ob *OrderBook) float64 {
	return ob.bestAsk.Price - ob.bestBid.Price
}

func computeBestAsk(ob *OrderBook) Entry {
	best := Entry{Price: math.MaxFloat64}

	for price, size := range ob.asks {
		if price < best.Price {
			best.Price = price
			best.Size = size
		}
	}

	return best
}

func computeBestBid(ob *OrderBook) Entry {
	best := Entry{}

	for price, size := range ob.bids {
		if price > best.Price {
			best.Price = price
			best.Size = size
		}
	}

	return best
}
