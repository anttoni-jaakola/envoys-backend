package spot

import (
	"context"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbspot"
	"github.com/cryptogateway/backend-envoys/server/service/v2/account"
	"github.com/cryptogateway/backend-envoys/server/service/v2/provider"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc/status"
	"strings"
)

// SetWithdraw - This code is a function written in the Go programming language.
// The purpose of the function is to process a request to withdraw a certain amount of a given currency from an account.
// The function handles authentication, validation, and the execution of the withdrawal request.
// It also handles logging the transaction in the database, setting the security code in the context, and returning the correct response.
func (e *Service) SetWithdraw(ctx context.Context, req *pbspot.SetRequestWithdrawal) (*pbspot.ResponseWithdrawal, error) {

	// The purpose of this code is to declare two variables: response and fees. The first variable, response, is of type
	// pbspot.ResponseWithdrawal, and the second variable, fees, is of type float64.
	var (
		response pbspot.ResponseWithdrawal
		fees     float64
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// Service account is a service that provides access to the account api. It provides methods to create, update, and delete accounts.
	_account := account.Service{
		Context: e.Context,
	}

	// The if statement is used to check if the GetRefresh() function is returning a truthy (true) value. If it is, the code
	// within the if block will be executed.
	if req.GetRefresh() {

		// This code is checking for an error when setting the secure flag to false. If there is an error, it returns a
		// response and the error context.
		if err := _account.WriteSecure(ctx, false); err != nil {
			return &response, err
		}

		return &response, nil
	}

	//This code is checking to see if the value of the platform in the request is valid. If it is not valid, it will return an error message.
	if err := types.Platform(req.GetPlatform()); err != nil {
		return &response, err
	}

	// This code is attempting to query a user using a given authentication (auth). The QueryUser function is likely part of
	// a migration library and returns a user object and an error object. If there is an error, it is returned with a nil
	// value for the user object.
	user, err := _account.QueryUser(auth)
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

	// This code is part of an error handling process. The if statement checks to see if the validateInternalAsset method
	// returns an error when given the address from the request. If it does return an error, the code will return the
	// response and log the error using the Context.Error() method.
	if err := e.queryValidateInternal(req.GetAddress()); err != nil {
		return &response, err
	}

	// provide is used to create a Service provider with the given Context.
	_provider := provider.Service{
		Context: e.Context,
	}

	// This code is used to get the chain with the specified ID from the e.getChain() function. If an error occurs, the code
	// returns an error message indicating the chain array by the specified ID is currently unavailable.
	chain, err := _provider.QueryChain(req.GetId(), true)
	if err != nil {
		return &response, status.Errorf(11584, "the chain array by id %v is currently unavailable", req.GetId())
	}

	// This code is used to get the currency of a request. It checks if the currency is available in the request and if it
	// is not available, it returns an error message (status.Errorf(10029, "the currency requested array by id %v is
	// currently unavailable", req.GetSymbol())).
	currency, err := _provider.QueryAsset(req.GetSymbol(), false)
	if err != nil {
		return &response, status.Errorf(10029, "the asset requested array by id %v is currently unavailable", req.GetSymbol())
	}

	// The purpose of the code above is to retrieve a contract from a blockchain given a symbol and chain ID. It does this
	// by calling the getContract() function on the e variable, passing in the symbol from the req variable and the chain ID
	// from the chain variable. The result of this call is then stored in the contract variable.
	contract, _ := _provider.QueryContract(req.GetSymbol(), chain.GetId())
	if len(contract.GetProtocol()) == 0 {
		contract.Protocol = types.ProtocolMainnet
	}

	// This code is checking to see if the function QuerySecure() returns an error. If an error is returned, the code is
	// returning a response and an error message.
	secure, err := _account.QuerySecure(ctx)
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
	if contract.GetProtocol() != types.ProtocolMainnet {

		// This code is used to get the price of a given asset. It takes two parameters, the base unit and the quote unit, and
		// requests the price of the asset. The GetPrice() function returns the price and an error if one is encountered, which
		// is then checked and handled.
		price, err := pbprovider.NewApiClient(e.Context.GrpcClient).GetPrice(context.Background(), &pbprovider.GetRequestPrice{
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
	if err := e.queryValidateWithdrawal(req.GetQuantity(), _provider.QueryReserve(req.GetSymbol(), req.GetPlatform(), contract.GetProtocol()), _provider.QueryBalance(req.GetSymbol(), types.TypeSpot, auth), currency.GetMaxWithdraw(), currency.GetMinWithdraw(), fees); err != nil {
		return &response, err
	}

	// This if statement is checking to see if the address given by the request is the same as the address that it is attempting to send the request to.
	// If they are the same, the code will return an error indicating that the user cannot send from an address to the same address.
	if address := _provider.QueryAddress(auth, req.GetPlatform()); address == strings.ToLower(req.GetAddress()) {
		return &response, status.Error(758690, "your cannot send from an address to the same address")
	}

	// This code is checking for an error when attempting to set a balance for a symbol with a given quantity. If there is
	// an error, the program will debug the error and return the response and an error.
	if err := _provider.WriteBalance(req.GetSymbol(), types.TypeSpot, auth, req.GetQuantity(), types.BalanceMinus); e.Context.Debug(err) {
		return &response, err
	}

	// This code snippet is used to insert data into the 'transactions' table in a database. The code is using the Exec
	// method of the database to execute an SQL insert statement. The values of the transaction being inserted are provided
	// as parameters in the insert statement. Finally, the code checks for any errors that may have occurred during the
	// insertion, and returns an appropriate response.
	if _, err := e.Context.Db.Exec(`insert into transactions (symbol, value, price, "to", chain_id, platform, protocol, fees, user_id, assignment, "group") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		req.GetSymbol(),
		req.GetQuantity(),
		req.GetPrice(),
		req.GetAddress(),
		req.GetId(),
		req.GetPlatform(),
		contract.GetProtocol(),
		chain.GetFees(),
		auth,
		types.AssignmentWithdrawal,
		currency.GetGroup(),
	); err != nil {
		return &response, status.Error(554322, "transaction hash is already in the list, please contact support")
	}

	// This code checks if an error occurs when the setSecure function is called. If an error occurs, it returns an error
	// response and logs the error.
	if err := _account.WriteSecure(ctx, true); err != nil {
		return &response, err
	}
	response.Success = true

	return &response, nil
}

// CancelWithdraw - This function is used to cancel a pending withdrawal request for a user. It checks for the user's ID and the request's
// ID in the database in order to validate the request, and if it is valid, it updates the status of the request to
// "CANCEL" and adds the withdrawn value back to the user's balance.
func (e *Service) CancelWithdraw(ctx context.Context, req *pbspot.CancelRequestWithdrawal) (*pbspot.ResponseWithdrawal, error) {

	// The purpose of this code is to declare a variable called response of type pbspot.ResponseWithdrawal. This variable is
	// used to store a response from a withdrawal request.
	var (
		response pbspot.ResponseWithdrawal
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// Creates a service provider to be used in the given context, providing the necessary services for the application.
	_provider := provider.Service{
		Context: e.Context,
	}

	// This query is used to select the specified row from the transactions table based on the given parameters: id, status
	// and user_id. The purpose of this query is to retrieve the specified row from the database for further processing,
	// such as updating the status of the transaction or displaying the information to the user. The row is then closed,
	// which releases any resources associated with the query.
	row, err := e.Context.Db.Query(`select id, user_id, symbol, value from transactions where id = $1 and status = $2 and user_id = $3 or id = $1 and status = $4 and user_id = $3 order by id`, req.GetId(), types.StatusPending, auth, types.StatusFailed)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The purpose of this code is to check if there is a next row in a database. The if condition will evaluate to true if
	// there is a row after the current row and false if there is no next row.
	if row.Next() {

		// The purpose of the above line of code is to declare a variable 'item' of type 'types.Transaction'. This is done in
		// order to create an instance of the types.Transaction class, which can then be used to store and manipulate data
		// related to a transaction.
		var (
			item types.Transaction
		)

		// This code is used to scan the rows of a database query and assign values to item.Id, item.UserId, item.Symbol and
		// item.Value variables. If an error occurs while scanning the rows, the error is returned in the response and the
		// function returns an error.
		if err = row.Scan(&item.Id, &item.UserId, &item.Symbol, &item.Value); err != nil {
			return &response, err
		}

		// This statement is updating the table "transactions" to set the status to "CANCEL" for a specific row identified by
		// "id" and "user_id". The purpose of this statement is to update the status of the transaction in the database.
		if _, err := e.Context.Db.Exec("update transactions set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), types.StatusCancel); err != nil {
			return &response, err
		}

		// This code is checking for an error when setting a balance for a user's account. If an error occurs, it will log the
		// error and return an error response.
		if err := _provider.WriteBalance(item.GetSymbol(), types.TypeSpot, item.GetUserId(), item.GetValue(), types.BalancePlus); e.Context.Debug(err) {
			return &response, err
		}

		response.Success = true
	}

	return &response, nil
}
