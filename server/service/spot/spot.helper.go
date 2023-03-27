package spot

import (
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto"

	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"google.golang.org/grpc/status"
	"strconv"
)

// helperSymbol - This function checks whether a given asset symbol exists and is enabled. It returns an error if the asset symbol
// does not exist or is disabled.
func (e *Service) helperSymbol(symbol string) error {

	// This code is checking if a given currency exists and is enabled. If either of these conditions are not met, it
	// returns an error with status 11580 and a message informing the user that the asset does not exist or is disabled.
	if _, err := e.getAsset(symbol, true); err != nil {
		return status.Errorf(11580, "this %v asset does not exist, or this currency is disabled", symbol)
	}

	return nil
}

// helperPair - This function is used to check for the availability of a certain trading pair. It takes two strings, base and quote,
// and queries the database. If the query returns an integer greater than 0, then the pair is temporarily unavailable and
// an error is returned to the caller. Otherwise, the function returns nil.
func (e *Service) helperPair(base, quote string) error {

	var (
		id int64
	)

	// This code is checking to see if a pair (base and quote units) exists in the database and is set to "false" for the
	// status. If it does exist and is set to false, it returns an error that the pair is temporarily unavailable.
	if _ = e.Context.Db.QueryRow("select id from pairs where base_unit = $1 and quote_unit = $2 and status = $3", base, quote, false).Scan(&id); id > 0 {
		return status.Errorf(21605, "this pair %v-%v is temporarily unavailable", base, quote)
	}

	return nil
}

// helperWithdraw - This function is used to validate a withdrawal request. It checks to make sure that the requested withdrawal amount is
// not greater than the reserve, the balance, the maximum, and the minimum, and it also takes into account any fees that
// the user might have to pay. If any of the conditions are not met, the function returns an error.
func (e *Service) helperWithdraw(quantity, reserve, balance, max, min, fees float64) error {

	// This statement is creating a new variable called proportion that stores the result of the operation
	// decimal.New(min).Add(fees).Float(). The operation decimal.New(min) creates a new decimal from the value of min, and
	// then the Add(fees) method adds the value of fees to the new decimal. Finally, the Float() method converts the
	// resulting decimal to a float value, which is then stored in the variable proportion.
	var (
		proportion = decimal.New(min).Add(fees).Float()
	)

	// This code is used to check if the claimed amount is greater than the reserve. If it is, it will return an error
	// message with status code 47784.
	if quantity > reserve {
		return status.Errorf(47784, "the claimed amount %v is greater than the reserve %v itself", quantity, reserve)
	}

	// This code checks if the requested quantity is more than the available balance. If it is greater than the balance, it
	// returns an error message with the status code 48584. This prevents users from spending more money than they have.
	if quantity > balance {
		return status.Errorf(48584, "the claimed amount %v is more than what you have on your balance %v", quantity, balance)
	}

	// This code checks if the quantity is less than the proportion and, if it is, it returns an error indicating that the
	// withdrawal amount must not be less than the minimum amount.
	if quantity < proportion {
		return status.Errorf(48880, "the withdrawal amount %v must not be less than the minimum amount: %v", quantity, proportion)
	}

	// This code is used to check if the quantity declared for withdrawal is greater than the maximum allowed. If it is, an
	// error is returned with an appropriate error message.
	if quantity > max {
		return status.Errorf(70083, "the amount %v declared for withdrawal should not be more than allowed %v", quantity, max)
	}

	return nil
}

// helperInternalAsset - This function is used to check if a given address is an internal asset. It queries the database to see if the address
// exists in the wallets table, and if it does, it returns an error indicating that the address is an internal asset and
// that another address should be used.
func (e *Service) helperInternalAsset(address string) error {
	var (
		exist bool
	)

	// This code is used to check if a particular address exists in the wallets table of a database. The code is querying
	// the database for a row with the same address as the one being passed to the query. The result of the query is then
	// stored in the bool variable exist.
	_ = e.Context.Db.QueryRow("select exists(select id from wallets where lower(address) = lower($1))::bool", address).Scan(&exist)

	// This code is checking to see if an address exists, and if it does, it will return an error message. The error message
	// tells the user that they cannot use the address as it is internal, and they should use another address.
	if exist {
		return status.Errorf(717883, "you cannot use this address %v, this address is internal, please use another address", address)
	}

	return nil
}

// helperOrder - This function is a helper function for placing orders in a Spot trading service. It performs checks on the order
// before it is submitted, such as checking the price is not 0, checking the user has enough funds to cover the order,
// and ensuring the quantity of the order is within the predetermined range. If any of these checks fail, an error is
// returned. Otherwise, the order is accepted and the quantity is returned.
func (e *Service) helperOrder(order *pbspot.Order) (summary float64, err error) {

	// This code checks if the order's price is 0, and if it is, it returns an error message (65790) with the impossible
	// price that is being requested. This helps to identify errors in the order's price and allows for more accurate
	// debugging of the program.
	if order.GetPrice() == 0 {
		return 0, status.Errorf(65790, "impossible price %v", order.GetPrice())
	}

	// This switch statement is used to check the value of the GetAssigning() method on the order object. Depending on the
	// value of GetAssigning(), different code blocks may be executed.
	switch order.GetAssigning() {
	case proto.Assigning_BUY:

		// The purpose of this code is to calculate the total cost of an order, given the quantity and price of a product. The
		// code uses the decimal library to get the quantity and price of the order, and then multiplies them together to
		// calculate the total cost.
		quantity := decimal.New(order.GetQuantity()).Mul(order.GetPrice()).Float()

		// This code is checking the range of a given quantity, and returning an error if the quantity is not within the
		// specified range. The min and max variables represent the minimum and maximum values of the quantity, while the ok
		// variable indicates whether the range is valid. If the range is invalid, the code will return an error with a message
		// containing the minimum and maximum values.
		if min, max, ok := e.getRange(order.GetQuoteUnit(), quantity); !ok {
			return 0, status.Errorf(11623, "[quote]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		// This line of code is used to get the balance of the user in a given quote unit. It is used to determine the amount
		// of funds available to the user for a particular order. The getBalance() method takes in two parameters, the quote
		// unit and the user id, and returns the balance for the user in the specified quote unit.
		balance := e.getBalance(order.GetQuoteUnit(), order.GetUserId())

		// This statement is an if-statement that is used to check if the quantity is greater than the balance or if the
		// order's quantity is equal to 0. If either of these conditions are true, then the statement will return a value of 0,
		// along with an error message. The purpose of this statement is to ensure that a user does not place an order with
		// insufficient funds and to inform them if they have attempted to do so.
		if quantity > balance || order.GetQuantity() == 0 {
			return 0, status.Error(11586, "[quote]: there is not enough funds on your asset balance to place an order")
		}

		return quantity, nil

	case proto.Assigning_SELL:

		// This statement retrieves the quantity of an order from the order object and assigns it to the variable "quantity".
		quantity := order.GetQuantity()

		// This code is checking to see if the order's base unit and quantity meet a certain minimum and maximum trading
		// amount. If the order does not meet the requirements, an error is returned with a message that states the minimum and
		// maximum trading amounts.
		if min, max, ok := e.getRange(order.GetBaseUnit(), order.GetQuantity()); !ok {
			return 0, status.Errorf(11587, "[base]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		// The purpose of this code is to get the balance of the user from the order object. The order object contains the base
		// unit and the userId of the user, which are used to call the getBalance() method of the e object to get the balance.
		balance := e.getBalance(order.GetBaseUnit(), order.GetUserId())

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
