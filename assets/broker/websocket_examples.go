package broker

import (
	"fmt"
	"os"
	"time"
)

type MessagePrinter struct{}

func (printer *MessagePrinter) ProcessMessage(data interface{}) {
	fmt.Println(data)
}

func BindToStream(connector *WebsocketConnector) {
	// bind to positions stream
	connector.Send(map[string]interface{}{
		"event": "bind",
		"feed":  "P",
	})

	// bind to books stream
	connector.Send(map[string]interface{}{
		"event":  "bind",
		"feed":   "B",
		"feedId": 42,
	})
}

func UnbindFromStream(connector *WebsocketConnector) {
	// unbind from positions stream
	connector.Send(map[string]interface{}{
		"event": "unbind",
		"feed":  "P",
	})

	// unbind from books stream
	connector.Send(map[string]interface{}{
		"event":  "unbind",
		"feed":   "B",
		"feedId": 42,
	})
}

func SendRequest(connector *WebsocketConnector) {
	// send new order
	connector.Send(map[string]interface{}{
		"event":  "request",
		"reqId":  12345,
		"method": "add",
		"content": map[string]interface{}{
			"instrument":    "BTC-USD",
			"clientOrderId": 123456,
			"price":         9000 * EfxUnit,
			"size":          2 * EfxUnit,
			"side":          "bid",
			"type":          "limitFOK",
			"cod":           true,
		},
	})

	// cancel it
	connector.Send(map[string]interface{}{
		"event":  "request",
		"reqId":  1234567,
		"method": "del",
		"content": map[string]interface{}{
			"clientOrderId": 123456,
		},
	})
}

func RunAllWebsocketsExamples() {
	connector, _ := NewWebsocketConnector(os.Getenv("KEY"),
		os.Getenv("KEY"), true, &MessagePrinter{})

	BindToStream(connector)
	time.Sleep(time.Second)
	UnbindFromStream(connector)
	time.Sleep(time.Second)
	SendRequest(connector)
	time.Sleep(time.Second)

}
