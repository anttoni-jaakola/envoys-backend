package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
	"github.com/wangluozhe/requests/utils"
	"strconv"
	"time"
)

func main() {

	method := "instruments"
	key := "GZizBSzOrsjHPquTlsJQCxRTx7fh6ERzd0YjXz8vMtY"
	secret := "rzC2HVdeJFddkQdKYlRg8mCriMxBVuMULERsMNGmeof"

	nonce := time.Now().Unix()
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	data := url.NewData()
	data.Set("nonce", strconv.FormatInt(nonce, 10))
	data.Set("timestamp", fmt.Sprint(timestamp))

	request := struct {
		Nonce     int64 `json:"nonce"`
		Timestamp int64 `json:"timestamp"`
	}{
		Nonce:     nonce,
		Timestamp: timestamp,
	}

	marshal, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}

	signature := utils.Btoa(utils.HmacSHA384(method+string(marshal), secret))

	headers := url.NewHeaders()
	headers.Set("EFX-Key", key)
	headers.Set("EFX-Sign", signature)

	req := url.NewRequest()
	req.Headers = headers
	req.Data = data

	post, err := requests.Post(fmt.Sprintf("https://test.finerymarkets.com/api/%s", method), req)
	if err != nil {
		panic(err)
	}

	fmt.Println("Signature:", signature)
	fmt.Println("Payload:", method+string(marshal))
	fmt.Println("Request:", data.Encode())
	spew.Dump(post)
}
