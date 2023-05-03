package future

import (
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
)

func (a *Service) closePosition(order *types.Future) {
	// publish exchange
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
			if order.GetPrice() >= item.GetPrice() {
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

}
