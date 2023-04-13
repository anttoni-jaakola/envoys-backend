package index

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbindex"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/types"
	"strings"
)

// GetStatistic - This function is used to get statistics from the database and return them in the form of a pbindex.ResponseStatistic.
// It scans the database for the values of accounts, pairs, chains, currencies, transactions and orders, and then
// calculates the values of reserves using the prices of pairs. It also converts the values of charged reserves to USD.
func (i *Service) GetStatistic(_ context.Context, _ *pbindex.GetRequestStatistic) (*pbindex.ResponseStatistic, error) {

	// The purpose of the variables above is to define all the necessary fields for a response statistic in a typescol
	// Buffers Index. Each of the variables defines a different field within the ResponseStatistic typescol buffer,
	// including the Account, Pair, Chain, Currency, Transaction, and Order fields. This will define the data structure for
	// the ResponseStatistic typescol buffer.
	var (
		response    pbindex.ResponseStatistic
		statistic   pbindex.Statistic
		account     pbindex.Statistic_Account
		pair        pbindex.Statistic_Pair
		chain       pbindex.Statistic_Chain
		currency    pbindex.Statistic_Currency
		transaction pbindex.Statistic_Transaction
		order       pbindex.Statistic_Order
	)

	// The purpose of the code is to count the number of accounts with a status of 'true' and 'false' and store the values
	// in the 'account.Enable' and 'account.Disable' variables.
	_ = i.Context.Db.QueryRow("select count(*) from accounts where status = $1", true).Scan(&account.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from accounts where status = $1", false).Scan(&account.Disable)

	// The purpose of the code is to query a database and count the number of records in the "pairs" table with a status of
	// true and false. It then stores the results in the "pair.Enable" and "pair.Disable" variables.
	_ = i.Context.Db.QueryRow("select count(*) from pairs where status = $1", true).Scan(&pair.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from pairs where status = $1", false).Scan(&pair.Disable)

	// The purpose of the code snippet is to query the database for the number of chains with status set to true and false,
	// respectively. It then stores the result of these queries in the variables chain.Enable and chain.Disable, respectively.
	_ = i.Context.Db.QueryRow("select count(*) from chains where status = $1", true).Scan(&chain.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from chains where status = $1", false).Scan(&chain.Disable)

	// The purpose of this code is to query a database for the number of currencies with a status of either true or false,
	// and assign the result to the variables currency.Enable and currency.Disable.
	_ = i.Context.Db.QueryRow("select count(*) from assets where status = $1", true).Scan(&currency.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from assets where status = $1", false).Scan(&currency.Disable)

	// The purpose of the code is to query the database for the number of transactions with a status of FILLED and PENDING
	// and store the results in the variables "transaction.Filled" and "transaction.Pending".
	_ = i.Context.Db.QueryRow("select count(*) from transactions where status = $1", types.StatusFilled).Scan(&transaction.Filled)
	_ = i.Context.Db.QueryRow("select count(*) from transactions where status = $1", types.StatusPending).Scan(&transaction.Pending)

	// The purpose of the two queries is to count the number of orders with a specific assigning and status. The assigning
	// and status are passed in as parameters. The result of the query is stored in the order.Sell and order.Buy variables.
	_ = i.Context.Db.QueryRow("select count(*) from orders where assigning = $1 and status = $2", types.AssigningSell, types.StatusPending).Scan(&order.Sell)
	_ = i.Context.Db.QueryRow("select count(*) from orders where assigning = $1 and status = $2", types.AssigningBuy, types.StatusPending).Scan(&order.Buy)

	// The purpose of this code is to assign values to the fields of the statistic struct. The accounts, pairs, chains,
	// currencies, transactions, and orders fields are assigned the memory address of the variables account, pair, chain, currency, transaction, and order, respectively.
	statistic.Accounts = &account
	statistic.Pairs = &pair
	statistic.Chains = &chain
	statistic.Currencies = &currency
	statistic.Transactions = &transaction
	statistic.Orders = &order

	// This code is used to query the database for a specific set of data. The query searches for the three fields symbol,
	// fees_charges, and fees_costs from the currencies table. The query is executed using the i.Context.Db.Query()
	// function, and the result is returned in the rows variable. If an error occurs during the query, the err variable is
	// populated with the error. The defer statement ensures
	// that the rows.Close() function is called after the query is complete, which closes the connection to the database.
	rows, err := i.Context.Db.Query("select symbol, fees_charges, fees_costs from assets")
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// This loop is used to iterate through a set of rows in a relational database. The loop will execute the code inside of
	// it for each row in the set of rows, until it reaches the end of the set.
	for rows.Next() {

		// The purpose of the above code is to declare two variables, reserve and price, which store the value of
		// pbindex.Statistic_Reserve and a float64 data type, respectively.
		var (
			reserve pbindex.Statistic_Reserve
			price   float64
		)

		// This code is checking for a potential error while scanning the rows of a database query. If an error is found, the
		// code is returning an error message and a response.
		if err := rows.Scan(&reserve.Symbol, &reserve.ValueCharged, &reserve.ValueCosts); err != nil {
			return &response, err
		}

		// This is a conditional statement checking to see if the sum of the 'value' field from the 'reserves' table is greater
		// than 0, where the 'symbol' field is equal to the value stored in the 'reserve.Symbol' variable. If the sum is
		// greater than 0, the code inside the curly braces will be executed.
		if _ = i.Context.Db.QueryRow("select sum(value) from reserves where symbol = $1", reserve.Symbol).Scan(&reserve.Value); reserve.Value > 0 {

			// The purpose of this code is to check if the symbol stored in the reserve variable is equal to "usdt". If the symbol
			// is not equal to "usdt", then this code will execute a specific action.
			if reserve.Symbol != "usdt" {

				// This code is attempting to query a database for the price of a particular currency pair. The query is looking for
				// the price of the currency pair where the base unit is the reserve.Symbol and the quote unit is "usd". If an error
				// occurs while attempting to query the database, an error is returned.
				if err := i.Context.Db.QueryRow("select price from pairs where base_unit = $1 and quote_unit = $2", reserve.Symbol, "usd").Scan(&price); err != nil {
					return &response, err
				}

				// This code is used to convert the value of a reserve (reserve.ValueCharged) to a decimal value, multiplied by a
				// given price (price), and then store the resulting float in the reserve.ValueChargedConvert variable.
				reserve.ValueChargedConvert = decimal.New(reserve.ValueCharged).Mul(price).Float()
			}
		}

		// The purpose of this line of code is to add a new item to the end of the existing slice of Reserves in the statistic
		// object. The new item added is the address of the reserve object.
		statistic.Reserves = append(statistic.Reserves, &reserve)
	}

	//The purpose of this statement is to set the Fields property of the response object to the Fields property of the statistic object. This allows the response object to access and modify the fields of the statistic object.
	response.Fields = &statistic

	return &response, nil
}

// GetMarkets - This function is a method of the Service struct, and its purpose is to get a list of pairs from the database and
// return them in the form of a ResponseMarket. It also retrieves additional information such as the ratio, high, low, and
// volume of each pair, as well as a list of closing prices for the candles. It also takes in a context and a
// GetRequestMarkets object as parameters, to limit the number of pairs returned and to filter the search.
func (i *Service) GetMarkets(ctx context.Context, req *pbindex.GetRequestMarkets) (*pbindex.ResponseMarket, error) {

	// The purpose of this code is to declare two variables, response and maps, both of which are of type
	// pbindex.ResponseMarket and []string, respectively.
	var (
		response pbindex.ResponseMarket
		maps     []string
	)

	//The purpose of this code is to create a new API client for the pbprovider package using the existing gRPC client in the context.
	migrate := pbprovider.NewApiClient(i.Context.GrpcClient)

	// This if statement serves to set a default value of 30 for the req.Limit variable if the req.GetLimit() method returns
	// a value of 0. This default value may be used in cases where the user has not specified a value for the req.Limit
	// variable.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This if statement checks if the request has a search query. If it does, it adds a WHERE clause to the SQL query that
	// looks for any base units or quote units that match the search query. The WHERE clause is appended to a list of other
	// clauses called 'maps'.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where base_unit like %[1]s or quote_unit like %[1]s", "'%"+strings.ToLower(req.GetSearch())+"%'"))
	}

	// This code is checking if the query returns any rows. The row count is stored in the variable response.Count, and the
	// variable response.GetCount() is then compared to 0. If the count is greater than 0, the code will execute whatever is
	// inside the if statement.
	if _ = i.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from pairs %s", strings.Join(maps, " "))).Scan(&response.Count); response.GetCount() > 0 {

		// This code is used to calculate the offset for a paginated request. The offset is used to indicate where the query
		// should start, in order to return the desired number of records. If the page is greater than 0, then the offset
		// should be the limit multiplied by the page minus 1. Otherwise, the offset should be the limit multiplied by the page.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is querying the database for specific data from the "pairs" table. The query is using a limit and an
		// offset to return a specific range, and it is ordered by the "status" column. The "rows" variable will contain the
		// returned data. If there is an error it will be handled in the "if err" statement and the response will be returned with the error.
		rows, err := i.Context.Db.Query(fmt.Sprintf("select id, base_unit, quote_unit, price, status from pairs %s order by status desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through the rows of a database table. The rows.Next() statement is used to access
		// the next row in the table and will execute the loop until the last row is reached. This allows the programmer to
		// access and manipulate the data in each row.
		for rows.Next() {

			// The above code declares a variable 'pair' of type pbindex.Market. This variable is used to store a pair of values
			// which is commonly used in programming. It can be used to store two related values such as a key-value pair.
			var (
				item pbindex.Market
			)

			// This code is used to scan the rows of a database query result and assigns the values to the variables in the order
			// they are specified. If any errors occur during the scanning, it returns an error.
			if err := rows.Scan(&item.Id, &item.BaseUnit, &item.QuoteUnit, &item.Price, &item.Status); err != nil {
				return &response, err
			}

			// This code is attempting to get the current price of a currency pair (pair) and set the pair.Ratio value to that
			// price. If there is an error encountered while getting the price (i.queryPrice(pair.GetBaseUnit(),
			// pair.GetQuoteUnit())), the function will return an error.
			item.Ratio, err = i.queryPrice(item.GetBaseUnit(), item.GetQuoteUnit())
			if err != nil {
				return &response, err
			}

			// This code is retrieving the latest 50 candles for a specific trading pair. The request is sent to the spot exchange
			// and if the request is successful, the candles are returned. If there is an error, it is returned with the err variable.
			ticker, err := migrate.GetTicker(ctx, &pbprovider.GetRequestTicker{
				Limit:     50,
				BaseUnit:  item.GetBaseUnit(),
				QuoteUnit: item.GetQuoteUnit(),
			})
			if err != nil {
				return &response, err
			}

			// This code is setting the values of the High, Low, and Volume properties of the 'pair' object to the corresponding
			// values from the 'candles' object's Stats object. This is likely being done to save the corresponding values from
			// the 'candles' object in the 'pair' object for later use.
			item.High = ticker.Stats.GetHigh()
			item.Low = ticker.Stats.GetLow()
			item.Volume = ticker.Stats.GetVolume()

			// This code is looping through an array of candles and appending each candle's close value to a pair.Ticker array.
			// The purpose is to add the closing values of each candle to a list of candles for a given pair.
			for i := 0; i < len(ticker.Fields); i++ {
				item.Ticker = append(item.Ticker, ticker.Fields[i].Close)
			}

			// This code appends a new field to the response object. The &pair parameter is added to the end of the
			// response.Fields array. This can be used to add additional information to the response object, such as a new field or value.
			response.Fields = append(response.Fields, &item)
		}

	}

	return &response, nil
}
