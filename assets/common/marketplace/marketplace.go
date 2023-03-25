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

// These are constants that define the URLs of various cryptocurrency exchanges. They are used to retrieve the latest
// trade history for a given currency pair from the respective exchanges.
// https://api3.binance.com/api/v3/ticker/price?symbol=ETHBTC
// https://api-pub.bitfinex.com/v2/trades/tETHUSD/hist?limit=1
// https://api.kucoin.com/api/v1/prices?base=USD&currencies=ETH
// https://poloniex.com/public?command=returnTradeHistory&currencyPair=USDT_ETH&limit=1
// https://api.kuna.io/v3/tickers?symbols=btcuah
// https://api.huobi.pro/market/detail/merged?symbol=ethusdt
const (
	ApiExchangePoloniex = "https://poloniex.com/public?command=returnTradeHistory&currencyPair=%v_%v&limit=1"
	ApiExchangeKucoin   = "https://api.kucoin.com/api/v1/prices?base=%v&currencies=%v"
	ApiExchangeBitfinex = "https://api-pub.bitfinex.com/v2/trades/t%v%v/hist?limit=1"
	ApiExchangeBinance  = "https://api3.binance.com/api/v3/ticker/price?symbol=%v%v"
	ApiExchangeKuna     = "https://api.kuna.io/v3/tickers?symbols=%v%v"
	ApiExchangeHuobi    = "https://api.huobi.pro/market/detail/merged?symbol=%v%v"
)

// The Marketplace struct is a data structure used to store information about an e-commerce marketplace. It contains a
// http.Client for making web requests, an int to keep track of the number of items in the marketplace, and a slice of
// float64s to represent the scale of the marketplace.
type Marketplace struct {
	client http.Client
	count  int
	scale  []float64
}

// Price - The purpose of this function is to create a new instance of the Marketplace struct and return a pointer to it.
func Price() *Marketplace {
	return &Marketplace{}
}

// Unit - This function is part of a larger program that is designed to calculate the market price of a currency pair. The
// function uses several exchanges to calculate the market price by taking the average of the results from each
// exchange. It takes a base and quote currency as parameters and returns the average market price of the currency pair.
func (p *Marketplace) Unit(base, quote string) float64 {

	// This code is used to initialize four variables in Go. The variables are base, quote, p.count and p.scale. The base
	// and quote variables are set to the uppercase versions of the strings passed in as arguments, while the p.count
	// variable is set to 0 and the p.scale variable is set to an empty slice of float64 values.
	base, quote, p.count, p.scale = strings.ToUpper(base), strings.ToUpper(quote), 0, []float64{}

	// This code is adding data to the "p.scale" array from the various cryptocurrency exchanges. The purpose of this is
	// likely to collect the data from the different exchanges and store it in the "p.scale" array for further analysis.
	p.scale = append(p.scale, p.getBinance(base, quote))
	p.scale = append(p.scale, p.getBitfinex(base, quote))
	p.scale = append(p.scale, p.getKucoin(base, quote))
	p.scale = append(p.scale, p.getPoloniex(base, quote))
	p.scale = append(p.scale, p.getKuna(base, quote))
	p.scale = append(p.scale, p.getHuobi(base, quote))

	var (
		price float64
	)

	// This loop iterates through a slice of prices (p.filter()) and adds each element of the slice to the variable price.
	// It is used to calculate the total cost of the items in the slice.
	for i := 0; i < len(p.filter()); i++ {
		price += p.filter()[i]
	}

	// The purpose of time.Sleep(1 * time.Second) is to pause the program for one second before continuing. This is useful
	// for when you need to pause a program to allow other processes to run or to slow down the execution speed.
	time.Sleep(1 * time.Second)

	// The purpose of this if statement is to check if the length of the list returned by the p.filter() function is greater
	// than 0. If the length is greater than 0, the statement will return the price divided by the length of the list
	// returned by the p.filter() function. Otherwise, the statement will not do anything.
	if len(p.filter()) > 0 {
		return price / float64(len(p.filter()))
	}

	return 0
}

// filter - This function is used to filter a Marketplace structure by removing all the zero values from the scale field. It
// takes the Marketplace structure as an input parameter and returns a slice of float64 values that only contains the
// non-zero elements.
func (p *Marketplace) filter() []float64 {
	var r []float64

	// This for loop is used to iterate through a collection of strings stored in the variable p.scale. For each string in
	// the collection, it checks if the string is not equal to zero, then appends it to the variable r.
	for _, str := range p.scale {
		if str != 0 {
			r = append(r, str)
		}
	}

	return r
}

// len - This function is used to determine if the number of items in the Marketplace is equal to one. It increments the count
// variable and returns true if the count is equal to one, and false otherwise.
func (p *Marketplace) len() bool {

	// The purpose of this code is to check if the count of a variable 'p' is equal to one. If it is, the code will return the boolean value 'true'.
	if p.count == 1 {
		return true
	}

	// This line of code increases the value of the variable count by 1. This is commonly used to keep track of how many
	// times a particular action has taken place.
	p.count += 1

	return false
}

// getHuobi - This function is used to get the exchange rate from Huobi for a given base and quote currency. It checks if the base
// and quote currency parameters contain USD, and if so, it changes them to USDT. It then makes a request to the Huobi
// API, and if the request was successful, it returns the exchange rate as a float. If the request failed, it checks if
// there is another exchange rate available for the reversed currency pair, and if so, it returns the inverse of that
// value instead. Otherwise, it returns 0.
func (p *Marketplace) getHuobi(base, quote string) float64 {

	// The purpose of the variable 'result' is to create a map of type string to interface. This means that it can store any
	// type of value (string, integer, boolean, etc.) as the value for each key. This type of data structure is often used
	// to store key-value pairs in a program.
	var (
		result map[string]interface{}
	)

	// This if statement checks to see if the string "base" contains the string "USD", and if it does, changes the value of
	// the string "base" to "USDT".
	if strings.Contains(base, "USD") {
		base = "USDT"
	}

	// This code checks if the string "quote" contains the substring "USD". If it does, it will replace the string "quote"
	// with the string "USDT". This is likely done to ensure that the quote is in the correct format.
	if strings.Contains(quote, "USD") {
		quote = "USDT"
	}

	// The purpose of this statement is to convert two strings (base and quote) to lower case letters and assign the new
	// strings to base and quote. strings.ToLower is a function from the strings package that converts strings to lower case letters.
	base, quote = strings.ToLower(base), strings.ToLower(quote)

	//This code is setting up an HTTP request to the Huobi Exchange API with specific base and quote parameters. It is
	//using the fmt.Sprintf() function to format the request string using the ApiExchangeHuobi string and the base and
	//quote variables. The request is then saved to the request variable, and an error is checked for. If there is an
	//error, the function will return 0.
	request, err := p.request(fmt.Sprintf(ApiExchangeHuobi, base, quote))
	if err != nil {
		return 0
	}

	// The purpose of this code is to attempt to unmarshal a JSON object stored in the variable 'request' into a variable
	// 'result'. If the unmarshal fails, the code will return 0.
	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	// This if statement is used to check if the string "status" contains the word "error". If the string "status" does
	// contain the word "error", then the code within the if statement will be executed.
	if strings.Contains(result["status"].(string), "error") {

		// The purpose of this code is to check if the length of the variable "p" is zero (0). If the length of the variable is
		// zero, the code returns 0.
		if p.len() {
			return 0
		}

		// This code is checking the price of a particular asset on the Huobi exchange. The "if" statement is checking if the
		// price is greater than 0. If it is, the code then uses the "Decimal" library to calculate the inverse of the price,
		// round it to 8 decimal places, and then return the rounded float.
		if price := p.getHuobi(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	// This code checks if the given result contains a "tick" key, which is a map of strings to interfaces. It then checks
	// if it contains a "close" key. If it does, it returns the price as a float64.
	if price, ok := result["tick"].(map[string]interface{})["close"]; ok {
		return price.(float64)
	}

	return 0
}

// getKuna - This function is part of a class named Marketplace and its purpose is to get the exchange rate between two currencies
// from the Kuna exchange. It first checks if the two currencies are US dollars and if not, it makes a request to the
// Kuna API to get the exchange rate. If the request fails, it will check if the rate is available in the reverse
// direction, and if not, it will return 0. If the request succeeds, it will return the exchange rate as a float.
func (p *Marketplace) getKuna(base, quote string) float64 {

	// The purpose of the variable result is to declare a 2-dimensional array of type interface{}. This type can hold any
	// value, allowing for the storage of multiple data types in the same array. This is useful for creating general-purpose
	// data structures that can store multiple different types of data.
	var (
		result [][]interface{}
	)

	// This code checks if the string "base" contains the substring "USDT" and if so, it sets the string "base" to "USD".
	// This could be used to normalize the input string and make sure it is always in the same format.
	if strings.Contains(base, "USDT") {
		base = "USD"
	}

	// This code snippet is checking if the string "quote" contains the substring "USDT". If it does, it sets the string
	// "quote" to "USD". This could be used to ensure that the value of "quote" is always set to "USD" when it includes the
	// substring "USDT".
	if strings.Contains(quote, "USDT") {
		quote = "USD"
	}

	// This code is creating an HTTP request using the fmt.Sprintf() function to generate a URL string with the parameters
	// base and quote. The request is then sent using the p.request() function. The purpose of this code is to retrieve
	// information from the Kuna cryptocurrency exchange using the API endpoint ApiExchangeKuna.
	request, err := p.request(fmt.Sprintf(ApiExchangeKuna, base, quote))
	if err != nil {

		// The purpose of this code is to check if the length of the variable "p" is zero (0). If the length of the variable is
		// zero, the code returns 0.
		if p.len() {
			return 0
		}

		// This code is checking the price of a given quote and base currency in kuna. If the price is greater than 0, it
		// returns a rounded decimal float with 8 decimal places.
		if price := p.getKuna(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	// The purpose of this code is to attempt to unmarshal a JSON object stored in the variable 'request' into a variable
	// 'result'. If the unmarshal fails, the code will return 0.
	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	// This if statement is used to check if the value stored in result[0][1] is a float64 type. If it is, the statement
	// returns the value stored in price. If not, the statement evaluates to false and nothing is returned.
	if price, ok := result[0][1].(float64); ok {
		return price
	}

	return 0
}

// getPoloniex - This function is part of a Marketplace struct that is used to retrieve the exchange rate between two different
// currencies from the Poloniex exchange. The function takes in two strings as parameters that represent the base and
// quote currencies and returns the exchange rate as a float. It checks if the base and quote currencies contain "USD"
// and replaces them with "USDT" (which is the currency used by Poloniex to represent USD). It then makes an API request
// with these currencies to the Poloniex exchange and retrieves a JSON response. It unmarshals the response into an
// interface and checks if the rate exists in it. Finally, it converts the rate into a float and returns it.
func (p *Marketplace) getPoloniex(base, quote string) float64 {

	// The purpose of the variable "result" is to store a slice of type interface{}. An interface is an abstract type that
	// contains no data or methods, but it is a type that any type can implement. A slice of type interface{} is a slice of
	// values of any type, which is useful for storing different types of values in the same slice.
	var (
		result []interface{}
	)

	// This if statement checks to see if the string "base" contains the string "USD", and if it does, changes the value of
	// the string "base" to "USDT".
	if strings.Contains(base, "USD") {
		base = "USDT"
	}

	// This code checks if the string "quote" contains the substring "USD". If it does, it will replace the string "quote"
	// with the string "USDT". This is likely done to ensure that the quote is in the correct format.
	if strings.Contains(quote, "USD") {
		quote = "USDT"
	}

	// This code is attempting to construct a request to the Poloniex exchange API in order to obtain an exchange rate
	// between two specified currencies (quote and base). The request is constructed using the fmt.Sprintf function, and
	// stored in the request variable. The request is then attempted with the p.request function, which will return an error
	// (stored in the err variable) if the request fails. If an error is returned, the function will return a value of 0.
	request, err := p.request(fmt.Sprintf(ApiExchangePoloniex, quote, base))
	if err != nil {
		return 0
	}

	// The purpose of this code is to check if the string "Invalid currency pair" is found in the request string. If it is
	// found, the code within the if statement will be executed.
	if strings.Contains(string(request), "Invalid currency pair") {

		// The purpose of this code is to check if the length of the variable "p" is zero (0). If the length of the variable is
		// zero, the code returns 0.
		if p.len() {
			return 0
		}

		// This code is used to get the price of a quote currency in terms of a base currency from the Poloniex exchange. If
		// the price is greater than 0 (i.e. a valid price exists), the code returns the price rounded to 8 decimal places as a float.
		if price := p.getPoloniex(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	// The purpose of this code is to attempt to unmarshal a JSON object stored in the variable 'request' into a variable
	// 'result'. If the unmarshal fails, the code will return 0.
	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	// The purpose of this code is to check if the result list contains at least one item. If it does, the code block that
	// follows the 'if' statement will be executed. If the result list is empty, the code block will be skipped.
	if len(result) > 0 {

		// The purpose of this code is to parse a rate value from a map of strings and interfaces, convert it to a float, and
		// then return the value. It uses an if statement to check whether the rate value is present in the map, and if it is,
		// attempts to convert it to a float using strconv.ParseFloat. If successful, the function returns the value.
		if price, ok := result[0].(map[string]interface{})["rate"]; ok {
			if price, err := strconv.ParseFloat(price.(string), 64); err == nil {
				return price
			}
		}

	}

	return 0
}

// getKucoin - This function is part of the Marketplace class and is used to get the exchange rate between two currencies from the
// Kucoin Exchange. It contains logic to convert USDT to USD, and also handles cases where the base currency is not
// supported. It then returns the exchange rate, rounded to 8 decimal places, as a float.
func (p *Marketplace) getKucoin(base, quote string) float64 {

	// The purpose of the above code is to create a variable named 'result' which is a map[string]interface{}. A map is a
	// type of data structure that allows for the storage of data in key-value pairs. The interface{} type is a special type
	// in Go which is used to represent any value. This allows for the map to store any type of data.
	var (
		result map[string]interface{}
	)

	// The purpose of this code is to check if a string, "base", contains the string "USDT". If it does, the value of base
	// will be changed to "USD".
	if strings.Contains(base, "USDT") {
		base = "USD"
	}

	// This code is checking if the string "quote" contains the substring "USDT". If it does, it will set the value of
	// "quote" to "USD" instead. This is likely being used to simplify the currency quoted to just "USD".
	if strings.Contains(quote, "USDT") {
		quote = "USD"
	}

	// This code is making an API request to the Kucoin exchange to get the price of a certain base and quote. It uses the
	// fmt.Sprintf function to format the API call, then passes it to the p.request function, which will make the API
	// request. If there is an error, the code will return 0.
	request, err := p.request(fmt.Sprintf(ApiExchangeKucoin, base, quote))
	if err != nil {
		return 0
	}

	// This code is used to check for errors when unmarshalling a JSON request. The JSON request is being unmarshalled into
	// the result variable, and the if statement checks is there was an error in this process. If there was an error, the code returns 0.
	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	// This code checks the result map to see if there is a key named "msg" and if so, checks to see if it contains the
	// string "Unsupported base currency". If both conditions are true, the code executes the following code. This could be
	// used to check for an error message in a program and handle it accordingly.
	if msg, err := result["msg"]; err && strings.Contains(msg.(string), "Unsupported base currency") {

		// The purpose of this code is to check if the length of the variable p is greater than 0. If it is, the code will return 0.
		if p.len() {
			return 0
		}

		// This code is checking the price of a certain asset on the Kucoin exchange and using this to calculate the inverse
		// price. If the price on Kucoin is greater than 0, it returns the inverse price with 8 decimal places of precision.
		if price := p.getKucoin(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	// This code checks if the map "result" contains a key called "data". If it does, it assigns the value of that key to
	// the variable "data" and assigns the boolean value of true to the variable "ok". This is a common pattern in Go to
	// check if a key exists in a map.
	if data, ok := result["data"]; ok {

		// The purpose of the code snippet above is to check if the data type contains a key-value pair with the key "quote".
		// If so, the value associated with the key is stored in the variable "price". The boolean variable "ok" is set to true
		// if the data contains the key and false if it does not.
		if price, ok := data.(map[string]interface{})[quote]; ok {

			// This code is used to parse a string, which presumably contains a price, into a float. It then uses the Decimal
			// package in Go to divide 1 by the price, round it to 8 decimal places, and return the result as a float.
			if price, err := strconv.ParseFloat(price.(string), 64); err == nil {
				return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
			}

		}
	}

	return 0
}

// getBitfinex - This function is part of a larger program that is used to get the exchange rate between two different currencies,
// using the Bitfinex API. The function takes two arguments, base and quote, which are strings representing the two
// currencies. It then uses the API to make a request, and if successful, it returns the exchange rate between the two
// currencies as a float. The function also checks if either of the currencies is USDT, and if so, it changes it to USD.
// If the API request fails, the function will try to request the exchange rate in the opposite direction. Finally, if
// the API request is successful, it returns the exchange rate as a float.
func (p *Marketplace) getBitfinex(base, quote string) float64 {

	// The purpose of this code is to declare a two-dimensional array of interfaces. This type of array is useful for
	// storing data that can be of different types, such as strings, integers, and booleans. The result array is initialized
	// to an empty array, and can be filled with data later.
	var (
		result [][]interface{}
	)

	// This code is checking if the string variable "base" contains the string "USDT". If it does, it will set the variable
	// "base" to "USD". This is likely to be used to ensure that the currency used is always in the correct format.
	if strings.Contains(base, "USDT") {
		base = "USD"
	}

	// This code is checking to see if the variable "quote" contains the string "USDT". If it does, then it is setting the
	// variable "quote" to equal "USD". This is likely being used to make sure that all quotes are in the same format,
	// likely USD, regardless of how it was originally input.
	if strings.Contains(quote, "USDT") {
		quote = "USD"
	}

	// This code is making an API request to the Bitfinex exchange to get the exchange rate between two currencies (base and
	// quote). The request is made using the sprintf function to format the API URL. The request function returns a request
	// object, as well as an error if something went wrong. If an error is returned, the function returns 0.
	request, err := p.request(fmt.Sprintf(ApiExchangeBitfinex, base, quote))
	if err != nil {
		return 0
	}

	// This code is checking to see if there was an error when using the json.Unmarshal() function. If there was an error,
	// the code will return 0. Otherwise, it will continue to execute the rest of the code.
	if err = json.Unmarshal(request, &result); err != nil {
		return 0
	}

	// This statement is checking if the length of the result list is 0, or if the length of the first element in the list
	// is not equal to 4. If either is true, then some action will be taken.
	if len(result) == 0 || len(result[0]) != 4 {

		// This code is checking to see if the length of the variable 'p' is greater than zero. If it is, the code will return
		// a value of 0. This is likely done to ensure that the value of 'p' is valid before proceeding with the rest of the code.
		if p.len() {
			return 0
		}

		// This code is checking the price of a given currency pair from an exchange (Bitfinex) and then returning the inverse
		// of that price. The code uses the decimal.New() function from the decimal library to prevent rounding errors and to
		// ensure a precise result. The code then rounds the result to 8 decimal places, and returns the result as a float.
		if price := p.getBitfinex(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	// This code is checking if the element located at result[0][3] is a float64 value. If it is, it returns the price. If
	// not, the code will continue to execute without returning the price.
	if price, ok := result[0][3].(float64); ok {
		return price
	}

	return 0
}

// getBinance - This is a function within the Marketplace type which uses the Binance API to get the exchange rate of a base currency
// to a quote currency. It performs a request to the Binance API and uses the response to return the exchange rate as a
// float. If the request is unsuccessful, it will check the exchange rate in the opposite direction (quote to base) and
// return the inverse.
func (p *Marketplace) getBinance(base, quote string) float64 {

	// The purpose of the above code is to declare a variable named "result" that is a map of strings to strings. This could
	// be used to store data that has a key-value structure, such as a dictionary.
	var (
		result map[string]string
	)

	// This code is part of a program that makes an API request to the Binance Exchange. The purpose of this code is to
	// construct the request with the given parameters (base and quote) and send the request to the Binance Exchange. The
	// request variable will store the response from the API request, and the err variable will contain any errors that
	// occur during the request.
	request, err := p.request(fmt.Sprintf(ApiExchangeBinance, base, quote))
	if err != nil {

		// The purpose of this code is to check if the length of p is greater than 0. If it is, the code will return 0.
		if p.len() {
			return 0
		}

		// The purpose of this code is to get the price of a quote currency and a base currency from the Binance exchange, and
		// then return the inverse of that price rounded to 8 decimal places.
		if price := p.getBinance(quote, base); price > 0 {
			return decimal.New(decimal.New(1).Div(price).Float()).Round(8).Float()
		}

		return 0
	}

	// This code is using the json.Unmarshal() function to convert the contents of a JSON request into an object, and assign
	// it to the 'result' variable. If there is an error in the process, the function will return 0.
	if err := json.Unmarshal(request, &result); err != nil {
		return 0
	}

	// This code is attempting to parse a float value from a map of values. The first if statement checks to see if the map
	// contains a value for the key "price". If so, it attempts to parse the value as a float and return it. If it is
	// unsuccessful, it returns nothing.
	if price, ok := result["price"]; ok {
		if price, err := strconv.ParseFloat(price, 64); err == nil {
			return price
		}
	}

	return 0
}

// request - This function is part of a Marketplace struct and is used to make a request (using HTTP) to a specified URL. It
// creates a new request, checks for any errors, reads the response body and returns the body as well as any errors.
func (p *Marketplace) request(url string) ([]byte, error) {

	// This code is creating an HTTP request with the http.NewRequest() function. The first argument is the HTTP method (GET
	// in this case), the second is the URL of the request, and the third is a body for the request (nil in this case, as it
	// is a GET request). If there is an error creating the request, it is returned and the function exits.
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// This code is used to make an HTTP request using the 'client' object of type http.Client. The request is passed as an
	// argument to the Do() function of the client, which returns a response and an error. If an error occurs, the function
	// returns nil and the error, otherwise the response is returned. The response body is closed after the function returns.
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// This if statement is checking the status code of a response object (resp). If the status code is not equal to 200
	// (which typically indicates a successful response), the function will return nil and an error with a message
	// containing a description of what the status code was. The purpose of this is to provide feedback to the user about why the response failed.
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Status code: %v", resp.StatusCode))
	}

	// The code is trying to read the response body from a response object (resp) and store it in the variable body. If
	// there is an error when trying to read the body, it will return nil and the error.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
