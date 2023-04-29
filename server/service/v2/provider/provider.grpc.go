package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/service/v2/account"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
)

// GetSymbol - This function is used to get the symbol of a given currency pair (base unit and quote unit). It first checks if the
// base and quote currency exist in the database, then it checks if the pair exists in the database. If both checks pass,
// the success field of response is set to true and the response is returned. If either of the checks fail, an error is returned.
func (a *Service) GetSymbol(_ context.Context, req *pbprovider.GetRequestSymbol) (*pbprovider.ResponseSymbol, error) {

	// The purpose of this code is to declare two variables, response of type pbprovider.ResponseSymbol and exist of type bool.
	// The response variable will store the response from a service that provides stock symbol information, while to exist
	// boolean will keep track of whether the symbol exists or not.
	var (
		response pbprovider.ResponseSymbol
		exist    bool
	)

	// This piece of code checks if the base unit of the request is valid. If it is not valid, an error is returned with a
	// status code and a message.
	if row, err := a.QueryAsset(req.GetBaseUnit(), false); err != nil {
		return &response, status.Errorf(11584, "this base asset does not exist, %v", row.GetSymbol())
	}

	// The purpose of this code is to check if the requested currency exists and if it does not, then return an error
	// message with the appropriate status code. The if statement checks to see if the requested currency exists by using
	// the function getCurrency() with the parameters req.GetQuoteUnit() and false. If an error is returned, the error
	// message is set with the status code 11582 and the currency symbol is included in the message.
	if row, err := a.QueryAsset(req.GetQuoteUnit(), false); err != nil {
		return &response, status.Errorf(11582, "this quote asset does not exist, %v", row.GetSymbol())
	}

	// The purpose of this code is to check if the pair (base_unit and quote_unit) provided by the request exists in the
	// database. It uses a SQL query to check if the pair exists and then stores the result of the query in the boolean
	// variable "exist". It then uses an if statement to check if the query was successful and if the "exist" variable is
	// false. If either of these conditions are not met, the code returns an error.
	if err := a.Context.Db.QueryRow("select exists(select id from pairs where base_unit = $1 and quote_unit = $2)::bool", req.GetBaseUnit(), req.GetQuoteUnit()).Scan(&exist); err != nil || !exist {
		return &response, status.Errorf(11585, "this pair %v-%v does not exist", req.GetBaseUnit(), req.GetQuoteUnit())
	}

	// The purpose of response.Success = true is to indicate that a successful response was received. This is typically used
	// in programming to indicate that a given operation was successful and that no errors occurred.
	response.Success = true

	return &response, nil
}

// SetOrder - This function is a method of the Service struct. It is used to set an order for a user. It checks the authentication
// and authorization, checks the validity of the input, sets the order and sends back the response. It also handles
// errors encountered during the process.
func (a *Service) SetOrder(ctx context.Context, req *pbprovider.SetRequestOrder) (*pbprovider.ResponseOrder, error) {

	// The purpose of this code is to declare two variables of type pbprovider.ResponseOrder and types.Order respectively.
	// Declaring the variables allows them to be used in the code.
	var (
		response pbprovider.ResponseOrder
		order    types.Order
	)

	// Validates the type of the request and returns an error if it is invalid.
	if err := types.Type(req.GetType()); err != nil {
		return &response, err
	}

	// This code snippet checks if the request is authenticated by calling the Auth() method on the Context object. If the
	// authentication fails, the code returns an error.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// Validate that the requested base and quote units and type are valid for the given configuration before proceeding with the request.
	if err := a.queryValidatePair(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetType()); err != nil {
		return &response, err
	}

	// The purpose of this code is to create a Service object that uses the context stored in the variable e. The Service
	// object is then assigned to the variable migrate.
	_account := account.Service{
		Context: a.Context,
	}

	// This code is attempting to query a user from migrate using the provided authentication credentials (auth). If the
	// query fails, an error is returned.
	user, err := _account.QueryUser(auth)
	if err != nil {
		return nil, err
	}

	// This code is checking the user's status. If the user's status is not valid (GetStatus() returns false), it returns an
	// error message informing the user that their account and assets have been blocked and instructing them to contact
	// technical support for any questions.
	if !user.GetStatus() {
		return &response, status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions")
	}

	// This is setting the order quantity and value based on the request quantity and price.
	// The request quantity is used to set the order quantity, order type, and the order value is calculated by multiplying the request quantity by the request price.
	order.Quantity = req.GetQuantity()
	order.Value = req.GetQuantity()
	order.Type = req.GetType()

	// This is a switch statement that is used to evaluate the trade type of the request object. Depending on the trade
	// type, different actions can be taken. For example, if the trade type is "buy", the code may execute a certain set of
	// instructions to purchase the item, and if the trade type is "sell", the code may execute a different set of instructions to sell the item.
	switch req.GetTrading() {
	case types.TradingMarket:

		// The purpose of this code is to set the price of the order (order.Price) to the market price of the requested base
		// and quote units, assigning, and price, which is retrieved from the "e.getMarket" function.
		order.Price = a.queryMarket(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetType(), req.GetAssigning(), req.GetPrice())

		// This if statement is checking to see if the request is to buy something. If it is, it is calculating the quantity
		// and value of the order by dividing the quantity by the price.
		if req.GetAssigning() == types.AssigningBuy {
			order.Quantity, order.Value = decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float(), decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float()
		}

	case types.TradingLimit:

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
	order.Trading = req.GetTrading()
	order.Status = types.StatusPending
	order.CreateAt = time.Now().UTC().Format(time.RFC3339)

	// This code is checking for an error in the queryValidateOrder() function and if one is found, it returns an error response
	// and calls the Context.Error() method with the error. The quantity variable is used to store the result of queryValidateOrder(), which is used to complete the order.
	quantity, err := a.queryValidateOrder(&order)
	if err != nil {
		return &response, err
	}

	// This is a conditional statement used to set a new order and check for any errors that might occur. If an error is
	// encountered, the statement will return a response and an Error context to indicate that an error has occurred.
	if order.Id, err = a.writeOrder(&order); err != nil {
		return &response, err
	}

	// The switch statement is used to evaluate the value of the expression "order.GetAssigning()" and execute the
	// corresponding case statement. It is a type of conditional statement that allows a program to make decisions based on different conditions.
	switch order.GetAssigning() {
	case types.AssigningBuy:

		// This code snippet is likely a part of a function that processes an order. The purpose of the code is to use the
		// function "writeAsset()" to set the base unit and user ID of the order to false. If an error occurs during the process,
		// the code will return the response and an error message.
		if err := a.writeAsset(order.GetBaseUnit(), order.GetType(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		// This code is checking the balance of a user and attempting to subtract the specified quantity from it. If the
		// operation is successful, it will continue with the program. If an error occurs, it will return an error response.
		if err := a.WriteBalance(order.GetQuoteUnit(), order.GetType(), order.GetUserId(), quantity, types.BalanceMinus); err != nil {
			return &response, err
		}

		a.trade(&order, types.AssigningSell)

		break
	case types.AssigningSell:

		// This code snippet is likely a part of a function that processes an order. The purpose of the code is to use the
		// function "writeAsset()" to set the base unit and user ID of the order to false. If an error occurs during the process,
		// the code will return the response and an error message.
		if err := a.writeAsset(order.GetQuoteUnit(), order.GetType(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		// This code is checking the balance of a user and attempting to subtract the specified quantity from it. If the
		// operation is successful, it will continue with the program. If an error occurs, it will return an error response.
		if err := a.WriteBalance(order.GetBaseUnit(), order.GetType(), order.GetUserId(), quantity, types.BalanceMinus); err != nil {
			return &response, err
		}

		a.trade(&order, types.AssigningBuy)

		break
	default:
		return &response, status.Error(11588, "invalid assigning trade position")
	}

	// This statement is used to append an element to the "Fields" slice of the "response" struct. The element being
	// appended is the "order" struct.
	response.Fields = append(response.Fields, &order)

	return &response, nil
}

// GetMarkers - This function is part of a service that is used to retrieve marker symbols from a database. It takes in a context and
// a GetRequestMarkers object as input, executes a SQL query on the database, and returns a ResponseMarker which contains
// a list of marker symbols and an error if one occurred.
func (a *Service) GetMarkers(_ context.Context, req *pbprovider.GetRequestMarkers) (*pbprovider.ResponseMarker, error) {

	// The above code is declaring a variable named "response" and assigning it the type of pbprovider.ResponseMarker. This
	// allows the program to create an object of type pbprovider.ResponseMarker, which is a type of structure used to store and
	// process data in a specific way.
	var (
		response pbprovider.ResponseMarker
		maps     []string
	)

	// This code checks the request type and adds the appropriate condition to the map for further query.
	// If the request type is of type spot, the marker is set to true and type to the request type. Else, the 'group' is set to the action group.
	if req.GetType() == types.TypeSpot || req.GetType() == types.TypeCross {
		maps = append(maps, fmt.Sprintf(`where marker = %[2]v and type = '%[1]v'`, types.TypeSpot, true))
	} else {
		maps = append(maps, fmt.Sprintf(`where "group" = '%[1]v'`, types.GroupFiat))
	}

	// This code is querying a database for a certain symbol from the currencies table. The purpose of the code is to query
	// the database and check for an error. If an error is present, it will return an error response. If no error is
	// present, the rows will be closed.
	rows, err := a.Context.Db.Query(fmt.Sprintf(`select symbol from assets %s`, strings.Join(maps, " ")))
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

// SetAsset - This function is used to set an asset for a user. It takes in a context, a request for an asset, and returns a
// response and an error. It uses the auth to get an entropy and generate a new record with asset wallet data, address,
// and entropy. It then checks if the asset address has already been generated and if not, inserts the asset and wallet
// into the database. It then sets the success to true and returns the response and no error.
func (a *Service) SetAsset(ctx context.Context, req *pbprovider.SetRequestAsset) (*pbprovider.ResponseAsset, error) {

	// Response and cross are used to store various variables used in pbprovider.ResponseAsset and keypair.CrossChain respectively.
	var (
		response pbprovider.ResponseAsset
		cross    keypair.CrossChain
	)

	// The purpose of this code is to retrieve the authentication information associated with the given context (ctx). If
	// there is an error with retrieving the authentication information, then the error is returned.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// Validate the Group field in the request and return an error if.
	if err := types.Group(req.GetGroup()); err != nil {
		return &response, err
	}

	// Validates the request type and returns an error if there is an issue.
	if err := types.Type(req.GetType()); err != nil {
		return &response, err
	}

	// CheckBalance attempts to detect if a balance entry exists for the given symbol, user_id and type, and returns the result in the Success field of the response.
	_ = a.Context.Db.QueryRow("select exists(select value as balance from balances where symbol = $1 and user_id = $2 and type = $3)::bool", req.GetSymbol(), auth, req.GetType()).Scan(&response.Success)

	// If the request group is of the type 'GroupCrypto', then perform the code following this comment.
	if req.GetGroup() == types.GroupCrypto {

		// Service account is a Service struct used by account package to store context.
		_account := account.Service{
			Context: a.Context,
		}

		// The purpose of this code is to check if the platform provided by the request is a valid platform using the Platform()
		// function from the types package. If the platform is not valid, an error is returned.
		if err := types.Platform(req.GetPlatform()); err != nil {
			return &response, err
		}

		// This code is used to get entropy from a given authentication (auth) and assign it to the variable entropy. If an
		// error is encountered during this process, the code returns a response and an error is logged.
		entropy, err := _account.QueryEntropy(auth)
		if err != nil {
			return &response, err
		}

		// The code is attempting to create a new address using a secret, entropy, and platform. If there is an error, the
		// function will return the response and an error.
		if response.Address, _, err = cross.New(fmt.Sprintf("%v-&*39~763@)", a.Context.Secrets[1]), entropy, req.GetPlatform()); err != nil {
			return &response, err
		}

		// The code above is checking if the address returned from the e.getAddress() method is empty. If the address is
		// empty, the code inside the if statement will be executed.
		if address := a.QueryAddress(auth, req.GetPlatform()); len(address) == 0 {

			// This code is performing an SQL INSERT statement to add a new record to the 'wallets' table. The values being
			// inserted are the address, platform, and user_id from the request parameters. The query is then
			// executed and if there is an error, an error message is returned.
			if _, err = a.Context.Db.Exec("insert into wallets (address, platform, user_id) values ($1, $2, $3)", response.GetAddress(), req.GetPlatform(), auth); err != nil {
				return &response, err
			}
		}
	}

	// Check if the response was successful, if not return an error message.
	if !response.GetSuccess() {

		// The purpose of this code is to set the asset for the given symbol, using the provided authentication details. If
		// there is an error encountered while attempting to set the asset, the response variable is returned and the error is logged.
		if err := a.writeAsset(req.GetSymbol(), req.GetType(), auth, true); err != nil {
			return &response, err
		}

		response.Success = true
	}

	return &response, nil
}

// GetAsset - This function is used to get an asset from a database and retrieve related information, such as the balance, volume,
// and fees associated with it. It also gets information about the chains associated with the asset, such as the
// reserves, address, existence, and contract. Finally, it returns the response asset which contains all the gathered information.
func (a *Service) GetAsset(ctx context.Context, req *pbprovider.GetRequestAsset) (*pbprovider.ResponseAsset, error) {

	// The variable 'response' is declared as a type of pbprovider.ResponseAsset. This is used to store a response asset, which
	// is typically used to store the response of an API request. This allows the response to be accessed and manipulated by the code.
	var (
		response pbprovider.ResponseAsset
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// Validates the request type and returns an error if there is an issue.
	if err := types.Type(req.GetType()); err != nil {
		return &response, err
	}

	// The code is checking to see if an error occurred while attempting to get a currency. If there is an error, the
	// function will return the response and the error.
	row, err := a.QueryAsset(req.GetSymbol(), false)
	if err != nil {
		return &response, err
	}

	// This line of code sets the value of the Balance attribute of the row object to the balance of the account associated
	// with the symbol in the request object, which is authenticated using the auth parameter.
	row.Balance = a.QueryBalance(req.GetSymbol(), req.GetType(), auth)

	// This query is used to calculate the volume of orders for a particular symbol, with the given assigning and status,
	// for the given user. It is checking the base unit and quote unit against the given symbol and using the price to
	// convert between the two if necessary. It is then adding them together and using the coalesce function to return 0.00
	// if there is no data returned. Finally, it is scanning the result into the row.Volume field.
	_ = a.Context.Db.QueryRow(`select coalesce(sum(case when base_unit = $1 then value when quote_unit = $1 then value * price end), 0.00) as volume from orders where base_unit = $1 and type = $2 and status = $3 and user_id = $4 or quote_unit = $1 and type = $2 and status = $3 and user_id = $4`, req.GetSymbol(), req.GetType(), types.StatusPending, auth).Scan(&row.Volume)

	// Check if the request type is of type.
	if req.GetType() == types.TypeSpot {

		// This is a for loop in Go. The purpose of the loop is to iterate over each element in the "row.GetFields()" array.
		// The loop starts at index 0 and continues until it reaches the last element in the array. On each iteration, the loop
		// will execute the code inside the block.
		for i := 0; i < len(row.GetFields()); i++ {

			// The code above is checking if a chain exists by checking its ID. If the ID is greater than 0, the chain exists.
			if chain, _ := a.QueryChain(row.GetFields()[i], false); chain.GetId() > 0 {

				// chain.Rpc and chain.Address are being assigned empty strings. This is likely to reset the values of these variables to their default state.
				chain.Rpc, chain.Address = "", ""

				// This is an if statement used to determine if a contract has an ID greater than 0.
				// If it does, it will execute the code inside the if statement.
				if contract, _ := a.QueryContract(row.GetSymbol(), row.GetFields()[i]); contract.GetId() > 0 {

					// The purpose of the code is to assign two variables with the same value. The first variable, chain.Fees, is
					// assigned the value returned by the function contract.GetFees(). The second variable, contract.FeesWithdraw, is assigned the value 0.
					chain.Fees, contract.Fees = contract.GetFees(), 0

					// This code is used to get the price of a requested symbol given a base unit. It uses the GetPrice method from the e
					// object and passes in a context.Background() and a GetRequestPrice object containing the base unit and the
					// requested symbol. If the GetPrice method returns an error, the error is returned in the response and the Context.Error() method handles the error.
					price, err := a.GetPrice(context.Background(), &pbprovider.GetRequestPrice{BaseUnit: chain.GetParentSymbol(), QuoteUnit: req.GetSymbol()})
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
				chain.Reserve = a.QueryReserve(req.GetSymbol(), chain.GetPlatform(), chain.Contract.GetProtocol())

				// Switch statement to set chain address and balance based on the group.
				switch row.GetGroup() {
				case types.GroupCrypto:

					// The purpose of this code is to get the address for a particular platform from an external
					// source (e.getAddress) and assign the address to the chain.Address variable.
					if chain.Address = a.QueryAddress(auth, chain.GetPlatform()); len(chain.Address) > 0 {

						//The purpose of this code is to query a database for a row that matches the given parameters, which include the symbol, user_id, and type. It then stores the result in the 'chain.Exist' boolean variable.
						_ = a.Context.Db.QueryRow("select exists(select value as balance from balances where symbol = $1 and user_id = $2 and type = $3)::bool", req.GetSymbol(), auth, types.TypeSpot).Scan(&chain.Exist)
					}

				case types.GroupFiat:

					//The purpose of this code is to query a database for a row that matches the given parameters, which include the symbol, user_id, and type. It then stores the result in the 'chain.Exist' boolean variable.
					_ = a.Context.Db.QueryRow("select exists(select value as balance from balances where symbol = $1 and user_id = $2 and type = $3)::bool", req.GetSymbol(), auth, types.TypeSpot).Scan(&chain.Exist)
				}

				// This statement is used to add a new item, "chain", to the end of an existing slice of items, "row.Chains". Append
				// is a built-in function that allows you to add items to the end of a slice.
				row.Chains = append(row.Chains, chain)
			}
		}

		// row.Fields = make([]int64, 0) is used to create a slice of int64 elements with zero length. It will initialize the
		// slice with no elements in it.
		row.Fields = make([]int64, 0)
	} else {

		//The purpose of this code is to query a database for a row that matches the given parameters, which include the symbol, user_id, and type. It then stores the result in the 'row.Exist' boolean variable.
		_ = a.Context.Db.QueryRow("select exists(select value as balance from balances where symbol = $1 and user_id = $2 and type = $3)::bool", req.GetSymbol(), auth, req.GetType()).Scan(&row.Exist)
	}

	// This statement is appending a row to the Fields slice in the response variable. This allows the user to add more values to the slice.
	response.Fields = append(response.Fields, row)

	return &response, nil
}

// GetAssets - This function is a method of the Service struct that is used to query the database for currencies and their associated
// balance if the user is authenticated. It takes in a context and a GetRequestAssetsManual and returns a ResponseAsset
// and an error. It iterates through the result of the query and appends the currency and its balance (if authenticated) to the response.
func (a *Service) GetAssets(ctx context.Context, req *pbprovider.GetRequestAssets) (*pbprovider.ResponseAsset, error) {

	// The purpose of the code is to declare a variable called response with the type pbprovider.ResponseAsset. This variable
	// can then be used in the program to store a response asset from the pbprovider API.
	var (
		response pbprovider.ResponseAsset
		maps     []string
	)

	// Generate a condition based on the group and type given in the request.
	// If the group is specified, set the condition to where "group" = <group>.
	// Otherwise, set the condition to where type = <type>..
	if len(req.GetGroup()) > 0 {
		maps = append(maps, fmt.Sprintf(`where "group" = '%v'`, req.GetGroup()))
	} else {

		// If the request type is "Margin", set the request type to "Spot".
		if req.GetType() == types.TypeSpot || req.GetType() == types.TypeCross {
			maps = append(maps, fmt.Sprintf(`where type = '%v'`, types.TypeSpot))
		}
	}

	// This code is querying the database to select the columns id, name, symbol, and status from the table currencies. The
	// purpose of the code is to retrieve the information from the table currencies and store them in the variables rows and
	// err. If there is an error, the code will return the response and an error message. Finally, the defer rows.Close()
	// will close the rows of information when the function is finished executing.
	rows, err := a.Context.Db.Query(fmt.Sprintf(`select id, name, symbol, status from assets %s`, strings.Join(maps, " ")))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for rows.Next() statement is used in SQL queries to loop through the results of a query. It retrieves the next
	// row from the result set, and assigns the values of the row to variables specified in the query. This allows the
	// programmer to iterate through the result set, one row at a time, and process the data as needed.
	for rows.Next() {

		// The purpose of this code is to declare a variable called asset of the type types.Asset. This allows the code to
		// reference this type of asset later in the code.
		var (
			asset types.Asset
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
		if auth, err := a.Context.Auth(ctx); err == nil {

			// This code is checking the balance of a certain asset (identified by the symbol) from the account of the user
			// (identified by the auth variable) and assigning the balance to the asset.Balance variable if the balance is greater than 0.
			if balance := a.QueryBalance(asset.GetSymbol(), req.GetType(), auth); balance > 0 {
				asset.Balance = balance
			}
		}

		// This statement is used to append a field to the response.Fields array. It is used to add a new element to an array.
		// The element being added is the asset variable.
		response.Fields = append(response.Fields, &asset)
	}

	return &response, nil
}

// GetPairs - This function is used to retrieve pairs from the database based on a given symbol. It retrieves all pairs that have a
// base or quote unit that matches the given symbol. For each row, it scans the columns and sets the corresponding fields
// in a Pair object. It also sets the ratio and price of the pair and the status of the pair. Finally, it appends the
// Pair object to the response and returns it.
func (a *Service) GetPairs(_ context.Context, req *pbprovider.GetRequestPairs) (*pbprovider.ResponsePair, error) {

	// The purpose of this code is to declare a variable called response of type pbprovider.ResponsePair. This variable can then
	// be used to store data related to a response to a request for information or a request for action.
	var (
		response pbprovider.ResponsePair
	)

	// If the request type is "Margin", set the request type to "Spot".
	if req.GetType() == types.TypeCross {
		req.Type = types.TypeSpot
	}

	// This code is used to query the database for information from the pairs table where either the base_unit or the
	// quote_unit is equal to the value in the req.GetSymbol() variable. The purpose of this code is to retrieve data from
	// the database and store it in the response variable. The defer rows.Close() statement ensures that the rows are closed once the function is completed.
	rows, err := a.Context.Db.Query("select id, base_unit, quote_unit, base_decimal, quote_decimal, type, status from pairs where type = $1 and (base_unit = $2 or quote_unit = $2)", req.GetType(), req.GetSymbol())
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for loop with rows.Next() is used to iterate through a database query result set. This allows the code to loop
	// through each row of the result set and do something with the data from each row.
	for rows.Next() {

		// The purpose of the declaration above is to create a variable called "pair" of type "types.Pair". This variable will
		// be used to store a pair of values, such as two strings, two numbers, or two objects. This is often used in
		// programming to store related data in a single object.
		var (
			pair types.Pair
		)

		// This is an if statement which is used to assign the scanned rows from the database to the corresponding variables.
		// If an error occurs while scanning the rows, the statement will return an error as part of the response.
		if err := rows.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit, &pair.BaseDecimal, &pair.QuoteDecimal, &pair.Type, &pair.Status); err != nil {
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
		// ratio is assigned to the pair. The if statement checks is the ratio is returned by the queryRatio() function, and if
		// it is, the ok variable will be true, and the ratio will be assigned to the pair.
		if ratio, ok := a.queryRatio(pair.GetBaseUnit(), pair.GetQuoteUnit()); ok {
			pair.Ratio = ratio
		}

		// This if statement is used to check if the getPrice function returns a value. If it does, it assigns that value to
		// the Price field of the pair variable. The ok variable is a boolean which is used to determine if the getPrice
		// function returns a value or not. The ok variable will be true if the getPrice function returns a value, and false otherwise.
		if price, ok := a.queryPrice(pair.GetBaseUnit(), pair.GetQuoteUnit()); ok {
			pair.Price = price
		}

		// _status is used to indicate the current status of a process.
		var (
			_status bool
		)

		// This code executes a database query to retrieve the status of an asset using the provided symbol, base unit, and quote unit values.
		// If the asset status is retrieved successfully, it will be saved in the pair.Status variable, otherwise the function will return an error.
		// The code demonstrates the use of the QueryRow method to execute the query and the Scan method to scan the query result.
		// Error handling in this code ensures that the function will only return an error if something goes wrong during the query execution.
		if err := a.Context.Db.QueryRow("select not (false in (select status from assets where symbol in ($1, $2))) as result;", pair.GetBaseUnit(), pair.GetQuoteUnit()).Scan(&_status); err != nil {
			return &response, err
		}

		// If '_status' is false, set the 'Status' property of 'pair' to '_status'.
		if !_status {
			pair.Status = _status
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
func (a *Service) GetPair(_ context.Context, req *pbprovider.GetRequestPair) (*pbprovider.ResponsePair, error) {

	// The purpose of this code is to declare a variable called response of type pbprovider.ResponsePair. This is used to store
	// the response from a server in the form of a key-value pair.
	var (
		response pbprovider.ResponsePair
	)

	// This code is querying a database for a specific row in the table. The query is looking for a row with the specified
	// base_unit and quote_unit from the 'parameters' req.GetBaseUnit() and req.GetQuoteUnit(). If an error occurs, the error.
	// Finally, the row is closed with the defer keyword so that it is properly released back to the server.
	row, err := a.Context.Db.Query(`select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs where base_unit = $1 and quote_unit = $2`, req.GetBaseUnit(), req.GetQuoteUnit())
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The if statement is used to test if the result of row.Next() returns true. If it does, the code within the if
	// statement will be executed. The purpose of row.Next() is to advance the row pointer to the next row in the result set.
	if row.Next() {

		// The purpose of this code is to declare a variable called 'pair' of type 'types.Pair'. This variable can then be
		// used to store values of type 'types.Pair'.
		var (
			pair types.Pair
		)

		// This code is part of a larger program which likely retrieves data from a database. The purpose of this code is to
		// scan each row of the retrieved data and store the relevant information into a structure called "pair", which likely
		// holds data regarding currency pairs. The "if" statement is a check to make sure that the data was successfully read
		// and stored into the structure, and if not, it will return an error.
		if err := row.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit, &pair.Price, &pair.BaseDecimal, &pair.QuoteDecimal, &pair.Status); err != nil {
			return &response, err
		}

		// _status is used to indicate the current status of a process.
		var (
			_status bool
		)

		// This code executes a database query to retrieve the status of an asset using the provided symbol, base unit, and quote unit values.
		// If the asset status is retrieved successfully, it will be saved in the pair.Status variable, otherwise the function will return an error.
		// The code demonstrates the use of the QueryRow method to execute the query and the Scan method to scan the query result.
		// Error handling in this code ensures that the function will only return an error if something goes wrong during the query execution.
		if err := a.Context.Db.QueryRow("select not (false in (select status from assets where symbol in ($1, $2))) as result;", pair.GetBaseUnit(), pair.GetQuoteUnit()).Scan(&_status); err != nil {
			return &response, err
		}

		// If '_status' is false, set the 'Status' property of 'pair' to '_status'.
		if !_status {
			pair.Status = _status
		}

		// This statement is appending a pointer to the variable 'pair' to the array stored in the 'Fields' property of the
		// 'response' variable. This statement is used to add a new element to the 'Fields' array.
		response.Fields = append(response.Fields, &pair)
	}

	return &response, nil
}

// GetTicker - The purpose of this code is to create a service that retrieves OHLC (open-high-low-close) data from a database and
// returns it in a response. It is used to set limits on the number of results returned, filter the results based on a
// time range, perform calculations on the data, and store the results in an array.
func (a *Service) GetTicker(_ context.Context, req *pbprovider.GetRequestTicker) (*pbprovider.ResponseTicker, error) {

	// The purpose of this code is to create three variables with zero values: response, limit and maps. The response
	// variable is of type pbprovider.ResponseTicker, the limit variable is of type string, and the maps variable is of type
	// slice of strings.
	var (
		response pbprovider.ResponseTicker
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
		maps = append(maps, fmt.Sprintf(`and to_char(o.create_at::timestamp, 'yyyy-mm-dd hh24:mi:ss') < to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss')`, req.GetTo()))
	}

	// This code is used to query the database to return OHLC (open-high-low-close) data. The SQL query is using the
	// fmt.Sprintf function to substitute the variables (req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "),
	// help.Resolution(req.GetResolution()), limit) into the query. The query is then executed, and the results are stored
	// in the rows variable. Finally, the rows variable is closed at the end of the code.
	rows, err := a.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', o.create_at))::integer buckettime, first(o.price, o.create_at) as open, last(o.price, o.create_at) as close, first(o.price, o.price) as low, last(o.price, o.price) as high, sum(o.quantity) as volume, avg(o.price) as avg_price, o.base_unit, o.quote_unit from ohlcv as o where o.base_unit = '%[1]s' and o.quote_unit = '%[2]s' %[3]s group by buckettime, o.base_unit, o.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The purpose of the for rows.Next() loop is to iterate through the rows in a database table. It is used to perform
	// some action on each row of the table. This could include retrieving data from the row, updating data in the row, or
	// deleting the row.
	for rows.Next() {

		// The purpose of the variable "item" is to store data of type types.Ticker. This could be used to store an array of
		// candles or other data related to types.Ticker.
		var (
			item types.Ticker
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

	// The purpose of the following code is to declare a variable called stats of the type pbprovider.Stats. This variable will
	// be used to store information related to the pbprovider.Stats data type.
	var (
		stats types.Stats
	)

	// This code is used to fetch and analyze data from a database. It uses the QueryRow() method to retrieve data from the
	// database and then scan it into the stats variable. The code is specifically used to get the count, volume, low, high,
	// first and last values from the trades table for a given base unit and quote unit.
	_ = a.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from ohlcv as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

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

// SetTicker - The purpose of this code is to retrieve two candles with a given resolution from a spot exchange, add a new row to a
// database table, publish a message to an exchange on a specific topic, and append the returned values to a response array.
func (a *Service) SetTicker(_ context.Context, req *pbprovider.SetRequestTicker) (*pbprovider.ResponseTicker, error) {

	// The purpose of this code is to declare a variable called 'response' of type 'pbprovider.Response'. This variable will be
	// used to store data that is returned from a function call involving the 'pbprovider' library.
	var (
		response pbprovider.ResponseTicker
	)

	// This code is checking if the key provided in the request (req.GetKey()) matches the secret stored in the context
	// (s.Context.Secrets[2]). If the keys don't match, the code is returning an error with the code 654333 and the message
	// "the access key is incorrect". This is likely part of an authorization process to make sure only authorized users are
	// able to access a certain resource.
	if req.GetKey() != a.Context.Secrets[2] {
		return &response, status.Error(654333, "the access key is incorrect")
	}

	// This piece of code is inserting data into a database table. The purpose of this code is to add a new row to the
	// "ohlcv" table, based on the values stored in the params array. The five columns in the table are assigning,
	// base_unit, quote_unit, price, and quantity, and each of these is being populated with the corresponding value from
	// the params array. The code then checks for any errors that may have occurred while executing the query and returns if any are found.
	if _, err := a.Context.Db.Exec(`insert into ohlcv (assigning, base_unit, quote_unit, price, quantity) values ($1, $2, $3, $4, $5)`, req.GetAssigning(), req.GetBaseUnit(), req.GetQuoteUnit(), req.GetPrice(), req.GetValue()); a.Context.Debug(err) {
		return &response, err
	}

	// The for loop is used to iterate through each element in the Depth() array. The underscore is used to assign the index
	// number to a variable that is not used in the loop. The interval variable is used to access the contents of each
	// element in the Depth() array.
	for _, interval := range help.Depth() {

		// This code is used to retrieve two candles with a given resolution from a spot exchange. The purpose of the migrate,
		// err := a.GetTicker() line is to make a request to the spot exchange using the BaseUnit, QuoteUnit, Limit, and
		// Resolution parameters provided. The if err != nil { return err } line is used to check if there was an error with
		// the request and return that error if necessary.
		migrate, err := a.GetTicker(context.Background(), &pbprovider.GetRequestTicker{BaseUnit: req.GetBaseUnit(), QuoteUnit: req.GetQuoteUnit(), Limit: 2, Resolution: interval})
		if err != nil {
			return &response, err
		}

		// This code is used to publish a message to an exchange on a specific topic. The message is "migrate" and the topic is
		// "trade/ticker:interval". The purpose of this code is to send a message to the exchange,
		// action based on the message. The if statement is used to check for any errors that may occur during the publishing
		// of the message. If an error is encountered, it will be returned.
		if err := a.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/ticker:%v", interval)); err != nil {
			return &response, err
		}

		// The purpose of this statement is to append the values of the migrate.Fields array to the response.Fields array. This
		// statement essentially adds the values of migrate.Fields array to the existing values in response.Fields array.
		response.Fields = append(response.Fields, migrate.Fields...)
	}

	return &response, nil
}

// GetPrice - This function is part of a service that is used to get the price of a given asset. It takes in a context and a
// GetRequestPriceManual request object as parameters. The purpose of the function is to get the price of the asset from
// the given base and quote units, then return a ResponsePrice object. It also checks if the price can be obtained from
// the quote and base units in the opposite order. If so, the price is then calculated by taking the inverse of the price
// and rounding it to 8 decimal places. Finally, the ResponsePrice object is returned.
func (a *Service) GetPrice(_ context.Context, req *pbprovider.GetRequestPrice) (*pbprovider.ResponsePrice, error) {

	// The code above declares two variables: response of type pbprovider.ResponsePrice and ok of type bool. The purpose of this
	// code is to create two variables that can be used in the code that follows.
	var (
		response pbprovider.ResponsePrice
		ok       bool
	)

	// The purpose of this code is to check whether the price of a certain item has been successfully retrieved, and if so,
	// return the response and a nil (no error) value. The line starts by attempting to get the price of the item based on
	// the given base and quote units, and then assigns the response and the boolean value of ok to the variables'
	// response.Price and ok, respectively. Finally, the code returns the response and a nil value if ok is true.
	if response.Price, ok = a.queryPrice(req.GetBaseUnit(), req.GetQuoteUnit()); ok {
		return &response, nil
	}

	// This code is checking if the price of a product or service can be obtained given the quote and base units. If the
	// price can be obtained, the response will be rounded to 8 decimal places and stored in the response.Price variable.
	if response.Price, ok = a.queryPrice(req.GetQuoteUnit(), req.GetBaseUnit()); ok {
		response.Price = decimal.New(decimal.New(1).Div(response.Price).Float()).Round(8).Float()
	}

	return &response, nil
}

// GetOrders - This function is used to get orders from the database based on the parameters provided. It uses the req
// *pbprovider.GetRequestOrders which contains parameters such as the limit and page, assigning, owner, user_id, status
// and base_unit and quote_unit to query the database. The function then uses the query results to build a response which
// is a *pbprovider.ResponseOrder. This includes the count and volume of the query results, as well as a list of all the orders that match the query criteria.
func (a *Service) GetOrders(ctx context.Context, req *pbprovider.GetRequestOrders) (*pbprovider.ResponseOrder, error) {

	// The purpose of this is to declare two variables: response and maps. The variable response is of type
	// pbprovider.ResponseOrder, while the variable maps is of type string array.
	var (
		response pbprovider.ResponseOrder
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
	case types.AssigningBuy:
		maps = append(maps, fmt.Sprintf("where assigning = '%v'", types.AssigningBuy))
	case types.AssigningSell:
		maps = append(maps, fmt.Sprintf("where assigning = '%v'", types.AssigningSell))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = '%v' or assigning = '%v')", types.AssigningBuy, types.AssigningSell))
	}

	// Append type to maps array if the length of req.GetType() is greater than 0.
	if len(req.GetType()) > 0 {

		// Checks that the request contains valid type and returns a response and an error if it does.
		if err := types.Type(req.GetType()); err != nil {
			return &response, err
		}

		maps = append(maps, fmt.Sprintf("and type = '%v'", req.GetType()))
	}

	// This checks to see if the request (req) has an owner. If it does, the code after this statement will be executed.
	if req.GetOwner() {

		// This code is used to check the authentication of the user. The auth variable is used to store the authentication
		// credentials of the user, and the err variable is used to store any errors that might occur during the authentication
		// process. If an error occurs, the response and error is returned.
		auth, err := a.Context.Auth(ctx)
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
	case types.StatusFilled:
		maps = append(maps, fmt.Sprintf("and status = '%v'", types.StatusFilled))
	case types.StatusPending:
		maps = append(maps, fmt.Sprintf("and status = '%v'", types.StatusPending))
	case types.StatusCancel:
		maps = append(maps, fmt.Sprintf("and status = '%v'", types.StatusCancel))
	}

	// This code checks if the length of the base unit and the quote unit in the request are greater than 0. If they are, it
	// appends a string to the maps variable which includes a formatted SQL query containing the base and quote unit. This
	// is likely part of a larger SQL query used to search for data in a database.
	if len(req.GetBaseUnit()) > 0 && len(req.GetQuoteUnit()) > 0 {
		maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))
	}

	// The purpose of this code is to query the database to count the number of orders and total value of the orders in the
	// database. It then stores the count and volume in the response variable.
	_ = a.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count, sum(value) as volume from orders %s", strings.Join(maps, " "))).Scan(&response.Count, &response.Volume)

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
		rows, err := a.Context.Db.Query(fmt.Sprintf("select id, assigning, price, value, quantity, base_unit, quote_unit, user_id, create_at, type, status from orders %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through the rows of the result set. The rows.Next() command will return true if the
		// iteration is successful and false if the iteration has reached the end of the result set. The loop will continue to
		// execute until the rows.Next() returns false.
		for rows.Next() {

			// The purpose of the above code is to declare a variable called item with the type types.Order. This allows the
			// program to create an object of type types.Order and assign it to the item variable.
			var (
				item types.Order
			)

			// This code is scanning the rows returned from a database query and assigning the values to the variables in the item
			// struct. If an error is encountered during the scanning process, an error is returned.
			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Value, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt, &item.Type, &item.Status); err != nil {
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

// GetTrades - This function is a method of the Service type. It is used to get a list of transfers from a database. It takes a
// context and a request object as parameters. The request object contains information about the requested transfers,
// such as the limit, order ID and whether the request should only return transfers for a specific user. The function
// then queries the database for the requested transfers, and returns a ResponseTransfer object containing the relevant transfers.
func (a *Service) GetTrades(ctx context.Context, req *pbprovider.GetRequestTrades) (*pbprovider.ResponseTrade, error) {

	// The purpose of this code is to declare two variables: a variable called "response" of type "pbprovider.ResponseTrade"
	// and a variable called "maps" of type "string slice". This allows the program to store values in these two variables
	// and access them throughout the code.
	var (
		response pbprovider.ResponseTrade
		maps     []string
	)

	// This code checks if the request's limit is 0. If it is, it sets the request's limit to 30. This is likely done to
	// ensure that the request is not given an unlimited amount of data, which could cause performance issues.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is used to generate a query for a database. The purpose of the switch statement is to determine the value
	// of the assigning parameter and then generate the appropriate query based on the value. For example, if the value of
	// req.GetAssigning() is types.AssigningBuy, then the query generated will be "where assigning =
	// types.AssigningBuy". If the value of req.GetAssigning() is not specified, then the query generated will be "where
	// (assigning = types.AssigningBuy or assigning = types.AssigningSell)".
	switch req.GetAssigning() {
	case types.AssigningBuy:
		maps = append(maps, fmt.Sprintf("where assigning = '%v'", types.AssigningBuy))
	case types.AssigningSell:
		maps = append(maps, fmt.Sprintf("where assigning = '%v'", types.AssigningSell))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = '%v' or assigning = '%v')", types.AssigningBuy, types.AssigningSell))
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
		auth, err := a.Context.Auth(ctx)
		if err != nil {
			return &response, err
		}

		// The purpose of this code is to create a map which contains a key-value pair of "user_id" and the value of the
		// variable "auth". The variable "maps" is a slice of type string, and the function "append" is used to append the map created by the fmt.Sprintf to the slice.
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", auth))
	}

	// This code is used to query the table 'transfers' with the given parameters. It uses a fmt.Sprintf statement to format the
	// query string with the given parameters, then it uses the a.Context.Db.Query() to execute the query and store the
	// results into the rows variable. If an error occurs, it returns an error response. Finally, it closes the rows variable.
	rows, err := a.Context.Db.Query(fmt.Sprintf("select id, user_id, base_unit, quote_unit, price, quantity, assigning, fees, maker, create_at from trades %s order by id desc limit %d", strings.Join(maps, " "), req.GetLimit()))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The for loop is used to iterate through the rows in a database. The rows.Next() statement is used to move to the next
	// row in the result set. This loop allows you to access the data in each row and process it as needed.
	for rows.Next() {

		// The purpose of this code is to declare a variable called item, and assign it to a value of type types.Trade.
		// This is used in programming to store data in the form of a variable and access it at a later time.
		var (
			item types.Trade
		)

		// This code is part of a function that retrieves data from a database. The purpose of the if statement is to scan the
		// rows of the database and assign each row's values to the corresponding variables. If an error occurs while scanning
		// the rows, the function will return an error.
		if err = rows.Scan(&item.Id, &item.UserId, &item.BaseUnit, &item.QuoteUnit, &item.Price, &item.Quantity, &item.Assigning, &item.Fees, &item.Maker, &item.CreateAt); err != nil {
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

// GetTransactions - This function is a method in a service struct used to get a list of transactions and associated data from a database.
// The function takes a context.Context and a *pbprovider.GetRequestTransactions as parameters. The function returns a
// response of type *pbprovider.ResponseTransaction and an error. The function sets a default value of 30 for the limit if it
// is set to 0 and then uses a switch statement to filter out the transactions by type. The function also filters by
// symbol and status. The function then queries for the number of transactions and their associated data and returns the results.
func (a *Service) GetTransactions(ctx context.Context, req *pbprovider.GetRequestTransactions) (*pbprovider.ResponseTransaction, error) {

	// The purpose of the code snippet is to declare two variables, response and maps, of types pbprovider.ResponseTransaction and []string respectively.
	var (
		response pbprovider.ResponseTransaction
		maps     []string
	)

	// The purpose of this code is to set a default limit value if the limit value requested (req.GetLimit()) is equal to
	// zero. In this case, the default limit value is set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This switch statement is used to create a query condition based on the transaction type. Depending on the transaction
	// type, the query string will be amended to include the appropriate condition. If the transaction type is deposit, the
	// query string will include the condition where assignment = types.AssignmentDeposit. If the transaction type is withdraws,
	// the query string will include the condition where assignment = types.AssignmentWithdrawal. If the transaction type is
	// neither deposit nor withdraws, the query string will include the condition where (assignment = types.AssignmentWithdrawal or assignment = types.AssignmentDeposit).
	switch req.GetAssignment() {
	case types.AssignmentDeposit:
		maps = append(maps, fmt.Sprintf("where assignment = '%v'", types.AssignmentDeposit))
	case types.AssignmentWithdrawal:
		maps = append(maps, fmt.Sprintf("where assignment = '%v'", types.AssignmentWithdrawal))
	default:
		maps = append(maps, fmt.Sprintf("where (assignment = '%v' or assignment = '%v')", types.AssignmentWithdrawal, types.AssignmentDeposit))
	}

	// This code is checking the length of the request's symbol and, if greater than zero, appending a string to the maps
	// variable. The string contains the symbol from the request. This allows the code to filter out requests that do not
	// have a symbol provided and process only those that do.
	if len(req.GetSymbol()) > 0 {
		maps = append(maps, fmt.Sprintf("and symbol = '%v'", req.GetSymbol()))
	}

	// This line of code is appending a string to the maps variable. The string is formatted using the fmt.Sprintf()
	// function, and is used to set the status of the variable to a specific value (in this case, types.StatusReserve).
	// This is likely being used to filter out specific values from a list or array of values.
	maps = append(maps, fmt.Sprintf("and status != '%v'", types.StatusReserve))

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This line of code is appending a string to the maps slice. The string is formatted with the auth variable. The
	// purpose of this line of code is to add a key-value pair to the maps slice, where the key is "user_id" and the value
	// is the auth variable.
	maps = append(maps, fmt.Sprintf("and user_id = %v", auth))

	// The purpose of this code is to query a database and retrieve a count of the transactions that meet the criteria
	// specified in the maps variable. The result of the query is then stored in the response.Count variable.
	_ = a.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from transactions %s", strings.Join(maps, " "))).Scan(&response.Count)

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
		rows, err := a.Context.Db.Query(fmt.Sprintf(`select id, symbol, hash, value, price, fees, confirmation, "to", chain_id, user_id, assignment, "group", platform, protocol, status, error, create_at from transactions %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through the rows returned from a database query. The rows.Next() method is used to
		// read the next row from the query result. The for loop allows the programmer to loop through each row of the query
		// result and perform an action on the data returned.
		for rows.Next() {

			// The purpose of this variable declaration is to declare a variable named "item" of type "types.Transaction". This
			// variable can then be used to store and manipulate data of type "types.Transaction".
			var (
				item types.Transaction
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
				&item.Group,
				&item.Platform,
				&item.Protocol,
				&item.Status,
				&item.Error,
				&item.CreateAt,
			); err != nil {
				return &response, err
			}

			// This code is retrieving the chain from the given chain ID and storing it in the item.Chain variable. It then checks
			// if an error occurred while doing so and if there was an error, it will return a nil value and the error.
			item.Chain, err = a.QueryChain(item.GetChainId(), false)
			if err != nil {
				return nil, err
			}

			// This code checks the protocol associated with the item. If the protocol is not equal to the mainnet protocol, then
			// the fees associated with the item are multiplied by the item's price, and the result is stored as a float.
			if item.GetProtocol() != types.ProtocolMainnet {
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

// CancelOrder - This function is used to cancel an order in a spot trading system. It takes in a context and a request object, and
// returns a response object and an error. It checks the status of the order, updates the order status to "CANCEL",
// updates the balance, and publishes a message.
func (a *Service) CancelOrder(ctx context.Context, req *pbprovider.CancelRequestOrder) (*pbprovider.ResponseOrder, error) {

	// The purpose of this code is to declare a variable named response of type pbprovider.ResponseOrder. This is a type used to
	// store information about a response to an order, such as the status, order ID, and other details.
	var (
		response pbprovider.ResponseOrder
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This query is used to fetch data from the orders table in the database. The query is parameterized to ensure that
	// only the desired records are returned. The parameters are the status, id, and user_id. The query also includes an
	// order by clause to ensure that the data is returned in a specific order. The data is then stored in the row variable
	// and the defer statement is used to close the row when the query is finished.
	row, err := a.Context.Db.Query(`select id, value, quantity, price, assigning, base_unit, quote_unit, user_id, type, create_at from orders where id = $1 and status = $2 and user_id = $3 order by id`, req.GetId(), types.StatusPending, auth)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The purpose of the following code is to check if there is a row available for retrieving data from. The `row.Next()`
	// method returns a boolean value indicating whether there is a row available. If the result is true, it means that a
	// row is available and can be used to retrieve data.
	if row.Next() {

		// The purpose of the 'var' statement is to declare a new variable, in this case "item", which is of type
		// "types.Order". This allows the program to use the variable "item" to store values of type "types.Order", such as
		// orders placed on an online store.
		var (
			item types.Order
		)

		// This code is used to scan the row of a database table and assign the values to the relevant variables. The if
		// statement checks for any errors that may occur during the scanning process, and if an error is found, it will return an error response.
		if err = row.Scan(&item.Id, &item.Value, &item.Quantity, &item.Price, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.Type, &item.CreateAt); err != nil {
			return &response, err
		}

		// The purpose of the code is to update the status of an order for a particular user in the database. It takes three
		// parameters: the ID of the order, the user ID, and the new status for the order. If the execution of the SQL query
		// fails, an error is returned.
		if _, err := a.Context.Db.Exec("update orders set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), types.StatusCancel); err != nil {
			return &response, err
		}

		// The switch statement is used to compare the value of a variable (in this case, item.Assigning) to a list of possible
		// values. If the value matches one of the values in the list, a specific action will be executed.
		switch item.Assigning {
		case types.AssigningBuy:

			// This code is setting the balance of a user for a given item. It is using the item's quote unit, user id, value and
			// price to calculate the new balance and then updating the balance using the types.Balance_PLUS parameter. If there
			// is an error setting the balance, an error is returned.
			if err := a.WriteBalance(item.GetQuoteUnit(), item.GetType(), item.GetUserId(), decimal.New(item.GetValue()).Mul(item.GetPrice()).Float(), types.BalancePlus); err != nil {
				return &response, err
			}

			break
		case types.AssigningSell:

			// This code is used to set a balance for a user in a particular base unit. The "if err" statement is used to check if
			// there is an error when setting the balance. If there is an error, the code will return an error message.
			if err := a.WriteBalance(item.GetBaseUnit(), item.GetType(), item.GetUserId(), item.GetValue(), types.BalancePlus); err != nil {
				return &response, err
			}

			break
		}

		// This code is intended to publish an item to an exchange with the routing key "order/cancel". If any errors occur
		// while attempting to publish the item, the error is returned and the response is returned.
		if err := a.Context.Publish(&item, "exchange", "order/cancel"); err != nil {
			return &response, err
		}

	} else {
		return &response, status.Error(11538, "the requested order does not exist")
	}
	response.Success = true

	return &response, nil
}
