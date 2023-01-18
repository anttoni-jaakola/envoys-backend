package marketplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// https://api3.binance.com/api/v3/ticker/price?symbol=ETHBTC
// https://api-pub.bitfinex.com/v2/trades/tETHUSD/hist?limit=1
// https://api.kucoin.com/api/v1/prices?base=USD&currencies=ETH
// https://poloniex.com/public?command=returnTradeHistory&currencyPair=USDT_ETH&limit=1
// https://api.kuna.io/v3/tickers?symbols=btcuah
// https://ftx.com/api/markets/BTC_USD/trades
// https://api.huobi.pro/market/detail/merged?symbol=ethusdt
const (
	ApiExchangePoloniex = "https://poloniex.com/public?command=returnTradeHistory&currencyPair=%v_%v&limit=1"
	ApiExchangeKucoin   = "https://api.kucoin.com/api/v1/prices?base=%v&currencies=%v"
	ApiExchangeBitfinex = "https://api-pub.bitfinex.com/v2/trades/t%v%v/hist?limit=1"
	ApiExchangeBinance  = "https://api3.binance.com/api/v3/ticker/price?symbol=%v%v"
	ApiExchangeKuna     = "https://api.kuna.io/v3/tickers?symbols=%v%v"
	ApiExchangeFtx      = "https://ftx.com/api/markets/%v_%v/trades"
	ApiExchangeHuobi    = "https://api.huobi.pro/market/detail/merged?symbol=%v%v"
)

type Marketplace struct {
	client http.Client
	count  int
	scale  []float64
}

func Price() *Marketplace {
	return &Marketplace{}
}

// Unit - price scale.
func (p *Marketplace) Unit(base, quote string) float64 {
	base, quote, p.count, p.scale = strings.ToUpper(base), strings.ToUpper(quote), 0, []float64{}

	p.scale = append(p.scale, p.getBinance(base, quote))
	p.scale = append(p.scale, p.getBitfinex(base, quote))
	p.scale = append(p.scale, p.getKucoin(base, quote))
	p.scale = append(p.scale, p.getPoloniex(base, quote))
	p.scale = append(p.scale, p.getKuna(base, quote))
	p.scale = append(p.scale, p.getFtx(base, quote))
	p.scale = append(p.scale, p.getHuobi(base, quote))

	var (
		price float64
	)

	for i := 0; i < len(p.filter()); i++ {
		price += p.filter()[i]
	}
	time.Sleep(1 * time.Second)

	if len(p.filter()) > 0 {
		return price / float64(len(p.filter()))
	}

	return 0
}

// filter - remove zero value.
func (p *Marketplace) filter() []float64 {
	var r []float64
	for _, str := range p.scale {
		if str != 0 {
			r = append(r, str)
		}
	}
	return r
}

// len - price reverse symbol.
func (p *Marketplace) len() bool {
	if p.count == 1 {
		return true
	}
	p.count += 1
	return false
}

// getHuobi - get price.
func (p *Marketplace) getHuobi(base, quote string) float64 {

	var (
		result map[string]interface{}
	)

	if strings.Contains(base, "USD") {
		base = "USDT"
	}

	if strings.Contains(quote, "USD") {
		quote = "USDT"
	}

	base, quote = strings.ToLower(base), strings.ToLower(quote)

	request, err := p.request(fmt.Sprintf(ApiExchangeHuobi, base, quote))
	if err != nil {
		return 0
	}

	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if strings.Contains(result["status"].(string), "error") {
		if p.len() {
			return 0
		}

		if price := p.getHuobi(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if price, ok := result["tick"].(map[string]interface{})["close"]; ok {
		return price.(float64)
	}

	return 0
}

// getFtx - get price.
func (p *Marketplace) getFtx(base, quote string) float64 {

	var (
		result map[string]interface{}
	)

	request, err := p.request(fmt.Sprintf(ApiExchangeFtx, base, quote))
	if err != nil {
		if p.len() {
			return 0
		}

		if price := p.getFtx(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if price, ok := result["result"].([]interface{})[0].(map[string]interface{})["price"]; ok {
		return price.(float64)
	}

	return 0
}

// getKuna - get price.
func (p *Marketplace) getKuna(base, quote string) float64 {

	var (
		result [][]interface{}
	)

	if strings.Contains(base, "USDT") {
		base = "USD"
	}

	if strings.Contains(quote, "USDT") {
		quote = "USD"
	}

	request, err := p.request(fmt.Sprintf(ApiExchangeKuna, base, quote))
	if err != nil {
		if p.len() {
			return 0
		}

		if price := p.getKuna(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if price, ok := result[0][1].(float64); ok {
		return price
	}

	return 0
}

// getPoloniex - get price.
func (p *Marketplace) getPoloniex(base, quote string) float64 {

	var (
		result []interface{}
	)

	if strings.Contains(base, "USD") {
		base = "USDT"
	}

	if strings.Contains(quote, "USD") {
		quote = "USDT"
	}

	request, err := p.request(fmt.Sprintf(ApiExchangePoloniex, quote, base))
	if err != nil {
		return 0
	}

	if strings.Contains(string(request), "Invalid currency pair") {
		if p.len() {
			return 0
		}

		if price := p.getPoloniex(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if len(result) > 0 {
		if price, ok := result[0].(map[string]interface{})["rate"]; ok {
			if price, err := strconv.ParseFloat(price.(string), 64); err == nil {
				return price
			}
		}
	}

	return 0
}

// getKucoin - get price.
func (p *Marketplace) getKucoin(base, quote string) float64 {

	var (
		result map[string]interface{}
	)

	if strings.Contains(base, "USDT") {
		base = "USD"
	}

	if strings.Contains(quote, "USDT") {
		quote = "USD"
	}

	request, err := p.request(fmt.Sprintf(ApiExchangeKucoin, base, quote))
	if err != nil {
		return 0
	}

	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if msg, err := result["msg"]; err && strings.Contains(msg.(string), "Unsupported base currency") {
		if p.len() {
			return 0
		}

		if price := p.getKucoin(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if data, ok := result["data"]; ok {
		if price, ok := data.(map[string]interface{})[quote]; ok {
			if price, err := strconv.ParseFloat(price.(string), 64); err == nil {
				return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
			}
		}
	}

	return 0
}

// getBitfinex - get price.
func (p *Marketplace) getBitfinex(base, quote string) float64 {

	var (
		result [][]interface{}
	)

	if strings.Contains(base, "USDT") {
		base = "USD"
	}

	if strings.Contains(quote, "USDT") {
		quote = "USD"
	}

	request, err := p.request(fmt.Sprintf(ApiExchangeBitfinex, base, quote))
	if err != nil {
		return 0
	}

	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if len(result) == 0 || len(result[0]) != 4 {
		if p.len() {
			return 0
		}

		if price := p.getBitfinex(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if price, ok := result[0][3].(float64); ok {
		return price
	}

	return 0
}

// getBinance - get price.
func (p *Marketplace) getBinance(base, quote string) float64 {

	var (
		result map[string]string
	)

	request, err := p.request(fmt.Sprintf(ApiExchangeBinance, base, quote))
	if err != nil {
		if p.len() {
			return 0
		}

		if price := p.getBinance(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	if err := json.Unmarshal(request, &result); err != nil {
		return 0
	}

	if price, ok := result["price"]; ok {
		if price, err := strconv.ParseFloat(price, 64); err == nil {
			return price
		}
	}

	return 0
}

// request - get request service.
func (p *Marketplace) request(url string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Status code: %v", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
