package okxv5

import (
	"encoding/json"
	"fmt"
	"gotrader/pkg/httpx"
	"gotrader/pkg/utils"
	"gotrader/trader/constant"
	"gotrader/trader/types"
	"net/http"
	"strconv"
	"time"
)

var httpClient = httpx.NewClient()

type RestClient struct {
	apiKey     string
	secretKey  string
	passPhrase string
}

type BaseOkRsp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func NewRestClient(apiKey, secretKey, passPhrase string) *RestClient {
	client := &RestClient{
		apiKey:     apiKey,
		secretKey:  secretKey,
		passPhrase: passPhrase,
	}
	return client
}

func (client *RestClient) Request(method string, uri string, payload []byte) ([]byte, *http.Response, error) {
	var param string
	if payload != nil {
		param = string(payload)
	}
	currentTime := IsoTime()
	toSignStr := currentTime + method + uri + param
	signature := utils.GenBase64Digest(utils.HmacSha256(toSignStr, client.secretKey))
	url := RestUrl + uri
	head := map[string]string{
		"Content-Type":         "application/json",
		"OK-ACCESS-KEY":        client.apiKey,
		"OK-ACCESS-SIGN":       signature,
		"OK-ACCESS-TIMESTAMP":  currentTime,
		"OK-ACCESS-PASSPHRASE": client.passPhrase,
	}
	args := &httpx.Request{
		Url:    url,
		Head:   head,
		Method: method,
		Body:   payload,
	}
	body, res, err := httpClient.Request(args)
	if err != nil {
		return nil, res, err
	}
	return *body, res, err
}

func (client *RestClient) HttpGet(url string) ([]byte, *http.Response, error) {
	body, res, err := httpClient.Get(url)
	if err != nil {
		return nil, res, err
	}
	return *body, res, nil
}

type TickerRsp struct {
	BaseOkRsp
	Data []Ticker `json:"data"`
}

func (t *TickerRsp) valid() bool {
	return t.Code == "0" && len(t.Data) > 0
}

type Ticker struct {
	InstType  string `json:"instType"`
	InstID    string `json:"instId"`
	Last      string `json:"last"`
	LastSz    string `json:"lastSz"`
	AskPx     string `json:"askPx"`
	AskSz     string `json:"askSz"`
	BidPx     string `json:"bidPx"`
	BidSz     string `json:"bidSz"`
	Open24H   string `json:"open24h"`
	High24H   string `json:"high24h"`
	Low24H    string `json:"low24h"`
	VolCcy24H string `json:"volCcy24h"`
	Vol24H    string `json:"vol24h"`
	Ts        string `json:"ts"`
	SodUtc0   string `json:"sodUtc0"`
	SodUtc8   string `json:"sodUtc8"`
}

func (client *RestClient) FetchTickers() ([]*types.BookTicker, error) {
	queryDict := map[string]interface{}{}
	queryDict["instType"] = "SWAP"
	payload := utils.UrlEncodeParams(queryDict)
	url := RestUrl + fmt.Sprintf(TickersRest, payload)

	body, _, err := client.HttpGet(url)
	if err != nil {
		log.Errorf("ok get /api/v5/market/tickers err:%v", err)
		return nil, err
	}

	response := new(TickerRsp)
	if err = json.Unmarshal(body, response); err != nil {
		log.Errorf("ok get /api/v5/market/tickers parser err:%v", err)
		return nil, err
	}

	if !response.valid() {
		err := fmt.Errorf("ok get /api/v5/market/tickers fail, code:%s, msg:%s", response.Code, response.Msg)
		return nil, err
	}

	if len(response.Data) == 0 {
		err := fmt.Errorf("ok get /api/v5/market/tickers empty")
		return nil, err
	}

	result, err := tickersTransform(response)
	if err != nil {
		err := fmt.Errorf("ok get /api/v5/market/tickers transform err:%s", err)
		return nil, err
	}

	return result, nil
}

func tickersTransform(response *TickerRsp) ([]*types.BookTicker, error) {
	result := make([]*types.BookTicker, 0, len(response.Data))
	for _, ticker := range response.Data {
		symbol := OkInstId2Symbol(ticker.InstID)
		ex := constant.OkxV5Future
		askPrice, _ := strconv.ParseFloat(ticker.AskPx, 64)
		bidPrice, _ := strconv.ParseFloat(ticker.BidPx, 64)
		askSize, _ := strconv.ParseFloat(ticker.AskSz, 64)
		bidSize, _ := strconv.ParseFloat(ticker.BidSz, 64)
		exchangeTs, _ := strconv.ParseInt(ticker.Ts, 10, 64)
		bookTicker := &types.BookTicker{
			Symbol:     symbol,
			Exchange:   ex,
			AskPrice:   askPrice,
			BidPrice:   bidPrice,
			AskQty:     askSize,
			BidQty:     bidSize,
			ExchangeTs: exchangeTs,
			Ts:         utils.Millisec(time.Now()),
		}
		result = append(result, bookTicker)
	}
	return result, nil
}