package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/s-frick/exchange/api"
	"github.com/s-frick/exchange/exchange"
)

var (
	clients       = make(map[*websocket.Conn]bool)
	subscriptions = make(map[*websocket.Conn]Subscription)
	inbound       = make(chan WSRequest)
	book_signal   = make(chan api.Market)
	// TODO: how to handle subs of user orders?
	// orders_signal = make(chan api.Market)
	outbound = make(chan MessageWithMarket)
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type (
	Order struct {
		UserID    string
		ID        string
		Price     float64
		Size      float64
		BID_ASK   api.BidAsk
		Timestamp int64
	}
	Message           struct{}
	MessageWithMarket struct {
		msg    []byte
		market api.Market
	}
)

type CryptoSettlementService struct {
}

func (c *CryptoSettlementService) Transfer(from string, to string, amount float64, market api.Market) {
	fmt.Printf("transfer %2.f from %s to %s, market [%s]", amount, from, to, market)
}

func Start() {
	e := echo.New()

	ex, err := exchange.New("", &CryptoSettlementService{}, book_signal)
	if err != nil {
		log.Fatal(err)
	}
	ex.AddOrderbook("ETH")

	e.GET("/book/:market", ex.HandleGetBook)
	e.GET("/book/:market/subscribe", handleConnections)
	e.POST("/order", ex.HandlePlaceOrder)
	e.DELETE("/order/:id", ex.HandelCancelOrder)

	go handleSubscriptions()
	go handleSignals(ex)
	go broadcast()

	fmt.Println(e.Start(":8080"))
}

func handleSignals(ex *exchange.Exchange) {
	for {
		market := <-book_signal
		fmt.Printf("handleSignal [%s]\n", market)
		book, err := ex.GetBook(market)
		if err != nil {
			continue
		}
		fmt.Printf("%+v\n", book)
		data, err := json.Marshal(book)
		if err != nil {
			continue
		}
		outbound <- MessageWithMarket{msg: []byte(data), market: market}
	}
}

func broadcast() {
	for {
		msg := <-outbound
		for c := range subscriptions {
			sub := subscriptions[c]
			fmt.Println("GUENTHER")
			switch s := sub.(type) {
			case OrderBookSubscription:
				if s.market == msg.market {
					w, err := c.NextWriter(websocket.TextMessage)
					if err != nil {
						log.Println(err)
						continue
					}
					w.Write(msg.msg)
					fmt.Printf("broadcast [%s]\n", msg.market)

					if err := w.Close(); err != nil {
						log.Println(err)
						continue
					}
					continue
				}
			default:
				continue
			}

		}
		// for client := range clients {
		// 	w, err := client.NextWriter(websocket.TextMessage)
		// 	if err != nil {
		// 		return
		// 	}
		// 	w.Write(msg.msg)
		//
		// 	if err := w.Close(); err != nil {
		// 		return
		// 	}
		// }
	}
}

func handleConnections(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		panic(err)
	}
	defer ws.Close()
	clients[ws] = true

	for {
		t, msg, err := ws.ReadMessage()
		if err != nil {
			delete(clients, ws)
			return err
		}
		if t == websocket.TextMessage {
			req, err := NewWsRequest(msg, ws)
			if err != nil {
				continue
			}
			inbound <- req
		}
	}
}

func handleSubscriptions() {
	for {
		req := <-inbound
		sub, err := NewSubscription(req)
		if err != nil {
			panic(err)
		}
		subscriptions[req.ws()] = sub
	}
}
