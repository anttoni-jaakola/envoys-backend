package broker

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"time"

	"github.com/gorilla/websocket"
	"net/http"
)

type IMessageProcessor interface {
	ProcessMessage(msg interface{})
}

type WebsocketConnector struct {
	key        string
	secret     string
	processor  IMessageProcessor
	connection *websocket.Conn
}

func NewWebsocketConnector(key string, secret string, isDemo bool, processor IMessageProcessor) (*WebsocketConnector, error) {
	var host string
	if isDemo {
		host = "wss://test.finerymarkets.com/ws"
	} else {
		host = "wss://trade.finerymarkets.com/ws"
	}

	// Generate signature for authentication
	content := map[string]interface{}{
		"nonce":     time.Now().UnixNano(),
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
	}
	contentString, _ := json.Marshal(content)

	hasher := hmac.New(sha512.New384, []byte(secret))
	hasher.Write(contentString)
	signature := hasher.Sum(nil)

	// Add authentication headers to ws connection request
	header := http.Header{}
	header.Add("EFX-Key", key)
	header.Add("EFX-Sign", base64.StdEncoding.EncodeToString(signature[:]))
	header.Add("EFX-Content", string(contentString))

	// Connect websocket
	conection, _, err := websocket.DefaultDialer.Dial(host, header)
	if err != nil {
		return nil, err
	}

	result := &WebsocketConnector{
		key:        key,
		secret:     secret,
		processor:  processor,
		connection: conection,
	}

	connected := make(chan struct{}, 1)
	go result.listen(connected)
	<-connected

	return result, nil
}

func (connector *WebsocketConnector) Send(data interface{}) {
	if err := connector.connection.WriteJSON(data); err != nil {
		fmt.Printf("failed to send '%+v': %v", data, err)
	}
}

func (connector *WebsocketConnector) listen(connected chan<- struct{}) {
	for {
		var data interface{}
		connector.connection.ReadJSON(&data)

		if listData, ok := data.([]interface{}); ok {
			if messageType, ok := listData[0].(string); ok && messageType == "X" {
				if errorCode, ok := listData[3].(float64); ok && errorCode == 0 {
					connected <- struct{}{}
					continue
				} else {
					fmt.Printf("failed to connect: %+v", listData[3])
					return
				}
			}
		}

		spew.Dump(data)

		//connector.processor.ProcessMessage(data)
	}
}
