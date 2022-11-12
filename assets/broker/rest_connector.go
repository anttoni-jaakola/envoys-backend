package broker

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type RestConnector struct {
	key    string
	secret string
	host   string
	nonce  int64
}

func NewRestConnector(key, secret string, isDemo bool) *RestConnector {
	var host string
	if isDemo {
		host = "https://test.finerymarkets.com/api/"
	} else {
		host = "https://trade.finerymarkets.com/api/"
	}

	return &RestConnector{
		key:    key,
		secret: secret,
		host:   host,
		nonce:  time.Now().UnixNano(),
	}
}

func (connector *RestConnector) Request(method string, content map[string]interface{}) (interface{}, error) {
	if content == nil {
		content = make(map[string]interface{})
	}

	content["nonce"] = atomic.AddInt64(&connector.nonce, 1)
	content["timestamp"] = time.Now().UnixNano() / int64(time.Millisecond)
	contentString, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	hasher := hmac.New(sha512.New384, []byte(connector.secret))
	hasher.Write([]byte(method))
	hasher.Write(contentString)
	signature := hasher.Sum(nil)

	request, err := http.NewRequest(http.MethodPost, connector.host+method,
		strings.NewReader(string(contentString)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("EFX-Key", connector.key)
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
