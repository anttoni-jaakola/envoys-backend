package main

import (
	"fmt"
	"os"
	"time"
)

func printResult(result interface{}, err error) {
	if err != nil {
		fmt.Printf("request failed: %v\n", err)
	} else {
		fmt.Println(result)
	}
}

func GetInstruments(connector *RestConnector) {
	printResult(connector.Request("instruments", nil))
}

func GetPositions(connector *RestConnector) {
	printResult(connector.Request("positions", nil))
}

func GetLimits(connector *RestConnector) {
	printResult(connector.Request("limits", nil))
}

func GetCounterpartyLimits(connector *RestConnector) {
	printResult(connector.Request("climits", nil))
}

func GetSettlmentRequests(connector *RestConnector) {
	printResult(connector.Request("settlementRequests", nil))
}

func GetSettlmentTransactions(connector *RestConnector) {
	printResult(connector.Request("settlementTransactions", nil))
}

func GetBook(connector *RestConnector) {
	params := map[string]interface{}{
		"instrument": "BTC-USD",
		"tradable":   true,
	}

	printResult(connector.Request("book", params))
}

func GetDealHistory(connector *RestConnector) {
	instrument := "BTC-USD" // get deals for BTC-USD only
	now := time.Now().UTC()
	start := now.Add(-time.Hour * 24 * 3) // get deals for last 3 days
	dealsTo := now.UnixNano() / int64(time.Millisecond)
	dealsFrom := start.UnixNano() / int64(time.Millisecond)
	var dealsTill *int
	var deals [][]interface{}
	for {
		params := map[string]interface{}{
			"from":       dealsFrom,
			"instrument": instrument,
			"limit":      250,
		}
		if dealsTill != nil {
			params["till"] = dealsTill
		} else {
			params["to"] = dealsTo
		}

		result, _ := connector.Request("dealHistory", params)
		castResult := result.([]interface{})
		if len(castResult) == 0 {
			break
		}

		for _, deal := range castResult {
			castDeal := deal.([]interface{})
			deals = append(deals, castDeal)
			dealID := castDeal[11].(int)
			if dealsTill == nil || *dealsTill < dealID {
				dealsTill = &dealID
			}
		}
	}

	for _, deal := range deals {
		fmt.Println(deal)
	}
}

// Buy 2 BTC-USD @ 9000
func AddOrder(connector *RestConnector) {
	instrument := "BTC-USD"
	clientOrderID := 42
	price := 9000 * EfxUnit
	size := 2 * EfxUnit
	side := "bid"
	orderType := "limitFOK"
	cancelOnDisconnect := true

	printResult(connector.Request("add", map[string]interface{}{
		"instrument":    instrument,
		"clientOrderId": clientOrderID,
		"price":         price,
		"size":          size,
		"side":          side,
		"type":          orderType,
		"cod":           cancelOnDisconnect,
	}))
}

func DelOrder(connector *RestConnector) {
	// del by client order id
	printResult(connector.Request("del", map[string]interface{}{
		"clientOrderId": 42,
	}))

	// del by exchange order id (from add response)
	printResult(connector.Request("del", map[string]interface{}{
		"orderId": 42,
	}))
}

func DelAllOrders(connector *RestConnector) {
	// delete all BTC-USD orders
	printResult(connector.Request("delAll", map[string]interface{}{
		"instrument": "BTC-USD",
	}))

	// delete all orders for all instruments
	printResult(connector.Request("delAll", nil))
}

func RunAllRestExamples() {
	connector := NewRestConnector(os.Getenv("KEY"),
		os.Getenv("SECRET"), true)
	GetInstruments(connector)
	GetPositions(connector)
	GetLimits(connector)
	GetCounterpartyLimits(connector)
	GetSettlmentRequests(connector)
	GetSettlmentTransactions(connector)
	GetBook(connector)
	GetDealHistory(connector)
	AddOrder(connector)
	DelOrder(connector)
	DelAllOrders(connector)
}
