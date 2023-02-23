package spot

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/blockchain"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"google.golang.org/grpc/status"
	"time"
)

// replayPriceScale - The purpose of this code is to update the prices of pairs at a specific time interval. It first loads the pair ids,
// prices, base units, and quote units from the pairs table where the status is active. It then gets the candles for the
// base and quote units and calculates the new price based on the data. Lastly, it updates the price of the pair in the database.
func (e *Service) replayPriceScale() {

	// The code above creates a new ticker that will run once a minute and then loop through each range of the ticker.C
	// channel. This is useful for running a certain task or operation on a regular interval of time.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code queries a database for pairs with a status of true and orders them by their ID. It uses the
			// e.Context.Db.Query() function to execute a query and assigns the output of the query to the rows variable. If there
			// is an error, it calls the e.Context.Debug() function to debug the error, and if successful, it will defer the
			// rows.Close() function to close the rows after the function call.
			rows, err := e.Context.Db.Query(`select id, price, base_unit, quote_unit from pairs where status = $1 order by id`, true)
			if e.Context.Debug(err) {
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
				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); e.Context.Debug(err) {
					return
				}

				// The purpose of this code is to retrieve two candles from a given pair of base and quote units. The GetCandles()
				// function is used to retrieve the candles and the returned value is stored in the migrate variable. If an error is
				// encountered, the code will skip the iteration and continue to the next one.
				migrate, err := e.GetCandles(context.Background(), &pbspot.GetRequestCandles{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2})
				if e.Context.Debug(err) {
					return
				}

				// This if statement checks if the price of a given pair of currencies is greater than 0. This is important to ensure
				// that the price is a valid number and is not negative.
				if price = marketplace.Price().Unit(pair.GetBaseUnit(), pair.GetQuoteUnit()); price > 0 {

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
				if _, err := e.Context.Db.Exec("update pairs set price = $3 where base_unit = $1 and quote_unit = $2;", pair.GetBaseUnit(), pair.GetQuoteUnit(), price); e.Context.Debug(err) {
					return
				}
			}
		}()
	}
}

// replayMarket - This function is used to replay market prices. The function is executed on a specific time interval and retrieves data
// from the database. It then inserts the data into the trades table, and publishes the data to exchange topics. This
// allows for the market data to be replayed at a specific interval.
func (e *Service) replayMarket() {

	// The code creates a ticker that triggers every minute and runs a loop that executes each time the ticker is triggered.
	// This allows code to be executed at regular intervals without the need for an explicit loop.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code allows the program to query a database and retrieve the values of the 'id', 'price', 'base_unit', and
			// 'quote_unit' columns from the 'pairs' table, where the 'status' column is equal to 'true'. The code then closes the
			// rows when the query is complete.
			rows, err := e.Context.Db.Query(`select id, price, base_unit, quote_unit from pairs where status = $1 order by id`, true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for rows.Next() loop is used to iterate over each row in a database result set. It allows you to access each
			// row one at a time, until all rows have been processed. This is useful for processing large result sets without
			// loading them all into memory at once.
			for rows.Next() {

				// This is a variable declaration statement. The variable 'pair' is being declared as type 'pbspot.Pair'. This allows
				// the variable to store a pair of values (e.g. two integers, two strings, two objects, etc.).
				var (
					pair pbspot.Pair
				)

				// This code is checking for an error when scanning the rows of a database table. The if statement scans the rows of
				// the database table using the Scan() method, and if it encounters an error, it will log the error and continue
				// scanning the remaining rows.
				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit); e.Context.Debug(err) {
					continue
				}

				// This code is part of a loop that is looping over a list of currency pairs. The purpose of this code is to insert
				// the information from each pair (assigning, base_unit, quote_unit, price, quantity, market) into a table called
				// trades. The code is checking for any errors and if there is an error, it is continuing the loop.
				if _, err := e.Context.Db.Exec(`insert into trades (assigning, base_unit, quote_unit, price, quantity, market) values ($1, $2, $3, $4, $5, $6)`, pbspot.Assigning_MARKET_PRICE, pair.GetBaseUnit(), pair.GetQuoteUnit(), pair.GetPrice(), 0, true); e.Context.Debug(err) {
					continue
				}

				// The code above is iterating over the Depth() function from the help package. The purpose of this code is to loop
				// through the values returned by the Depth() function and assign them to the interval variable. This allows the code
				// to perform certain operations on each of the values returned by the function.
				for _, interval := range help.Depth() {

					// The purpose of this code is to get the most recent two candles for a given currency pair and interval, using the
					// pbspot.GetRequestCandles object. If an error occurs during the process, the code will continue execution, but the
					// error will be logged in the debug logs.
					migrate, err := e.GetCandles(context.Background(), &pbspot.GetRequestCandles{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2, Resolution: interval})
					if e.Context.Debug(err) {
						continue
					}

					// This code is part of a loop that is attempting to publish a message to an exchange. The if statement checks the
					// result of the Publish method and, if an error occurs, the code continues to the next iteration of the loop.
					if err := e.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/candles:%v", interval)); e.Context.Debug(err) {
						continue
					}
				}
			}
		}()
	}
}

// replayChainStatus - This function is used to replay the status of chains stored in a database. It loads at a specific time interval and
// queries the database for chains that have been stored. It then uses the 'help.Ping' function to check whether each
// chain is available or not. If the chain is not available, its status is set to false, and then updated in the database.
func (e *Service) replayChainStatus() {

	// The code is creating a new ticker that will fire every minute (time.Minute * 1). The for loop will continually
	// execute until the ticker is stopped or the program exits. This code is useful for creating a repeating task at a
	// regular interval. For example, if you wanted to perform a task every minute, you could use this code to do so.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code snippet is querying a database table to retrieve data. The purpose of this code is to query the "chains"
			// table in the database and retrieve the columns "id", "rpc", and "status". The variable "rows" will store the result
			// of the query. The variable "err" is used to catch any errors that may occur during the query. If an error is
			// caught, the code will print the error and return. The code also uses "defer rows.Close()" to ensure that the rows
			// are closed after the query is finished.
			rows, err := e.Context.Db.Query(`select id, rpc, status from chains`)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for loop is used to iterate over the rows in the result set of a query. The .Next() method is used to move the
			// cursor to the next row in the result set. Each iteration of the loop will assign the values in the row to the
			// variables given in the query.
			for rows.Next() {

				// This variable declaration is creating a variable named "item", which is of type "pbspot.Chain". This is a way of
				// creating a new variable and assigning it a data type.
				var (
					item pbspot.Chain
				)

				// This code is likely part of a loop that is iterating through the rows of a database query. The purpose of the if
				// statement is to scan each row, store the values in the item object, and check for any errors. If an error is
				// detected, the loop will continue to the next row.
				if err = rows.Scan(&item.Id, &item.Rpc, &item.Status); e.Context.Debug(err) {
					continue
				}

				// The purpose of this code is to check if a remote procedure call (RPC) is functioning correctly. The help.Ping()
				// method is used to ping the RPC, and if the ping is unsuccessful, the item's status is set to false.
				if ok := help.Ping(item.Rpc); !ok {
					item.Status = false
				}

				// This code is updating the status of an item in a database. The if statement is used to check for errors when
				// executing the update query, and the "continue" keyword is used to skip any further processing if an error is
				// encountered.
				if _, err := e.Context.Db.Exec("update chains set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
					continue
				}
			}

		}()
	}
}

// replayTradeInit - This function is used to replay a trade init. It takes an order and a side (BID or ASK) as parameters. It then queries
// the database for orders with the same base unit, quote unit and user ID, and with a status of "PENDING". It then
// iterates through the results and checks if the order's price is higher than the item's price for a BID position and
// lower for an ASK position. If this is the case, it calls the replayTradeProcess() function. Finally, it logs any matches or failed matches.
func (e *Service) replayTradeInit(order *pbspot.Order, side pbspot.Side) {

	// This code is checking for an error when publishing to the exchange. If an error occurs, the code is printing out the
	// error and returning.
	if err := e.Context.Publish(order, "exchange", "order/create"); e.Context.Debug(err) {
		return
	}

	// This code is querying the "orders" table in a database for data that matches the given parameters. It is using the
	// parameters given to query for a specific set of data from the "orders" table. It is using the $1, $2, $3, $4, $5 and
	// $6 to represent the given parameters. The query is also ordering the results by the "id" column. It is checking for
	// errors and deferring the closing of the rows.
	rows, err := e.Context.Db.Query(`select id, assigning, base_unit, quote_unit, value, quantity, price, user_id, status from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and user_id != $4 and status = $5 and type = $6 order by id`, side, order.GetBaseUnit(), order.GetQuoteUnit(), order.GetUserId(), pbspot.Status_PENDING, pbspot.OrderType_SPOT)
	if e.Context.Debug(err) {
		return
	}
	defer rows.Close()

	// The purpose of the for loop is to iterate over a set of rows from a database query. The rows.Next() function advances
	// the iterator to the next row in the result set, returning false when there are no more rows to iterate over.
	for rows.Next() {

		// The purpose of this line of code is to declare a variable named 'item' of type 'pbspot.Order'. This will allow the
		// programmer to use the 'item' variable to store data of type 'pbspot.Order' in the program.
		var (
			item pbspot.Order
		)

		// This code is attempting to scan the rows of a database table, assigning each column value to a variable (item.Id,
		// item.Assigning, etc.). The if statement is checking for any errors that might occur when scanning the rows and
		// returning any errors that might be present.
		if err = rows.Scan(&item.Id, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.Value, &item.Quantity, &item.Price, &item.UserId, &item.Status); err != nil {
			return
		}

		// This code is used to query a database for a specific row. The query is looking for an entry with a specific ID and
		// status. The two parameters (order.GetId() and pbspot.Status_PENDING) are used to filter the query results. The row
		// variable will store the results of the query, and the err variable will store any errors that occur.
		row, err := e.Context.Db.Query("select value from orders where id = $1 and status = $2", order.GetId(), pbspot.Status_PENDING)
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

		case pbspot.Side_BID: // Buy at BID price.

			// This code checks whether the price of an order is greater than or equal to the price of an item. If it is, it will
			// log a message and call the replayTradeProcess function. If it is not, it will log another message.
			if order.GetPrice() >= item.GetPrice() {
				e.Context.Logger.Infof("[BID]: (item [%v]) >= (order [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				// This function is used to replay a trade process. It takes two parameters, an order and an item, and replays the
				// trade process associated with them. The order and item parameters contain the necessary information needed to
				// replay the trade process, allowing the process to be repeated in order to confirm the accuracy of the trade.
				e.replayTradeProcess(order, &item)

			} else {
				e.Context.Logger.Infof("[BID]: no matches found: (item [%v]) >= (order [%v])", order.GetPrice(), item.GetPrice())
			}

			break

		case pbspot.Side_ASK: // Sell at ASK price.

			// This code is checking if the price of an order is lower than or equal to the price of an item. If it is, it will
			// log an informational message and call the replayTradeProcess() method. If not, it will log a different
			// informational message.
			if order.GetPrice() <= item.GetPrice() {
				e.Context.Logger.Infof("[ASK]: (order [%v]) <= (item [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				// This function is used to replay a trade process. It takes two parameters, an order and an item, and replays the
				// trade process associated with them. The order and item parameters contain the necessary information needed to
				// replay the trade process, allowing the process to be repeated in order to confirm the accuracy of the trade.
				e.replayTradeProcess(order, &item)

			} else {
				e.Context.Logger.Infof("[ASK]: no matches found: (order [%v]) <= (item [%v])", order.GetPrice(), item.GetPrice())
			}

			break
		default:
			if err := e.Context.Debug(status.Error(11589, "invalid assigning trade position")); err {
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

// replayTradeProcess - This function is used to replay a trade process. It updates two orders with different amounts to determine the result
// of a trade. It updates the order status in the database with pending in to filled, updates the balance by adding the
// amount of the order to the balance, and sends a mail. In addition, it logs information about the trade.
func (e *Service) replayTradeProcess(params ...*pbspot.Order) {

	// The purpose of the code snippet is to declare two variables, value of type float64 and migrate of type query.Migrate.
	// To migrate variable is initialized with the context from the environment (e.Context).
	var (
		value   float64
		migrate = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of a trade matching algorithm for a spot market. The purpose of the code is to match two orders and
	// settle the trade according to the parameters of each order. The code first checks if the amount of the current order
	// is greater than or equal to the amount of the found order. If so, then it checks if the amount of the found order is
	// greater than zero and proceeds to update the order values and statuses accordingly. It also sets the fees and
	// parameters for each order. If the amount of the current order is less than the amount of the found order, the code
	// does the same thing but with the amount of the current order. Finally, it logs the details of the trade.
	if params[0].GetValue() >= params[1].GetValue() {

		// This code is used to update an order status from pending to filled when the order is completed. It also updates the
		// quantity of the orders and sets the necessary parameters for the order. Finally, it logs the parameters of the order.
		if params[1].GetValue() > 0 {

			// This code is performing an update query on the database. The query is updating the "value" field in the "orders"
			// table. The parameters for the query are the ID (params[0].GetId()), the value to subtract (params[1].GetValue()),
			// and the current status (pbspot.Status_PENDING). The code is also returning the updated value (value) after the
			// query is executed. The "if err" statement is used to check for any errors that may have occurred during the query
			// execution. If an error is found, the code returns.
			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[0].GetId(), params[1].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// This code is used to update the status of an order to "filled" and to send an email with the order details to the
			// user who placed the order. The code checks if the value is 0, and if it is, it executes an SQL query that updates the order status to "filled".
			if value == 0 {

				// This code is updating the status of an order in a database. Specifically, it is setting the status of the order
				// with the ID specified in params[0].GetId() to FILLED. It is using the Exec() method of the Db object to execute
				// the update statement. The if statement is checking if there is an error in executing the statement and, if so,
				// returning to prevent further execution.
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[0].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				// The purpose of this code is to send an email when an order is filled. The parameters passed in include the user
				// ID, the order ID, the quantity, the price, the base unit, the quote unit, and the assigning details. This code is
				// used to notify the user that their order has been filled and to provide them with details about the order.
				go migrate.SendMail(params[0].GetUserId(), "order_filled", params[0].GetId(), e.getQuantity(params[0].GetAssigning(), params[0].GetQuantity(), params[0].GetPrice(), false), params[0].GetBaseUnit(), params[0].GetQuoteUnit(), params[0].GetAssigning())
			}

			// This code is performing an update query on the 'orders' table in a database according to the provided parameters.
			// Specifically, the query is updating a value in the table by subtracting the value in the second parameter from the
			// original value. It is then returning the updated value. The 'if' statement is checking to ensure that the query is
			// working properly and returning an error if it fails.
			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[1].GetId(), params[1].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// The purpose of this code is to update an order's status in the database, and then email the user
			// associated with the order with specific details about the order. The code first checks to make sure the value is 0,
			// and then it executes a database query to update the order's status.
			if value == 0 {

				// The purpose of this code is to execute an update statement on the database associated with the context of 'e'. The
				// update statement is setting the status of an order with a given ID to 'filled'. If there is an error in executing
				// the statement, the code will return.
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[1].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				// The purpose of this code is to send an email when an order is filled. The parameters passed in include the user
				// ID, the order ID, the quantity, the price, the base unit, the quote unit, and the assigning details. This code is
				// used to notify the user that their order has been filled and to provide them with details about the order.
				go migrate.SendMail(params[1].GetUserId(), "order_filled", params[1].GetId(), e.getQuantity(params[1].GetAssigning(), params[1].GetQuantity(), params[1].GetPrice(), false), params[1].GetBaseUnit(), params[1].GetQuoteUnit(), params[1].GetAssigning())
			}

			// The purpose of this code is to update the balance of two users when they take part in a spot market trade. The
			// switch statement checks the assigning of the trade and updates the corresponding balances accordingly. It also sets
			// the parameters such as fees, maker, turn and equal in the order of each user.
			switch params[1].GetAssigning() {
			case pbspot.Assigning_BUY:

				// The purpose of this code is to calculate the total quantity and fees of a given quote unit, value, and price. It
				// takes the parameters of the quote unit, value, and price as inputs, then it calculates the total quantity and fees
				// using the getSum function and returns the calculated values.
				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.New(params[1].GetValue()).Mul(params[1].GetPrice()).Float(), false)

				// The purpose of this code is to set the balance of a user with the specified quantity and type. The function checks
				// for any errors, and if an error is encountered, the function returns.
				if err := e.setBalance(params[0].GetQuoteUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters of a pbspot order. The parameters include the fees, whether the
				// order is a maker order, whether the order is a turn order, and whether the order is an equal order.
				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: false, Equal: true}

				// The purpose of this line of code is to calculate the quantity and fees of a given parameter. The getSum() function
				// takes in three arguments: the base unit of the first parameter, the value of the second parameter, and a boolean
				// value that indicates if the quantity should be rounded up. The function then returns two values: the quantity and the fees associated with the parameter.
				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[1].GetValue(), true)

				// The purpose of this code is to set the balance of a user with the specified quantity and type. The function checks
				// for any errors, and if an error is encountered, the function returns.
				if err := e.setBalance(params[0].GetBaseUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters for a pbspot order. It sets the fees, whether the order is a
				// maker order, whether it has a turn and whether it is equal.
				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: true, Equal: true}

				break
			case pbspot.Assigning_SELL:

				// The purpose of this code is to calculate the total quantity and fees of a given base unit, value, and price. It
				// takes the parameters of the base unit, value, and price as inputs, then it calculates the total quantity and fees
				// using the getSum function and returns the calculated values.
				quantity, fees := e.getSum(params[0].GetBaseUnit(), params[1].GetValue(), false)

				// This code is used to set the balance of a user by adding a specified quantity to the user's existing balance. The
				// parameters used are the base unit, user ID, quantity, and the balance operation (in this case, PLUS). If an error
				// is encountered, the function returns without further execution.
				if err := e.setBalance(params[0].GetBaseUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters of a pbspot order. The parameters include the fees, whether the
				// order is a maker order, whether the order is a turn order, and whether the order is an equal order.
				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: true, Equal: true}

				// The purpose of this line of code is to calculate the quantity and fees of a given parameter. The getSum() function
				// takes in three arguments: the quote unit of the first parameter, the value of the second parameter, and a boolean
				// value that indicates if the quantity should be rounded up. The function then returns two values: the quantity and the fees associated with the parameter.
				quantity, fees = e.getSum(params[0].GetQuoteUnit(), decimal.New(params[1].GetValue()).Mul(params[0].GetPrice()).Float(), true)

				// The purpose of this code is to set the balance of a user with the specified quantity and type. The function checks
				// for any errors, and if an error is encountered, the function returns.
				if err := e.setBalance(params[0].GetQuoteUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters for a pbspot order. It sets the fees, whether the order is a
				// maker order, whether it has a turn and whether it is equal.
				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: false, Equal: true}

				break
			}

			e.Context.Logger.Infof("[(A)%v]: (Assigning: %v), (Price: [%v]-[%v]), (Value: [%v]-[%v]), (UserID: [%v]-[%v])", params[0].GetAssigning(), params[1].GetAssigning(), params[0].GetPrice(), params[1].GetPrice(), params[0].GetValue(), params[1].GetValue(), params[0].GetUserId(), params[1].GetUserId())
		}

		// If the amount of the current order is less than the amount of the found order.
	} else if params[0].GetValue() < params[1].GetValue() {

		// This code is part of a larger program that is used to handle orders in a spot trading system. The purpose of this
		// code is to check if the value of an order is greater than 0 and, if it is, then to update the order status and
		// update the order's user's balance. It also sends an email notification to the user if the order is filled. Finally,
		// it logs information about the order.
		if params[0].GetValue() > 0 {

			// This is a SQL query within an if statement used to update a record in a database. The purpose of this statement is
			// to update the value of a record with the given id, whose status is PENDING, by subtracting the given value
			// (params[0].GetValue()). The statement also returns the new value (value) of the updated record. If any error occurs
			// during the query, the statement returns without executing.
			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[0].GetId(), params[0].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// This code is part of an if statement checking if the value is equal to 0. If the value is equal to 0, the code
			// updates the orders table in the database with a status of FILLED and sends an email to the user with details about the order.
			if value == 0 {

				// This is an example of an if statement used to update a record in a database. The statement is checking if there is
				// an error when executing the update query. If there is an error, the statement returns. Otherwise, the query is
				// executed and the record is updated.
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[0].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				// This line of code is calling a function called migrate.SendMail, which is used to send an email notification to a
				// user when a certain event has occurred. The parameters being passed to the function are the user ID, the type of
				// event (in this case, "order_filled"), the order ID, the quantity of the order, the base unit and quote unit of the
				// order, and the assigning of the order. This information is used to email the user informing them that their order has been filled.
				go migrate.SendMail(params[0].GetUserId(), "order_filled", params[0].GetId(), e.getQuantity(params[0].GetAssigning(), params[0].GetQuantity(), params[0].GetPrice(), false), params[0].GetBaseUnit(), params[0].GetQuoteUnit(), params[0].GetAssigning())
			}

			// This code is attempting to update a row in the 'orders' table with a new value, based on the parameters received.
			// The code is first checking to see that the order has a status of 'PENDING' before attempting to update the value in
			// the row. After the update, the 'value' field is being returned. If there is an error encountered during this process, the code returns.
			if err := e.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and status = $3 returning value;", params[1].GetId(), params[0].GetValue(), pbspot.Status_PENDING).Scan(&value); err != nil {
				return
			}

			// This code is part of an if-statement. The purpose of this statement is to update an order's status to "FILLED" in
			// the database and to email the user to indicate that the order has been filled.
			if value == 0 {

				// This is an example of an if statement used to update a record in a database. The statement is checking if there is
				// an error when executing the update query. If there is an error, the statement returns. Otherwise, the query is
				// executed and the record is updated.
				if _, err := e.Context.Db.Exec("update orders set status = $2 where id = $1;", params[1].GetId(), pbspot.Status_FILLED); err != nil {
					return
				}

				// The purpose of this code is to send an email when an order is filled. The parameters passed in include the user
				// ID, the order ID, the quantity, the price, the base unit, the quote unit, and the assigning details. This code is
				// used to notify the user that their order has been filled and to provide them with details about the order.
				go migrate.SendMail(params[1].GetUserId(), "order_filled", params[1].GetId(), e.getQuantity(params[1].GetAssigning(), params[1].GetQuantity(), params[1].GetPrice(), false), params[1].GetBaseUnit(), params[1].GetQuoteUnit(), params[1].GetAssigning())
			}

			// The purpose of this code is to update the balance of two users when they take part in a spot market trade. The
			// switch statement checks the assigning of the trade and updates the corresponding balances accordingly. It also sets
			// the parameters such as fees, maker, turn and equal in the order of each user.
			switch params[1].GetAssigning() {
			case pbspot.Assigning_BUY:

				// The purpose of this code is to calculate the total quantity and fees of a given quote unit, value, and price. It
				// takes the parameters of the quote unit, value, and price as inputs, then it calculates the total quantity and fees
				// using the getSum function and returns the calculated values.
				quantity, fees := e.getSum(params[0].GetQuoteUnit(), decimal.New(params[0].GetValue()).Mul(params[1].GetPrice()).Float(), false)

				// The purpose of this code is to set the balance of a user with the specified quantity and type. The function checks
				// for any errors, and if an error is encountered, the function returns.
				if err := e.setBalance(params[0].GetQuoteUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters of a pbspot order. The parameters include the fees, whether the
				// order is a maker order, whether the order is a turn order, and whether the order is an equal order.
				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: false, Equal: false}

				// The purpose of this line of code is to calculate the quantity and fees of a given parameter. The getSum() function
				// takes in three arguments: the base unit of the first parameter, the value of the second parameter, and a boolean
				// value that indicates if the quantity should be rounded up. The function then returns two values: the quantity and the fees associated with the parameter.
				quantity, fees = e.getSum(params[0].GetBaseUnit(), params[0].GetValue(), true)

				// The purpose of this code is to set the balance of a user with the specified quantity and type. The function checks
				// for any errors, and if an error is encountered, the function returns.
				if err := e.setBalance(params[0].GetBaseUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters for a pbspot order. It sets the fees, whether the order is a
				// maker order, whether it has a turn and whether it is equal.
				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: true, Equal: false}

				break
			case pbspot.Assigning_SELL:

				// The purpose of this code is to calculate the total quantity and fees of a given base unit, value, and price. It
				// takes the parameters of the base unit, value, and price as inputs, then it calculates the total quantity and fees
				// using the getSum function and returns the calculated values.
				quantity, fees := e.getSum(params[0].GetBaseUnit(), params[0].GetValue(), false)

				// This code is used to set the balance of a user by adding a specified quantity to the user's existing balance. The
				// parameters used are the base unit, user ID, quantity, and the balance operation (in this case, PLUS). If an error
				// is encountered, the function returns without further execution.
				if err := e.setBalance(params[0].GetBaseUnit(), params[0].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters of a pbspot order. The parameters include the fees, whether the
				// order is a maker order, whether the order is a turn order, and whether the order is an equal order.
				params[0].Param = &pbspot.Order_Param{Fees: fees, Maker: false, Turn: true, Equal: false}

				// The purpose of this line of code is to calculate the quantity and fees of a given parameter. The getSum() function
				// takes in three arguments: the quote unit of the first parameter, the value of the second parameter, and a boolean
				// value that indicates if the quantity should be rounded up. The function then returns two values: the quantity and the fees associated with the parameter.
				quantity, fees = e.getSum(params[0].GetQuoteUnit(), decimal.New(params[0].GetValue()).Mul(params[0].GetPrice()).Float(), true)

				// The purpose of this code is to set the balance of a user with the specified quantity and type. The function checks
				// for any errors, and if an error is encountered, the function returns.
				if err := e.setBalance(params[0].GetQuoteUnit(), params[1].GetUserId(), quantity, pbspot.Balance_PLUS); err != nil {
					return
				}

				// The purpose of this code is to set the parameters for a pbspot order. It sets the fees, whether the order is a
				// maker order, whether it has a turn and whether it is equal.
				params[1].Param = &pbspot.Order_Param{Fees: fees, Maker: true, Turn: false, Equal: false}

				break
			}

			e.Context.Logger.Infof("[(B)%v]: (Assigning: %v), (Price: [%v]-[%v]), (Value: [%v]-[%v]), (UserID: [%v]-[%v])", params[0].GetAssigning(), params[1].GetAssigning(), params[0].GetPrice(), params[1].GetPrice(), params[0].GetValue(), params[1].GetValue(), params[0].GetUserId(), params[1].GetUserId())
		}

	}

	if err := e.setTrade(params[0], params[1]); err != nil {
		return
	}
}

// replayDeposit - The purpose of this code is to replay deposits on different chains. It retrieves details of the chain from the
// database and depending on the platform (Ethereum or Tron) it calls the depositEthereum or depositTron functions. After
// that it sleeps for 500 milliseconds and replays the confirmation deposits.
func (e *Service) replayDeposit() {

	// e.run, e.wait, and e.block are all maps in the program. The purpose of these maps is to store boolean, boolean, and
	// int64 values respectively. These values can be referenced and modified by their associated key which is an int64
	// value. The maps allow the program to store and access the values quickly and easily.
	e.run, e.wait, e.block = make(map[int64]bool), make(map[int64]bool), make(map[int64]int64)

	for {

		func() {

			// The purpose of this code is to declare a variable called 'chain' of type 'pbspot.Chain'. This is known as a
			// declaration statement, which is used to declare a variable in a program. The variable can then be used to store
			// data like a string, an integer, or any other type of data.
			var (
				chain pbspot.Chain
			)

			// This code is querying the chains table in a database and returning the id, rpc, platform, block, network,
			// confirmation and parent_symbol fields from each row where the status field is true. The purpose of this code is to
			// query the database for records with a true status and get the associated fields for each. The Context.Debug()
			// function is used to check for errors, and the defer rows.Close() statement is used to close the rows object when the function is complete.
			rows, err := e.Context.Db.Query("select id, rpc, platform, block, network, confirmation, parent_symbol from chains where status = $1", true)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for rows.Next() loop is used to iterate through the rows of a query result in a database. It is typically used
			// with a SQL query that has been prepared and executed, and the result set is stored in a Rows object. The loop will
			// iterate over each row, and can be used to access and process the data in each row.
			for rows.Next() {

				// This code snippet is checking for an error while scanning the row of data and continuing if there is an error. The
				// purpose of the if statement is to ensure that the data is scanned correctly and that the program can continue if
				// there is an error.
				if err := rows.Scan(&chain.Id, &chain.Rpc, &chain.Platform, &chain.Block, &chain.Network, &chain.Confirmation, &chain.ParentSymbol); e.Context.Debug(err) {
					continue
				}

				// This code is setting the value of the variable "chain.Block" to "1" if the value of "chain.GetBlock()" is
				// currently "0". This is likely to be used to initialize the value of the "chain.Block" variable to a known value if
				// it is currently not set.
				if chain.GetBlock() == 0 {
					chain.Block = 1
				}

				// This if block is used to check if the block in the chain exists in the e.block map and if it is equal to the
				// chain's GetBlock() method. If either of these conditions are not met, the loop will continue.
				if block, ok := e.block[chain.GetId()]; !ok && block == chain.GetBlock() {
					continue
				}

				// This code is checking to see if a given chain is running. If it is running, it will set the wait value for that
				// chain to false. If it is not running, it will set the run value for that chain to true.
				if e.run[chain.GetId()] {

					//This statement is used to check if a particular key, in this case chain.GetId(), exists in the map e.wait. If the
					//key does not exist, the loop will continue to the next iteration.
					if _, ok := e.wait[chain.GetId()]; !ok {
						continue
					}

					e.wait[chain.GetId()] = false
				} else {
					e.run[chain.GetId()] = true
				}

				// This switch statement is used to differentiate between two different blockchain platforms, Ethereum and Tron. It
				// will allow the code to take different actions depending on which platform the chain is connected to.
				switch chain.GetPlatform() {
				case pbspot.Platform_ETHEREUM:

					// The purpose of this statement is to deposit Ethereum into a blockchain. It is used to send the Ethereum to the
					// chain and to store it securely.
					go e.depositEthereum(&chain)
					break
				case pbspot.Platform_TRON:

					// The purpose of this code is to deposit Tron (a cryptocurrency) on a blockchain platform. It is used to transfer
					// funds from one account to another and keep a record of the transaction on the blockchain.
					go e.depositTron(&chain)
					break
				}

				time.Sleep(500 * time.Millisecond)
			}

			// Confirmation deposits assets - The e.replayConfirmation() function is used to confirm that a replay has been recorded and saved. It is typically
			// used to ensure that a replay can be accessed and replayed later.
			e.replayConfirmation()

		}()
	}

}

// replayWithdraw - This function is used to replay pending withdraw transactions. It checks for transactions with a status of pending, a
// transaction type of withdraws, and a financial type of crypto in the database. It then loops through these
// transactions and attempts to transfer the funds. It also handles cases where there are fees to be paid, by attempting
// to transfer funds from a reserve asset with the same platform, symbol, and protocol. It is repeated every 10 seconds.
func (e *Service) replayWithdraw() {

	// The purpose of this code is to handle a panic and recover gracefully. To defer keyword will execute the following
	// code whenever the function it is contained in ends, even if the function ends in error. The recover() function is
	// used to catch and handle any panic that may have occurred in the function. If a panic is caught, the code will call
	// e.Context.Debug(r) to output the panic information, and then return.
	defer func() {
		if r := recover(); e.Context.Debug(r) {
			return
		}
	}()

	for {

		func() {

			// This code is querying a database for transactions with specific parameters. The code uses the sql Query method to
			// query the database, passing in the parameters as variables. The query will return rows, which are stored in the
			// rows variable. The error from the query is stored in the err variable, and an error is printed out if err is not
			// nil. The rows returned by the query are then closed when the function is finished executing.
			rows, err := e.Context.Db.Query(`select id, symbol, "to", chain_id, fees, value, price, platform, protocol, claim from transactions where status = $1 and tx_type = $2 and fin_type = $3`, pbspot.Status_PENDING, pbspot.TxType_WITHDRAWS, pbspot.FinType_CRYPTO)
			if e.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for loop with rows.Next() is used to loop through the rows of a result set from a database query. The .Next()
			// method advances the cursor to the next row and returns true if there is another row, or false if there are no more
			// rows. The for loop will continue looping through the result set until the .Next() method returns false.
			for rows.Next() {

				// The purpose of the code is to declare three variables, item, reserve, and insure, of type pbspot.Transaction. This
				// allows the program to use those three variables to interact with the pbspot.Transaction type.
				var (
					item, reserve, insure pbspot.Transaction
				)

				// This code is used to scan a row of data from a database and store each of the values in variables. The if
				// statement checks for an error while scanning and logs the error with the context.Debug() method. If an error
				// occurs, the loop will continue, otherwise the values are stored in the variables.
				if err := rows.Scan(&item.Id, &item.Symbol, &item.To, &item.ChainId, &item.Fees, &item.Value, &item.Price, &item.Platform, &item.Protocol, &item.Claim); e.Context.Debug(err) {
					return
				}

				// This code is setting up a chain and checking for errors. If an error is encountered, the code will continue on
				// without executing the rest of the code. This allows the code to continue running in the event of an error.
				chain, err := e.getChain(item.GetChainId(), true)
				if e.Context.Debug(err) {
					return
				}

				// This if statement is used to check if the item's protocol is set to mainnet. Mainnet is the original and most
				// widely used network for transactions to take place on. If the item's protocol is set to mainnet, then the code
				// inside the if statement will execute.
				if item.GetProtocol() == pbspot.Protocol_MAINNET {

					// Find the reserve asset from which funds will be transferred, by its platform, as well as by protocol, symbol, and number of funds.
					// This code is checking to see if the query returns a row with a value greater than 0. The query is looking for a
					// specific combination of values in the reserves table that match the item values passed in. The code is searching
					// for a row with a value greater than 0 and if one is found, it stores the value and user_id in the reserve object.
					if _ = e.Context.Db.QueryRow("select value, user_id from reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), false).Scan(&reserve.Value, &reserve.UserId); reserve.GetValue() > 0 {

						// This code is updating the status of a transaction in a database based on the item ID. The if statement is used
						// to check for any errors that occur during the update process. If an error is found, the loop will continue
						// without executing any further code.
						if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), pbspot.Status_PROCESSING); e.Context.Debug(err) {
							return
						}

						// This code is part of a loop that is looping through a list of items. The purpose of this code is to set a
						// reserve lock for each item in the list. If it is successful, the loop will continue to the next item. If there
						// is an error, the loop will skip the current item and move on to the next one.
						if err := e.setReserveLock(reserve.GetUserId(), item.GetSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
							return
						}

						//This switch statement is used to determine which platform the item is associated with. Depending on whether the
						//item is associated with Ethereum or Tron, different operations can be performed.
						switch item.GetPlatform() {
						case pbspot.Platform_ETHEREUM:

							// The purpose of this code is to transfer Ethereum from a user's account to a designated recipient. It takes the
							// user's ID, the item's ID, symbol, recipient, value, protocol, and blockchain as parameters. The 'true'
							// parameter indicates that the transaction should be executed immediately.
							e.transferEthereum(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), 0, item.GetProtocol(), chain, true)
							break
						case pbspot.Platform_TRON:

							// This code is used to transfer Tron from one user's account to another user's account. The parameters passed in
							// this function are the user IDs of the sender and receiver, the item ID, symbol, to address, value, protocol,
							// and chain. The last parameter, 'true', indicates that this transfer is enabled.
							e.transferTron(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), 0, item.GetProtocol(), chain, true)
							break
						}

					}

				} else {

					// Find the reserve asset from which funds will be transferred,
					// by its platform, as well as by protocol, symbol, and number of funds.
					// This code is part of a transaction process. The purpose of the code is to find funds in a reserve asset to use
					// for a transaction, and to find funds in a reserve asset to use for a fee. If the fee is not found, the transaction is reversed. The code is also responsible for setting locks on the funds in the reserve asset to prevent them from being used for another transaction.
					if _ = e.Context.Db.QueryRow("select a.value, a.user_id from reserves a inner join reserves b on case when a.protocol > 0 then b.user_id = a.user_id and b.symbol = $6 and b.platform = a.platform and b.protocol = 0 and b.value >= $5 and b.lock = $7 end where a.symbol = $1 and a.value >= $2 and a.platform = $3 and a.protocol = $4 and a.lock = $7", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), item.GetFees(), chain.GetParentSymbol(), false).Scan(&reserve.Value, &reserve.UserId); reserve.GetValue() > 0 {

						// This code is updating the status of a specific transaction in the database. The if statement is used to check
						// for any errors that may occur when executing the update command and continue the loop if an error is found.
						if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), pbspot.Status_PROCESSING); e.Context.Debug(err) {
							return
						}

						// This code is checking for an error when setting a reserve lock. If an error is found, the code continues without
						// taking any action. This is usually done to prevent the code from crashing due to an unexpected error.
						if err := e.setReserveLock(reserve.GetUserId(), item.GetSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
							return
						}

						// This switch statement is used to check the platform of a given item. Depending on which platform the item is on
						// (Ethereum or Tron), the appropriate action will be taken.
						switch item.GetPlatform() {
						case pbspot.Platform_ETHEREUM:

							// This function is used to transfer Ethereum from one user to another. The parameters include user ID, item ID,
							// item symbol, receiver address, value, price, protocol, and a boolean indicating whether the transaction is
							// confirmed or not. This function allows for the transfer of Ethereum between users, enabling them to purchase goods and services on the blockchain.
							e.transferEthereum(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), item.GetPrice(), item.GetProtocol(), chain, true)
							break
						case pbspot.Platform_TRON:

							// This is a function used to transfer Tron from one user to another, usually as part of a transaction. The
							// parameters represent the user ID, item ID, item symbol, recipient address, value, price, protocol (Tron or
							// another blockchain), and the chain code, respectively. The last parameter is a boolean value that specifies whether the transaction should be broadcast or not.
							e.transferTron(reserve.GetUserId(), item.GetId(), item.GetSymbol(), item.GetTo(), item.GetValue(), item.GetPrice(), item.GetProtocol(), chain, true)
							break
						}

					} else {

						//This code block is used to check if a transaction can be reversed. If the fees for the transaction are greater
						//than or equal to two times the amount of the transaction's fees, the code will attempt to find a reserve asset
						//from which funds can be transferred. If it finds a reserve asset, it will then initiate a transfer of funds from
						//the reserve asset to the address associated with the transaction. Finally, the code will set the "claim" field of the transaction to true.
						if !item.GetClaim() {

							// This code is attempting to retrieve the currency for a given parent symbol, and is handling the case where an
							// error occurs when attempting to get the currency. If the call to getCurrency() returns an error, the function
							// returns without attempting to do anything else.
							currency, err := e.getCurrency(chain.GetParentSymbol(), false)
							if err != nil {
								return
							}

							// This code is part of a larger program, and its purpose is to transfer funds from a reserve asset to pay fees
							// associated with a transaction. It checks if the currency has fees that are greater than or equal to double the
							// item's fees, and if so, it queries the reserves to find the correct asset, platform, protocol, symbol, and
							// value of the funds to be transferred. It then sets the asset's lock to true, and uses either the Ethereum or Tron transfer functions to move the funds. Finally, it updates the transaction's claim to true.
							if currency.GetFeesCharges() >= item.GetFees()*2 {

								// Find the reserve asset from which funds will be transferred,
								// by its platform, as well as by protocol, symbol, and number of funds.
								// This code is likely part of a larger program that is transferring money from one user to another.
								// Specifically, it is attempting to find enough funds in a reserve to cover the cost of the transaction, and if
								// enough funds are found, it will transfer the money to the specified user using either Ethereum or Tron, depending on the platform specified. It will then update the transaction to set the "claim" field to true.
								if _ = e.Context.Db.QueryRow("select value, address from reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), false).Scan(&reserve.Value, &reserve.To); reserve.GetValue() > 0 {

									// The purpose of this code is to set a reserve lock on a user's chain, platform, and protocol, transfer funds
									// depending on the platform of the item, and update the 'claim' column of a row in the 'transactions' table to
									// 'true'. If any errors are encountered while performing these actions, the code will skip the current
									// iteration of the loop it is in and continue looping.
									if _ = e.Context.Db.QueryRow("select value, user_id from reserves where symbol = $1 and value >= $2 and platform = $3 and protocol = $4 and lock = $5", chain.GetParentSymbol(), item.GetFees()*2, item.GetPlatform(), pbspot.Protocol_MAINNET, false).Scan(&insure.Value, &insure.UserId); insure.GetValue() > 0 {

										// This code is setting a reserve lock on a user's chain, platform, and protocol. If an error is encountered
										// while setting the lock, the code will continue without further action.
										if err := e.setReserveLock(insure.GetUserId(), chain.GetParentSymbol(), item.GetPlatform(), item.GetProtocol()); e.Context.Debug(err) {
											return
										}

										// The purpose of this switch statement is to check the platform of the item.
										// If the item is on the Ethereum platform, then the code within the first case statement will be executed.
										// If the item is on the Tron platform, then the code within the second case statement will be executed.
										switch item.GetPlatform() {
										case pbspot.Platform_ETHEREUM:
											e.transferEthereum(insure.GetUserId(), item.GetId(), chain.GetParentSymbol(), reserve.GetTo(), item.GetFees(), 0, pbspot.Protocol_MAINNET, chain, false)
											break
										case pbspot.Platform_TRON:
											e.transferTron(insure.GetUserId(), item.GetId(), chain.GetParentSymbol(), reserve.GetTo(), item.GetFees(), 0, pbspot.Protocol_MAINNET, chain, false)
											break
										}

										// This code is updating a row in a database table called 'transactions' by setting the 'claim' column to
										// 'true' for a specific row identified by the 'id' column. The 'if' statement is checking for any errors that
										// occur during the update process. If an error is detected, the code will skip the current iteration of the loop it is in and continue looping.
										if _, err := e.Context.Db.Exec("update transactions set claim = $2 where id = $1;", item.GetId(), true); e.Context.Debug(err) {
											return
										}
									}
								}

							} else {
								e.Context.Logger.Info("[REVERSE]: there are no fees to pay the reverse fee")
							}
						}
					}
				}
			}
		}()

		// The time.Sleep() function is used to pause the current goroutine for a specified duration. In this case, the
		// goroutine would be paused for 10 seconds.
		time.Sleep(10 * time.Second)
	}

}

// replayConfirmation - This function is used to check the status of pending deposits. It queries the database for transactions with a status
// of PENDING and tx type of DEPOSIT. It then checks the status of the hash associated with the transaction on the
// relevant blockchain. If the status is successful, the deposit is credited to the local wallet address and the status
// of the transaction is changed to FILLED. If the status is unsuccessful, the status is changed to FAILED. If the number
// of confirmations is not yet met, the number of confirmations is updated in the database.
func (e *Service) replayConfirmation() {

	// This code is performing a SQL query to select information from a database. The purpose is to select a specific set of
	// information from the database based on the parameters of the query. The query is selecting the fields id, hash,
	// symbol, "to", fees, chain_id, user_id, value, confirmation, block, platform, protocol, and create_at where the status
	// is equal to pbspot.Status_PENDING and the tx_type is equal to pbspot.TxType_DEPOSIT. The code also checks for an error and closes the rows when finished.
	rows, err := e.Context.Db.Query(`select id, hash, symbol, "to", fees, chain_id, user_id, value, confirmation, block, platform, protocol, create_at from transactions where status = $1 and tx_type = $2`, pbspot.Status_PENDING, pbspot.TxType_DEPOSIT)
	if e.Context.Debug(err) {
		return
	}
	defer rows.Close()

	// The purpose of the for loop is to iterate through each row of a result set from a database query. The rows.Next()
	// function is used to move to the next row in the result set.
	for rows.Next() {

		// The above code is declaring a variable called "item" of type "pbspot.Transaction". This means that the variable
		// "item" will be used to store information related to a pbspot transaction.
		var (
			item pbspot.Transaction
		)

		// This code is part of a loop that is iterating over results from a database query. The purpose of the code is to scan
		// each row of the query result into their corresponding variables. If an error is encountered while scanning, the loop
		// continues to the next row. The e.Context.Debug() function logs the error but does not cause the program to stop.
		if err := rows.Scan(&item.Id, &item.Hash, &item.Symbol, &item.To, &item.Fees, &item.ChainId, &item.UserId, &item.Value, &item.Confirmation, &item.Block, &item.Platform, &item.Protocol, &item.CreateAt); e.Context.Debug(err) {
			return
		}

		// The purpose of this code is to get a chain from the "e" object, using the item's chain ID. If an error occurs, the
		// function will return, and the error will be printed if debugging is enabled.
		chain, err := e.getChain(item.GetChainId(), true)
		if e.Context.Debug(err) {
			return
		}

		// This code is used to connect to a blockchain using the GetRpc and GetPlatform methods of the chain object. The
		// client object is used to make requests to the blockchain and the err object will be used to check for any errors
		// that occurred in the process. If an error is encountered, the code will return to avoid any further issues.
		client, err := blockchain.Dial(chain.GetRpc(), chain.GetPlatform())
		if e.Context.Debug(err) {
			return
		}

		// This code is part of a deposit process. The purpose of this code is to check a deposit's status, which is tracked
		// using the client.Status(item.Hash) function. If the deposit is confirmed, the code credits the new deposit to the
		// local wallet address, updates the deposits pending status to success status, and publishes the status to the
		// exchange. If the deposit is not confirmed, it updates the confirmation number in the database. If the deposit fails, it updates the status in the database and publishes the status to the exchange.
		if client.Status(item.Hash) {

			// The purpose of this code is to check if the difference between the current block and the item block is greater than
			// or equal to the confirmation number of the chain and if the item confirmation is greater than or equal to the
			// chain's confirmation number. If both conditions are true, then the subsequent code will execute.
			if (chain.GetBlock()-item.GetBlock()) >= chain.GetConfirmation() && item.GetConfirmation() >= chain.GetConfirmation() {

				// This code is checking if there was an error when calling the getCurrency() function. If there was an error, it
				// will be printed out using the Debug() function and the function will return.
				currency, err := e.getCurrency(item.GetSymbol(), false)
				if e.Context.Debug(err) {
					return
				}

				// The purpose of the if statement is to check if the value of the item is greater than or equal to the minimum
				// deposit amount for the currency. If the condition evaluates to true, then the code inside the statement will execute.
				if item.GetValue() >= currency.GetMinDeposit() {

					// Crediting a new deposit to the local wallet address.
					// This code is updating the balance of an asset with a given symbol and user ID. The purpose is to update the
					// balance with a given value (item.GetValue()) for the user and symbol combination. The code is using the Exec
					// function on the database object and passing in the appropriate values. If there is an error, the code continues.
					if _, err := e.Context.Db.Exec("update assets set balance = balance + $3 where symbol = $2 and user_id = $1;", item.GetUserId(), item.GetSymbol(), item.GetValue()); e.Context.Debug(err) {
						return
					}

					// The purpose of the item.Hook = true statement is to indicate that the item is hooked in a certain way. This can
					// be used to indicate whether an item has been connected to another item, or has been "hooked" in some way.
					// The item.Status = pbspot.Status_FILLED statement is used to set the status of the item to "FILLED". This indicates that the item has been filled with the necessary information, such as the item's name, description, or other details.
					item.Hook = true
					item.Status = pbspot.Status_FILLED

					// This code is from a function that is publishing a message to an exchange with a certain routing key.  The purpose
					// of this code is to attempt to publish the message to the exchange.  If an error is encountered, the context debug
					// method is called with the error and the function returns.
					if err := e.Context.Publish(&item, "exchange", "deposit/open", "deposit/status"); e.Context.Debug(err) {
						return
					}
				} else {

					// The purpose of this line of code is to assign a status to an item. In this case, the status is set to the
					// constant 'Status_RESERVE', which is a pre-defined constant. This constant is likely used to denote that the item
					// has been reserved or set aside for a specific purpose.
					item.Status = pbspot.Status_RESERVE
				}

				// The purpose of this code is to set a reserve for a specified user, symbol, value, platform, and protocol. If an
				// error occurs, the code will continue to execute. The e.Context.Debug(err) line logs the error for debugging purposes.
				if err := e.setReserve(item.GetUserId(), item.GetTo(), item.GetSymbol(), item.GetValue(), item.GetPlatform(), item.GetProtocol(), pbspot.Balance_PLUS); e.Context.Debug(err) {
					return
				}

				// This code is part of a loop, and it is used to update the status of a transaction in a database. The first two
				// arguments in the Exec() function are the ID and status of the transaction. The third argument is a function that
				// will debug any error that may occur during execution. If an error occurs, the code will skip the current iteration of the loop and continue on to the next one.
				if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
					return
				}

			} else {

				// This code is updating the 'confirmation' column of the 'transactions' table with the difference between the
				// current block and the block of the item. The purpose of this code is to track the number of blocks that have
				// passed since the transaction was confirmed. If an error is encountered, the code will continue without halting.
				if _, err := e.Context.Db.Exec("update transactions set confirmation = $2 where id = $1;", item.GetId(), chain.GetBlock()-item.GetBlock()); e.Context.Debug(err) {
					return
				}
			}

		} else {

			// The item.Hook = true statement is used to indicate that an item has been hooked, meaning that it has been linked or
			// attached to something else. The item.Status = pbspot.Status_FAILED statement is used to set the status of the item
			// to "Failed", which indicates that the item has not been successful in performing its intended task.
			item.Hook = true
			item.Status = pbspot.Status_FAILED

			// This statement is an example of an if statement that is used to update a database record with a specific status.
			// The if statement checks for an error, and if one is found, the loop will continue. The purpose of this statement is
			// to ensure that the database is updated without any errors.
			if _, err := e.Context.Db.Exec("update transactions set status = $2 where id = $1;", item.GetId(), item.GetStatus()); e.Context.Debug(err) {
				return
			}

			// This code is checking for an error when publishing an item to an exchange. The exchange is specified as "exchange"
			// and the routing keys are "deposit/open" and "deposit/status". If an error occurs, it is logged and the function returns.
			if err := e.Context.Publish(&item, "exchange", "deposit/open", "deposit/status"); e.Context.Debug(err) {
				return
			}
		}
	}
}
