package dydx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/rorokimdim/obclient/orderbook"
)

func New(websocketURI string) *DyDx {
	return &DyDx{websocketURI: websocketURI}
}

func (dydx *DyDx) SubscribeToOrderBook(ctx context.Context, pairId string) chan DyDxOrderBookMessageResult {
	resultCh := make(chan DyDxOrderBookMessageResult)

	go func() {
		conn, _, err := websocket.DefaultDialer.Dial(dydx.websocketURI, nil)
		if err != nil {
			resultCh <- DyDxOrderBookMessageResult{Err: err}
			close(resultCh)
			return
		}
		defer conn.Close()

		if err := dydx.subscribe(conn, "v4_orderbook", pairId); err != nil {
			resultCh <- DyDxOrderBookMessageResult{Err: err}
			close(resultCh)
			return
		}

		for {
			if err := ctx.Err(); err != nil {
				log.Printf("Closing websocket connection...")
				werr := conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				)
				if werr != nil {
					resultCh <- DyDxOrderBookMessageResult{
						Err: fmt.Errorf("could not send close message on websocket: %w", werr),
					}
				}

				resultCh <- DyDxOrderBookMessageResult{Err: err}
				close(resultCh)
				break
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				resultCh <- DyDxOrderBookMessageResult{Err: err}
				close(resultCh)
				break
			}

			contents, err := parseMessage(message)
			if err != nil {
				resultCh <- DyDxOrderBookMessageResult{Err: err}
				close(resultCh)
				break
			}

			if len(contents.Asks) > 0 || len(contents.Bids) > 0 {
				resultCh <- DyDxOrderBookMessageResult{Message: contents}
			}
		}
	}()

	return resultCh
}

func (dydx *DyDx) subscribe(conn *websocket.Conn, topic string, pairId string) error {
	subscribeMessage := fmt.Sprintf(`
	{
	  "type": "subscribe",
	  "channel": "%s",
	  "id": "%s"
	}
	`, topic, pairId)

	return conn.WriteMessage(websocket.TextMessage, []byte(subscribeMessage))
}

func parseMessage(message []byte) (DyDxOrderBookContents, error) {
	m := DyDxMessage{}

	if err := json.Unmarshal(message, &m); err != nil {
		return DyDxOrderBookContents{}, err
	}

	if m.Type == "connected" {
		return DyDxOrderBookContents{}, nil
	} else if m.Type == "subscribed" {
		return parseOnSubscribeMessage(m.MessageId, m.Contents)
	} else if m.Type == "channel_data" {
		return parseOnUpdateMessage(m.MessageId, m.Contents)
	} else {
		return DyDxOrderBookContents{}, fmt.Errorf("unexpected message type: %s", m.Type)
	}
}

func parseOnUpdateMessage(messageId int, message []byte) (DyDxOrderBookContents, error) {
	m := DyDxOrderBookContents{}

	onUpdateMessage := DyDxOnUpdateMessageContents{MessageId: messageId}
	if err := json.Unmarshal(message, &onUpdateMessage); err != nil {
		return m, err
	}

	m, err := onUpdateMessage.toDyDxMessage()
	if err != nil {
		return m, err
	}

	return m, nil
}

func parseOnSubscribeMessage(messageId int, message []byte) (DyDxOrderBookContents, error) {
	m := DyDxOrderBookContents{MessageId: messageId}

	if err := json.Unmarshal(message, &m); err != nil {
		return m, err
	}

	return m, nil
}

func (umsg DyDxOnUpdateMessageContents) toDyDxMessage() (DyDxOrderBookContents, error) {
	m := DyDxOrderBookContents{}

	toEntries := func(xs [][2]string) ([]orderbook.Entry, error) {
		entries := []orderbook.Entry{}
		for _, x := range xs {
			entry := orderbook.Entry{}
			price, err := strconv.ParseFloat(x[0], 64)
			if err != nil {
				return nil, fmt.Errorf("expected price to be a float; got %v", x[0])

			}

			size, err := strconv.ParseFloat(x[1], 64)
			if err != nil {
				return nil, fmt.Errorf("expected size to be a float; got %v", x[1])

			}
			entry.Price = price
			entry.Size = size
			entries = append(entries, entry)
		}
		return entries, nil
	}

	asks, err := toEntries(umsg.Asks)
	if err != nil {
		return m, err
	}

	bids, err := toEntries(umsg.Bids)
	if err != nil {
		return m, err
	}

	m.Asks = asks
	m.Bids = bids
	m.MessageId = umsg.MessageId

	return m, nil
}
