package service

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	proto "github.com/cryptogateway/backend-envoys/server/proto"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"strings"
	"sync/atomic"
	"time"

	"net/http"
)

type MarketService struct {
	Context *assets.Context

	host, socket string
	nonce        int64
	connect      *WebsocketConnector
}

// Initialization - perform actions.
func (m *MarketService) Initialization() {
	if m.Context.MarketTest {
		m.host = "https://test.finerymarkets.com/api/"
	} else {
		m.host = "https://trade.finerymarkets.com/api/"
	}
	m.nonce = time.Now().UnixNano()

	go func() {
		connector, err := m.WebsocketConnector()
		if err != nil {
			return
		}
		m.connect = connector
	}()
}

// request - new connect to apis.
func (m *MarketService) request(method string, content map[string]interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); m.Context.Debug(r) {
			return
		}
	}()

	if content == nil {
		content = make(map[string]interface{})
	}

	content["nonce"] = atomic.AddInt64(&m.nonce, 1)
	content["timestamp"] = time.Now().UnixNano() / int64(time.Millisecond)
	contentString, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	hasher := hmac.New(sha512.New384, []byte(m.Context.MarketSecret))
	hasher.Write([]byte(method))
	hasher.Write(contentString)
	signature := hasher.Sum(nil)

	request, err := http.NewRequest(http.MethodPost, m.host+method, strings.NewReader(string(contentString)))
	if err != nil {
		return nil, err
	}
	defer request.Body.Close()

	request.Header.Add("EFX-Key", m.Context.MarketKey)
	request.Header.Add("EFX-Sign", base64.StdEncoding.EncodeToString(signature[:]))
	request.Header.Add("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	responseString, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(responseString, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type WebsocketConnector struct {
	key        string
	secret     string
	m          *MarketService
	connection *websocket.Conn
}

func (m *MarketService) message(data interface{}) {
	if err := m.Context.Publish(data, "exchange", "broker/depth"); err != nil {
		return
	}
}

func (m *MarketService) WebsocketConnector() (*WebsocketConnector, error) {

	if m.Context.MarketTest {
		m.socket = "wss://test.finerymarkets.com/ws"
	} else {
		m.socket = "wss://trade.finerymarkets.com/ws"
	}

	// Generate signature for authentication
	content := map[string]interface{}{
		"nonce":     m.nonce,
		"timestamp": time.Now().UnixNano() / int64(time.Millisecond),
	}
	contentString, _ := json.Marshal(content)

	hasher := hmac.New(sha512.New384, []byte(m.Context.MarketSecret))
	hasher.Write(contentString)
	signature := hasher.Sum(nil)

	// Add authentication headers to ws connection request
	header := http.Header{}
	header.Add("EFX-Key", m.Context.MarketKey)
	header.Add("EFX-Sign", base64.StdEncoding.EncodeToString(signature[:]))
	header.Add("EFX-Content", string(contentString))

	// Connect websocket
	connection, _, err := websocket.DefaultDialer.Dial(m.socket, header)
	if err != nil {
		return nil, err
	}

	result := &WebsocketConnector{
		key:        m.Context.MarketKey,
		secret:     m.Context.MarketSecret,
		m:          m,
		connection: connection,
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
		var request interface{}
		connector.connection.ReadJSON(&request)

		if listData, ok := request.([]interface{}); ok {
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

		if fields, ok := request.([]interface{}); ok {

			var (
				response proto.MarketDepth
			)
			response.Symbol = fields[1].(string)

			if extract, ok := fields[3].([]interface{}); ok {

				for i := 0; i < len(extract[0].([]interface{})); i++ {

					if bid, ok := extract[0].([]interface{}); ok {
						columns := bid[i].([]interface{})
						if len(columns) == 2 {
							response.Fields = append(response.Fields, &proto.MarketDepth_Attributes{
								Assigning: "BID",
								Price:     columns[0].(float64) / 100000000,
								Size:      columns[1].(float64) / 100000000,
							})
						}
						if len(columns) == 3 {
							response.Fields = append(response.Fields, &proto.MarketDepth_Attributes{
								Assigning: "BID",
								Action:    columns[0].(string),
								Price:     columns[1].(float64) / 100000000,
								Size:      columns[2].(float64) / 100000000,
							})
						}
					}

				}

				for i := 0; i < len(extract[1].([]interface{})); i++ {

					if ask, ok := extract[1].([]interface{}); ok {
						columns := ask[i].([]interface{})
						if len(columns) == 2 {
							response.Fields = append(response.Fields, &proto.MarketDepth_Attributes{
								Assigning: "ASK",
								Price:     columns[0].(float64) / 100000000,
								Size:      columns[1].(float64) / 100000000,
							})
						}
						if len(columns) == 3 {
							response.Fields = append(response.Fields, &proto.MarketDepth_Attributes{
								Assigning: "ASK",
								Action:    columns[0].(string),
								Price:     columns[1].(float64) / 100000000,
								Size:      columns[2].(float64) / 100000000,
							})
						}
					}

				}

				connector.m.message(&response)
			}
		}
	}
}
