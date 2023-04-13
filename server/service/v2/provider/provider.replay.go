package provider

import (
	"context"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/svarlamov/goyhfin"
	"strings"
	"time"
)

// market - This function is used to replay market prices. The function is executed on a specific time interval and retrieves data
// from the database. It then inserts the data into the trades table, and publishes the data to exchange topics. This
// allows for the market data to be replayed at a specific interval.
func (a *Service) market() {

	// The code creates a ticker that triggers every minute and runs a loop that executes each time the ticker is triggered.
	// This allows code to be executed at regular intervals without the need for an explicit loop.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code allows the program to query a database and retrieve the values of the 'id', 'price', 'base_unit', and
			// 'quote_unit' columns from the 'pairs' table, where the 'status' column is equal to 'true'. The code then closes the
			// rows when the query is complete.
			rows, err := a.Context.Db.Query(`select id, price, base_unit, quote_unit from pairs where status = $1 order by id`, true)
			if a.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for rows.Next() loop is used to iterate over each row in a database result set. It allows you to access each
			// row one at a time, until all rows have been processed. This is useful for processing large result sets without
			// loading them all into memory at once.
			for rows.Next() {

				// This is a variable declaration statement. The variable 'pair' is being declared as type 'types.Pair'. This allows
				// the variable to store a pair of values (e.g. two integers, two strings, two objects, etc.).
				var (
					item types.Pair
				)

				// This code is checking for an error when scanning the rows of a database table. The if statement scans the rows of
				// the database table using the Scan() method, and if it encounters an error, it will log the error and continue
				// scanning the remaining rows.
				if err := rows.Scan(&item.Id, &item.Price, &item.BaseUnit, &item.QuoteUnit); a.Context.Debug(err) {
					continue
				}

				// This code is setting a ticker for a currency pair in order to track its price. The context.Background() is used to
				// create a basic context, the SetRequest contains the key, price, base unit, quote unit, and assigning type.
				// Finally, the e.Context.Debug(err) is used to debug any errors that may occur during the process. If an error occurs, the code will continue.
				if _, err := a.SetTicker(context.Background(), &pbprovider.SetRequestTicker{Key: a.Context.Secrets[2], Price: item.GetPrice(), BaseUnit: item.GetBaseUnit(), QuoteUnit: item.GetQuoteUnit(), Assigning: types.AssigningSupply}); a.Context.Debug(err) {
					continue
				}
			}
		}()
	}
}

// price - The purpose of this code is to update the prices of pairs at a specific time interval. It first loads the pair ids,
// prices, base units, and quote units from the pairs table where the status is active. It then gets the candles for the
// base and quote units and calculates the new price based on the data. Lastly, it updates the price of the pair in the database.
func (a *Service) price() {

	// The code above creates a new ticker that will run once a minute and then loop through each range of the ticker.C
	// channel. This is useful for running a certain task or operation on a regular interval of time.
	ticker := time.NewTicker(time.Minute * 1)
	for range ticker.C {

		func() {

			// This code queries a database for pairs with a status of true and orders them by their ID. It uses the
			// e.Context.Db.Query() function to execute a query and assigns the output of the query to the rows variable. If there
			// is an error, it calls the e.Context.Debug() function to debug the error, and if successful, it will defer the
			// rows.Close() function to close the rows after the function call.
			rows, err := a.Context.Db.Query(`select id, price, base_unit, quote_unit, type from pairs where status = $1 order by id`, true)
			if a.Context.Debug(err) {
				return
			}
			defer rows.Close()

			// The for loop with the rows.Next() statement is used to loop through a result set of an SQL query. The rows.Next()
			// statement advances the current row pointer to the next row and returns true if it was successful. This loop is used
			// to iterate through each row in the result set and perform an action on it.
			for rows.Next() {

				// This code creates two variables, pair and price, of the types.Pair and float64 respectively. This allows
				// the program to store and use data of these two types.
				var (
					pair  types.Pair
					price float64
				)

				// This code is used to scan the rows of data from a database and assign the values to variables. The if statement is
				// used to check for any errors that may occur when scanning the rows. If an error is found, the code will skip to
				// the next row. The e.Context.Debug() function is used to provide more information about what caused the error,
				// which can help with debugging.
				if err := rows.Scan(&pair.Id, &pair.Price, &pair.BaseUnit, &pair.QuoteUnit, &pair.Type); a.Context.Debug(err) {
					return
				}

				// The purpose of this code is to retrieve two candles from a given pair of base and quote units. The GetTicker()
				// function is used to retrieve the candles and the returned value is stored in the migrate variable. If an error is
				// encountered, the code will skip the iteration and continue to the next one.
				ticker, err := a.GetTicker(context.Background(), &pbprovider.GetRequestTicker{BaseUnit: pair.GetBaseUnit(), QuoteUnit: pair.GetQuoteUnit(), Limit: 2})
				if a.Context.Debug(err) {
					return
				}

				switch pair.GetType() {
				case types.TypeStock:

					// This code retrieves the closing price of a given pair from the YHFIN API and stores it in the "price" variable.
					if resp, _ := goyhfin.GetTickerData(strings.ToUpper(pair.GetBaseUnit()), goyhfin.OneMinute, goyhfin.OneMinute, false); len(resp.Quotes) > 0 {
						price = resp.Quotes[0].Close
					}

					break
				case types.TypeSpot:

					// Check if the unit price of the pair is available in the marketplace and set the price if available.
					if resp := marketplace.Price().Unit(pair.GetBaseUnit(), pair.GetQuoteUnit()); resp > 0 {
						price = resp
					}

					break
				}

				// This is an if statement that checks whether the variable price is equal to 0. If it is, the code inside the curly
				// braces will be executed. Otherwise, it will be skipped.
				if price == 0 {

					// This code is used to calculate the price of an item. It checks if to migrate.Fields array has any elements in
					// it. If it does, it takes the first element, gets its price, adds that to the price of the pair, and divides the
					// sum by 2. If the array is empty, it just returns the price of the pair.
					if len(ticker.Fields) > 0 {
						price = (ticker.Fields[0].GetPrice() + pair.GetPrice()) / 2
					} else {
						price = pair.GetPrice()
					}

				} else {

					// This code is calculating the price of an item. The purpose of the if statement is to check if there are any
					// "ticker.Fields" present. If there are, then the price is calculated by taking the average of the price, the
					// pair's price, and the price of the first field in to migrate.Fields array. If there are no migrate.Fields, then
					// the price is calculated by taking the average of the price and the pair's price.
					if len(ticker.Fields) > 0 {
						price = (price + pair.GetPrice() + ticker.Fields[0].GetPrice()) / 3
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

				// This code is attempting to update a row in the database table "pairs" with the given values. The if statement is
				// checking to see if there is an error and if there is, the code will continue without changing the values.
				if _, err := a.Context.Db.Exec("update pairs set price = $3 where base_unit = $1 and quote_unit = $2;", pair.GetBaseUnit(), pair.GetQuoteUnit(), price); a.Context.Debug(err) {
					return
				}
			}
		}()
	}
}
