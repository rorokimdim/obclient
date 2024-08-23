package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/rorokimdim/obclient/dydx"
	"github.com/rorokimdim/obclient/orderbook"
)

// See https://docs.dydx.exchange/api_integration-indexer/indexer_websocket
//
// testnet: 'wss://indexer.v4testnet.dydx.exchange/v4/ws'
// staging: wss://indexer.v4staging.dydx.exchange/v4/ws
// real?: 'wss://indexer.dydx.trade/v4/ws'
const defaultDydxURI = "wss://indexer.dydx.trade/v4/ws"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	dydxURI, ok := os.LookupEnv("DYDX_WSS_URL")
	if !ok {
		dydxURI = defaultDydxURI
	}

	d := dydx.New(dydxURI)
	ob := orderbook.New()

	go func() {
		for r := range d.SubscribeToOrderBook(ctx, "ETH-USD") {
			if r.Err != nil {
				if r.Err != context.Canceled {
					log.Printf("An error occurred: %v", r.Err)
				}
				close(done)
			}

			updated := ob.Update(r.Message.MessageId, r.Message.Asks, r.Message.Bids)
			if updated {
				fmt.Println(ob.String())
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Exiting. Please wait...")
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
