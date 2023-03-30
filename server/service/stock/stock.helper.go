package stock

import (
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto"

	"google.golang.org/grpc/status"
)

// helperOrder - This function is a helper function for placing orders in a Spot trading service. It performs checks on the order
// before it is submitted, such as checking the price is not 0, checking the user has enough funds to cover the order,
// and ensuring the quantity of the order is within the predetermined range. If any of these checks fail, an error is
// returned. Otherwise, the order is accepted and the quantity is returned.
func (s *Service) helperOrder(order *proto.Order) (summary float64, err error) {

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

		// This line of code is used to get the balance of the user in a given quote unit. It is used to determine the amount
		// of funds available to the user for a particular order. The getBalance() method takes in two parameters, the quote
		// unit and the user id, and returns the balance for the user in the specified quote unit.
		balance := s.getBalance(order.GetQuoteUnit(), order.GetUserId())

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

		// The purpose of this code is to get the balance of the user from the order object. The order object contains the base
		// unit and the userId of the user, which are used to call the getBalance() method of the e object to get the balance.
		balance := s.getBalance(order.GetBaseUnit(), order.GetUserId())

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
