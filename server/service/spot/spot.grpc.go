package spot

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/server/proto/pbasset"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/service/account"
	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

// GetSymbol - This function is used to get the symbol of a given currency pair (base unit and quote unit). It first checks if the
// base and quote currency exist in the database, then it checks if the pair exists in the database. If both checks pass,
// the success field of response is set to true and the response is returned. If either of the checks fail, an error is returned.
func (e *Service) GetSymbol(_ context.Context, req *pbspot.GetRequestSymbol) (*pbspot.ResponseSymbol, error) {

	// The purpose of this code is to declare two variables, response of type pbspot.ResponseSymbol and exist of type bool.
	// The response variable will store the response from a service that provides stock symbol information, while the exist
	// boolean will keep track of whether the symbol exists or not.
	var (
		response pbspot.ResponseSymbol
		exist    bool
	)

	// This piece of code checks if the base unit of the request is valid. If it is not valid, an error is returned with a
	// status code and a message.
	if row, err := e.getCurrency(req.GetBaseUnit(), false); err != nil {
		return &response, status.Errorf(11584, "this base currency does not exist, %v", row.GetSymbol())
	}

	// The purpose of this code is to check if the requested currency exists and if it does not, then return an error
	// message with the appropriate status code. The if statement checks to see if the requested currency exists by using
	// the function getCurrency() with the parameters req.GetQuoteUnit() and false. If an error is returned, the error
	// message is set with the status code 11582 and the currency symbol is included in the message.
	if row, err := e.getCurrency(req.GetQuoteUnit(), false); err != nil {
		return &response, status.Errorf(11582, "this quote currency does not exist, %v", row.GetSymbol())
	}

	// The purpose of this code is to check if the pair (base_unit and quote_unit) provided by the request exists in the
	// database. It uses a SQL query to check if the pair exists and then stores the result of the query in the boolean
	// variable "exist". It then uses an if statement to check if the query was successful and if the "exist" variable is
	// false. If either of these conditions are not met, the code returns an error.
	if err := e.Context.Db.QueryRow("select exists(select id from pairs where base_unit = $1 and quote_unit = $2)::bool", req.GetBaseUnit(), req.GetQuoteUnit()).Scan(&exist); err != nil || !exist {
		return &response, status.Errorf(11585, "this pair %v-%v does not exist", req.GetBaseUnit(), req.GetQuoteUnit())
	}

	// The purpose of response.Success = true is to indicate that a successful response was received. This is typically used
	// in programming to indicate that a given operation was successful and that no errors occurred.
	response.Success = true

	return &response, nil
}

// GetAnalysis - This function is used to get an analysis for a given request. It takes in a request containing limit, page, base unit
// and quote unit as input. It then queries the database to get the count of all pairs with status = true. It then
// calculates the offset and queries the database to get the id, base unit, quote unit and price of the pairs. It then
// calls the GetCandles() function to get the candles and the GetOrders() function to get the buy and sell ratio.
// Finally, it stores all the data in a ResponseAnalysis and returns it.
func (e *Service) GetAnalysis(ctx context.Context, req *pbspot.GetRequestAnalysis) (*pbspot.ResponseAnalysis, error) {

	// The variable 'response' is used to store a response object of type 'pbspot.ResponseAnalysis', which is a type of
	// struct used to store data from a response from a server. By declaring the variable with the keyword 'var', it is
	// initialized to its zero value, which is an empty struct.
	var (
		response pbspot.ResponseAnalysis
	)

	// The purpose of this code is to set a default limit of 30 if the request limit is set to 0. This ensures that requests
	// always have a limit and prevents requests from returning an unlimited amount of data.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is used to authenticate a user before they can access a resource. The auth variable is assigned the result
	// of the Auth() function, which is then checked for any errors. If there is an error, the Error() function is called to
	// handle it.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to query the database for the amount of pairs with a status of true. The result is then
	// stored in the response.Count variable.
	_ = e.Context.Db.QueryRow(`select count(*) as count from pairs where status = $1`, true).Scan(&response.Count)

	// The purpose of this code is to check if the response object contains any values in it. The GetCount() method is used
	// to retrieve the number of values stored in the response object. If this number is greater than 0, then the response
	// object contains at least one value.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for a paginated query. The offset is the number of records to skip when
		// returning the results from the query. The offset is calculated by multiplying the limit (number of records to
		// return) by the page number. If the page number is greater than 0, then the offset is calculated as the limit
		// multiplied by one less than the page number.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query a database for a specific set of data. The code is executing a SQL query to select
		// certain columns from the table "pairs" where the status is equal to true. The limit and offset parameters of the
		// query are being specified by the req.GetLimit() and offset values. The query result is being stored in the rows
		// variable and any errors are being saved to the err variable. The defer statement is used to ensure that the
		// rows.Close() method is called after the query execution is completed.
		rows, err := e.Context.Db.Query(`select id, base_unit, quote_unit, price from pairs where status = $3 order by id desc limit $1 offset $2`, req.GetLimit(), offset, true)
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// This is a loop that is used to iterate through the rows of a result set. The Next() method is used to move the
		// cursor to the next row in the result set and returns true if there is another row to process, or false if there are no more rows.
		for rows.Next() {

			// The purpose of this code is to declare a variable named analysis of type pbspot.Analysis. This variable is used to
			// store an instance of the pbspot.Analysis class which is used to perform analysis on data.
			var (
				analysis pbspot.Analysis
			)

			// This code is a part of an if statement which checks for an error when scanning the rows of a query. The purpose of
			// the code is to scan the rows and assign the values to the variables, analysis.Id, analysis.BaseUnit,
			// analysis.QuoteUnit, and analysis.Price. If an error is encountered, the function will return a response and an error.
			if err := rows.Scan(&analysis.Id, &analysis.BaseUnit, &analysis.QuoteUnit, &analysis.Price); err != nil {
				return &response, err
			}

			// This code is used to request 40 candles from a given base and quote unit. The migrate variable is used to store the
			// response of the GetCandles function and the err variable is used to store any errors that occur. If an error
			// occurs, the response is returned and the error is logged.
			migrate, err := e.GetCandles(context.Background(), &pbspot.GetRequestCandles{BaseUnit: analysis.GetBaseUnit(), QuoteUnit: analysis.GetQuoteUnit(), Limit: 40})
			if err != nil {
				return &response, err
			}

			// This loop iterates through the fields in the migrate object and appends the prices of each field to the Candles
			// field of the analysis object.
			for i := 0; i < len(migrate.Fields); i++ {
				analysis.Candles = append(analysis.Candles, migrate.Fields[i].GetPrice())
			}

			// The for loop is used to execute a set of statements repeatedly until a certain condition is met. In this example,
			// the loop will be executed twice. The loop will begin with the initial value of i being 0 and end when the condition
			// i < 2 is no longer true (i.e. when i reaches 2). Each time the loop is executed, the value of i will be incremented by 1.
			for i := 0; i < 2; i++ {

				// The purpose of the above code is to declare a variable called assigning which is of type pbspot.Assigning. This
				// variable can then be used to store an instance of the pbspot.Assigning class.
				var (
					assigning pbspot.Assigning
				)

				// The above code is using a switch statement to assign a value to the variable "assigning". Depending on the value
				// of "i", the variable "assigning" will be assigned either the value of "pbspot.Assigning_BUY" or "pbspot.Assigning_SELL".
				switch i {
				case 0:
					assigning = pbspot.Assigning_BUY
				case 1:
					assigning = pbspot.Assigning_SELL
				}

				// This code is getting orders from a service with a given set of parameters. It is using the
				// pbspot.GetRequestOrders function to get orders with a specific base unit, quote unit, assigning, status,
				// user ID, and limit. The response is saved in the migrate variable, and if there is an error, it is handled by
				// returning an error and the response.
				migrate, err := e.GetOrders(context.Background(), &pbspot.GetRequestOrders{
					BaseUnit:  analysis.GetBaseUnit(),
					QuoteUnit: analysis.GetQuoteUnit(),
					Assigning: assigning,
					Status:    pbspot.Status_FILLED,
					UserId:    auth,
					Limit:     2,
				})
				if err != nil {
					return &response, err
				}

				// This code segment is used to calculate the buy and sell ratios of a given analysis. The buy ratio is calculated by
				// taking the difference between the current price and the previous price, divided by the previous price, and then
				// multiplying by 100. The sell ratio is calculated in the same manner. The switch statement is used to distinguish between the buy and sell ratios.
				if len(migrate.GetFields()) == 2 {
					switch i {
					case 0:

						// The purpose of this line of code is to calculate the Buy Ratio of an analysis by taking the difference between
						// the current price and the previous price and dividing it by the previous price. The result is then rounded to
						// two decimal places and converted to a float value.
						analysis.BuyRatio = decimal.New(100 * (analysis.GetPrice() - migrate.Fields[0].GetPrice()) / migrate.Fields[0].GetPrice()).Round(2).Float()
					case 1:

						// The purpose of this line of code is to calculate the Sell Ratio of an analysis by taking the difference between
						// the current price and the previous price and dividing it by the previous price. The result is then rounded to
						// two decimal places and converted to a float value.
						analysis.SelRatio = decimal.New(100 * (analysis.GetPrice() - migrate.Fields[0].GetPrice()) / migrate.Fields[0].GetPrice()).Round(2).Float()
					}
				}
			}

			// This code is adding an item to the response.Fields slice. The item being added is the analysis variable. This is
			// likely being done so that the response.Fields slice can be used for further processing or for returning the results of the analysis.
			response.Fields = append(response.Fields, &analysis)
		}
	}

	return &response, nil
}

// GetMarkers - This function is part of a service that is used to retrieve marker symbols from a database. It takes in a context and
// a GetRequestMarkers object as input, executes a SQL query on the database, and returns a ResponseMarker which contains
// a list of marker symbols and an error if one occurred.
func (e *Service) GetMarkers(_ context.Context, _ *pbspot.GetRequestMarkers) (*pbspot.ResponseMarker, error) {

	// The above code is declaring a variable named "response" and assigning it the type of pbspot.ResponseMarker. This
	// allows the program to create an object of type pbspot.ResponseMarker, which is a type of structure used to store and
	// process data in a specific way.
	var (
		response pbspot.ResponseMarker
	)

	// This code is querying a database for a certain symbol from the currencies table. The purpose of the code is to query
	// the database and check for an error. If an error is present, it will return an error response. If no error is
	// present, the rows will be closed.
	rows, err := e.Context.Db.Query("select symbol from currencies where marker = $1", true)
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for rows.Next() loop is used to iterate through the rows of a result set returned from a query. It allows you to
	// access each row of the result set one at a time, so you can process the data accordingly.
	for rows.Next() {

		var (
			symbol string
		)

		// This is an if statement used to check for an error during the process of scanning a row. If an error is encountered,
		// then the function will return the response with the Error() method applied to the context.
		if err := rows.Scan(&symbol); err != nil {
			return &response, err
		}

		// This code adds the symbol to the end of the response Fields array. This is likely being done to provide additional data in the response.
		response.Fields = append(response.Fields, symbol)
	}

	return &response, nil
}

// GetPairs - This function is used to retrieve pairs from the database based on a given symbol. It retrieves all pairs that have a
// base or quote unit that matches the given symbol. For each row, it scans the columns and sets the corresponding fields
// in a Pair object. It also sets the ratio and price of the pair and the status of the pair. Finally, it appends the
// Pair object to the response and returns it.
func (e *Service) GetPairs(_ context.Context, req *pbspot.GetRequestPairs) (*pbspot.ResponsePair, error) {

	// The purpose of this code is to declare a variable called response of type pbspot.ResponsePair. This variable can then
	// be used to store data related to a response to a request for information or a request for action.
	var (
		response pbspot.ResponsePair
	)

	// This code is used to query the database for information from the pairs table where either the base_unit or the
	// quote_unit is equal to the value in the req.GetSymbol() variable. The purpose of this code is to retrieve data from
	// the database and store it in the response variable. The defer rows.Close() statement ensures that the rows are closed once the function is completed.
	rows, err := e.Context.Db.Query("select id, base_unit, quote_unit, base_decimal, quote_decimal, status from pairs where base_unit = $1 or quote_unit = $1", req.GetSymbol())
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for loop with rows.Next() is used to iterate through a database query result set. This allows the code to loop
	// through each row of the result set and do something with the data from each row.
	for rows.Next() {

		// The purpose of the declaration above is to create a variable called "pair" of type "pbspot.Pair". This variable will
		// be used to store a pair of values, such as two strings, two numbers, or two objects. This is often used in
		// programming to store related data in a single object.
		var (
			pair pbspot.Pair
		)

		// This is an if statement which is used to assign the scanned rows from the database to the corresponding variables.
		// If an error occurs while scanning the rows, the statement will return an error as part of the response.
		if err := rows.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit, &pair.BaseDecimal, &pair.QuoteDecimal, &pair.Status); err != nil {
			return &response, err
		}

		// This code is checking the request symbol against the pair symbol and then setting the pair symbol to either the base
		// unit or the quote unit depending on the request symbol. This is likely being done in order to ensure the pair symbol
		// matches the request symbol so the right data is returned.
		if req.GetSymbol() == pair.GetQuoteUnit() {
			pair.Symbol = pair.GetBaseUnit()
		} else {
			pair.Symbol = pair.GetQuoteUnit()
		}

		// The purpose of this code snippet is to check if the exchange (e) has the ratio of the given pair (pair). If so, the
		// ratio is assigned to the pair. The if statement checks is the ratio is returned by the getRatio() function, and if
		// it is, the ok variable will be true, and the ratio will be assigned to the pair.
		if ratio, ok := e.getRatio(pair.GetBaseUnit(), pair.GetQuoteUnit()); ok {
			pair.Ratio = ratio
		}

		// This if statement is used to check if the getPrice function returns a value. If it does, it assigns that value to
		// the Price field of the pair variable. The ok variable is a boolean which is used to determine if the getPrice
		// function returns a value or not. The ok variable will be true if the getPrice function returns a value, and false otherwise.
		if price, ok := e.getPrice(pair.GetBaseUnit(), pair.GetQuoteUnit()); ok {
			pair.Price = price
		}

		// This code checks the status of a pair (a combination of two units) and sets the status of the pair to false if the
		// status of the pair is not ok.
		if ok := e.getStatus(pair.GetBaseUnit(), pair.GetQuoteUnit()); !ok {
			pair.Status = false
		}

		// This statement is appending the value of the variable "pair" to the list of fields in the "response" variable. This
		// is likely part of a function that is building a response containing multiple fields.
		response.Fields = append(response.Fields, &pair)
	}

	return &response, nil
}

// GetPair - This function is part of a service and is used to get a pair from the database given the base unit and quote unit. It
// will return the response pair with fields populated with the row from the database or an error. The function will also
// check the status of the pair and update it in the response.
func (e *Service) GetPair(_ context.Context, req *pbspot.GetRequestPair) (*pbspot.ResponsePair, error) {

	// The purpose of this code is to declare a variable called response of type pbspot.ResponsePair. This is used to store
	// the response from a server in the form of a key-value pair.
	var (
		response pbspot.ResponsePair
	)

	// This code is querying a database for a specific row in the table. The query is looking for a row with the specified
	// base_unit and quote_unit from the 'parameters' req.GetBaseUnit() and req.GetQuoteUnit(). If an error occurs, the error.
	// Finally, the row is closed with the defer keyword so that it is properly released back to the server.
	row, err := e.Context.Db.Query(`select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs where base_unit = $1 and quote_unit = $2`, req.GetBaseUnit(), req.GetQuoteUnit())
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The if statement is used to test if the result of row.Next() returns true. If it does, the code within the if
	// statement will be executed. The purpose of row.Next() is to advance the row pointer to the next row in the result set.
	if row.Next() {

		// The purpose of this code is to declare a variable called 'pair' of type 'pbspot.Pair'. This variable can then be
		// used to store values of type 'pbspot.Pair'.
		var (
			pair pbspot.Pair
		)

		// This code is part of a larger program which likely retrieves data from a database. The purpose of this code is to
		// scan each row of the retrieved data and store the relevant information into a structure called "pair", which likely
		// holds data regarding currency pairs. The "if" statement is a check to make sure that the data was successfully read
		// and stored into the structure, and if not, it will return an error.
		if err := row.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit, &pair.Price, &pair.BaseDecimal, &pair.QuoteDecimal, &pair.Status); err != nil {
			return &response, err
		}

		// The purpose of this code is to check the status of a given request, with the base unit and quote unit specified,
		// before setting the status of the pair to false. The function e.getStatus() returns a boolean value which indicates
		// whether the given request is valid. If the status is not valid, the pair.Status is set to false.
		if ok := e.getStatus(req.GetBaseUnit(), req.GetQuoteUnit()); !ok {
			pair.Status = false
		}

		// This statement is appending a pointer to the variable 'pair' to the array stored in the 'Fields' property of the
		// 'response' variable. This statement is used to add a new element to the 'Fields' array.
		response.Fields = append(response.Fields, &pair)
	}

	return &response, nil
}

// SetOrder - This function is a method of the Service struct. It is used to set an order for a user. It checks the authentication
// and authorization, checks the validity of the input, sets the order and sends back the response. It also handles
// errors encountered during the process.
func (e *Service) SetOrder(ctx context.Context, req *pbspot.SetRequestOrder) (*pbspot.ResponseOrder, error) {

	// The purpose of this code is to declare two variables of type pbspot.ResponseOrder and pbspot.Order respectively.
	// Declaring the variables allows them to be used in the code.
	var (
		response pbspot.ResponseOrder
		order    pbspot.Order
	)

	// This code snippet checks if the request is authenticated by calling the Auth() method on the Context object. If the
	// authentication fails, the code returns an error.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to create a Service object that uses the context stored in the variable e. The Service
	// object is then assigned to the variable migrate.
	migrate := account.Service{
		Context: e.Context,
	}

	// This code is attempting to query a user from migrate using the provided authentication credentials (auth). If the
	// query fails, an error is returned.
	user, err := migrate.QueryUser(auth)
	if err != nil {
		return nil, err
	}

	// This code is checking the user's status. If the user's status is not valid (GetStatus() returns false), it returns an
	// error message informing the user that their account and assets have been blocked and instructing them to contact
	// technical support for any questions.
	if !user.GetStatus() {
		return &response, status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions")
	}

	// This code is part of an if statement that is checking to see if an error has been returned from the e.helperPair()
	// function. If an error has been returned, the code will return the response variable and call the Context.Error() function to log the error.
	if err := e.helperPair(req.GetBaseUnit(), req.GetQuoteUnit()); err != nil {
		return &response, err
	}

	// This is setting the order quantity and value based on the request quantity and price.
	// The request quantity is used to set the order quantity, and the order value is calculated by multiplying the request quantity by the request price.
	order.Quantity = req.GetQuantity()
	order.Value = req.GetQuantity()

	// This is a switch statement that is used to evaluate the trade type of the request object. Depending on the trade
	// type, different actions can be taken. For example, if the trade type is "buy", the code may execute a certain set of
	// instructions to purchase the item, and if the trade type is "sell", the code may execute a different set of instructions to sell the item.
	switch req.GetTrading() {
	case pbspot.Trading_MARKET:

		// The purpose of this code is to set the price of the order (order.Price) to the market price of the requested base
		// and quote units, assigning, and price, which is retrieved from the "e.getMarket" function.
		order.Price = e.getMarket(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetAssigning(), req.GetPrice())

		// This if statement is checking to see if the request is to buy something. If it is, it is calculating the quantity
		// and value of the order by dividing the quantity by the price.
		if req.GetAssigning() == pbspot.Assigning_BUY {
			order.Quantity, order.Value = decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float(), decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float()
		}

	case pbspot.Trading_LIMIT:

		// The purpose of this code is to set the value of the order.Price variable to the value returned by the GetPrice()
		// method of the req object.
		order.Price = req.GetPrice()
	default:
		return &response, status.Error(82284, "invalid type trade position")
	}

	// The purpose of these lines of code is to assign the values of certain variables to the corresponding values from a
	// request object.  This is typically done when creating an order object from the request information.  In this case,
	// the values of the order object are set to the UserId, BaseUnit, QuoteUnit, Assigning, Status, and CreateAt variables
	// in the request object.  The Status is set to PENDING and the CreateAt is set to the current time.
	order.UserId = user.GetId()
	order.BaseUnit = req.GetBaseUnit()
	order.QuoteUnit = req.GetQuoteUnit()
	order.Assigning = req.GetAssigning()
	order.Status = pbspot.Status_PENDING
	order.Trading = req.GetTrading()
	order.CreateAt = time.Now().UTC().Format(time.RFC3339)

	// This code is checking for an error in the helperOrder() function and if one is found, it returns an error response
	// and calls the Context.Error() method with the error. The quantity variable is used to store the result of helperOrder(), which is used to complete the order.
	quantity, err := e.helperOrder(&order)
	if err != nil {
		return &response, err
	}

	// This is a conditional statement used to set a new order and check for any errors that might occur. If an error is
	// encountered, the statement will return a response and an Error context to indicate that an error has occurred.
	if order.Id, err = e.setOrder(&order); err != nil {
		return &response, err
	}

	// The switch statement is used to evaluate the value of the expression "order.GetAssigning()" and execute the
	// corresponding case statement. It is a type of conditional statement that allows a program to make decisions based on different conditions.
	switch order.GetAssigning() {
	case pbspot.Assigning_BUY:

		// This code snippet is likely a part of a function that processes an order. The purpose of the code is to use the
		// function "setAsset()" to set the base unit and user ID of the order to false. If an error occurs during the process,
		// the code will return the response and an error message.
		if err := e.setAsset(order.GetBaseUnit(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		// This code is checking the balance of a user and attempting to subtract the specified quantity from it. If the
		// operation is successful, it will continue with the program. If an error occurs, it will return an error response.
		if err := e.setBalance(order.GetQuoteUnit(), order.GetUserId(), quantity, pbspot.Balance_MINUS); err != nil {
			return &response, err
		}

		// The purpose of e.trade(&order, pbspot.Side_BID) is to replay a trade initiation with the given order and
		// side (BID). This is typically used when the trade is being initiated manually by an operator or trader. It allows
		// the trade to be replayed with the same parameters for accuracy and consistency.
		e.trade(&order, pbspot.Side_BID)

		break
	case pbspot.Assigning_SELL:

		// This code snippet is likely a part of a function that processes an order. The purpose of the code is to use the
		// function "setAsset()" to set the base unit and user ID of the order to false. If an error occurs during the process,
		// the code will return the response and an error message.
		if err := e.setAsset(order.GetQuoteUnit(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		// This code is checking the balance of a user and attempting to subtract the specified quantity from it. If the
		// operation is successful, it will continue with the program. If an error occurs, it will return an error response.
		if err := e.setBalance(order.GetBaseUnit(), order.GetUserId(), quantity, pbspot.Balance_MINUS); err != nil {
			return &response, err
		}

		// The purpose of e.trade(&order, pbspot.Side_ASK) is to replay a trade initiation with the given order and
		// side (ASK). This is typically used when the trade is being initiated manually by an operator or trader. It allows
		// the trade to be replayed with the same parameters for accuracy and consistency.
		e.trade(&order, pbspot.Side_ASK)

		break
	default:
		return &response, status.Error(11588, "invalid assigning trade position")
	}

	// This statement is used to append an element to the "Fields" slice of the "response" struct. The element being
	// appended is the "order" struct.
	response.Fields = append(response.Fields, &order)

	return &response, nil
}

// GetOrders - This function is used to get orders from the database based on the parameters provided. It uses the req
// *pbspot.GetRequestOrders which contains parameters such as the limit and page, assigning, owner, user_id, status
// and base_unit and quote_unit to query the database. The function then uses the query results to build a response which
// is a *pbspot.ResponseOrder. This includes the count and volume of the query results, as well as a list of all the orders that match the query criteria.
func (e *Service) GetOrders(ctx context.Context, req *pbspot.GetRequestOrders) (*pbspot.ResponseOrder, error) {

	// The purpose of this is to declare two variables: response and maps. The variable response is of type
	// pbspot.ResponseOrder, while the variable maps is of type string array.
	var (
		response pbspot.ResponseOrder
		maps     []string
	)

	// This code checks if the limit of the request is set to 0. If it is, then it sets the limit to 30. This is likely done
	// so that a request has a sensible limit, even if one wasn't specified.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// The purpose of this switch statement is to generate a SQL query with the correct assignment clause. Depending on
	// the value of req.GetAssigning(), the maps slice will be appended with the corresponding formatted string.
	switch req.GetAssigning() {
	case pbspot.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %[1]d and type = %[2]d", pbspot.Assigning_BUY, pbasset.Type_SPOT))
	case pbspot.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %[1]d and type = %[2]d", pbspot.Assigning_SELL, pbasset.Type_SPOT))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %[1]d and type = %[3]d or assigning = %[2]d and type = %[3]d)", pbspot.Assigning_BUY, pbspot.Assigning_SELL, pbasset.Type_SPOT))
	}

	// This checks to see if the request (req) has an owner. If it does, the code after this statement will be executed.
	if req.GetOwner() {

		// This code is used to check the authentication of the user. The auth variable is used to store the authentication
		// credentials of the user, and the err variable is used to store any errors that might occur during the authentication
		// process. If an error occurs, the response and error is returned.
		auth, err := e.Context.Auth(ctx)
		if err != nil {
			return &response, err
		}

		// The purpose of this code is to append a formatted string to a slice of strings (maps). The string will include the
		// value of the auth variable and will be of the format "and user_id = '%v'", where %v is a placeholder for the value of auth.
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", auth))

		//	The code snippet is most likely within an if statement, and the purpose of the else if statement is to check if the
		//	user ID of the request is greater than 0. This could be used to check if the user is logged in or has an active
		//	session before performing a certain action.
	} else if req.GetUserId() > 0 {

		//This code is appending a string to a slice of strings (maps) which includes a formatted string containing the user
		//ID from a request object (req). This is likely part of an SQL query being built, with the user ID being used to filter the results.
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", req.GetUserId()))
	}

	// The purpose of this switch statement is to add a condition to a query string based on the status of the request
	// (req.GetStatus()). Depending on the value of the status, a string is added to the maps slice using the fmt.Sprintf()
	// function. This string contains a condition that will be used in the query string.
	switch req.GetStatus() {
	case pbspot.Status_FILLED:
		maps = append(maps, fmt.Sprintf("and status = %d", pbspot.Status_FILLED))
	case pbspot.Status_PENDING:
		maps = append(maps, fmt.Sprintf("and status = %d", pbspot.Status_PENDING))
	case pbspot.Status_CANCEL:
		maps = append(maps, fmt.Sprintf("and status = %d", pbspot.Status_CANCEL))
	}

	// This code checks if the length of the base unit and the quote unit in the request are greater than 0. If they are, it
	// appends a string to the maps variable which includes a formatted SQL query containing the base and quote unit. This
	// is likely part of a larger SQL query used to search for data in a database.
	if len(req.GetBaseUnit()) > 0 && len(req.GetQuoteUnit()) > 0 {
		maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))
	}

	// The purpose of this code is to query the database to count the number of orders and total value of the orders in the
	// database. It then stores the count and volume in the response variable.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count, sum(value) as volume from orders %s", strings.Join(maps, " "))).Scan(&response.Count, &response.Volume)

	// This statement is testing if the response from a user has a count that is greater than 0. If the response has a count
	// greater than 0, then something else will occur.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for a page of results in a request. It calculates the offset by
		// multiplying the limit (number of results per page) by the page number. If the page number is greater than 0, then
		// the offset is recalculated by multiplying the limit by one minus the page number.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to perform a SQL query on a database. It is used to select certain columns from the orders table
		// and to order them by the id in descending order. The limit and offset parameters are used to limit the number of
		// rows returned and to specify where in the result set to start returning rows from. The strings.Join function is used to join the "maps" parameter which is an array of strings.
		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, assigning, price, value, quantity, base_unit, quote_unit, user_id, create_at, status from orders %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through the rows of the result set. The rows.Next() command will return true if the
		// iteration is successful and false if the iteration has reached the end of the result set. The loop will continue to
		// execute until the rows.Next() returns false.
		for rows.Next() {

			// The purpose of the above code is to declare a variable called item with the type pbspot.Order. This allows the
			// program to create an object of type pbspot.Order and assign it to the item variable.
			var (
				item pbspot.Order
			)

			// This code is scanning the rows returned from a database query and assigning the values to the variables in the item
			// struct. If an error is encountered during the scanning process, an error is returned.
			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Value, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt, &item.Status); err != nil {
				return &response, err
			}

			// The purpose of this statement is to add an item to the existing list of fields in a response object. The
			// response.Fields list is appended with the item, which is passed as an argument to the append function.
			response.Fields = append(response.Fields, &item)
		}

		// This code checks for an error in the rows object. If an error is found, the function will return a response and an error message.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// GetAssets - This function is a method of the Service struct that is used to query the database for currencies and their associated
// balance if the user is authenticated. It takes in a context and a GetRequestAssetsManual and returns a ResponseAsset
// and an error. It iterates through the result of the query and appends the currency and its balance (if authenticated) to the response.
func (e *Service) GetAssets(ctx context.Context, _ *pbspot.GetRequestAssetsManual) (*pbspot.ResponseAsset, error) {

	// The purpose of the code is to declare a variable called response with the type pbspot.ResponseAsset. This variable
	// can then be used in the program to store a response asset from the pbspot API.
	var (
		response pbspot.ResponseAsset
	)

	// This code is querying the database to select the columns id, name, symbol, and status from the table currencies. The
	// purpose of the code is to retrieve the information from the table currencies and store them in the variables rows and
	// err. If there is an error, the code will return the response and an error message. Finally, the defer rows.Close()
	// will close the rows of information when the function is finished executing.
	rows, err := e.Context.Db.Query("select id, name, symbol, status from currencies")
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for rows.Next() statement is used in SQL queries to loop through the results of a query. It retrieves the next
	// row from the result set, and assigns the values of the row to variables specified in the query. This allows the
	// programmer to iterate through the result set, one row at a time, and process the data as needed.
	for rows.Next() {

		// The purpose of this code is to declare a variable called asset of the type pbspot.Currency. This allows the code to
		// reference this type of asset later in the code.
		var (
			asset pbspot.Currency
		)

		// This is a snippet of code used to query a database. The purpose of this code is to scan the rows of the database and
		// assign each value to a variable. The "if err" statement is used to check for any errors that may occur while running
		// the query, and returns an error if one is found.
		if err := rows.Scan(&asset.Id, &asset.Name, &asset.Symbol, &asset.Status); err != nil {
			return &response, err
		}

		// The purpose of this statement is to check if the authentication is successful before proceeding with the code. The
		// statement is checking if the authentication is successful by assigning the authentication to the variable auth and
		// then checking if the error is equal to nil. If the error is equal to nil, then the authentication was successful and the code can proceed.
		if auth, err := e.Context.Auth(ctx); err == nil {

			// This code is checking the balance of a certain asset (identified by the symbol) from the account of the user
			// (identified by the auth variable) and assigning the balance to the asset.Balance variable if the balance is greater than 0.
			if balance := e.getBalance(asset.GetSymbol(), auth); balance > 0 {
				asset.Balance = balance
			}
		}

		// This statement is used to append a field to the response.Fields array. It is used to add a new element to an array.
		// The element being added is the asset variable.
		response.Fields = append(response.Fields, &asset)
	}

	return &response, nil
}

// SetAsset - This function is used to set an asset for a user. It takes in a context, a request for an asset, and returns a
// response and an error. It uses the auth to get an entropy and generate a new record with asset wallet data, address,
// and entropy. It then checks if the asset address has already been generated and if not, inserts the asset and wallet
// into the database. It then sets the success to true and returns the response and no error.
func (e *Service) SetAsset(ctx context.Context, req *pbspot.SetRequestAsset) (*pbspot.ResponseAsset, error) {

	// The purpose of this code is to declare a variable called "response" of type "pbspot.ResponseAsset". This variable
	// will be used to store data that is returned from a request made to a pbspot API.
	var (
		response pbspot.ResponseAsset
	)

	// The purpose of this code is to retrieve the authentication information associated with the given context (ctx). If
	// there is an error with retrieving the authentication information, then the error is returned.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The switch statement is used to check the value of the req.GetType() expression and then execute the relevant case
	// statement depending on the result. The purpose of this is to allow the code to behave differently depending on the
	// value of the expression. For example, if the value of req.GetType() is "fixed", it will run the first case
	// statement, if it is "variable" it will run the second, and so on.
	switch req.GetType() {
	case pbspot.Type_CRYPTO:

		// The purpose of this code is to declare a variable named "cross" and set it to a CrossChain object. This allows the
		// code to reference the CrossChain object and use it in the program.
		var (
			cross keypair.CrossChain
		)

		// This code is used to get entropy from a given authentication (auth) and assign it to the variable entropy. If an
		// error is encountered during this process, the code returns a response and an error is logged.
		entropy, err := e.getEntropy(auth)
		if err != nil {
			return &response, err
		}

		// The code is attempting to create a new address using a secret, entropy, and platform. If there is an error, the
		// function will return the response and an error.
		if response.Address, _, err = cross.New(fmt.Sprintf("%v-&*39~763@)", e.Context.Secrets[1]), entropy, req.GetPlatform()); err != nil {
			return &response, err
		}

		// This code is querying a database for an asset with a given symbol and user_id. The row variable is used to store the
		// result of the query, and err is used to store any errors that may occur. The if err != nil statement checks for any
		// errors that may have occurred and, if one is found, the error is returned. To defer row.Close() statement closes
		// the database query at the end of the function, regardless of how the function ends.
		row, err := e.Context.Db.Query(`select id from assets where symbol = $1 and user_id = $2 and type = $3`, req.GetSymbol(), auth, pbasset.Type_SPOT)
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The if row.Next() statement is used to check if there is another row available in a result set from a database
		// query. If there is another row, it returns true and can be used to loop through the result set.
		if row.Next() {

			// The code above is checking if the address returned from the e.getAddress() method is empty. If the address is
			// empty, the code inside the if statement will be executed.
			if address := e.getAddress(auth, req.GetSymbol(), req.GetPlatform(), req.GetProtocol()); len(address) == 0 {

				// This code is performing an SQL INSERT statement to add a new record to the 'wallets' table. The values being
				// inserted are the address, symbol, platform, protocol, and user_id from the request parameters. The query is then
				// executed and if there is an error, an error message is returned.
				if _, err = e.Context.Db.Exec("insert into wallets (address, symbol, platform, protocol, user_id) values ($1, $2, $3, $4, $5)", response.GetAddress(), req.GetSymbol(), req.GetPlatform(), req.GetProtocol(), auth); err != nil {
					return &response, err
				}

				return &response, nil
			}

			return &response, status.Error(700990, "the asset address has already been generated")
		}

		// This code snippet is used to execute an SQL query to insert a row of data into the 'assets' table. The first
		// parameter is the user_id from the auth variable, and the second parameter is the symbol from the req.GetSymbol()
		// variable. If there is an error during the execution of the query, an error will be returned.
		if _, err = e.Context.Db.Exec("insert into assets (user_id, symbol) values ($1, $2);", auth, req.GetSymbol()); err != nil {
			return &response, err
		}

		// This code is used to insert values into the wallets table in a database. The values being inserted are the address,
		// symbol, platform, protocol, and user_id, which are passed in as arguments. It checks for any errors that may occur
		// while inserting the values and returns an error if one is encountered.
		if _, err = e.Context.Db.Exec("insert into wallets (address, symbol, platform, protocol, user_id) values ($1, $2, $3, $4, $5)", response.GetAddress(), req.GetSymbol(), req.GetPlatform(), req.GetProtocol(), auth); err != nil {
			return &response, err
		}

	case pbspot.Type_FIAT:

		// The purpose of this code is to set the asset for the given symbol, using the provided authentication details. If
		// there is an error encountered while attempting to set the asset, the response variable is returned and the error is logged.
		if err := e.setAsset(req.GetSymbol(), auth, true); err != nil {
			return &response, err
		}

	}
	response.Success = true

	return &response, nil
}

// GetAsset - This function is used to get an asset from a database and retrieve related information, such as the balance, volume,
// and fees associated with it. It also gets information about the chains associated with the asset, such as the
// reserves, address, existance, and contract. Finally, it returns the response asset which contains all the gathered information.
func (e *Service) GetAsset(ctx context.Context, req *pbspot.GetRequestAsset) (*pbspot.ResponseAsset, error) {

	// The variable 'response' is declared as a type of pbspot.ResponseAsset. This is used to store a response asset, which
	// is typically used to store the response of an API request. This allows the response to be accessed and manipulated by the code.
	var (
		response pbspot.ResponseAsset
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The code is checking to see if an error occurred while attempting to get a currency. If there is an error, the
	// function will return the response and the error.
	row, err := e.getCurrency(req.GetSymbol(), false)
	if err != nil {
		return &response, err
	}

	// This line of code sets the value of the Balance attribute of the row object to the balance of the account associated
	// with the symbol in the request object, which is authenticated using the auth parameter.
	row.Balance = e.getBalance(req.GetSymbol(), auth)

	// This query is used to calculate the volume of orders for a particular symbol, with the given assigning and status,
	// for the given user. It is checking the base unit and quote unit against the given symbol and using the price to
	// convert between the two if necessary. It is then adding them together and using the coalesce function to return 0.00
	// if there is no data returned. Finally, it is scanning the result into the row.Volume field.
	_ = e.Context.Db.QueryRow(`select coalesce(sum(case when base_unit = $1 then value when quote_unit = $1 then value * price end), 0.00) as volume from orders where base_unit = $1 and assigning = $2 and status = $4 and type = $5 and user_id = $6 or quote_unit = $1 and assigning = $3 and status = $4 and type = $5 and user_id = $6`, req.GetSymbol(), pbspot.Assigning_SELL, pbspot.Assigning_BUY, pbspot.Status_PENDING, pbasset.Type_SPOT, auth).Scan(&row.Volume)

	// This is a for loop in Go. The purpose of the loop is to iterate over each element in the "row.GetChainsIds()" array.
	// The loop starts at index 0 and continues until it reaches the last element in the array. On each iteration, the loop
	// will execute the code inside the block.
	for i := 0; i < len(row.GetChainsIds()); i++ {

		// The code above is checking if a chain exists by checking its ID. If the ID is greater than 0, the chain exists.
		if chain, _ := e.getChain(row.GetChainsIds()[i], false); chain.GetId() > 0 {

			// chain.Rpc and chain.Address are being assigned empty strings. This is likely to reset the values of these variables to their default state.
			chain.Rpc, chain.Address = "", ""

			// This is an if statement used to determine if a contract has an ID greater than 0.
			// If it does, it will execute the code inside the if statement.
			if contract, _ := e.getContract(row.GetSymbol(), row.GetChainsIds()[i]); contract.GetId() > 0 {

				// The purpose of the code is to assign two variables with the same value. The first variable, chain.Fees, is
				// assigned the value returned by the function contract.GetFees(). The second variable, contract.FeesWithdraw, is assigned the value 0.
				chain.Fees, contract.Fees = contract.GetFees(), 0

				// This code is used to get the price of a requested symbol given a base unit. It uses the GetPrice method from the e
				// object and passes in a context.Background() and a GetRequestPriceManual object containing the base unit and the
				// requested symbol. If the GetPrice method returns an error, the error is returned in the response and the Context.Error() method handles the error.
				price, err := e.GetPrice(context.Background(), &pbspot.GetRequestPriceManual{BaseUnit: chain.GetParentSymbol(), QuoteUnit: req.GetSymbol()})
				if err != nil {
					return &response, err
				}

				// The purpose of this code is to calculate the fees for withdrawing from a particular chain. The chain.FeesWithdraw
				// variable is assigned to a decimal value which is calculated by multiplying the chain.GetFeesWithdraw() value with
				// the price.GetPrice() value. The result is then converted to a floating point number.
				chain.Fees = decimal.New(chain.GetFees()).Mul(price.GetPrice()).Float()

				// The purpose of this statement is to set the contract for the chain. This statement is typically used in a
				// blockchain context and assigns the contract object to the chain object. This allows the chain to access the
				// functions and variables defined in the contract.
				chain.Contract = contract
			}

			// The purpose of this code is to set the reserve of the chain to the reserve of the asset that is requested from the
			// symbol, platform, and protocol. The code is retrieving the reserve of the asset in order to set the reserve of the chain.
			chain.Reserve = e.getReserve(req.GetSymbol(), chain.GetPlatform(), chain.Contract.GetProtocol())

			// The purpose of this code is to get the address for a particular symbol, platform, and protocol from an external
			// source (e.getAddress) and assign the address to the chain.Address variable.
			chain.Address = e.getAddress(auth, req.GetSymbol(), chain.GetPlatform(), chain.Contract.GetProtocol())

			// The purpose of this code is to retrieve an asset from the chain with the symbol provided in the request. The Exist
			// variable is being set to the result of calling the getAsset() method on the chain object, which takes the symbol from the request and the authentication details as parameters.
			chain.Exist = e.getAsset(req.GetSymbol(), auth)

			// This statement is used to add a new item, "chain", to the end of an existing slice of items, "row.Chains". Append
			// is a built-in function that allows you to add items to the end of a slice.
			row.Chains = append(row.Chains, chain)
		}
	}

	// row.ChainsIds = make([]int64, 0) is used to create a slice of int64 elements with zero length. It will initialize the
	// slice with no elements in it.
	row.ChainsIds = make([]int64, 0)

	// This statement is appending a row to the Fields slice in the response variable. This allows the user to add more values to the slice.
	response.Fields = append(response.Fields, row)

	return &response, nil
}

// GetCandles - This function is a method of the service struct and serves to provide a response of candles for a given request. It is
// responsible for querying the database for the candle data, and then formatting it into a response struct. It also
// calculates some statistics based on the requested data and adds them to the response struct.
func (e *Service) GetCandles(_ context.Context, req *pbspot.GetRequestCandles) (*pbspot.ResponseCandles, error) {

	// The purpose of this code is to create three variables with zero values: response, limit and maps. The response
	// variable is of type pbspot.ResponseCandles, the limit variable is of type string, and the maps variable is of type
	// slice of strings.
	var (
		response pbspot.ResponseCandles
		limit    string
		maps     []string
	)

	// This code checks if the limit of the request is set to 0. If it is, then it sets the limit to 30. This is likely done
	// so that a request has a sensible limit, even if one wasn't specified.
	if req.GetLimit() == 0 {
		req.Limit = 500
	}

	// This code is used to set a limit to the request. It checks if req.GetLimit() is greater than 0. If so, it sets the
	// limit variable to a string with the limit set to that amount. This is likely used to set a limit on the amount of
	// data that will be returned in the response.
	if req.GetLimit() > 0 {
		limit = fmt.Sprintf("limit %d", req.GetLimit())
	}

	// This code is checking to see if the "From" and "To" values in the request are greater than 0. If they are, a
	// formatted string will be appended to the "maps" array containing a timestamp that is less than the "To" value in the
	// request. This code is likely used to filter a query based on a time range.
	if req.GetTo() > 0 {
		maps = append(maps, fmt.Sprintf(`and to_char(ohlc.create_at::timestamp, 'yyyy-mm-dd hh24:mi:ss') < to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss')`, req.GetTo()))
	}

	// This code is used to query the database to return OHLC (open-high-low-close) data. The SQL query is using the
	// fmt.Sprintf function to substitute the variables (req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "),
	// help.Resolution(req.GetResolution()), limit) into the query. The query is then executed, and the results are stored
	// in the rows variable. Finally, the rows variable is closed at the end of the code.
	rows, err := e.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', ohlc.create_at))::integer buckettime, first(ohlc.price, ohlc.create_at) as open, last(ohlc.price, ohlc.create_at) as close, first(ohlc.price, ohlc.price) as low, last(ohlc.price, ohlc.price) as high, sum(ohlc.quantity) as volume, avg(ohlc.price) as avg_price, ohlc.base_unit, ohlc.quote_unit from trades as ohlc where ohlc.base_unit = '%[1]s' and ohlc.quote_unit = '%[2]s' %[3]s group by buckettime, ohlc.base_unit, ohlc.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The purpose of the for rows.Next() loop is to iterate through the rows in a database table. It is used to perform
	// some action on each row of the table. This could include retrieving data from the row, updating data in the row, or
	// deleting the row.
	for rows.Next() {

		// The purpose of the variable "item" is to store data of type pbspot.Candles. This could be used to store an array of
		// candles or other data related to pbspot.Candles.
		var (
			item pbspot.Candles
		)

		// This code is checking for errors while scanning a row of data from a database. It is assigning the values of the row
		// to the variables item.Time, item.Open, item.Close, item.Low, item.High, item.Volume, item.Price, item.BaseUnit, and
		// item.QuoteUnit. If an error occurs during the scan, the code will return an error response.
		if err = rows.Scan(&item.Time, &item.Open, &item.Close, &item.Low, &item.High, &item.Volume, &item.Price, &item.BaseUnit, &item.QuoteUnit); err != nil {
			return &response, err
		}

		// This code is likely appending an item to a response.Fields array. It is likely used to add an item to the array and
		// modify the array.
		response.Fields = append(response.Fields, &item)
	}

	// The purpose of the following code is to declare a variable called stats of the type pbspot.Stats. This variable will
	// be used to store information related to the pbspot.Stats data type.
	var (
		stats pbspot.Stats
	)

	// This code is used to fetch and analyze data from a database. It uses the QueryRow() method to retrieve data from the
	// database and then scan it into the stats variable. The code is specifically used to get the count, volume, low, high,
	// first and last values from the trades table for a given base unit and quote unit.
	_ = e.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from trades as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

	// This code checks if the length of the 'response.Fields' array is greater than 1. If so, it assigns the 'Close' value
	// of the second element in the 'response.Fields' array to the 'Previous' field of the 'stats' object.
	if len(response.Fields) > 1 {
		stats.Previous = response.Fields[1].Close
	}

	//The purpose of this statement is to assign the pointer stats to the Stats field of the response object. This allows
	//the response object to access the data stored in the stats variable.
	response.Stats = &stats

	return &response, nil
}

// GetTransfers - This function is a method of the Service type. It is used to get a list of transfers from a database. It takes a
// context and a request object as parameters. The request object contains information about the requested transfers,
// such as the limit, order ID and whether the request should only return transfers for a specific user. The function
// then queries the database for the requested transfers, and returns a ResponseTransfer object containing the relevant transfers.
func (e *Service) GetTransfers(ctx context.Context, req *pbspot.GetRequestTransfers) (*pbspot.ResponseTransfer, error) {

	// The purpose of this code is to declare two variables: a variable called "response" of type "pbspot.ResponseTransfer"
	// and a variable called "maps" of type "string slice". This allows the program to store values in these two variables
	// and access them throughout the code.
	var (
		response pbspot.ResponseTransfer
		maps     []string
	)

	// This code checks if the request's limit is 0. If it is, it sets the request's limit to 30. This is likely done to
	// ensure that the request is not given an unlimited amount of data, which could cause performance issues.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is used to generate a query for a database. The purpose of the switch statement is to determine the value
	// of the assigning parameter and then generate the appropriate query based on the value. For example, if the value of
	// req.GetAssigning() is pbspot.Assigning_BUY, then the query generated will be "where assigning =
	// pbspot.Assigning_BUY". If the value of req.GetAssigning() is not specified, then the query generated will be "where
	// (assigning = pbspot.Assigning_BUY or assigning = pbspot.Assigning_SELL)".
	switch req.GetAssigning() {
	case pbspot.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_BUY))
	case pbspot.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_SELL))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %d or assigning = %d)", pbspot.Assigning_BUY, pbspot.Assigning_SELL))
	}

	// This line of code is adding a string to a slice of strings (maps) which contains a formatted variable
	// (req.GetOrderId()). The purpose of this code is to add a condition to a SQL query which includes the value of the
	// req.GetOrderId() variable.
	maps = append(maps, fmt.Sprintf("and order_id = '%v'", req.GetOrderId()))

	// The "if req.GetOwner()" statement is checking if a request has an owner associated with it. If it does, the code
	// inside the if statement will execute. If not, it will skip over it.
	if req.GetOwner() {

		// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
		// an error. This is necessary to ensure that only authorized users are accessing certain resources.
		auth, err := e.Context.Auth(ctx)
		if err != nil {
			return &response, err
		}

		// The purpose of this code is to create a map which contains a key-value pair of "user_id" and the value of the
		// variable "auth". The variable "maps" is a slice of type string, and the function "append" is used to append the map created by the fmt.Sprintf to the slice.
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", auth))
	}

	// This code is used to query the table 'transfers' with the given parameters. It uses a fmt.Sprintf statement to format the
	// query string with the given parameters, then it uses the e.Context.Db.Query() to execute the query and store the
	// results into the rows variable. If an error occurs, it returns an error response. Finally, it closes the rows variable.
	rows, err := e.Context.Db.Query(fmt.Sprintf("select id, user_id, base_unit, quote_unit, price, quantity, assigning, fees, create_at from transfers %s order by id desc limit %d", strings.Join(maps, " "), req.GetLimit()))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for loop is used to iterate through the rows in a database. The rows.Next() statement is used to move to the next
	// row in the result set. This loop allows you to access the data in each row and process it as needed.
	for rows.Next() {

		// The purpose of this code is to declare a variable called item, and assign it to a value of type pbspot.Transfer.
		// This is used in programming to store data in the form of a variable and access it at a later time.
		var (
			item pbspot.Transfer
		)

		// This code is part of a function that retrieves data from a database. The purpose of the if statement is to scan the
		// rows of the database and assign each row's values to the corresponding variables. If an error occurs while scanning
		// the rows, the function will return an error.
		if err = rows.Scan(&item.Id, &item.UserId, &item.BaseUnit, &item.QuoteUnit, &item.Price, &item.Quantity, &item.Assigning, &item.Fees, &item.CreateAt); err != nil {
			return &response, err
		}

		// This statement is appending a new item to the Fields array of the response object. The purpose of this statement is
		// to add a new item to the response's Fields array.
		response.Fields = append(response.Fields, &item)
	}

	// This code is checking for an error in the rows object. If there is an error, the code will return an empty response
	// and an error object. This is likely part of a larger function that is retrieving data from a database and returning
	// it in a response. The if statement is making sure that the query was successful and that the response is valid.
	if err = rows.Err(); err != nil {
		return &response, err
	}

	return &response, nil
}

// GetTrades - This function is used to retrieve trades from the database. It takes a GetRequestTrades request as an argument and
// returns a ResponseTrades. The function uses a switch statement to filter the data based on the assigning field in the
// GetRequestTrades request. It then checks the count of trades that match the filter and returns them in the
// ResponseTrades. The limit and page values are used to limit the number of trades returned.
func (e *Service) GetTrades(_ context.Context, req *pbspot.GetRequestTrades) (*pbspot.ResponseTrades, error) {

	// The purpose of the above code is to declare two variables, response and maps, of the types pbspot.ResponseTrades and
	// []string respectively. This allows for the response to store information related to trades, such as the trade price
	// and quantity, and the maps to store an array of strings.
	var (
		response pbspot.ResponseTrades
		maps     []string
	)

	// The purpose of this code is to check if the value of the "req.GetLimit()" is equal to 0. If it is equal to 0, then
	// the value of "req.Limit" is set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This switch statement is used to conditionally append a string to an array (maps) based on the value of the
	// req.GetAssigning() function. req.GetAssigning() returns a value that is part of the pbspot.Assigning enumeration,
	// which can be either BUY or SELL. Depending on what value is returned, the corresponding string is appended to the
	// array. If neither BUY nor SELL is returned, then a string that includes both BUY and SELL is appended.
	switch req.GetAssigning() {
	case pbspot.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_BUY))
	case pbspot.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_SELL))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %d or assigning = %d)", pbspot.Assigning_BUY, pbspot.Assigning_SELL))
	}

	// The purpose of this code is to append a string to the maps variable. The string contains a formatted version of the
	// BaseUnit and QuoteUnit from the req variable. This is likely part of a larger SQL query that is being built.
	maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))

	// This code is used to query a database for the number of trades in a table and then store the result in a variable
	// called response.Count. The query is made up of a string of SQL code as well as strings.Join() which adds the
	// information in the maps variable to the query. The Scan() function is then used to store the result of the query in
	// the response.Count variable.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from trades %s", strings.Join(maps, " "))).Scan(&response.Count)

	// The purpose of this statement is to check if the response object has any elements in it. If the response object has
	// more than 0 elements, then the code inside the if statement will execute.
	if response.GetCount() > 0 {

		// This code is used to calculate an offset based on the page and limit provided. The page is usually used to determine
		// which page of records to return from a database, and the limit is used to determine how many records to return. The
		// offset is the starting index of the records to return. The code checks if the page is greater than 0, and if so,
		// subtracts 1 from the page to get the correct offset.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query the database. It uses the `Db.Query` function to select data from the trades table, with
		// the given conditions in the `strings.Join(maps, " ")`, limit, and offset. If an error occurs, it returns an error
		// response. The `defer rows.Close()` statement ensures that the rows will be closed when the function returns.
		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, assigning, price, quantity, base_unit, quote_unit, create_at from trades %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate over a database query result set. It will loop through each row in the
		// result set and perform the instructions within the loop. This is a common way to access data from a database.
		for rows.Next() {

			// The purpose of this code is to declare a variable called item with a type of pbspot.Trade. This allows the item
			// variable to be used to store any type of data associated with the pbspot.Trade type.
			var (
				item pbspot.Trade
			)

			// This is a line of code used to scan the rows of a database table and assign the columns of each row to variables.
			// It scans the columns (ID, Assigning, Price, Quantity, BaseUnit, QuoteUnit, and CreateAt) of the row, and assigns
			// them to the corresponding variables (item.Id, item.Assigning, item.Price, item.Quantity, item.BaseUnit,
			// item.QuoteUnit, and item.CreateAt). If an error occurs during the scanning process, it will return an error message.
			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.CreateAt); err != nil {
				return &response, err
			}

			// This statement adds an item to the response.Fields slice. The response.Fields slice stores a list of items, and
			// this statement appends a new item to the end of the list.
			response.Fields = append(response.Fields, &item)
		}

		// The purpose of this code is to check for an error when querying a database. If an error is returned, the code will
		// return a response and an error message.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// GetPrice - This function is part of a service that is used to get the price of a given asset. It takes in a context and a
// GetRequestPriceManual request object as parameters. The purpose of the function is to get the price of the asset from
// the given base and quote units, then return a ResponsePrice object. It also checks if the price can be obtained from
// the quote and base units in the opposite order. If so, the price is then calculated by taking the inverse of the price
// and rounding it to 8 decimal places. Finally, the ResponsePrice object is returned.
func (e *Service) GetPrice(_ context.Context, req *pbspot.GetRequestPriceManual) (*pbspot.ResponsePrice, error) {

	// The code above declares two variables: response of type pbspot.ResponsePrice and ok of type bool. The purpose of this
	// code is to create two variables that can be used in the code that follows.
	var (
		response pbspot.ResponsePrice
		ok       bool
	)

	// The purpose of this code is to check whether the price of a certain item has been successfully retrieved, and if so,
	// return the response and a nil (no error) value. The line starts by attempting to get the price of the item based on
	// the given base and quote units, and then assigns the response and the boolean value of ok to the variables
	// response.Price and ok, respectively. Finally, the code returns the response and a nil value if ok is true.
	if response.Price, ok = e.getPrice(req.GetBaseUnit(), req.GetQuoteUnit()); ok {
		return &response, nil
	}

	// This code is checking if the price of a product or service can be obtained given the quote and base units. If the
	// price can be obtained, the response will be rounded to 8 decimal places and stored in the response.Price variable.
	if response.Price, ok = e.getPrice(req.GetQuoteUnit(), req.GetBaseUnit()); ok {
		response.Price = decimal.New(decimal.New(1).Div(response.Price).Float()).Round(8).Float()
	}

	return &response, nil
}

// GetTransactions - This function is a method in a service struct used to get a list of transactions and associated data from a database.
// The function takes a context.Context and a *pbspot.GetRequestTransactionsManual as parameters. The function returns a
// response of type *pbspot.ResponseTransaction and an error. The function sets a default value of 30 for the limit if it
// is set to 0 and then uses a switch statement to filter out the transactions by type. The function also filters by
// symbol and status. The function then queries for the number of transactions and their associated data and returns the results.
func (e *Service) GetTransactions(ctx context.Context, req *pbspot.GetRequestTransactionsManual) (*pbspot.ResponseTransaction, error) {

	// The purpose of the code snippet is to declare two variables, response and maps, of types pbspot.ResponseTransaction and []string respectively.
	var (
		response pbspot.ResponseTransaction
		maps     []string
	)

	// The purpose of this code is to set a default limit value if the limit value requested (req.GetLimit()) is equal to
	// zero. In this case, the default limit value is set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This switch statement is used to create a query condition based on the transaction type. Depending on the transaction
	// type, the query string will be amended to include the appropriate condition. If the transaction type is deposit, the
	// query string will include the condition where assignment = pbspot.Assignment_DEPOSIT. If the transaction type is withdraws,
	// the query string will include the condition where assignment = pbspot.Assignment_WITHDRAWS. If the transaction type is
	// neither deposit nor withdraws, the query string will include the condition where (assignment = pbspot.Assignment_WITHDRAWS or assignment = pbspot.Assignment_DEPOSIT).
	switch req.GetAssignment() {
	case pbspot.Assignment_DEPOSIT:
		maps = append(maps, fmt.Sprintf("where assignment = %d", pbspot.Assignment_DEPOSIT))
	case pbspot.Assignment_WITHDRAWS:
		maps = append(maps, fmt.Sprintf("where assignment = %d", pbspot.Assignment_WITHDRAWS))
	default:
		maps = append(maps, fmt.Sprintf("where (assignment = %d or assignment = %d)", pbspot.Assignment_WITHDRAWS, pbspot.Assignment_DEPOSIT))
	}

	// This code is checking the length of the request's symbol and, if greater than zero, appending a string to the maps
	// variable. The string contains the symbol from the request. This allows the code to filter out requests that do not
	// have a symbol provided and process only those that do.
	if len(req.GetSymbol()) > 0 {
		maps = append(maps, fmt.Sprintf("and symbol = '%v'", req.GetSymbol()))
	}

	// This line of code is appending a string to the maps variable. The string is formatted using the fmt.Sprintf()
	// function, and is used to set the status of the variable to a specific value (in this case, pbspot.Status_RESERVE).
	// This is likely being used to filter out specific values from a list or array of values.
	maps = append(maps, fmt.Sprintf("and status != %d", pbspot.Status_RESERVE))

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This line of code is appending a string to the maps slice. The string is formatted with the auth variable. The
	// purpose of this line of code is to add a key-value pair to the maps slice, where the key is "user_id" and the value
	// is the auth variable.
	maps = append(maps, fmt.Sprintf("and user_id = %v", auth))

	// The purpose of this code is to query a database and retrieve a count of the transactions that meet the criteria
	// specified in the maps variable. The result of the query is then stored in the response.Count variable.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from transactions %s", strings.Join(maps, " "))).Scan(&response.Count)

	// The purpose of this statement is to check if the response from a particular operation contains at least one element.
	// If the response has more than one element, the condition will evaluate to true; otherwise, it will evaluate to false.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for pagination. It takes the limit and page from the request and
		// multiplies them together. If the page is greater than 0 (so the first page), it subtracts one from the page before
		// multiplying. This is because the offset for the first page is 0, not the limit.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query data from a database. The fmt.Sprintf function is used to build an SQL query string. The
		// query string includes fields from the transactions table, a WHERE clause generated from the maps variable, a limit
		// (req.GetLimit()), and an offset (offset). The rows, err variable is used to execute the query and return the
		// results. To defer rows.Close() statement is used to ensure that the database connection is closed when the query is done.
		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, symbol, hash, value, price, fees, confirmation, "to", chain_id, user_id, assignment, type, platform, protocol, status, create_at from transactions %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through the rows returned from a database query. The rows.Next() method is used to
		// read the next row from the query result. The for loop allows the programmer to loop through each row of the query
		// result and perform an action on the data returned.
		for rows.Next() {

			// The purpose of this variable declaration is to declare a variable named "item" of type "pbspot.Transaction". This
			// variable can then be used to store and manipulate data of type "pbspot.Transaction".
			var (
				item pbspot.Transaction
			)

			// This code is used to scan the rows of a database table and assign the values to the corresponding fields of the
			// item struct. The if statement is used to check for any errors that occur while scanning the rows. If an error is
			// encountered, the error is passed to the Error function and the response is returned.
			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.Hash,
				&item.Value,
				&item.Price,
				&item.Fees,
				&item.Confirmation,
				&item.To,
				&item.ChainId,
				&item.UserId,
				&item.Assignment,
				&item.Type,
				&item.Platform,
				&item.Protocol,
				&item.Status,
				&item.CreateAt,
			); err != nil {
				return &response, err
			}

			// This code is retrieving the chain from the given chain ID and storing it in the item.Chain variable. It then checks
			// if an error occurred while doing so and if there was an error, it will return a nil value and the error.
			item.Chain, err = e.getChain(item.GetChainId(), false)
			if err != nil {
				return nil, err
			}

			// This code checks the protocol associated with the item. If the protocol is not equal to the mainnet protocol, then
			// the fees associated with the item are multiplied by the item's price, and the result is stored as a float.
			if item.GetProtocol() != pbspot.Protocol_MAINNET {
				item.Fees = decimal.New(item.GetFees()).Mul(item.GetPrice()).Float()
			}

			// This statement is appending an item to the Fields slice of the response variable. This adds the item to the end of
			// the slice, increasing its length by one. The purpose of this statement is to add an item to the Fields slice.
			response.Fields = append(response.Fields, &item)
		}

		// This is a check to make sure that the operation on the rows was successful. If it was not successful, it returns an error.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// SetWithdraw - This code is a function written in the Go programming language.
// The purpose of the function is to process a request to withdraw a certain amount of a given currency from an account.
// The function handles authentication, validation, and the execution of the withdrawal request.
// It also handles logging the transaction in the database, setting the security code in the context, and returning the correct response.
func (e *Service) SetWithdraw(ctx context.Context, req *pbspot.SetRequestWithdraw) (*pbspot.ResponseWithdraw, error) {

	// The purpose of this code is to declare two variables: response and fees. The first variable, response, is of type
	// pbspot.ResponseWithdraw, and the second variable, fees, is of type float64.
	var (
		response pbspot.ResponseWithdraw
		fees     float64
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The if statement is used to check if the GetRefresh() function is returning a truthy (true) value. If it is, the code
	// within the if block will be executed.
	if req.GetRefresh() {

		// This code is checking for an error when setting the secure flag to false. If there is an error, it returns a
		// response and the error context.
		if err := e.setSecure(ctx, false); err != nil {
			return &response, err
		}

		return &response, nil
	}

	// The purpose of this code is to create a new Service with the context set to the context in the e variable. Migrate is
	// a variable of type Service which is being instantiated. This Service will be used to perform certain operations
	// related to the account.
	migrate := account.Service{
		Context: e.Context,
	}

	// This code is attempting to query a user using a given authentication (auth). The QueryUser function is likely part of
	// a migration library and returns a user object and an error object. If there is an error, it is returned with a nil
	// value for the user object.
	user, err := migrate.QueryUser(auth)
	if err != nil {
		return nil, err
	}

	// The purpose of this code is to check the status of the user and return an error if the user's status is not valid. If
	// the user's status is not valid, the code returns an error message to the user, indicating that their account and
	// assets have been blocked and that they should contact technical support for any questions.
	if !user.GetStatus() {
		return &response, status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions")
	}

	// This code is checking to make sure that the address provided in the request is a valid crypto address for the
	// specified platform. If the address is not valid, the error is returned to the caller.
	if err := keypair.ValidateCryptoAddress(req.GetAddress(), req.GetPlatform()); err != nil {
		return &response, err
	}

	// This code is part of an error handling process. The if statement checks to see if the helperInternalAsset method
	// returns an error when given the address from the request. If it does return an error, the code will return the
	// response and log the error using the Context.Error() method.
	if err := e.helperInternalAsset(req.GetAddress()); err != nil {
		return &response, err
	}

	// This code is used to get the chain with the specified ID from the e.getChain() function. If an error occurs, the code
	// returns an error message indicating the chain array by the specified ID is currently unavailable.
	chain, err := e.getChain(req.GetId(), true)
	if err != nil {
		return &response, status.Errorf(11584, "the chain array by id %v is currently unavailable", req.GetId())
	}

	// This code is used to get the currency of a request. It checks if the currency is available in the request and if it
	// is not available, it returns an error message (status.Errorf(10029, "the currency requested array by id %v is
	// currently unavailable", req.GetSymbol())).
	currency, err := e.getCurrency(req.GetSymbol(), false)
	if err != nil {
		return &response, status.Errorf(10029, "the currency requested array by id %v is currently unavailable", req.GetSymbol())
	}

	// The purpose of the code above is to retrieve a contract from a blockchain given a symbol and chain ID. It does this
	// by calling the getContract() function on the e variable, passing in the symbol from the req variable and the chain ID
	// from the chain variable. The result of this call is then stored in the contract variable.
	contract, _ := e.getContract(req.GetSymbol(), chain.GetId())

	// This code is checking to see if the function getSecure() returns an error. If an error is returned, the code is
	// returning a response and an error message.
	secure, err := e.getSecure(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of the code snippet is to ensure that the email code provided is 6 numbers long. If it is not, an error
	// with the code 16763 will be returned.
	if len(req.GetEmailCode()) != 6 {
		return &response, status.Error(16763, "the code must be 6 numbers")
	}

	// This if statement is used to check if the security code provided by the user matches the security code associated
	// with the user's email address. If the code is incorrect or empty, an error is returned.
	if secure != req.GetEmailCode() || secure == "" {
		return &response, status.Errorf(58990, "security code %v is incorrect", req.GetEmailCode())
	}

	// The purpose of this statement is to check if the user has enabled two-factor authentication. If the user has enabled
	// two-factor authentication, the statement will return true, and the code following this statement will be executed.
	if user.GetFactorSecure() {

		// The purpose of this code is to verify a two-factor authentication (2FA) code. The code is compared to a user's 2FA
		// secret, and if it does not match, an error is returned.
		if !totp.Validate(req.GetFactorCode(), user.GetFactorSecret()) {
			return &response, status.Error(115654, "invalid 2fa secure code")
		}
	}

	// This code checks to see if the protocol used by the contract is not the mainnet protocol. This is important to ensure
	// that the contract uses the correct protocol, as different protocols have different rules and requirements.
	if contract.GetProtocol() != pbspot.Protocol_MAINNET {

		// This code is used to get the price of a given asset. It takes two parameters, the base unit and the quote unit, and
		// requests the price of the asset. The GetPrice() function returns the price and an error if one is encountered, which
		// is then checked and handled.
		price, err := e.GetPrice(context.Background(), &pbspot.GetRequestPriceManual{
			BaseUnit:  chain.GetParentSymbol(),
			QuoteUnit: req.GetSymbol(),
		})
		if err != nil {
			return &response, err
		}

		// The purpose of this line of code is to assign the value returned by the GetPrice() function to the req.Price
		// variable. This variable may be used later in the program to calculate the total cost of a purchase, or to determine
		// the cost of an individual item.
		req.Price = price.GetPrice()

		// The purpose of this statement is to assign the value returned from the GetFees() method of the contract
		// object to the GetFees property of the chain object.
		chain.Fees = contract.GetFees()

		// The purpose of this statement is to calculate the fees that are associated with a withdrawal request. The statement
		// uses the GetFeesWithdraw() and GetPrice() functions to retrieve the fee rate and the price of the request,
		// respectively. It then uses decimal.New to create a new decimal object and multiply it by the price to get the fees
		// associated with the withdrawal request. Finally, it uses the Float() function to convert the resulting decimal
		// object into a float value for further calculations.
		fees = decimal.New(contract.GetFees()).Mul(req.GetPrice()).Float()

	} else {

		// The purpose of this line of code is to get the fee associated with withdrawing funds from a chain, such as a
		// blockchain. This line of code calls the GetFeesWithdraw() method, which retrieves the fee associated with
		// withdrawing funds from the chain.
		fees = chain.GetFees()
	}

	// This code is checking if any errors arise when withdrawing a certain quantity of a certain currency from a certain
	// platform or protocol. If an error occurs, the code returns an error response.
	if err := e.helperWithdraw(req.GetQuantity(), e.getReserve(req.GetSymbol(), req.GetPlatform(), contract.GetProtocol()), e.getBalance(req.GetSymbol(), auth), currency.GetMaxWithdraw(), currency.GetMinWithdraw(), fees); err != nil {
		return &response, err
	}

	// This if statement is checking to see if the address given by the request is the same as the address that it is attempting to send the request to.
	// If they are the same, the code will return an error indicating that the user cannot send from an address to the same address.
	if address := e.getAddress(auth, req.GetSymbol(), req.GetPlatform(), contract.GetProtocol()); address == strings.ToLower(req.GetAddress()) {
		return &response, status.Error(758690, "your cannot send from an address to the same address")
	}

	// This code is checking for an error when attempting to set a balance for a symbol with a given quantity. If there is
	// an error, the program will debug the error and return the response and an error.
	if err := e.setBalance(req.GetSymbol(), auth, req.GetQuantity(), pbspot.Balance_MINUS); e.Context.Debug(err) {
		return &response, err
	}

	// This code snippet is used to insert data into the 'transactions' table in a database. The code is using the Exec
	// method of the database to execute an SQL insert statement. The values of the transaction being inserted are provided
	// as parameters in the insert statement. Finally, the code checks for any errors that may have occurred during the
	// insertion, and returns an appropriate response.
	if _, err := e.Context.Db.Exec(`insert into transactions (symbol, value, price, "to", chain_id, platform, protocol, fees, user_id, assignment, type) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		req.GetSymbol(),
		req.GetQuantity(),
		req.GetPrice(),
		req.GetAddress(),
		req.GetId(),
		req.GetPlatform(),
		contract.GetProtocol(),
		chain.GetFees(),
		auth,
		pbspot.Assignment_WITHDRAWS,
		currency.GetType(),
	); err != nil {
		return &response, status.Error(554322, "transaction hash is already in the list, please contact support")
	}

	// This code checks if an error occurs when the setSecure function is called. If an error occurs, it returns an error
	// response and logs the error.
	if err := e.setSecure(ctx, true); err != nil {
		return &response, err
	}
	response.Success = true

	return &response, nil
}

// CancelWithdraw - This function is used to cancel a pending withdrawal request for a user. It checks for the user's ID and the request's
// ID in the database in order to validate the request, and if it is valid, it updates the status of the request to
// "CANCEL" and adds the withdrawn value back to the user's balance.
func (e *Service) CancelWithdraw(ctx context.Context, req *pbspot.CancelRequestWithdraw) (*pbspot.ResponseWithdraw, error) {

	// The purpose of this code is to declare a variable called response of type pbspot.ResponseWithdraw. This variable is
	// used to store a response from a withdrawal request.
	var (
		response pbspot.ResponseWithdraw
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This query is used to select the specified row from the transactions table based on the given parameters: id, status
	// and user_id. The purpose of this query is to retrieve the specified row from the database for further processing,
	// such as updating the status of the transaction or displaying the information to the user. The row is then closed,
	// which releases any resources associated with the query.
	row, err := e.Context.Db.Query(`select id, user_id, symbol, value from transactions where id = $1 and status = $2 and user_id = $3 order by id`, req.GetId(), pbspot.Status_PENDING, auth)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The purpose of this code is to check if there is a next row in a database. The if condition will evaluate to true if
	// there is a row after the current row and false if there is no next row.
	if row.Next() {

		// The purpose of the above line of code is to declare a variable 'item' of type 'pbspot.Transaction'. This is done in
		// order to create an instance of the pbspot.Transaction class, which can then be used to store and manipulate data
		// related to a transaction.
		var (
			item pbspot.Transaction
		)

		// This code is used to scan the rows of a database query and assign values to item.Id, item.UserId, item.Symbol and
		// item.Value variables. If an error occurs while scanning the rows, the error is returned in the response and the
		// function returns an error.
		if err = row.Scan(&item.Id, &item.UserId, &item.Symbol, &item.Value); err != nil {
			return &response, err
		}

		// This statement is updating the table "transactions" to set the status to "CANCEL" for a specific row identified by
		// "id" and "user_id". The purpose of this statement is to update the status of the transaction in the database.
		if _, err := e.Context.Db.Exec("update transactions set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), pbspot.Status_CANCEL); err != nil {
			return &response, err
		}

		// This code is checking for an error when setting a balance for a user's account. If an error occurs, it will log the
		// error and return an error response.
		if err := e.setBalance(item.GetSymbol(), item.GetUserId(), item.GetValue(), pbspot.Balance_PLUS); e.Context.Debug(err) {
			return &response, err
		}
	}

	return &response, nil
}

// CancelOrder - This function is used to cancel an order in a spot trading system. It takes in a context and a request object, and
// returns a response object and an error. It checks the status of the order, updates the order status to "CANCEL",
// updates the balance, and publishes a message.
func (e *Service) CancelOrder(ctx context.Context, req *pbspot.CancelRequestOrder) (*pbspot.ResponseOrder, error) {

	// The purpose of this code is to declare a variable named response of type pbspot.ResponseOrder. This is a type used to
	// store information about a response to an order, such as the status, order ID, and other details.
	var (
		response pbspot.ResponseOrder
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This query is used to fetch data from the orders table in the database. The query is parameterized to ensure that
	// only the desired records are returned. The parameters are the status, id, and user_id. The query also includes an
	// order by clause to ensure that the data is returned in a specific order. The data is then stored in the row variable
	// and the defer statement is used to close the row when the query is finished.
	row, err := e.Context.Db.Query(`select id, value, quantity, price, assigning, base_unit, quote_unit, user_id, create_at from orders where status = $1 and type = $2 and id = $3 and user_id = $4 order by id`, pbspot.Status_PENDING, pbasset.Type_SPOT, req.GetId(), auth)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The purpose of the following code is to check if there is a row available for retrieving data from. The `row.Next()`
	// method returns a boolean value indicating whether there is a row available. If the result is true, it means that a
	// row is available and can be used to retrieve data.
	if row.Next() {

		// The purpose of the 'var' statement is to declare a new variable, in this case "item", which is of type
		// "pbspot.Order". This allows the program to use the variable "item" to store values of type "pbspot.Order", such as
		// orders placed on an online store.
		var (
			item pbspot.Order
		)

		// This code is used to scan the row of a database table and assign the values to the relevant variables. The if
		// statement checks for any errors that may occur during the scanning process, and if an error is found, it will return an error response.
		if err = row.Scan(&item.Id, &item.Value, &item.Quantity, &item.Price, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt); err != nil {
			return &response, err
		}

		// The purpose of the code is to update the status of an order for a particular user in the database. It takes three
		// parameters: the ID of the order, the user ID, and the new status for the order. If the execution of the SQL query
		// fails, an error is returned.
		if _, err := e.Context.Db.Exec("update orders set status = $1 where type = $2 and id = $3 and user_id = $4;", pbspot.Status_CANCEL, pbasset.Type_SPOT, item.GetId(), item.GetUserId()); err != nil {
			return &response, err
		}

		// The switch statement is used to compare the value of a variable (in this case, item.Assigning) to a list of possible
		// values. If the value matches one of the values in the list, a specific action will be executed.
		switch item.Assigning {
		case pbspot.Assigning_BUY:

			// This code is setting the balance of a user for a given item. It is using the item's quote unit, user id, value and
			// price to calculate the new balance and then updating the balance using the pbspot.Balance_PLUS parameter. If there
			// is an error setting the balance, an error is returned.
			if err := e.setBalance(item.GetQuoteUnit(), item.GetUserId(), decimal.New(item.GetValue()).Mul(item.GetPrice()).Float(), pbspot.Balance_PLUS); err != nil {
				return &response, err
			}

			break
		case pbspot.Assigning_SELL:

			// This code is used to set a balance for a user in a particular base unit. The "if err" statement is used to check if
			// there is an error when setting the balance. If there is an error, the code will return an error message.
			if err := e.setBalance(item.GetBaseUnit(), item.GetUserId(), item.GetValue(), pbspot.Balance_PLUS); err != nil {
				return &response, err
			}

			break
		}

		// This code is intended to publish an item to an exchange with the routing key "order/cancel". If any errors occur
		// while attempting to publish the item, the error is returned and the response is returned.
		if err := e.Context.Publish(&item, "exchange", "order/cancel"); err != nil {
			return &response, err
		}

	} else {
		return &response, status.Error(11538, "the requested order does not exist")
	}
	response.Success = true

	return &response, nil
}
