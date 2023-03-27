package stock

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/svarlamov/goyhfin"
	"strings"

	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"time"
)

func (s *Service) price() {

	// The code above creates a new ticker that will run once a minute and then loop through each range of the ticker.C
	// channel. This is useful for running a certain task or operation on a regular interval of time.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code queries a database for pairs with a status of true and orders them by their ID. It uses the
			// e.Context.Db.Query() function to execute a query and assigns the output of the query to the rows variable. If there
			// is an error, it calls the e.Context.Debug() function to debug the error, and if successful, it will defer the
			// rows.Close() function to close the rows after the function call.
			rows, err := s.Context.Db.Query(`select id, price, symbol, zone from stocks where status = $1 order by id`, true)
			if s.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for loop with the rows.Next() statement is used to loop through a result set of an SQL query. The rows.Next()
			// statement advances the current row pointer to the next row and returns true if it was successful. This loop is used
			// to iterate through each row in the result set and perform an action on it.
			for rows.Next() {

				// This code creates two variables, pair and price, of the types pbspot.Pair and float64 respectively. This allows
				// the program to store and use data of these two types.
				var (
					pair  pbspot.Pair
					price float64
				)

				// This code is used to scan the rows of data from a database and assign the values to variables. The if statement is
				// used to check for any errors that may occur when scanning the rows. If an error is found, the code will skip to
				// the next row. The e.Context.Debug() function is used to provide more information about what caused the error,
				// which can help with debugging.
				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); s.Context.Debug(err) {
					return
				}

				// The purpose of this code is to retrieve two candles from a given pair of base and quote units. The GetTicker()
				// function is used to retrieve the candles and the returned value is stored in the migrate variable. If an error is
				// encountered, the code will skip the iteration and continue to the next one.
				migrate, err := s.GetTicker(context.Background(), &pbstock.GetRequestTicker{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2})
				if s.Context.Debug(err) {
					return
				}

				// This code is used to get the ticker data from the Yahoo Finance API. The parameters passed to the GetTickerData
				// function are the base unit (in uppercase), the time period, the interval and a boolean indicating if the data
				// should be adjusted for splits. If an error is encountered, it is displayed using the Context.Debug method. If an
				// error is encountered, the function will return.
				resp, err := goyhfin.GetTickerData(strings.ToUpper(pair.GetBaseUnit()), goyhfin.OneMinute, goyhfin.OneMinute, false)
				if s.Context.Debug(err) {
					return
				}

				// This is a conditional statement that checks the length of the list "resp.Quotes" and only executes the code within
				// the if statement if the length is greater than 0.
				if len(resp.Quotes) > 0 {

					// The purpose of this code is to get the closing price of a stock from the response object. The response object is
					// assumed to contain a "Quotes" array with the information about the stock quote. The code retrieves the closing
					// price of the stock from the first element of the array.
					price = resp.Quotes[0].Close

					// This code is calculating the price of an item. The purpose of the if statement is to check if there are any
					// "migrate.Fields" present. If there are, then the price is calculated by taking the average of the price, the
					// pair's price, and the price of the first field in to migrate.Fields array. If there are no migrate.Fields, then
					// the price is calculated by taking the average of the price and the pair's price.
					if len(migrate.Fields) > 0 {
						price = (price + pair.GetPrice() + migrate.Fields[0].GetPrice()) / 3
					} else {
						price = (price + pair.GetPrice()) / 2
					}

					// This piece of code is calculating the price of a pair of something.  The if statement is checking if the price of
					// the pair is more than 100. If it is, then the price is reduced by 1/8 of the difference between the initial price
					// and the new price.
					if (price - pair.GetPrice()) > 100 {
						price -= (price - pair.GetPrice()) - (price-pair.GetPrice())/8
					}

				}

				// This is an if statement that checks whether the variable price is equal to 0. If it is, the code inside the curly
				// braces will be executed. Otherwise, it will be skipped.
				if price == 0 {

					// This code is used to calculate the price of an item. It checks if the migrate.Fields array has any elements in
					// it. If it does, it takes the first element, gets its price, adds that to the price of the pair, and divides the
					// sum by 2. If the array is empty, it just returns the price of the pair.
					if len(migrate.Fields) > 0 {
						price = (migrate.Fields[0].GetPrice() + pair.GetPrice()) / 2
					} else {
						price = pair.GetPrice()
					}

				}

				// This code is attempting to update a row in the database table "pairs" with the given values. The if statement is
				// checking to see if there is an error and if there is, the code will continue without changing the values.
				if _, err := s.Context.Db.Exec("update stocks set price = $3 where symbol = $1 and zone = $2;", pair.GetBaseUnit(), pair.GetQuoteUnit(), price); s.Context.Debug(err) {
					return
				}

				time.Sleep(time.Second * 1)
			}
		}()
	}
}

// market - This function is used to replay market prices. The function is executed on a specific time interval and retrieves data
// from the database. It then inserts the data into the trades table, and publishes the data to exchange topics. This
// allows for the market data to be replayed at a specific interval.
func (s *Service) market() {

	// The code creates a ticker that triggers every minute and runs a loop that executes each time the ticker is triggered.
	// This allows code to be executed at regular intervals without the need for an explicit loop.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code is performing a database query in order to retrieve data from the stocks table. The query is selecting
			// the fields id, price, symbol, and zone where the status is equal to true (in this case, it is likely a boolean
			// value). The query is then ordered by the id field. The rows variable is used to store the results of the database
			// query and the err variable is used to store any errors that occur during the query. If an error occurs, the debug
			// method is called, which likely prints out the error and stops the program. The rows variable is then closed as part
			// of the defer statement.
			rows, err := s.Context.Db.Query(`select id, price, symbol, zone from stocks where status = $1 order by id`, true)
			if s.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for rows.Next() loop is used to iterate over each row in a database result set. It allows you to access each
			// row one at a time, until all rows have been processed. This is useful for processing large result sets without
			// loading them all into memory at once.
			for rows.Next() {

				// The above statement is declaring a variable named 'item' of type 'pbstock.Asset'. This statement is used to create
				// a variable that can hold values of type 'pbstock.Asset', which is a type of asset such as stocks, bonds, etc. that
				// can be traded on the stock market.
				var (
					item pbstock.Asset
				)

				// This code is checking for an error when scanning the rows of a database table. The if statement scans the rows of
				// the database table using the Scan() method, and if it encounters an error, it will log the error and continue
				// scanning the remaining rows.
				if err := rows.Scan(&item.Id, &item.Price, &item.Symbol, &item.Zone); s.Context.Debug(err) {
					continue
				}

				// This code is part of a loop that is looping over a list of currency pairs. The purpose of this code is to insert
				// the information from each pair (assigning, base_unit, quote_unit, price, quantity, market) into a table called
				// trades. The code is checking for any errors and if there is an error, it is continuing the loop.
				if _, err := s.Context.Db.Exec(`insert into ohlcv (assigning, base_unit, quote_unit, price, quantity, market) values ($1, $2, $3, $4, $5, $6);`, proto.Assigning_MARKET_PRICE, item.GetSymbol(), item.GetZone(), item.GetPrice(), 0, true); s.Context.Debug(err) {
					continue
				}

				// The code above is iterating over the Depth() function from the help package. The purpose of this code is to loop
				// through the values returned by the Depth() function and assign them to the interval variable. This allows the code
				// to perform certain operations on each of the values returned by the function.
				for _, interval := range help.Depth() {

					// The purpose of this code is to get the most recent two candles for a given currency pair and interval, using the
					// pbstock.GetRequestTicker object. If an error occurs during the process, the code will continue execution, but the
					// error will be logged in the debug logs.
					migrate, err := s.GetTicker(context.Background(), &pbstock.GetRequestTicker{BaseUnit: item.GetSymbol(), QuoteUnit: item.GetZone(), Limit: 2, Resolution: interval})
					if s.Context.Debug(err) {
						continue
					}

					// This code is part of a loop that is attempting to publish a message to an exchange. The if statement checks the
					// result of the Publish method and, if an error occurs, the code continues to the next iteration of the loop.
					if err := s.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/ticker:%v", interval)); s.Context.Debug(err) {
						continue
					}
				}
			}
		}()
	}
}

// trade - This function is used to replay a trade init. It takes an order and a side (BID or ASK) as parameters. It then queries
// the database for orders with the same base unit, quote unit and user ID, and with a status of "PENDING". It then
// iterates through the results and checks if the order's price is higher than the item's price for a BID position and
// lower for an ASK position. If this is the case, it calls the replayTradeProcess() function. Finally, it logs any matches or failed matches.
func (s *Service) trade(order *pbstock.Order, side proto.Side) {

	// This code is checking for an error when publishing to the exchange. If an error occurs, the code is printing out the
	// error and returning.
	if err := s.Context.Publish(order, "exchange", "order/create"); s.Context.Debug(err) {
		return
	}

	// This code is querying the "orders" table in a database for data that matches the given parameters. It is using the
	// parameters given to query for a specific set of data from the "orders" table. It is using the $1, $2, $3, $4, $5 and
	// $6 to represent the given parameters. The query is also ordering the results by the "id" column. It is checking for
	// errors and deferring the closing of the rows.
	rows, err := s.Context.Db.Query(`select id, assigning, base_unit, quote_unit, value, quantity, price, user_id, status from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and user_id != $4 and status = $5 and type = $6 order by id`, side, order.GetBaseUnit(), order.GetQuoteUnit(), order.GetUserId(), proto.Status_PENDING, proto.Type_STOCK)
	if s.Context.Debug(err) {
		return
	}
	defer rows.Close()

	// The purpose of the for loop is to iterate over a set of rows from a database query. The rows.Next() function advances
	// the iterator to the next row in the result set, returning false when there are no more rows to iterate over.
	for rows.Next() {

		// The purpose of this line of code is to declare a variable named 'item' of type 'pbstock.Order'. This will allow the
		// programmer to use the 'item' variable to store data of type 'pbstock.Order' in the program.
		var (
			item pbstock.Order
		)

		// This code is attempting to scan the rows of a database table, assigning each column value to a variable (item.Id,
		// item.Assigning, etc.). The if statement is checking for any errors that might occur when scanning the rows and
		// returning any errors that might be present.
		if err = rows.Scan(&item.Id, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.Value, &item.Quantity, &item.Price, &item.UserId, &item.Status); err != nil {
			return
		}

		// This code is used to query a database for a specific row. The query is looking for an entry with a specific ID and
		// status. The two parameters (order.GetId() and pbstock.Status_PENDING) are used to filter the query results. The row
		// variable will store the results of the query, and the err variable will store any errors that occur.
		row, err := s.Context.Db.Query("select value from orders where id = $1 and status = $2 and type = $3", order.GetId(), proto.Status_PENDING, proto.Type_STOCK)
		if err != nil {
			return
		}

		// The purpose of this code is to check if there is any data in the "row" and if there is, attempt to scan it and get
		// the "Value" from the row. If there is no data, then set the "Value" to 0. Finally, close the row to free up any resources.
		if row.Next() {
			if err = row.Scan(&order.Value); err != nil {
				return
			}
		} else {
			order.Value = 0
		}
		row.Close()

		// This switch statement is used to check for a match between the order and item prices, depending on the side of the
		// trade (bid or ask). If the order and item prices match, the trade process is replayed. If not, a message is logged
		// for the user. If the side is invalid, an error is returned.
		switch side {

		case proto.Side_BID: // Buy at BID price.

			// This code checks whether the price of an order is greater than or equal to the price of an item. If it is, it will
			// log a message and call the replayTradeProcess function. If it is not, it will log another message.
			if order.GetPrice() >= item.GetPrice() {
				s.Context.Logger.Infof("[BID]: (item [%v]) >= (order [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				// This function is used to replay a trade process. It takes two parameters, an order and an item, and replays the
				// trade process associated with them. The order and item parameters contain the necessary information needed to
				// replay the trade process, allowing the process to be repeated in order to confirm the accuracy of the trade.
				s.process(order, &item)

			} else {
				s.Context.Logger.Infof("[BID]: no matches found: (item [%v]) >= (order [%v])", order.GetPrice(), item.GetPrice())
			}

			break

		case proto.Side_ASK: // Sell at ASK price.

			// This code is checking if the price of an order is lower than or equal to the price of an item. If it is, it will
			// log an informational message and call the replayTradeProcess() method. If not, it will log a different
			// informational message.
			if order.GetPrice() <= item.GetPrice() {
				s.Context.Logger.Infof("[ASK]: (order [%v]) <= (item [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				// This function is used to replay a trade process. It takes two parameters, an order and an item, and replays the
				// trade process associated with them. The order and item parameters contain the necessary information needed to
				// replay the trade process, allowing the process to be repeated in order to confirm the accuracy of the trade.
				s.process(order, &item)

			} else {
				s.Context.Logger.Infof("[ASK]: no matches found: (order [%v]) <= (item [%v])", order.GetPrice(), item.GetPrice())
			}

			break
		default:
			if err := s.Context.Debug(status.Error(11589, "invalid assigning trade position")); err {
				return
			}
		}

	}

	// The purpose of this code is to check for errors when using the rows.Err() function. It returns an error if there is
	// one and returns the function if this is the case.
	if err = rows.Err(); err != nil {
		return
	}
}

// process - This function is used to replay a trade process. It updates two orders with different amounts to determine the result
// of a trade. It updates the order status in the database with pending in to filled, updates the balance by adding the
// amount of the order to the balance, and sends a mail. In addition, it logs information about the trade.
func (s *Service) process(params ...*pbstock.Order) {

	// The purpose of this code is to declare and initialize a set of four variables. The first three variables are "value",
	// "equal", and "instance", and they are all declared as type float64, bool, and int, respectively. The last variable is
	// "migrate", and is declared as type query.Migrate, and is initialized with a Context set to e.Context.
	var (
		value    float64
		instance int
	)

	// This code is comparing two values (from the params array) to determine if they are equal or not. If they are equal,
	// it sets the 'equal' boolean to true and the 'instance' variable to 1. If they are not equal, it sets the 'equal'
	// boolean to false and the 'instance' variable to 0.
	if params[0].GetValue() >= params[1].GetValue() {
		instance = 1
	} else {
		instance = 0
	}

	spew.Dump(params[instance].GetValue())

	// This code is used to update an order status from pending to filled when the order is completed. It also updates the
	// quantity of the orders and sets the necessary parameters for the order. Finally, it logs the parameters of the order.
	if params[instance].GetValue() > 0 {

		// The purpose of the for loop is to iterate over the parameters passed in and update the "value" of the specified
		// order in the database. It also sets the status of the order to FILLED if the value is equal to 0. The code also
		// checks for any errors that may occur during the process. Lastly, the code sends an email to the user associated with the order once the order is filled.
		for i := 0; i < 2; i++ {

			// This if statement is used to update the "value" of a particular order in the database. The parameters passed in are
			// used in the query to find the specific order to update. If the query is successful, the "value" of the order is
			// stored in the "value" variable and the function will continue. If the query fails, the function will return.
			if err := s.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 and type = $4 returning value;", params[i].GetId(), params[instance].GetValue(), proto.Status_PENDING, proto.Type_STOCK).Scan(&value); err != nil {
				return
			}

			if value == 0 {

				// This code is performing an update on the orders table in a database. It is setting the status of the order with the
				// specified ID to the specified status (in this case, FILLED). The code is also checking for any errors that may
				// occur during the process. If an error is found, the code will return without proceeding.
				if _, err := s.Context.Db.Exec("update orders set status = $2 where id = $1 and type = $3;", params[i].GetId(), proto.Status_FILLED, proto.Type_STOCK); err != nil {
					return
				}
			}
		}

		switch params[1].GetAssigning() {
		case proto.Assigning_BUY:

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error,
			// the function will return without doing anything.
			if err := s.setBalance(params[0].GetQuoteUnit(), params[0].GetUserId(), decimal.New(params[instance].GetValue()).Mul(params[1].GetPrice()).Float(), proto.Balance_PLUS); err != nil {
				return
			}

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error,
			// the function will return without doing anything.
			if err := s.setBalance(params[0].GetBaseUnit(), params[1].GetUserId(), params[instance].GetValue(), proto.Balance_PLUS); err != nil {
				return
			}

			break
		case proto.Assigning_SELL:

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error,
			// the function will return without doing anything.
			if err := s.setBalance(params[0].GetBaseUnit(), params[0].GetUserId(), params[instance].GetValue(), proto.Balance_PLUS); err != nil {
				return
			}

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error,
			// the function will return without doing anything.
			if err := s.setBalance(params[0].GetQuoteUnit(), params[1].GetUserId(), decimal.New(params[instance].GetValue()).Mul(params[0].GetPrice()).Float(), proto.Balance_PLUS); err != nil {
				return
			}

			break
		}
	}

	// The purpose of this code is to check for an error when the setTrade method is called with the given parameters. If
	// there is an error, it will return without continuing with the code.
	if err := s.setTrade(params[0], params[1]); err != nil {
		return
	}
}
