package stock

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto"

	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"google.golang.org/grpc/status"
)

// Service - The purpose of this code is to create a "Service" struct that contains a pointer to an assets.Context. This allows the
// service to access the context and any of the assets within the context.
type Service struct {
	Context *assets.Context
}

func (s *Service) Initialization() {
	go s.market()
}

// getAgent - This function is used to get an Agent based on the userId provided. It uses a SQL query to search for an Agent with
// the given userId and returns the Agent's details. It also handles errors in case there is no Agent with the given userId.
func (s *Service) getAgent(userId int64) (*pbstock.Agent, error) {

	var (
		response pbstock.Agent
	)

	// This block of code is used to query a database and return information based on a userId as an input. The query looks
	// for a row in the "agents" table that matches the userId. If there is a match, the code will scan the row and store
	// the values in the "response" variable, which is then returned. If there is no match, an error is returned.
	if err := s.Context.Db.QueryRow("select a.id, a.user_id, case when a.broker_id > 0 then b.name else a.name end as agent_name, a.broker_id, a.type, a.status, a.create_at from agents a left join agents b on b.id = a.broker_id where a.user_id = $1", userId).Scan(&response.Id, &response.UserId, &response.Name, &response.BrokerId, &response.Type, &response.Status, &response.CreateAt); err != nil {
		return &response, err
	}

	return &response, nil
}

// setBalance - This function is used to update the balance of a user in a database. Depending on the cross parameter, either the
// balance is increased (proto.Balance_PLUS) or decreased (proto.Balance_MINUS) by a given quantity. The balance is
// updated in the assets table of the database, using a query. Finally, an error is returned if an error occurred during the update.
func (s *Service) setBalance(symbol string, userId int64, quantity float64, cross proto.Balance) error {

	switch cross {
	case proto.Balance_PLUS:

		// The code above is an if statement that is used to update the balance of an asset with a given symbol and user_id in
		// a database. The statement executes an update query, passing in the values of symbol, quantity, and userId as
		// parameters to the query. If the query fails to execute, the if statement will return an error.
		if _, err := s.Context.Db.Exec("update assets set balance = balance + $2 where symbol = $1 and user_id = $3 and type = $4;", symbol, quantity, userId, proto.Type_STOCK); err != nil {
			return err
		}
		break
	case proto.Balance_MINUS:

		// This code is used to update the balance of a user's assets in a database. The code updates the user's balance by
		// subtracting the quantity given. The values being used to update the balance are stored in variables, and are passed
		// into the code as parameters ($1, $2, and $3). The code also checks for errors and returns an error if one is found.
		if _, err := s.Context.Db.Exec("update assets set balance = balance - $2 where symbol = $1 and user_id = $3 and type = $4;", symbol, quantity, userId, proto.Type_STOCK); err != nil {
			return err
		}
		break
	}

	return nil
}

// setAsset - This function is used to set a new asset for a given user. It takes in three parameters - a string symbol to identify
// the asset, an int64 userId to identify the user, and a boolean error indicating whether an error should be returned if
// the asset already exists. The function checks if the asset already exists in the database, and if it does not exist,
// it inserts it into the database. If the error boolean is true, it will return an error if the asset already exists. If
// the error boolean is false, it will return no error regardless of the asset's existence.
func (s *Service) setAsset(symbol string, userId int64, error bool) error {

	// The purpose of this code is to query the database for a specific asset with a given symbol and userId. The query is
	// then stored in a row variable and an error is checked for. If there is an error, it will be returned. Finally, the
	// row is closed when the code is finished.
	row, err := s.Context.Db.Query(`select id from assets where symbol = $1 and user_id = $2 and type = $3`, symbol, userId, proto.Type_STOCK)
	if err != nil {
		return err
	}
	defer row.Close()

	// The code is used to check if the row is valid. The '!' operator is used to check if the row is not valid. If the row
	// is not valid, the code will execute.
	if !row.Next() {

		// This code is inserting values into a database table called "assets" with the specific columns "user_id" and
		// "symbol". The purpose of this code is to save the values of userId and symbol into the table for future reference.
		if _, err = s.Context.Db.Exec("insert into assets (user_id, symbol, type) values ($1, $2, $3)", userId, symbol, proto.Type_STOCK); err != nil {
			return err
		}

		return nil
	}

	// The purpose of this code is to return an error status to indicate that a fiat asset has already been generated. The
	// code uses the status.Error() function to return an error with a specific status code (700991) and an error message
	// ("the fiat asset has already been generated").
	if error {
		return status.Error(700991, "the fiat asset has already been generated")
	}

	return nil
}

// getBalance - This function is used to query the balance of a user's assets by symbol. It takes a symbol and userID as parameters
// and queries the assets table in the database for the balance associated with that symbol and userID, then returns the balance.
func (s *Service) getBalance(symbol string, userId int64) (balance float64) {

	// This line of code is used to retrieve the balance from the assets table in a database. It takes in two parameters
	// (symbol and userId) and uses them to query the database. The result is then stored in the variable balance.
	_ = s.Context.Db.QueryRow("select balance as balance from assets where symbol = $1 and user_id = $2 and type = $3", symbol, userId, proto.Type_STOCK).Scan(&balance)
	return balance
}

// setOrder - This function is used to set an order in the database. It takes in a pointer to a pbstock.Order which contains the
// order's details, and inserts the data into the 'orders' table. It then returns the id of the newly created order and any potential errors.
func (s *Service) setOrder(order *pbstock.Order) (id int64, err error) {

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.getAgent(order.GetUserId())
	if err != nil {
		return id, err
	}

	if agent.GetBrokerId() > 0 {
		agent.Id = agent.GetBrokerId()
	}

	if err := s.Context.Db.QueryRow("insert into orders (assigning, base_unit, quote_unit, price, value, quantity, user_id, trading, broker_id, type) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id", order.GetAssigning(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetPrice(), order.GetQuantity(), order.GetValue(), order.GetUserId(), order.GetTrading(), agent.GetId(), proto.Type_STOCK).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

// getQuantity - This function is used to calculate the quantity of a financial asset based on its price and whether it is a
// cross-trade or not. The function takes in the assigning (buy or sell), the quantity, the price, and a boolean value to
// check if it is a cross-trade. If it is a cross-trade, the function will divide the quantity by the price. Otherwise,
// it will multiply the quantity by the price. The function then returns the calculated quantity.
func (s *Service) getQuantity(assigning proto.Assigning, quantity, price float64, cross bool) float64 {

	if cross {

		// The purpose of this code is to calculate the quantity of an item by dividing it by its price. This switch statement
		// checks the assigning value to make sure it is set to "BUY", and then uses the decimal.New() method to divide the
		// quantity by the price and convert it to a float.
		switch assigning {
		case proto.Assigning_BUY:
			quantity = decimal.New(quantity).Div(price).Float()
		}

		return quantity

	} else {

		// This switch statement is used to determine the quantity of a purchase. In this case, if the assigning variable is
		// set to proto.Assigning_BUY, then the quantity will be multiplied by the price to determine the total cost of the
		// purchase.
		switch assigning {
		case proto.Assigning_BUY:
			quantity = decimal.New(quantity).Mul(price).Float()
		}

		return quantity
	}
}

// setTrade - This function is used to set a trade in a database. It takes a series of orders (param) as an argument and performs
// various operations including inserting data into the database, calculating fees, and publishing order status and trade candles.
func (s *Service) setTrade(param ...*pbstock.Order) error {

	// This is a conditional statement that is used to check the value of the parameter at the index of 0. If the value of
	// the parameter at index 0 is equal to 0, then the function will return nil.
	if param[0].GetValue() == 0 {
		return nil
	}

	// This code is used to insert a new row of data into the trades table of a database. The values for the new row are
	// taken from the param[0] variable. If the insertion fails, an error is returned.
	if _, err := s.Context.Db.Exec(`insert into trades (assigning, base_unit, quote_unit, price, quantity) values ($1, $2, $3, $4, $5)`, param[0].GetAssigning(), param[0].GetBaseUnit(), param[0].GetQuoteUnit(), param[0].GetPrice(), param[0].GetValue()); err != nil {
		return err
	}

	// The purpose of this "for" loop is to loop through a sequence of numbers (in this case, 0 and 1) and execute a certain
	// set of instructions a certain number of times (in this case, twice).
	for i := 0; i < 2; i++ {

		// The purpose of the code snippet is to publish a particular order to an exchange with the routing key "order/status".
		// The if statement checks for any errors encountered while publishing the order, and returns an error if one occurs.
		if err := s.Context.Publish(s.getOrder(param[i].GetId()), "exchange", "order/status"); err != nil {
			return err
		}
	}

	// The for loop is used to iterate through each element in the Depth() array. The underscore is used to assign the index
	// number to a variable that is not used in the loop. The interval variable is used to access the contents of each
	// element in the Depth() array.
	for _, interval := range help.Depth() {

		// This code is used to retrieve two candles with a given resolution from a spot exchange. The purpose of the migrate,
		// err := e.GetCandles() line is to make a request to the spot exchange using the BaseUnit, QuoteUnit, Limit, and
		// Resolution parameters provided. The if err != nil { return err } line is used to check if there was an error with
		// the request and return that error if necessary.
		migrate, err := s.GetCandles(context.Background(), &pbstock.GetRequestCandles{BaseUnit: param[0].GetBaseUnit(), QuoteUnit: param[1].GetQuoteUnit(), Limit: 2, Resolution: interval})
		if err != nil {
			return err
		}

		// This code is used to publish a message to an exchange on a specific topic. The message is "migrate" and the topic is
		// "trade/candles:interval". The purpose of this code is to send a message to the exchange,
		// action based on the message. The if statement is used to check for any errors that may occur during the publishing
		// of the message. If an error is encountered, it will be returned.
		if err := s.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/candles:%v", interval)); err != nil {
			return err
		}
	}

	return nil
}

// getOrder - This function is used to retrieve an order from a database by its ID. It takes an int64 (id) as a parameter and
// returns a pointer to a "pbstock.Order" type. It uses the "QueryRow" method of the database to scan the selected row
// into the "order" variable and then returns the pointer to the order.
func (s *Service) getOrder(id int64) *pbstock.Order {

	var (
		order pbstock.Order
	)

	// This code is used to query a database for a single row of data matching the specified criteria (in this case, the "id
	// = $1" condition) and then assign the returned values to the specified variables (in this case, the fields of the
	// "order" struct). This allows the program to retrieve data from the database and store it in a convenient and organized format.
	_ = s.Context.Db.QueryRow("select id, value, quantity, price, assigning, user_id, base_unit, quote_unit, status, create_at from orders where id = $1 and type = $2", id, proto.Type_STOCK).Scan(&order.Id, &order.Value, &order.Quantity, &order.Price, &order.Assigning, &order.UserId, &order.BaseUnit, &order.QuoteUnit, &order.Status, &order.CreateAt)
	return &order
}
