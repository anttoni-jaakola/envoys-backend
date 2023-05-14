package provider

import (
	"context"

	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
)

// trade - This function is used to replay a trade init. It takes an order and a side (BID or ASK) as parameters. It then queries
// the database for orders with the same base unit, quote unit and user ID, and with a status of "PENDING". It then
// iterates through the results and checks if the order's price is higher than the item's price for a BID position and
// lower for an ASK position. If this is the case, it calls the replayTradeProcess() function. Finally, it logs any matches or failed matches.
func (a *Service) trade(order *types.Order, assigning string) {

	// This code is checking for an error when publishing to the exchange. If an error occurs, the code is printing out the
	// error and returning.
	if err := a.Context.Publish(order, "exchange", "order/create"); a.Context.Debug(err) {
		return
	}

	// This code is querying the "orders" table in a database for data that matches the given parameters. It is using the
	// parameters given to query for a specific set of data from the "orders" table. It is using the $1, $2, $3, $4, $5 and
	// $6 to represent the given parameters. The query is also ordering the results by the "id" column. It is checking for
	// errors and deferring the closing of the rows.
	rows, err := a.Context.Db.Query(`select id, assigning, base_unit, quote_unit, value, quantity, price, user_id, type, status from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and user_id != $4 and type = $5 and status = $6 order by id`, assigning, order.GetBaseUnit(), order.GetQuoteUnit(), order.GetUserId(), order.GetType(), types.StatusPending)
	if a.Context.Debug(err) {
		return
	}
	defer rows.Close()

	// The purpose of the for loop is to iterate over a set of rows from a database query. The rows.Next() function advances
	// the iterator to the next row in the result set, returning false when there are no more rows to iterate over.
	for rows.Next() {

		// Item is a variable of type Order from the types package, used to store a reference to an Order.
		var (
			item types.Order
		)

		// This code is attempting to scan the rows of a database table, assigning each column value to a variable (item.Id,
		// item.Assigning, etc.). The if statement is checking for any errors that might occur when scanning the rows and
		// returning any errors that might be present.
		if err = rows.Scan(&item.Id, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.Value, &item.Quantity, &item.Price, &item.UserId, &item.Type, &item.Status); a.Context.Debug(err) {
			return
		}

		// Check if the order is in a pending status and update the order value accordingly.
		if row := a.queryOrder(order.GetId()); row.GetStatus() == types.StatusPending {
			order.Value = row.GetValue()
		}

		// This switch statement is used to check for a match between the order and item prices, depending on the side of the
		// trade (bid or ask). If the order and item prices match, the trade process is replayed. If not, a message is logged
		// for the user. If the side is invalid, an error is returned.
		switch assigning {

		case types.AssigningBuy: // Buy at BID price.

			// This code checks whether the price of an order is greater than or equal to the price of an item. If it is, it will
			// log a message and call the replayTradeProcess function. If it is not, it will log another message.
			if order.GetPrice() >= item.GetPrice() {
				a.Context.Logger.Infof("[BID]: (item [%v]) >= (order [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				// A switch statement is used to evaluate the type of order provided and executes the appropriate processing.
				// If the order type is either Spot or Stock, the defaultProcess function is called; otherwise the marginProcess function is called if the order type is Margin.
				switch order.GetType() {
				case types.TypeSpot, types.TypeStock:
					a.defaultProcess(types.AssigningBuy, order, &item)
					break
				case types.TypeCross:
					a.marginProcess(types.AssigningBuy, order, &item)
					break
				}

			} else {
				a.Context.Logger.Infof("[BID]: no matches found: (item [%v]) >= (order [%v])", order.GetPrice(), item.GetPrice())
			}

			break

		case types.AssigningSell: // Sell at ASK price.

			// This code is checking if the price of an order is lower than or equal to the price of an item. If it is, it will
			// log an informational message and call the replayTradeProcess() method. If not, it will log a different
			// informational message.
			if order.GetPrice() <= item.GetPrice() {
				a.Context.Logger.Infof("[ASK]: (order [%v]) <= (item [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				// A switch statement is used to evaluate the type of order provided and executes the appropriate processing.
				// If the order type is either Spot or Stock, the defaultProcess function is called; otherwise the marginProcess function is called if the order type is Margin.
				switch order.GetType() {
				case types.TypeSpot, types.TypeStock:
					a.defaultProcess(types.AssigningSell, order, &item)
					break
				case types.TypeCross:
					a.marginProcess(types.AssigningSell, order, &item)
					break
				}

			} else {
				a.Context.Logger.Infof("[ASK]: no matches found: (order [%v]) <= (item [%v])", order.GetPrice(), item.GetPrice())
			}

			break
		default:
			if err := a.Context.Debug(status.Error(11589, "invalid assigning trade position")); err {
				return
			}
		}
	}

	// The purpose of this code is to check for errors when using the rows.Err() function. It returns an error if there is
	// one and returns the function if this is the case.
	if err = rows.Err(); a.Context.Debug(err) {
		return
	}
}

// defaultProcess - This function is used to replay a trade process. It updates two orders with different amounts to determine the result
// of a trade. It updates the order status in the database with pending in to filled, updates the balance by adding the
// amount of the order to the balance, and sends a mail. In addition, it logs information about the trade.
func (a *Service) defaultProcess(assigning string, params ...*types.Order) {

	// The purpose of this code is to declare two variables, instance and migrate. The variable instance is declared as an
	// integer, and migrate is declared as a query.Migrate object with the Context field set to the value of the variable e.Context.
	var (
		price    float64
		instance int
		migrate  = query.Migrate{
			Context: a.Context,
		}
	)

	// This code is checking whether the value of the first parameter is greater than or equal to the value of the second
	// parameter. If it is, the instance variable is set to 1.
	if params[0].GetValue() >= params[1].GetValue() {
		instance = 1
	}

	// This switch statement is used to determine the price based on the side (bid or ask) of an order. The switch statement
	// checks for the side of the order and assigns the price accordingly, using the params array. If the side is BID, it
	// will assign the price from the second element in the params array. If the side is ASK, it will assign the price from
	// the first element in the params array.
	switch assigning {
	case types.AssigningBuy:
		price = params[1].GetPrice()
	case types.AssigningSell:
		price = params[0].GetPrice()
	}

	// This code is used to update an order status from pending to filled when the order is completed. It also updates the
	// quantity of the orders and sets the necessary parameters for the order. Finally, it logs the parameters of the order.
	if params[instance].GetValue() > 0 {

		// The purpose of the for loop is to iterate over the parameters passed in and update the "value" of the specified
		// order in the database. It also sets the status of the order to FILLED if the value is equal to 0. The code also
		// checks for any errors that may occur during the process. Lastly, the code sends an email to the user associated with the order once the order is filled.
		for i := 0; i < 2; i++ {

			// The purpose of this code is to declare a variable named "value" of type float64. This variable can be used to store a decimal number, such as 3.14159.
			var (
				value float64
			)

			// This if statement is used to update the "value" of a particular order in the database. The parameters passed in are
			// used in the query to find the specific order to update. If the query is successful, the "value" of the order is
			// stored in the "value" variable and the function will continue. If the query fails, the function will return.
			if err := a.Context.Db.QueryRow("update orders set value = value - $2 where id = $1 and type = $3 and status = $4 returning value;", params[i].GetId(), params[instance].GetValue(), params[i].GetType(), types.StatusPending).Scan(&value); a.Context.Debug(err) {
				return
			}

			if value == 0 {

				// This code is performing an update on the orders table in a database. It is setting the status of the order with the
				// specified ID to the specified status (in this case, FILLED). The code is also checking for any errors that may
				// occur during the process. If an error is found, the code will return without proceeding.
				if _, err := a.Context.Db.Exec("update orders set status = $3 where id = $1 and type = $2;", params[i].GetId(), params[i].GetType(), types.StatusFilled); a.Context.Debug(err) {
					return
				}

				go migrate.SendMail(params[i].GetUserId(), "order_filled", params[i].GetId(), a.queryQuantity(params[i].GetAssigning(), params[i].GetQuantity(), price, false), params[i].GetBaseUnit(), params[i].GetQuoteUnit(), params[i].GetAssigning())
			}
		}

		switch params[1].GetAssigning() {
		case types.AssigningBuy:

			// Order trades logs.
			quantity, err := a.writeTrade(params[0].GetId(), params[0].GetQuoteUnit(), params[instance].GetValue(), price, true)
			if a.Context.Debug(err) {
				return
			}

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error, the function will return without doing anything.
			if err := a.WriteBalance(params[0].GetQuoteUnit(), params[0].GetType(), params[0].GetUserId(), quantity, types.BalancePlus); a.Context.Debug(err) {
				return
			}

			// Order trades logs.
			quantity, err = a.writeTrade(params[1].GetId(), params[0].GetBaseUnit(), params[instance].GetValue(), price, false)
			if a.Context.Debug(err) {
				return
			}

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error, the function will return without doing anything.
			if err := a.WriteBalance(params[0].GetBaseUnit(), params[0].GetType(), params[1].GetUserId(), quantity, types.BalancePlus); err != nil {
				return
			}

			break
		case types.AssigningSell:

			// Order trades logs.
			quantity, err := a.writeTrade(params[0].GetId(), params[0].GetBaseUnit(), params[instance].GetValue(), price, false)
			if a.Context.Debug(err) {
				return
			}

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error, the function will return without doing anything.
			if err := a.WriteBalance(params[0].GetBaseUnit(), params[0].GetType(), params[0].GetUserId(), quantity, types.BalancePlus); a.Context.Debug(err) {
				return
			}

			// Order trades logs.
			quantity, err = a.writeTrade(params[1].GetId(), params[0].GetQuoteUnit(), params[instance].GetValue(), price, true)
			if a.Context.Debug(err) {
				return
			}

			// This code is part of a function that allows the user to set the balance of a certain item to a certain quantity.
			// The purpose of the if statement is to check if there is an error when setting the balance. If there is an error, the function will return without doing anything.
			if err := a.WriteBalance(params[0].GetQuoteUnit(), params[0].GetType(), params[1].GetUserId(), quantity, types.BalancePlus); a.Context.Debug(err) {
				return
			}

			break
		}
	}

	//The purpose of this code is to create a new API client for the pbprovider package using the existing gRPC client in the context.
	if _, err := a.SetTicker(context.Background(), &pbprovider.SetRequestTicker{Key: a.Context.Secrets[2], Price: params[0].GetPrice(), Value: params[0].GetValue(), BaseUnit: params[0].GetBaseUnit(), QuoteUnit: params[0].GetQuoteUnit(), Assigning: params[0].GetAssigning()}); a.Context.Debug(err) {
		return
	}
}

// marginProcess - This function is used to replay a trade process. It updates two orders with different amounts to determine the result
// of a trade. It updates the order status in the database with pending in to filled, updates the balance by adding the
// amount of the order to the balance, and sends a mail. In addition, it logs information about the trade.
func (a *Service) marginProcess(assigning string, params ...*types.Order) {
	// TODO margin trade process.
}
