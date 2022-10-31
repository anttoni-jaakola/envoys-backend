package service

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"github.com/cryptogateway/backend-envoys/assets"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type MarketService struct {
	Context *assets.Context

	host  string
	nonce int64
}

// Initialization - perform actions.
func (m *MarketService) Initialization() {
	if m.Context.MarketTest {
		m.host = "https://test.finerymarkets.com/api/"
	} else {
		m.host = "https://trade.finerymarkets.com/api/"
	}
	m.nonce = time.Now().UnixNano()
}

// request - new connect to apis.
func (m *MarketService) request(method string, content map[string]interface{}) (interface{}, error) {
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

	request, err := http.NewRequest(http.MethodPost, m.host+method,
		strings.NewReader(string(contentString)))
	if err != nil {
		return nil, err
	}

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
