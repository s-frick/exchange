package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/s-frick/exchange/api"
)

type (
	Order struct {
		ID        api.OrderID
		Limit     *Limit
		UserID    api.UserID
		Size      api.Size
		BidAsk    api.BidAsk
		Timestamp int64
	}
	Orders []*Order

	Match struct {
		Ask        *Order
		Bid        *Order
		SizeFilled api.Size
		Price      api.Price
	}

	Orderbook struct {
		asks Limits
		bids Limits

		mu     sync.RWMutex
		Limits map[api.Price]*Limit
		Orders map[api.OrderID]*Order
	}
)

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }

func NewOrder(bidAsk api.BidAsk, size api.Size, userID api.UserID) *Order {
	// bidAsk := api.BID
	// if !bid {
	// 	bidAsk = api.ASK
	// }
	return &Order{
		ID:        api.OrderID(int64(rand.Intn(10000000000000))),
		UserID:    api.UserID(userID),
		Size:      api.Size(size),
		BidAsk:    bidAsk,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		asks:   []*Limit{},
		bids:   []*Limit{},
		Limits: make(map[api.Price]*Limit),
		Orders: make(map[api.OrderID]*Order),
	}
}

func (ob *Orderbook) PlaceLimitOrder(price api.Price, o *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	limit := ob.Limits[price]

	if limit == nil {
		limit = NewLimit(price)

		if o.BidAsk == api.BID {
			ob.bids = append(ob.bids, limit)
		} else {
			ob.asks = append(ob.asks, limit)
		}
		ob.Limits[price] = limit
	}
	ob.Orders[o.ID] = o
	limit.AddOrder(o)
	fmt.Printf("limit: %v\n", limit)
}

func (ob *Orderbook) PlaceMarketOrder(o *Order) []Match {
	if o.BidAsk == api.BID {
		return ob.matchOrder(o, ob.Asks())
	} else {
		return ob.matchOrder(o, ob.Bids())
	}
}

func (ob *Orderbook) cancelOrder(o *Order) {
	limit := o.Limit
	limit.DeleteOrder(o)
	delete(ob.Orders, o.ID)
	if len(limit.Orders) == 0 {
		ob.clearLimit(limit)
	}
}

func (ob *Orderbook) CancelOrder(id api.OrderID) {
	order := ob.Orders[id]
	ob.cancelOrder(order)
}

func (ob *Orderbook) matchOrder(o *Order, limits Limits) []Match {
	totalVolume := ob.totalVolume(limits)
	matches := []Match{}
	if o.Size > totalVolume {
		panic(fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", totalVolume, o.Size))
	}

	for _, limit := range limits {
		limitMatches := limit.Fill(o)
		matches = append(matches, limitMatches...)

		if len(limit.Orders) == 0 {
			ob.clearLimit(limit)
		}
	}
	return matches
}

func (ob *Orderbook) BidTotalVolume() api.Size {
	return ob.totalVolume(ob.bids)
}

func (ob *Orderbook) AskTotalVolume() api.Size {
	return ob.totalVolume(ob.asks)
}

func (ob *Orderbook) totalVolume(limits Limits) api.Size {
	totalVolume := api.Size(0.0)

	for i := 0; i < len(limits); i++ {
		totalVolume += limits[i].TotalVolume
	}
	return totalVolume
}

func (ob *Orderbook) Asks() Limits {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}

func (ob *Orderbook) Bids() Limits {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}

func (ob *Orderbook) clearLimit(l *Limit) {
	delete(ob.Limits, l.Price)
	for i := 0; i < len(ob.bids); i++ {
		if ob.bids[i] == l {
			ob.bids[i] = ob.bids[len(ob.bids)-1]
			ob.bids = ob.bids[:len(ob.bids)-1]
			return
		}
	}
	for i := 0; i < len(ob.asks); i++ {
		if ob.asks[i] == l {
			ob.asks[i] = ob.asks[len(ob.asks)-1]
			ob.asks = ob.asks[:len(ob.asks)-1]
			return
		}
	}
}
