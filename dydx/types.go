package dydx

import (
	"encoding/json"

	"github.com/rorokimdim/obclient/orderbook"
)

type DyDx struct {
	websocketURI string
}

type DyDxMessage struct {
	Id        string
	Type      string
	MessageId int `json:"message_id"`
	Contents  json.RawMessage
}

type DyDxOrderBookContents struct {
	MessageId int `json:"-"`
	Bids      []orderbook.Entry
	Asks      []orderbook.Entry
}

type DyDxOnUpdateMessageContents struct {
	MessageId int `json:"-"`
	Bids      [][2]string
	Asks      [][2]string
}

type DyDxOrderBookMessageResult struct {
	Message DyDxOrderBookContents
	Err     error
}
