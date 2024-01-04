package exchange

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/s-frick/exchange/api"
	"github.com/s-frick/exchange/orderbook"
)

type SettlementService interface {
	Transfer(from string, to string, amount float64, market api.Market)
}

type Exchange struct {
	PrivateKey   string
	Client       SettlementService
	orderbooks   map[api.Market]*orderbook.Orderbook
	ordersMarket map[api.OrderID]api.Market
	subscriber   chan api.Market
}

func New(privateKey string, client SettlementService, subscriber chan api.Market) (*Exchange, error) {
	orderbooks := make(map[api.Market]*orderbook.Orderbook)
	ordersMarket := make(map[api.OrderID]api.Market)

	return &Exchange{
		PrivateKey:   privateKey,
		Client:       client,
		orderbooks:   orderbooks,
		ordersMarket: ordersMarket,
		subscriber:   subscriber,
	}, nil
}

func (ex *Exchange) AddOrderbook(market api.Market) {
	ex.orderbooks[market] = orderbook.NewOrderbook()
}

func (ex *Exchange) HandleGetBook(c echo.Context) error {
	market := c.Param("market")
	data, err := ex.GetBook(api.Market(market))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": err.Error()})
	}
	return c.JSON(http.StatusOK, data)
}

type MarketNotFoundError struct {
	market string
}

func (e *MarketNotFoundError) Error() string {
	return "market not found: " + e.market
}

func (ex *Exchange) GetBook(market api.Market) (api.OrderBookData, error) {
	ob, ok := ex.orderbooks[api.Market(market)]
	if !ok {
		return api.OrderBookData{}, &MarketNotFoundError{market: string(market)}
	}

	orderbookData := api.OrderBookData{
		TotalAskVolume: ob.AskTotalVolume(),
		TotalBidVolume: ob.BidTotalVolume(),
		Asks:           collectOrders(ob.Asks()),
		Bids:           collectOrders(ob.Bids()),
	}

	return orderbookData, nil
}

func (ex *Exchange) HandlePlaceOrder(c echo.Context) error {
	var placeOrderData api.PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := api.Market(placeOrderData.Market)
	order := orderbook.NewOrder(placeOrderData.BidAsk, placeOrderData.Size, placeOrderData.UserID)
	ob := ex.orderbooks[market]
	ex.ordersMarket[order.ID] = market

	switch placeOrderData.Type {
	case api.LIMIT:
		ob.PlaceLimitOrder(placeOrderData.Price, order)

		log.Printf("new LIMIT order => [%+v] | price [%.2f] | size [%.2f]\n", placeOrderData.BidAsk, order.Limit.Price, order.Size)
		resp := &api.PlaceOrderResponse{
			OrderID: order.ID,
		}

		ex.subscriber <- market
		return c.JSON(http.StatusOK, resp)

	case api.MARKET:
		log.Printf("new MARKET order => [%+v] | size [%.2f]\n", placeOrderData.BidAsk, order.Size)
		matches := ob.PlaceMarketOrder(order)
		if err := ex.handleMatches(matches, market); err != nil {
			return err
		}
		resp := &api.PlaceOrderResponse{
			OrderID: order.ID,
		}

		ex.subscriber <- market
		return c.JSON(http.StatusOK, resp)
	}

	return c.JSON(http.StatusBadRequest, "")
}

func (ex *Exchange) HandelCancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	market := ex.ordersMarket[api.OrderID(id)]
	ob := ex.orderbooks[market]
	ob.CancelOrder(api.OrderID(id))
	log.Println("order canceled id => ", id)

	return c.JSON(http.StatusOK, map[string]any{"msg": "order deleted"})
}

func (ex *Exchange) handleMatches(matches []orderbook.Match, market api.Market) error {
	for _, match := range matches {
		fromUser := match.Ask.UserID
		toUser := match.Bid.UserID

		ex.Client.Transfer(string(fromUser), string(toUser), float64(match.SizeFilled), market)
	}
	return nil
}

func collectOrders(limits []*orderbook.Limit) []api.Limit {
	ls := make([]api.Limit, 0)
	for _, limit := range limits {
		orders := make([]api.Order, 0)
		for _, order := range limit.Orders {
			order := api.Order{
				Price:     order.Limit.Price,
				Size:      order.Size,
				BidAsk:    order.BidAsk,
				Timestamp: order.Timestamp,
			}
			orders = append(orders, order)
		}
		l := api.Limit{
			Price:       limit.Price,
			Orders:      orders,
			TotalVolume: limit.TotalVolume,
		}
		ls = append(ls, l)
	}
	fmt.Printf("collectOrders: %+v\n", ls)
	return ls
}
