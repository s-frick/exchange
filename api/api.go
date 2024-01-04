package api

type (
	OrderID   int64
	UserID    string
	Price     float64
	Size      float64
	Market    string
	BidAsk    int
	OrderType int

	PlaceOrderRequest struct {
		UserID UserID    `json:"user_id"`
		Type   OrderType `json:"type"`
		BidAsk BidAsk    `json:"bid_ask"`
		Size   Size      `json:"size"`
		Price  Price     `json:"price"`
		Market Market    `json:"market"`
	}
	PlaceOrderResponse struct {
		OrderID OrderID `json:"order_id"`
	}

	Limit struct {
		Price       Price   `json:"price"`
		Orders      []Order `json:"orders"`
		TotalVolume Size    `json:"total_volume"`
	}

	OrderBookData struct {
		TotalAskVolume Size    `json:"total_ask_volume"`
		TotalBidVolume Size    `json:"total_bid_volume"`
		Asks           []Limit `json:"asks"`
		Bids           []Limit `json:"bids"`
	}

	Order struct {
		Price     Price  `json:"price"`
		Size      Size   `json:"size"`
		BidAsk    BidAsk `json:"bid_ask"`
		Timestamp int64  `json:"timestamp"`
	}
)

const (
	ASK BidAsk = iota
	BID
)

const (
	LIMIT OrderType = iota
	MARKET
)

func (b BidAsk) String() string {
	if b == BID {
		return "BID"
	}
	return "ASK"
}

func (o OrderType) String() string {
	if o == LIMIT {
		return "LIMIT"
	}
	return "MARKET"
}
