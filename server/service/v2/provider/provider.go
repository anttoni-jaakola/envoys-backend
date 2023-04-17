package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/status"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Service struct {
	Context *assets.Context
}

// Initialization - The code initializes a Service object and runs six concurrent functions: chain(), price(), market().
func (a *Service) Initialization() {
	go a.chain()
	go a.price()
	go a.market()
}

// queryRatio - This function is used to calculate the ratio of a given base and quote. It takes in two strings, base and quote, as
// parameters and returns a float64 representing the ratio and a boolean to indicate whether the ratio was successfully
// calculated. It uses the GetCandles function to retrieve the last 2 candles and then calculates the ratio by taking the
// difference between the first and second close prices and dividing it by the second close price.
func (a *Service) queryRatio(base, quote string) (ratio float64, ok bool) {

	// This code is part of a function that is attempting to get the ratio of two different currencies. The code is
	// attempting to get two candles from the e (which is an exchange) with the given base and quote units. If an error is
	// encountered, the function will return the ratio and ok.
	migrate, err := a.GetTicker(context.Background(), &pbprovider.GetRequestTicker{BaseUnit: base, QuoteUnit: quote, Limit: 2})
	if err != nil {
		return ratio, ok
	}

	// This code checks to see if the "migrate" has two "Fields", and if it does, it calculates the ratio of the closing
	// prices between the two Fields. The ratio is calculated by subtracting the closing price of the first Field from the
	// closing price of the second Field, then dividing by the closing price of the second Field and multiplying by 100.
	if len(migrate.Fields) == 2 {
		ratio = ((migrate.Fields[0].Close - migrate.Fields[1].Close) / migrate.Fields[1].Close) * 100
	}

	return ratio, true
}

// queryPrice - This function is used to query a database for the price of a currency pair given the base and quote units. It takes
// two parameters, base and quote, which are strings and returns a float value and a boolean. The function uses the
// QueryRow() method to execute the query, and the Scan() method to store the returned value in the price variable. If an
// error occurs, the ok boolean is returned as false, otherwise it is returned as true.
func (a *Service) queryPrice(base, quote string) (price float64, ok bool) {

	// This code is used to query and retrieve a price from a database. The "if err" statement is used to check for any
	// errors that may occur during the query and retrieve process. If an error is encountered, the code will return the price and ok.
	if err := a.Context.Db.QueryRow("select price from pairs where base_unit = $1 and quote_unit = $2", base, quote).Scan(&price); err != nil {
		return price, ok
	}

	return price, true
}

// querySum - The purpose of this code is to calculate the final value of a given value after subtracting fees. It queries the
// database for the corresponding currency's fees_trade and fees_discount columns, and checks the status of an order
// based on an id. If the order is a maker order, the discount is subtracted from the fees. Finally, the actual value
// after subtracting fees and the rounded value after subtracting fees are returned.
func (a *Service) querySum(id int64, symbol string, value float64) (b, f float64, m bool, err error) {

	// The purpose of this code is to declare three variables of different types: d is a float64, m is a boolean, and s is a
	// types.Status. This can be used to assign values to these variables and use them in your program.
	var (
		d float64
		s string
	)

	// This code is used to query a database for a particular record associated with the given symbol. It then scans the
	// result and stores the values of the fees_trade and fees_discount columns in the variables fees and discount
	// respectively. If an error occurs during the query, it returns the balance and fees variables.
	if err := a.Context.Db.QueryRow("select fees_trade, fees_discount from assets where symbol = $1", symbol).Scan(&f, &d); err != nil {
		return b, f, m, err
	}

	// The purpose of this code is to query a database for the status of an order based on the id and store the result in a
	// variable. If there is an error with the query, an error is returned.
	if err := a.Context.Db.QueryRow("select status from orders where id = $1;", id).Scan(&s); err != nil {
		return b, f, m, err
	}

	// This if statement assigns the boolean value true to the variable m if the variable s is equal to the constant
	// types.Status_PENDING. This can be used to evaluate a condition or determine if a specific value is present in a given set.
	if s == types.StatusPending {
		m = true
	}

	// This code is checking if the variable "maker" is true, and if it is, it is subtracting the value of "discount" from
	// "fees" and storing the result in "fees" as a float. This is likely being done to calculate a discounted fee for a maker order.
	if m {
		f = decimal.New(f).Sub(d).Float()
	}

	// This code is used to calculate the final value of a given value after subtracting fees. The two return values
	// represent the actual value after subtracting fees and the rounded value after subtracting fees.
	return decimal.New(value).Sub(decimal.New(decimal.New(value).Mul(f).Float()).Div(100).Float()).Float(), decimal.New(value).Sub(decimal.New(value).Sub(decimal.New(decimal.New(value).Mul(f).Float()).Div(100).Float()).Float()).Float(), m, nil
}

// queryMarket - This function is used to get the market price for a given base and quote currency. It takes in the base, quote,
// assigning (buy/sell), and current price as parameters. It then gets the current price from the getPrice() function
// and, depending on the assigning, queries the database for either the minimum or maximum price that is greater than or
// less than the current price and is in the pending status. Finally, it returns the market price.
func (a *Service) queryMarket(base, quote, _type string, assigning string, price float64) float64 {

	var (
		ok bool
	)

	// This code is checking for the existence of a price by attempting to get it from e.queryPrice(), which takes in two
	// parameters, base and quote. If the price exists (indicated by the ok return value), then it will be returned. If the
	// price does not exist (indicated by the !ok return value), then it will not be returned.
	if price, ok = a.queryPrice(base, quote); !ok {
		return price
	}

	// The switch statement is used to evaluate an expression and determine which statement should be executed based on the
	// value of the expression. The switch statement assigns the expression to a variable called assigning, which is then
	// used to make the determination of which statement to execute.
	switch assigning {
	case types.AssigningBuy:

		// The purpose of this code is to query the database for the minimum price of a particular order that has a specific
		// assigning, base unit, quote unit, price, and status. The result is then stored in the variable 'price'.
		_ = a.Context.Db.QueryRow("select min(price) as price from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and price >= $4 and status = $5 and type = $6", types.AssigningSell, base, quote, price, types.StatusPending, _type).Scan(&price)

	case types.AssigningSell:

		// The purpose of this code is to query a database for the maximum price from orders that meet certain criteria
		// (assigning, base unit, quote unit, price and status) and scan the result into the variable "price".
		_ = a.Context.Db.QueryRow("select max(price) as price from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and price <= $4 and status = $5 and type = $6", types.AssigningBuy, base, quote, price, types.StatusPending, _type).Scan(&price)
	}

	return price
}

// queryRange - This function is used to retrieve the minimum and maximum trade value of a given currency symbol from a database and
// to check if a given value is within the range. If the given value is within the range, it will return the min and max
// trade values, as well as a boolean value indicating whether the given value is within the range.
func (a *Service) queryRange(symbol string, value float64) (min, max float64, ok bool) {

	// This if statement is used to query a database for a row containing the min_trade and max_trade columns for the
	// currency with the symbol given as an argument. If the query is successful, the values for min_trade and max_trade are
	// stored in the variables min and max. If the query fails, an error is returned and the function returns min, max, and ok.
	if err := a.Context.Db.QueryRow("select min_trade, max_trade from assets where symbol = $1", symbol).Scan(&min, &max); err != nil {
		return min, max, ok
	}

	// This statement is checking to see if a given value is within a minimum and maximum range. If the value is between the
	// min and max values, then the function returns the min and max values, along with a boolean value of true.
	if value >= min && value <= max {
		return min, max, true
	}

	return min, max, ok
}

// queryQuantity - This function is used to calculate the quantity of a financial asset based on its price and whether it is a
// cross-trade or not. The function takes in the assigning (buy or sell), the quantity, the price, and a boolean value to
// check if it is a cross-trade. If it is a cross-trade, the function will divide the quantity by the price. Otherwise,
// it will multiply the quantity by the price. The function then returns the calculated quantity.
func (a *Service) queryQuantity(assigning string, quantity, price float64, cross bool) float64 {

	if cross {

		// The purpose of this code is to calculate the quantity of an item by dividing it by its price. This switch statement
		// checks the assigning value to make sure it is set to "BUY", and then uses the decimal.New() method to divide the
		// quantity by the price and convert it to a float.
		switch assigning {
		case types.AssigningBuy:
			quantity = decimal.New(quantity).Div(price).Float()
		}

		return quantity

	} else {

		// This switch statement is used to determine the quantity of a purchase. In this case, if the assigning variable is
		// set to types.Assigning_BUY, then the quantity will be multiplied by the price to determine the total cost of the
		// purchase.
		switch assigning {
		case types.AssigningSell:
			quantity = decimal.New(quantity).Mul(price).Float()
		}

		return quantity
	}
}

// queryValidateOrder - This function is a helper function for placing orders in a Spot trading service. It performs checks on the order
// before it is submitted, such as checking the price is not 0, checking the user has enough funds to cover the order,
// and ensuring the quantity of the order is within the predetermined range. If any of these checks fail, an error is
// returned. Otherwise, the order is accepted and the quantity is returned.
func (a *Service) queryValidateOrder(order *types.Order) (summary float64, err error) {

	// This code checks if the order's price is 0, and if it is, it returns an error message (65790) with the impossible
	// price that is being requested. This helps to identify errors in the order's price and allows for more accurate
	// debugging of the program.
	if order.GetPrice() == 0 {
		return 0, status.Errorf(65790, "impossible price %v", order.GetPrice())
	}

	// This switch statement is used to check the value of the GetAssigning() method on the order object. Depending on the
	// value of GetAssigning(), different code blocks may be executed.
	switch order.GetAssigning() {
	case types.AssigningBuy:

		// The purpose of this code is to calculate the total cost of an order, given the quantity and price of a product. The
		// code uses the decimal library to get the quantity and price of the order, and then multiplies them together to
		// calculate the total cost.
		quantity := decimal.New(order.GetQuantity()).Mul(order.GetPrice()).Float()

		// This code is checking the range of a given quantity, and returning an error if the quantity is not within the
		// specified range. The min and max variables represent the minimum and maximum values of the quantity, while the ok
		// variable indicates whether the range is valid. If the range is invalid, the code will return an error with a message
		// containing the minimum and maximum values.
		if min, max, ok := a.queryRange(order.GetQuoteUnit(), quantity); !ok {
			return 0, status.Errorf(11623, "[quote]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		// This line of code is used to get the balance of the user in a given quote unit. It is used to determine the amount
		// of funds available to the user for a particular order. The getBalance() method takes in two parameters, the quote
		// unit and the user id, and returns the balance for the user in the specified quote unit.
		balance := a.QueryBalance(order.GetQuoteUnit(), order.GetType(), order.GetUserId())

		// This statement is an if-statement that is used to check if the quantity is greater than the balance or if the
		// order's quantity is equal to 0. If either of these conditions are true, then the statement will return a value of 0,
		// along with an error message. The purpose of this statement is to ensure that a user does not place an order with
		// insufficient funds and to inform them if they have attempted to do so.
		if quantity > balance || order.GetQuantity() == 0 {
			return 0, status.Error(11586, "[quote]: there is not enough funds on your asset balance to place an order")
		}

		return quantity, nil

	case types.AssigningSell:

		// This statement retrieves the quantity of an order from the order object and assigns it to the variable "quantity".
		quantity := order.GetQuantity()

		// This code is checking to see if the order's base unit and quantity meet a certain minimum and maximum trading
		// amount. If the order does not meet the requirements, an error is returned with a message that states the minimum and
		// maximum trading amounts.
		if min, max, ok := a.queryRange(order.GetBaseUnit(), order.GetQuantity()); !ok {
			return 0, status.Errorf(11587, "[base]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		// The purpose of this code is to get the balance of the user from the order object. The order object contains the base
		// unit and the userId of the user, which are used to call the getBalance() method of the e object to get the balance.
		balance := a.QueryBalance(order.GetBaseUnit(), order.GetType(), order.GetUserId())

		// This code is testing whether the quantity of an order is greater than the balance of the asset used to place the
		// order. If the quantity is greater than the balance or the order quantity is 0, it will return an error message
		// indicating that there is not enough funds on the asset balance to place the order.
		if quantity > balance || order.GetQuantity() == 0 {
			return 0, status.Error(11624, "[base]: there is not enough funds on your asset balance to place an order")
		}

		return quantity, nil
	}

	return 0, status.Error(11596, "invalid input parameter")
}

// queryValidatePair - checks if a pair with given base and quote unit, and type exists in the DB. If not, an error is returned.
func (a *Service) queryValidatePair(base, quote, _type string) error {

	// exist is a boolean variable used to store a true or false value.
	var (
		exist bool
	)

	// Convert TypeCross to TypeSpot to ensure the correct type is used.
	if _type == types.TypeCross {
		_type = types.TypeSpot
	}

	// Check if the (base, quote, type) pair exists in the "pairs" table. If not, return an error.
	if err := a.Context.Db.QueryRow("select exists(select id from pairs where base_unit = $1 and quote_unit = $2 and type = $3)::bool", base, quote, _type).Scan(&exist); err != nil || !exist {
		return status.Errorf(11585, "this pair %v-%v does not exist", base, quote)
	}

	return nil
}

// queryOrder - This function is used to retrieve an order from a database by its ID. It takes an int64 (id) as a parameter and
// returns a pointer to a "types.Order" type. It uses the "QueryRow" method of the database to scan the selected row
// into the "order" variable and then returns the pointer to the order.
func (a *Service) queryOrder(id int64) *types.Order {

	var (
		order types.Order
	)

	// This code is used to query a database for a single row of data matching the specified criteria (in this case, the "id
	// = $1" condition) and then assign the returned values to the specified variables (in this case, the fields of the
	// "order" struct). This allows the program to retrieve data from the database and store it in a convenient and organized format.
	_ = a.Context.Db.QueryRow("select id, value, quantity, price, assigning, user_id, base_unit, quote_unit, status, create_at from orders where id = $1", id).Scan(&order.Id, &order.Value, &order.Quantity, &order.Price, &order.Assigning, &order.UserId, &order.BaseUnit, &order.QuoteUnit, &order.Status, &order.CreateAt)
	return &order
}

// writeAsset - This function is used to set a new asset for a given user. It takes in three parameters - a string symbol to identify
// the asset, an int64 userId to identify the user, and a boolean error indicating whether an error should be returned if
// the asset already exists. The function checks if the asset already exists in the database, and if it does not exist,
// it inserts it into the database. If the error boolean is true, it will return an error if the asset already exists. If
// the error boolean is false, it will return no error regardless of the asset's existence.
func (a *Service) writeAsset(symbol, _type string, userId int64, error bool) error {

	// The purpose of this code is to query the database for a specific asset with a given symbol and userId. The query is
	// then stored in a row variable and an error is checked for. If there is an error, it will be returned. Finally, the
	// row is closed when the code is finished.
	row, err := a.Context.Db.Query(`select id from balances where symbol = $1 and user_id = $2 and type = $3`, symbol, userId, _type)
	if err != nil {
		return err
	}
	defer row.Close()

	// The code is used to check if the row is valid. The '!' operator is used to check if the row is not valid. If the row
	// is not valid, the code will execute.
	if !row.Next() {

		// This code is inserting values into a database table called "assets" with the specific columns "user_id" and
		// "symbol". The purpose of this code is to save the values of userId and symbol into the table for future reference.
		if _, err = a.Context.Db.Exec("insert into balances (user_id, symbol, type) values ($1, $2, $3)", userId, symbol, _type); err != nil {
			return err
		}

		return nil
	}

	// The purpose of this code is to return an error status to indicate that a fiat asset has already been generated. The
	// code uses the status.Error() function to return an error with a specific status code (700991) and an error message
	// ("the fiat asset has already been generated").
	if error {
		return status.Error(700991, "the fiat asset has already been generated")
	}

	return nil
}

// writeOrder - This function is used to set an order in the database. It takes in a pointer to a types.Order which contains the
// order's details, and inserts the data into the 'orders' table. It then returns the id of the newly created order and any potential errors.
func (a *Service) writeOrder(order *types.Order) (id int64, err error) {

	if err := a.Context.Db.QueryRow("insert into orders (assigning, base_unit, quote_unit, price, value, quantity, user_id, type, trading) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id", order.GetAssigning(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetPrice(), order.GetQuantity(), order.GetValue(), order.GetUserId(), order.GetType(), order.GetTrading()).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

// writeTrade - The purpose of this code is to set a trade by converting a given value to a decimal number multiplied by a given
// price, get the sum of a given order, symbol, and value, insert the data into a database, update the "fees_charges"
// column in the "currencies" table in a database, and publish a particular order to an exchange.
func (a *Service) writeTrade(id int64, symbol string, value, price float64, convert bool) (float64, error) {

	// The purpose of this code is to retrieve an order from a database, given its ID. The variable 'order' will store the
	// order object that is returned from the queryOrder() method.
	order := a.queryOrder(id)
	order.Value = value

	// This code is used to convert a given value to a decimal number multiplied by a given price. The result is then stored
	// as a floating point number. This is likely used for some kind of financial calculation or to convert a given value to
	// a currency amount.
	if convert {
		value = decimal.New(value).Mul(price).Float()
	}

	// This code is attempting to get the sum of a given order, symbol and value. The variables s and f are used to store
	// the sum and any error encountered, respectively. The if statement checks for any errors that may have occurred and
	// returns 0 and the error if one is encountered.
	s, f, maker, err := a.querySum(id, symbol, value)
	if err != nil {
		return 0, err
	}

	// This code is used to calculate the fee for an order based on the assigned type. If the order is assigned to be a
	// SELL, the fee is calculated by dividing the fee (f) by the price. If the order is assigned to be something else, the
	// fee is simply set to be f.
	if order.GetAssigning() == types.AssigningSell {
		order.Fees = decimal.New(f).Div(price).Float()
	} else {
		order.Fees = f
	}

	// This code is used to insert data into the "transfers" table in a database using the parameters provided in the array
	// "param". The code first checks for any errors in the insertion process, and if there are any, it will return an error.
	if _, err := a.Context.Db.Exec(`insert into trades (order_id, assigning, user_id, base_unit, quote_unit, quantity, fees, price, maker) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, order.GetId(), order.GetAssigning(), order.GetUserId(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetValue(), order.GetFees(), price, maker); err != nil {
		return 0, err
	}

	// This statement is checking to see if the value of the parameter at index i in the param array is greater than 0. If
	// it is, then the code within the if statement will be executed. This is likely being used to check if a fee is
	// associated with the parameter at index i.
	if f > 0 {

		// This code is updating the "fees_charges" column in the "currencies" table in a database. The "symbol" and
		// "fee" are parameters that are passed into the statement. If an error occurs during the
		// execution of the statement, the function will return the error.
		if _, err := a.Context.Db.Exec("update assets set fees_charges = fees_charges + $2 where symbol = $1;", symbol, f); err != nil {
			return 0, err
		}
	}

	// The purpose of the code snippet is to publish a particular order to an exchange with the routing key "order/status".
	// The if statement checks for any errors encountered while publishing the order, and returns an error if one occurs.
	if err := a.Context.Publish(a.queryOrder(order.GetId()), "exchange", "order/status"); err != nil {
		return 0, err
	}

	return s, nil
}

// QueryPair - This function is used to get a specific pair from the database, based on the id and status passed as arguments. The
// function returns a pointer to a 'types.Pair' struct and an error if any. It prepares a query to select the specified
// pair from the database, based on the given id and status. It then scans the results and stores them in the struct, and
// finally returns the struct and an error if any.
func (a *Service) QueryPair(id int64, _type string, status bool) (*types.Pair, error) {

	var (
		chain types.Pair
		maps  []string
	)

	// The purpose of this code is to append a string to a list of maps if a certain condition is true. In this case, if the
	// variable "status" is true, the string "and status = %v" with "true" as the placeholder value is added to the list of maps.
	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	// Check if type is not empty then append a query to maps with the value of.
	if len(_type) > 0 {
		maps = append(maps, fmt.Sprintf("and type = '%v'", _type))
	}

	// This code is used to query a database and retrieve information about a pair with a specified id. The query is formed
	// using the fmt.Sprintf() function, and it is a combination of a string and the id parameter. The retrieved information
	// is then assigned to the chain struct. Finally, the code returns the chain struct and an error if it fails.
	if err := a.Context.Db.QueryRow(fmt.Sprintf("select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs where id = %[1]d %[2]s", id, strings.Join(maps, " "))).Scan(
		&chain.Id,
		&chain.BaseUnit,
		&chain.QuoteUnit,
		&chain.Price,
		&chain.BaseDecimal,
		&chain.QuoteDecimal,
		&chain.Status,
	); err != nil {
		return &chain, err
	}

	return &chain, nil
}

// QueryAddress - This function is used to get the address associated with a userId, symbol, platform and protocol. It does this by
// querying the assets and wallets tables in the database for a matching userId, symbol and platform, and
// returns the address associated with the query if one is found.
func (a *Service) QueryAddress(userId int64, platform string) (address string) {

	// This statement is used to query a database to get an address associated with a user, platform and symbol.
	// The purpose of using `coalesce` is to return a blank string if the address is null. The purpose of using `QueryRow`
	// is to limit the query to a single row. The purpose of using `Scan` is to store the result of the query into the `address` variable.
	_ = a.Context.Db.QueryRow("select coalesce(w.address, '') from balances a inner join wallets w on w.platform = $1 and w.user_id = a.user_id where a.user_id = $2 and a.type = $3", platform, userId, types.TypeSpot).Scan(&address)
	return address
}

// QueryBalance - This function is used to query the balance of a user's assets by symbol. It takes a symbol and userID as parameters
// and queries the assets table in the database for the balance associated with that symbol and userID, then returns the balance.
func (a *Service) QueryBalance(symbol, _type string, userId int64) (balance float64) {

	// This line of code is used to retrieve the balance from the assets table in a database. It takes in two parameters
	// (symbol and userId) and uses them to query the database. The result is then stored in the variable balance.
	_ = a.Context.Db.QueryRow("select value as balance from balances where symbol = $1 and user_id = $2 and type = $3", symbol, userId, _type).Scan(&balance)
	return balance
}

// QueryAsset - This function is used to retrieve asset information from a database. It takes a currency symbol and a status
// boolean as arguments. It then queries the database to retrieve information about the currency and stores it in the
// 'response' variable. It then checks for the existence of the asset icon and stores the result in the 'icon' field
// of the 'response' variable. Finally, it returns the currency and an error value, if any.
func (a *Service) QueryAsset(symbol string, status bool) (*types.Asset, error) {

	var (
		response types.Asset
		maps     []string
		storage  []string
		chains   []byte
	)

	// The purpose of this code is to append an item to a list of maps if a certain condition is met. In this case, if the
	// "status" variable is true, a string will be appended to the list of maps.
	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	// This code is performing a query of a database table called "currencies" and scanning the results into a response
	// object. The query is using the symbol parameter to filter the results and strings.Join(maps, " ") to join any
	// additional parameters. If the query fails, an error is returned.
	if err := a.Context.Db.QueryRow(fmt.Sprintf(`select id, name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, "group", type, create_at, chains from assets where symbol = '%v' %s`, symbol, strings.Join(maps, " "))).Scan(
		&response.Id,
		&response.Name,
		&response.Symbol,
		&response.MinWithdraw,
		&response.MaxWithdraw,
		&response.MinTrade,
		&response.MaxTrade,
		&response.FeesTrade,
		&response.FeesDiscount,
		&response.FeesCharges,
		&response.FeesCosts,
		&response.Marker,
		&response.Status,
		&response.Group,
		&response.Type,
		&response.CreateAt,
		&chains,
	); err != nil {
		return &response, err
	}

	// The purpose of the code is to add a string to a storage slice. The string is made up of elements from the
	// e.Context.StoragePath, the word "static", the word "icon", and a string created from the response.GetSymbol()
	// function. The ... in the code indicates that the elements of the slice are being "unpacked" into to append() call.
	storage = append(storage, []string{a.Context.StoragePath, "static", "icon", fmt.Sprintf("%v.png", response.GetSymbol())}...)

	// This statement is checking to see if a file at the given filepath exists. If it does, then the response.Icon will be
	// set to true. This statement is used in order to prevent the creation of unnecessary files.
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		response.Icon = true
	}

	// The purpose of this code is to unmarshal a json object into a response object. This is done using the
	// json.Unmarshal() function. The function takes the json object (chains) and a reference to the response.Fields
	// object. If there is an error, it will be returned with the error context.
	if err := json.Unmarshal(chains, &response.Fields); err != nil {
		return &response, err
	}

	return &response, nil
}

// QueryChain - This function is used to query a row from a database table "chains" with the given id and status. It then scans the
// row into a types.Chain struct. If there is an error, it returns the struct with an error. Otherwise, it returns the
// struct with no error.
func (a *Service) QueryChain(id int64, status bool) (*types.Chain, error) {

	var (
		chain types.Chain
		maps  []string
	)

	// The purpose of this code is to add the string "and status = true" to the maps slice, if the status variable is set to true.
	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	// This code is used to query a database for a row of data which matches the given id. The query is built by joining the
	// strings in the maps array and is passed to the QueryRow method. The data is then scanned into the chain object and
	// returned. If there is an error, it will be returned instead.
	if err := a.Context.Db.QueryRow(fmt.Sprintf("select id, name, rpc, block, network, explorer_link, platform, confirmation, time_withdraw, fees, tag, parent_symbol, decimals, status from chains where id = %[1]d %[2]s", id, strings.Join(maps, " "))).Scan(
		&chain.Id,
		&chain.Name,
		&chain.Rpc,
		&chain.Block,
		&chain.Network,
		&chain.ExplorerLink,
		&chain.Platform,
		&chain.Confirmation,
		&chain.TimeWithdraw,
		&chain.Fees,
		&chain.Tag,
		&chain.ParentSymbol,
		&chain.Decimals,
		&chain.Status,
	); err != nil {
		return &chain, errors.New("chain not found or chain network off")
	}

	return &chain, nil
}

// QueryContractById - This function is used to retrieve a contract from a database by its ID. It queries the database for the contract with
// the specified ID, and then scans the query result into the fields of the pbspot.Contract struct. The function then
// returns the contract and any errors that may have occurred.
func (a *Service) QueryContractById(id int64) (*types.Contract, error) {

	var (
		contract types.Contract
	)

	// This code is a SQL query used to retrieve data from a database. The purpose of the query is to select data from the
	// contracts and chains tables for a specific contract with a given ID. The query will return information such as the
	// symbol, chain ID, address, fees withdraw, protocol, decimals, and platform of the contract. The query uses the Scan()
	// method to store the retrieved data in the contract variable. The if statement is used to check for errors and return
	// the contract along with an error if one occurs.
	if err := a.Context.Db.QueryRow(`select c.id, c.symbol, c.chain_id, c.address, c.fees, c.protocol, c.decimals, n.platform from contracts c inner join chains n on n.id = c.chain_id where c.id = $1`, id).Scan(&contract.Id, &contract.Symbol, &contract.ChainId, &contract.Address, &contract.Fees, &contract.Protocol, &contract.Decimals, &contract.Platform); err != nil {
		return &contract, err
	}

	return &contract, nil
}

// QueryContract - This function retrieves a contract from the database based on a given symbol and chain ID. It then sets the fields of
// a types.Contract struct with the values from the database and returns it. Any errors encountered while retrieving the
// contract are returned as the second value.
func (a *Service) QueryContract(symbol string, cid int64) (*types.Contract, error) {

	var (
		contract types.Contract
	)

	// This code is checking the database for a contract with the specified symbol and chain ID and then storing the results
	// of the query in a contract struct. If the query fails, to err is returned.
	if err := a.Context.Db.QueryRow(`select id, address, fees, protocol, decimals from contracts where symbol = $1 and chain_id = $2`, symbol, cid).Scan(&contract.Id, &contract.Address, &contract.Fees, &contract.Protocol, &contract.Decimals); err != nil {
		return &contract, err
	}

	return &contract, nil
}

// QueryReserve - This function is used to get the total reserve for a given symbol, platform, and protocol from a database. It takes
// three parameters (symbol, platform, and protocol) and uses a SQL query to get the sum of the values from the reserves
// table where the symbol, platform, and protocol match the provided parameters. Finally, it returns the total reserve as a float64.
func (a *Service) QueryReserve(symbol, platform, protocol string) (reserve float64) {

	if len(protocol) == 0 {
		protocol = types.ProtocolMainnet
	}

	// The purpose of this code is to query a database for the sum of values from a specific set of reserves (symbol,
	// platform, and protocol) and store the result in the reserve variable.
	_ = a.Context.Db.QueryRow(`select sum(value) from reserves where symbol = $1 and platform = $2 and protocol = $3`, symbol, platform, protocol).Scan(&reserve)
	return reserve
}

// QueryReverse - This code is used to get the reverse of a certain user's address for a certain platform and symbol from a database. It
// takes in the userId, address, symbol, and platform as parameters, and returns the reverse value stored in the database.
func (a *Service) QueryReverse(userId int64, address, symbol, platform string) (reverse float64) {

	// The purpose of this code is to query a database for the sum of values from a specific set of reserves (symbol,
	// platform, and protocol) and store the result in the reserve variable.
	_ = a.Context.Db.QueryRow(`select reverse from reserves where user_id = $1 and address = $2 and symbol = $3 and platform = $4 and protocol = $5`, userId, address, symbol, platform, types.ProtocolMainnet).Scan(&reverse)
	return reverse
}

// WriteBalance - This function is used to update the balance of a user in a database. Depending on the cross parameter, either the
// balance is increased (types.Balance_PLUS) or decreased (types.Balance_MINUS) by a given quantity. The balance is
// updated in the assets table of the database, using a query. Finally, an error is returned if an error occurred during the update.
func (a *Service) WriteBalance(symbol, _type string, userId int64, quantity float64, cross string) error {

	switch cross {
	case types.BalancePlus:

		// The code above is an if statement that is used to update the balance of an asset with a given symbol and user_id in
		// a database. The statement executes an update query, passing in the values of symbol, quantity, and userId as
		// parameters to the query. If the query fails to execute, the if statement will return an error.
		if _, err := a.Context.Db.Exec("update balances set value = value + $2 where symbol = $1 and user_id = $3 and type = $4;", symbol, quantity, userId, _type); err != nil {
			return err
		}
		break
	case types.BalanceMinus:

		// This code is used to update the balance of a user's assets in a database. The code updates the user's balance by
		// subtracting the quantity given. The values being used to update the balance are stored in variables, and are passed
		// into the code as parameters ($1, $2, and $3). The code also checks for errors and returns an error if one is found.
		if _, err := a.Context.Db.Exec("update balances set value = value - $2 where symbol = $1 and user_id = $3 and type = $4;", symbol, quantity, userId, _type); err != nil {
			return err
		}
		break
	}

	return nil
}

// WriteTransaction - The purpose of this code is to set the transaction of a service. It checks if a transaction exists, then generates a
// unique identifier if it does not. It then inserts the transaction information into a database table, and sets the
// chain associated with the transaction. It then sets the chain.RPC and chainId values to empty strings and zero
// respectively. It then checks if the transaction has a parent, and if it does, it updates the allocation and status of
// the transaction and its parent in the database. Finally, it returns the transaction if the operation was successful, or nil if not.
func (a *Service) WriteTransaction(transaction *types.Transaction) (*types.Transaction, error) {

	// The purpose of the code snippet is to declare two variables, exist and err. exist is of type bool and err is of type
	// error. This is typically used when programming to indicate the existence of an error or to store the value of an error.
	var (
		exist bool
		err   error
	)

	// The purpose of this code is to check if the transaction exists in e.getTransactionExist(transaction.GetHash()). If
	// the transaction does not exist (i.e. !exist is true) then the code will execute the statements that follow.
	if _ = a.Context.Db.QueryRow("select exists(select id from transactions where hash = $1)::bool", transaction.GetHash()).Scan(&exist); !exist {

		// This code is used to generate a unique identifier (in this case a UUID) for a transaction if it doesn't already have
		// one. This UUID can be used to identify the transaction uniquely and ensure that it is not a duplicate of another transaction.
		if len(transaction.GetHash()) == 0 {
			transaction.Hash = uuid.NewV1().String()
		}

		// This code is a SQL query to insert transaction information into a database table called "transactions". It is
		// assigning values to each of the 13 columns in the table, and then returning the id, CreateAt, and Status columns in
		// the same row. It is then using the Scan() function to assign the returned values to the transaction object.
		if err := a.Context.Db.QueryRow(`insert into transactions (symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, assignment, "group", platform, protocol, allocation, parent) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) returning id, create_at, status;`,
			transaction.GetSymbol(),
			transaction.GetHash(),
			transaction.GetValue(),
			transaction.GetFees(),
			transaction.GetConfirmation(),
			transaction.GetTo(),
			transaction.GetBlock(),
			transaction.GetChainId(),
			transaction.GetUserId(),
			transaction.GetAssignment(),
			transaction.GetGroup(),
			transaction.GetPlatform(),
			transaction.GetProtocol(),
			transaction.GetAllocation(),
			transaction.GetParent(),
		).Scan(&transaction.Id, &transaction.CreateAt, &transaction.Status); err != nil {
			return transaction, err
		}

		// This code is getting the chain associated with the transaction. The first line is getting the chain, and the second
		// line checks if there has been an error. If there is an error, the code returns nil and the error.
		transaction.Chain, err = a.QueryChain(transaction.GetChainId(), false)
		if err != nil {
			return transaction, err
		}

		// This code sets the Chain.Rpc and ChainId values of the transaction variable to empty strings and zero respectively.
		// This is likely used to reset the transaction variable to its default values.
		transaction.Chain.Rpc, transaction.ChainId = "", 0

		return transaction, nil
	} else {

		// This code is used to update the status of a transaction in a database. The first `if` statement checks if the
		// transaction has a parent. If it does, the code will execute two `Exec` commands in order to update the allocation
		// and status of the transaction and its parent in the database.
		if _ = a.Context.Db.QueryRow("select exists(select id from transactions where hash = $1 and allocation = $2)::bool", transaction.GetHash(), types.AllocationInternal).Scan(&exist); exist {

			// This code is used to update a transaction in a database. It sets the assignment to DEPOSIT and the status to
			// PENDING by using the hash of the transaction as an identifier. The if statement is used to check for any errors
			// that may occur while executing the query. If an error occurs, the transaction is returned without any changes.
			if _, err := a.Context.Db.Exec("update transactions set assignment = $1, status = $2 where hash = $3;", types.AssignmentDeposit, types.StatusPending, transaction.GetHash()); err != nil {
				return transaction, err
			}

		}
	}

	return transaction, nil
}

// WriteReserve - This function is used to set a reserve for a user in a database. It takes the userId, address, symbol, value,
// platform, protocol, and cross as parameters. It first checks if the reserve already exists in the database. If it
// does, it updates it depending on the value of "cross." If the reserve does not exist, it inserts a new row into the database.
func (a *Service) WriteReserve(userId int64, address, symbol string, value float64, platform, protocol, cross string) error {

	if len(protocol) == 0 {
		protocol = types.ProtocolMainnet
	}

	// This code is querying a database for a specific set of information. The code is using placeholders ($1, $2, etc.) to
	// make the query more secure by preventing SQL injection. The row variable is the result of the query, and the defer
	// statement ensures that the connection to the database is closed after the query has finished. Finally, the if
	// statement is used to check for any errors that might occur during the query.
	row, err := a.Context.Db.Query("select id from reserves where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5", userId, symbol, platform, protocol, address)
	if err != nil {
		return err
	}
	defer row.Close()

	// The purpose of this statement is to check if there is a row available to be read from a database. If so, the code
	// within the if block will run. If not, it will skip the code within the if block.
	if row.Next() {

		switch cross {
		case types.BalancePlus:

			// This code is updating a database table called "reserves" with values provided by the user. The code first checks to
			// see if the update is successful, and if it fails, it returns an error.
			if _, err := a.Context.Db.Exec("update reserves set value = value + $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, protocol, address, value); err != nil {
				return err
			}
			break
		case types.BalanceMinus:

			// This code is updating a database table called "reserves" with values provided by the user. The code first checks to
			// see if the update is successful, and if it fails, it returns an error.
			if _, err := a.Context.Db.Exec("update reserves set value = value - $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, protocol, address, value); err != nil {
				return err
			}
			break
		}

		return nil
	}

	// This code is performing an SQL query to insert data into a database table called "reserves". The data being inserted
	// consists of user_id, symbol, platform, protocol, address, and value. If there is an error in executing the query, the
	// function will return the error.
	if _, err = a.Context.Db.Exec("insert into reserves (user_id, symbol, platform, protocol, address, value) values ($1, $2, $3, $4, $5, $6)", userId, symbol, platform, protocol, address, value); err != nil {
		return err
	}

	return nil
}

// WriteReserveLock - This function is used to update a record in the 'reserves' table in a database. It sets the 'lock' column to 'true'
// for a record that matches the given userId, symbol, platform and protocol. This will help ensure that only one process
// can access this record in the table at any given time, allowing for concurrent access to the table without conflicts.
func (a *Service) WriteReserveLock(userId int64, symbol, platform, protocol string) error {

	if len(protocol) == 0 {
		protocol = types.ProtocolMainnet
	}

	// This code is used to update the "lock" field in the "reserves" table in the database. The specific record to be
	// updated is identified using the userId, symbol, platform, and protocol values provided as arguments to the Exec()
	// function. If the update operation is not successful, an error is returned.
	if _, err := a.Context.Db.Exec("update reserves set lock = $5 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4;", userId, symbol, platform, protocol, true); err != nil {
		return err
	}
	return nil
}

// WriteReserveUnlock - The purpose of this function is to update the "lock" field of the "reserves" table in the database to false, where the
// user_id, symbol, platform, and protocol fields all match the given parameters. This is likely used to allow users to
// access the reserves for a given symbol on a given platform and protocol.
func (a *Service) WriteReserveUnlock(userId int64, symbol, platform, protocol string) error {

	if len(protocol) == 0 {
		protocol = types.ProtocolMainnet
	}

	// This code is part of an if statement and is used to update a database table. The code is used to update a "reserves"
	// table in the database, setting the "lock" field to false where the user ID, symbol, platform, and protocol all match
	// the given parameters. If any errors occur while executing the query, the code will return an error.
	if _, err := a.Context.Db.Exec("update reserves set lock = $5 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4;", userId, symbol, platform, protocol, false); err != nil {
		return err
	}
	return nil
}

// WriteReverse - The purpose of this code is to query a database for a specific set of information, insert data into a database table,
// and update the value of an existing record in the database. The code uses SQL queries to perform these operations and
// also checks for errors that may occur in the process.
func (a *Service) WriteReverse(userId int64, address, symbol string, value float64, platform, cross string) error {

	// This code is querying a database for a specific set of information. The code is using placeholders ($1, $2, etc.) to
	// make the query more secure by preventing SQL injection. The row variable is the result of the query, and the defer
	// statement ensures that the connection to the database is closed after the query has finished. Finally, the if
	// statement is used to check for any errors that might occur during the query.
	row, err := a.Context.Db.Query("select id from reserves where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5", userId, symbol, platform, types.ProtocolMainnet, address)
	if err != nil {
		return err
	}
	defer row.Close()

	// The purpose of this statement is to check if there is a row available to be read from a database. If so, the code
	// within the if block will run. If not, it will skip the code within the if block.
	if row.Next() {

		switch cross {
		case types.BalancePlus:

			// This code is updating a database entry with the given parameters. The specific entry is the 'reverse' field of the
			// 'reserves' table, and the update is an increase by the amount of 'value'. The other parameters (userId, symbol,
			// platform, protocol, address) are used to specify which entry should be updated. If an error is encountered, the
			// code will return an error.
			if _, err := a.Context.Db.Exec("update reserves set reverse = reverse + $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, types.ProtocolMainnet, address, value); err != nil {
				return err
			}
			break
		case types.BalanceMinus:

			// This code is updating a database entry with the given parameters. The specific entry is the 'reverse' field of the
			// 'reserves' table, and the update is an increase by the amount of 'value'. The other parameters (userId, symbol,
			// platform, protocol, address) are used to specify which entry should be updated. If an error is encountered, the
			// code will return an error.
			if _, err := a.Context.Db.Exec("update reserves set reverse = reverse - $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, types.ProtocolMainnet, address, value); err != nil {
				return err
			}
			break
		}

		return nil
	}

	// This code is performing an SQL query to insert data into a database table called "reserves". The data being inserted
	// consists of user_id, symbol, platform, protocol, address, and value. If there is an error in executing the query, the
	// function will return the error.
	if _, err = a.Context.Db.Exec("insert into reserves (user_id, symbol, platform, protocol, address, reverse) values ($1, $2, $3, $4, $5, $6)", userId, symbol, platform, types.ProtocolMainnet, address, value); err != nil {
		return err
	}

	return nil
}
