package server

import (
	"errors"
	"fmt"

	"github.com/s-frick/exchange/api"
)

type Subscription interface {
	footprint_method()
}

type OrderBookSubscription struct {
	market api.Market
}

func (o OrderBookSubscription) footprint_method() {}

type OrdersSubscription struct {
	userID api.UserID
}

func (o OrdersSubscription) footprint_method() {}

func NewSubscription(r WSRequest) (Subscription, error) {
	fmt.Printf("r: %T\n", r)
	switch v := r.(type) {
	case *orderBookRequest:
		return OrderBookSubscription{market: v.Market}, nil
	case *ordersRequest:
		return OrdersSubscription{userID: api.UserID("todo")}, nil
	}

	return nil, errors.New("invalid request")
}
