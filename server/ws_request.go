package server

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/s-frick/exchange/api"
)

type WSRequest interface {
	req_type() wsRequestType
	ws() *websocket.Conn
}

type wsRequestType string

const (
	orderBook wsRequestType = "order_book"
	orders                  = "orders"
)

type orderBookRequest struct {
	Type   wsRequestType
	Market api.Market
	client *websocket.Conn
}

func (o orderBookRequest) req_type() wsRequestType {
	return orderBook
}
func (o orderBookRequest) ws() *websocket.Conn {
	return o.client
}

func NewOrderBookRequest(market api.Market, client *websocket.Conn) WSRequest {
	return &orderBookRequest{
		Type:   orderBook,
		Market: market,
		client: client,
	}
}

type ordersRequest struct {
	Type   wsRequestType
	client *websocket.Conn
}

func (o ordersRequest) req_type() wsRequestType {
	return orders
}
func (o ordersRequest) ws() *websocket.Conn {
	return o.client
}
func NewOrdersRequest(client *websocket.Conn) WSRequest {
	return &ordersRequest{
		Type:   orders,
		client: client,
	}
}

func NewWsRequest(jsonData []byte, client *websocket.Conn) (WSRequest, error) {
	var r map[string]interface{}
	json.Unmarshal(jsonData, &r)

	if r["Type"] != nil {
		switch r["Type"].(string) {
		case string(orderBook):
			var o orderBookRequest
			json.Unmarshal(jsonData, &o)
			return NewOrderBookRequest(o.Market, client), nil
		case string(orders):
			return NewOrdersRequest(client), nil
		}
	}

	return nil, errors.New("invalid request")
}
