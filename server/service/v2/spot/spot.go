package spot

import (
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"google.golang.org/grpc/status"
)

// Service - The purpose of the Service struct is to store data related to a service, such as the Context, run and wait maps, and
// the block map. The Context is a pointer to an assets Context, which contains information about the service. The run
// and wait maps are booleans that indicate whether the service is running or waiting for an action. The block map is an
// integer that stores the block number associated with a particular service.
type Service struct {
	Context *assets.Context

	run, wait map[int64]bool
	block     map[int64]int64
}

// Initialization - The code initializes a Service object and runs six concurrent functions: deposit(), withdrawal(), and reward().
func (e *Service) Initialization() {
	go e.deposit()
	go e.withdrawal()
	go e.reward()
}

// queryValidateWithdraw - This function is used to validate a withdrawal request. It checks to make sure that the requested withdrawal amount is
// not greater than the reserve, the balance, the maximum, and the minimum, and it also takes into account any fees that
// the user might have to pay. If any of the conditions are not met, the function returns an error.
func (e *Service) queryValidateWithdrawal(quantity, reserve, balance, max, min, fees float64) error {

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

// queryValidateInternal - This function is used to check if a given address is an internal asset. It queries the database to see if the address
// exists in the wallets table, and if it does, it returns an error indicating that the address is an internal asset and
// that another address should be used.
func (e *Service) queryValidateInternal(address string) error {
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

// done - This function is used to mark an item with a given ID as done. The wait map is a collection of items with an
// associated boolean value indicating whether it is done or not. The function sets the value of the item with the given
// ID to true, thus marking it as done.
func (e *Service) done(id int64) {
	e.wait[id] = true
}
