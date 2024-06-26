package trader

import "github.com/wsg011/gotrader/trader/types"

type Strategy interface {
	GetName() string
	GetSymbol() string
	GetHedgeSymbol() string
	OnBookTicker(bookticker *types.BookTicker)
	OnOrderBook(orderbook *types.OrderBook)
	OnTrade(trade *types.Trade)
	OnOrder(order []*types.Order)
	Run()
	Start()
	Close()
}
