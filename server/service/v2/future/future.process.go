package future

import (
	"context"
	"fmt"

	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
)

func (a *Service) closePosition(order *types.Future) {
	// publish exchange
	fmt.Println("PUBLISH EXCHANGE")
	if err := a.Context.Publish(order, "exchange", "future/create"); a.Context.Debug(err) {
		return
	}

	var position string

	if order.Position == types.PositionLong {
		position = types.PositionShort
	} else {
		position = types.PositionLong
	}

	// get future orders by position
	rows, err := a.Context.Db.Query(`select id, assigning, position, base_unit, quote_unit, quantity, price, user_id, status from futures where assigning = $1 and base_unit = $2 and quote_unit = $3 and user_id != $4 and status = $5 and position = $6 order by id`, "close", order.GetBaseUnit(), order.GetQuoteUnit(), order.GetUserId(), types.StatusPending, position)

	if a.Context.Debug(err) {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var (
			item types.Future
		)

		if err = rows.Scan(&item.Id, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.Quantity, &item.Price, &item.UserId, &item.Status); a.Context.Debug(err) {
			return
		}

		// check order status and update value

		// swtich position
		switch position {
		//case long
		case types.PositionLong:
			if order.GetPrice() <= item.GetPrice() {
				a.Context.Logger.Infof("[BID]: (item [%v]) >= (order [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())

				//process trade
				a.handleFutureTrade(position, order, &item)
			} else {
				a.Context.Logger.Infof("[BID]: no matches found: (item [%v]) >= (order [%v])", order.GetPrice(), item.GetPrice())
			}
			break
			//case short
		case types.PositionShort:
			if order.GetPrice() >= item.GetPrice() {
				a.Context.Logger.Infof("[ASK]: (order [%v]) <= (item [%v]), order ID: %v", order.GetPrice(), item.GetPrice(), item.GetId())
				//process trade

				a.handleFutureTrade(position, order, &item)
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
}

func (a *Service) handleFutureTrade(position string, params ...*types.Future) {

	var (
		price    float64
		instance int
		migrate  = query.Migrate{
			Context: a.Context,
		}
	)

	if params[0].GetValue() >= params[1].GetValue() {
		instance = 1
	}

	switch position {
	case types.PositionLong:
		price = params[1].GetPrice()
	case types.PositionShort:
		price = params[0].GetPrice()
	}

	if params[instance].GetValue() > 0 {

		for i := 0; i < 2; i++ {

			var (
				value float64
			)

			if err := a.Context.Db.QueryRow("update futures set value = value - $2 where id = $1 and status = $3 returning value;", params[i].GetId(), params[instance].GetValue(), types.StatusPending).Scan(&value); a.Context.Debug(err) {
				return
			}

			if value == 0 {

				if _, err := a.Context.Db.Exec("update futures set status = $2 where id = $1", params[i].GetId(), types.StatusFilled); a.Context.Debug(err) {
					return
				}

				go migrate.SendMail(params[i].GetUserId(), "order_filled", params[i].GetId(), a.queryQuantity(params[i].GetAssigning(), params[i].GetPosition(), params[i].GetQuantity(), price, false), params[i].GetBaseUnit(), params[i].GetQuoteUnit(), params[i].GetAssigning())
			}
		}

		switch params[1].GetPosition() {
		case types.PositionLong:

			quantity, err := a.writeTrade(params[0].GetId(), params[0].GetQuoteUnit(), params[instance].GetValue(), price, true)
			if a.Context.Debug(err) {
				return
			}

			if err := a.WriteBalance(params[0].GetQuoteUnit(), "future", params[0].GetUserId(), quantity, types.BalancePlus); a.Context.Debug(err) {
				return
			}

			quantity, err = a.writeTrade(params[1].GetId(), params[0].GetBaseUnit(), params[instance].GetValue(), price, false)
			if a.Context.Debug(err) {
				return
			}

			if err := a.WriteBalance(params[0].GetBaseUnit(), "future", params[1].GetUserId(), quantity, types.BalancePlus); err != nil {
				return
			}

			break
		case types.PositionShort:

			quantity, err := a.writeTrade(params[0].GetId(), params[0].GetBaseUnit(), params[instance].GetValue(), price, false)
			if a.Context.Debug(err) {
				return
			}

			if err := a.WriteBalance(params[0].GetBaseUnit(), "future", params[0].GetUserId(), quantity, types.BalancePlus); a.Context.Debug(err) {
				return
			}

			quantity, err = a.writeTrade(params[1].GetId(), params[0].GetQuoteUnit(), params[instance].GetValue(), price, true)
			if a.Context.Debug(err) {
				return
			}

			if err := a.WriteBalance(params[0].GetQuoteUnit(), "future", params[1].GetUserId(), quantity, types.BalancePlus); a.Context.Debug(err) {
				return
			}

			break
		}
	}

	if _, err := a.SetTicker(context.Background(), &pbfuture.SetRequestTicker{Key: a.Context.Secrets[2], Price: params[0].GetPrice(), Value: params[0].GetValue(), BaseUnit: params[0].GetBaseUnit(), QuoteUnit: params[0].GetQuoteUnit(), Assigning: params[0].GetAssigning()}); a.Context.Debug(err) {
		return
	}
}
