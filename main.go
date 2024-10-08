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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Syntax: %s pair-id\n", os.Args[0])
		fmt.Println("\nExamples:")
		fmt.Printf("%s ETH-USD\n", os.Args[0])
		fmt.Printf("%s BTC-USD\n", os.Args[0])
		return
	}

	pairId := os.Args[1]
	config := readConfig()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	d := dydx.New(config.dydxURI)
	ob := orderbook.New()

	go func() {
		for r := range d.SubscribeToOrderBook(ctx, pairId) {
			if r.Err != nil {
				if r.Err != context.Canceled {
					log.Printf("An error occurred: %v", r.Err)
				}
				close(done)
			}

			updated := ob.Update(
				r.Message.MessageId,
				r.Message.Asks,
				r.Message.Bids,
				config.uncrossOrderBook,
			)
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

type Config struct {
	dydxURI          string
	uncrossOrderBook bool
}

func readConfig() Config {
	dydxURI, ok := os.LookupEnv("DYDX_WSS_URL")
	if !ok {
		dydxURI = defaultDydxURI
	}

	uncrossOrderBook := true
	v, ok := os.LookupEnv("UNCROSS_ORDER_BOOK")
	if ok {
		uncrossOrderBook = v == "1"
	}

	return Config{
		dydxURI:          dydxURI,
		uncrossOrderBook: uncrossOrderBook,
	}
}

// See https://docs.dydx.exchange/api_integration-indexer/indexer_websocket
//
// testnet: wss://indexer.v4testnet.dydx.exchange/v4/ws
// staging: wss://indexer.v4staging.dydx.exchange/v4/ws
// real: wss://indexer.dydx.trade/v4/ws
const defaultDydxURI = "wss://indexer.dydx.trade/v4/ws"
