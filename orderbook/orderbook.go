package orderbook

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
)

func New() *OrderBook {
	return &OrderBook{
		bids:    map[Price]sizeMessageId{},
		asks:    map[Price]sizeMessageId{},
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

func (ob *OrderBook) Update(messageId int, asks []Entry, bids []Entry, uncross bool) bool {
	prevBestAsk := ob.bestAsk
	prevBestBid := ob.bestBid
	prevSpread := computeSpread(prevBestAsk, prevBestBid)

	for _, a := range asks {
		if a.Size == 0 {
			delete(ob.asks, a.Price)
		} else {
			ob.asks[a.Price] = sizeMessageId{size: a.Size, messageId: messageId}
		}
	}

	for _, b := range bids {
		if b.Size == 0 {
			delete(ob.bids, b.Price)
		} else {
			ob.bids[b.Price] = sizeMessageId{size: b.Size, messageId: messageId}
		}
	}

	bestAsk, bestBid, spread := compute(ob, uncross)
	ob.bestAsk = bestAsk
	ob.bestBid = bestBid

	bestAskChanged := ob.bestAsk != prevBestAsk
	bestBidChanged := ob.bestBid != prevBestBid
	spreadChanged := spread != prevSpread

	return bestAskChanged || bestBidChanged || spreadChanged
}

func compute(ob *OrderBook, uncross bool) (bestAsk Entry, bestBid Entry, spread float64) {
	bestAsk = computeBestAsk(ob)
	bestBid = computeBestBid(ob)
	spread = computeSpread(bestAsk, bestBid)

	// Uncross
	// See https://docs.dydx.exchange/api_integration-guides/how_to_uncross_orderbook#how-to-uncross
	//
	uncrossCount := 0
	for uncross && spread <= 0 && len(ob.bids) > 0 && len(ob.asks) > 0 {
		uncrossCount += 1
		log.Printf("crossing detected; uncrossing count=%d", uncrossCount)

		bestAskMessageId := ob.asks[bestAsk.Price].messageId
		bestBidMessageId := ob.bids[bestBid.Price].messageId

		if bestBidMessageId < bestAskMessageId {
			delete(ob.bids, bestBid.Price)
		} else if bestBidMessageId > bestAskMessageId {
			delete(ob.asks, bestAsk.Price)
		} else {
			bestAskSize := bestAsk.Size
			bestBidSize := bestBid.Size
			if bestBidSize > bestAskSize {
				delete(ob.asks, bestAsk.Price)
				ob.bids[bestBid.Price] = sizeMessageId{
					size:      bestBidSize - bestAskSize,
					messageId: bestBidMessageId,
				}
			} else if bestBidSize < bestAskSize {
				delete(ob.bids, bestBid.Price)
				ob.asks[bestAsk.Price] = sizeMessageId{
					size:      bestAskSize - bestBidSize,
					messageId: bestAskMessageId,
				}
			} else {
				delete(ob.asks, bestAsk.Price)
				delete(ob.bids, bestBid.Price)
			}
		}

		bestAsk = computeBestAsk(ob)
		bestBid = computeBestBid(ob)
		spread = computeSpread(bestAsk, bestBid)
	}

	return bestAsk, bestBid, spread
}

func computeBestAsk(ob *OrderBook) Entry {
	best := Entry{Price: math.MaxFloat64}

	for price, sizeMessageId := range ob.asks {
		if price < best.Price {
			best.Price = price
			best.Size = sizeMessageId.size
		}
	}

	return best
}

func computeBestBid(ob *OrderBook) Entry {
	best := Entry{}

	for price, sizeMessageId := range ob.bids {
		if price > best.Price {
			best.Price = price
			best.Size = sizeMessageId.size
		}
	}

	return best
}

func computeSpread(bestAsk Entry, bestBid Entry) float64 {
	return bestAsk.Price - bestBid.Price
}
