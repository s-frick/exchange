package orderbook

import (
	"reflect"
	"testing"

	"github.com/s-frick/exchange/api"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(api.BID, 5, "0")
	buyOrderB := NewOrder(api.BID, 8, "0")
	buyOrderC := NewOrder(api.BID, 10, "0")

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(api.ASK, 10, "0")
	sellOrderB := NewOrder(api.ASK, 5, "0")
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(9_000, sellOrderB)

	assert(t, len(ob.Orders), 2)
	assert(t, ob.Orders[sellOrderA.ID], sellOrderA)
	assert(t, ob.Orders[sellOrderB.ID], sellOrderB)
	assert(t, len(ob.asks), 2)

}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(api.ASK, 20, "0")
	ob.PlaceLimitOrder(10_000, sellOrderA)

	buyOrderA := NewOrder(api.BID, 10, "0")
	matches := ob.PlaceMarketOrder(buyOrderA)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), api.Size(10.0))
	assert(t, matches[0].Ask, sellOrderA)
	assert(t, matches[0].Bid, buyOrderA)
	assert(t, matches[0].SizeFilled, api.Size(10.0))
	assert(t, matches[0].Price, api.Price(10_000.0))
	assert(t, buyOrderA.IsFilled(), true)

}

func TestPlaceOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buyOrderA := NewOrder(api.BID, 5, "0")
	buyOrderB := NewOrder(api.BID, 8, "0")
	buyOrderC := NewOrder(api.BID, 10, "0")
	buyOrderD := NewOrder(api.BID, 1, "0")

	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(5_000, buyOrderD)

	assert(t, ob.BidTotalVolume(), api.Size(24.00))

	sellOrder := NewOrder(api.ASK, 20, "0")
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidTotalVolume(), api.Size(4.0))
	assert(t, len(matches), 3)
	assert(t, len(ob.bids), 1)

}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(api.BID, 4, "0")
	ob.PlaceLimitOrder(10_000, buyOrder)

	assert(t, len(ob.bids), 1)
	assert(t, ob.BidTotalVolume(), api.Size(4.0))

	ob.CancelOrder(buyOrder.ID)

	assert(t, len(ob.bids), 0)
	assert(t, ob.BidTotalVolume(), api.Size(0.0))

	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)
}
