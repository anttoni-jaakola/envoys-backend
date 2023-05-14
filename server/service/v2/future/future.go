package future

import (
	"fmt"
	"strconv"

	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
)

type Service struct {
	Context *assets.Context
}

func (a *Service) Initialization() {

}

func (a *Service) queryValidatePair(base, quote, _type string) error {
	var (
		exist bool
	)

	if err := a.Context.Db.QueryRow("select exists(select id from pairs where base_unit = $1 and quote_unit = $2 and type = $3)::bool", base, quote, _type).Scan(&exist); err != nil || !exist {
		return status.Errorf(11585, "this pair %v-%v does not exist", base, quote)
	}
	return nil
}

func (a *Service) queryMarket(base, quote, _type string, assigning string, price float64) float64 {

	var (
		ok bool
	)

	if price, ok = a.queryPrice(base, quote); !ok {
		return price
	}

	switch assigning {
	case types.AssigningBuy:

		_ = a.Context.Db.QueryRow("select min(price) as price from futures where assigning = $1 and base_unit = $2 and quote_unit = $3 and price >= $4 and status = $5 and type = $6", types.AssigningSell, base, quote, price, types.StatusPending, _type).Scan(&price)

	case types.AssigningSell:

		_ = a.Context.Db.QueryRow("select max(price) as price from futures where assigning = $1 and base_unit = $2 and quote_unit = $3 and price <= $4 and status = $5 and type = $6", types.AssigningBuy, base, quote, price, types.StatusPending, _type).Scan(&price)
	}

	return price
}

func (a *Service) queryPrice(base, quote string) (price float64, ok bool) {

	if err := a.Context.Db.QueryRow("select price from pairs where base_unit = $1 and quote_unit = $2", base, quote).Scan(&price); err != nil {
		return price, ok
	}

	return price, true
}

func (a *Service) queryValidateOrder(order *types.Future) (summary float64, err error) {

	if order.GetPrice() == 0 {
		return 0, status.Errorf(65790, "impossible price %v", order.GetPrice())
	}

	switch order.GetAssigning() {
	case types.AssigningOpen:

		quantity := decimal.New(order.GetQuantity()).Mul(order.GetPrice()).Float()

		if min, max, ok := a.queryRange(order.GetQuoteUnit(), quantity); !ok {
			return 0, status.Errorf(11623, "[quote]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		balance := a.QueryBalance(order.GetQuoteUnit(), "future", order.GetUserId())

		if decimal.New(quantity).Div(order.GetLeverage()).Float() > balance || order.GetQuantity() == 0 {
			return 0, status.Error(11586, "[quote]: there is not enough funds on your asset balance to place an order")
		}

		return quantity, nil

	case types.AssigningClose:

		quantity := order.GetQuantity()

		_, openQuantity := a.queryOrders(order.UserId, types.AssigningOpen, order.GetPosition(), order.GetBaseUnit(), order.GetQuoteUnit())

		fmt.Println("TOTAL OPEN QUANTITY IS ", openQuantity)
		if min, max, ok := a.queryRange(order.GetBaseUnit(), order.GetQuantity()); !ok {
			return 0, status.Errorf(11587, "[base]: minimum trading amount: %v~%v, maximum trading amount: %v", min, strconv.FormatFloat(decimal.New(min).Mul(2).Float(), 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64))
		}

		// balance := a.QueryBalance(order.GetBaseUnit(), "future", order.GetUserId())

		if quantity > openQuantity || order.GetQuantity() == 0 {
			return 0, status.Error(11624, "[base]: there is not enough funds locking in your orders")
		}

		return quantity, nil
	}

	return 0, status.Error(11596, "invalid input parameter")
}

func (a *Service) writeOrder(order *types.Future) (id int64, err error) {

	if err := a.Context.Db.QueryRow("insert into futures (position, trading, base_unit, quote_unit, price, quantity, leverage, take_profit, stop_loss, fees, status, user_id, assigning, value) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) returning id", order.GetPosition(), order.GetOrderType(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetPrice(), order.GetQuantity(), order.GetLeverage(), order.GetTakeProfit(), order.GetStopLoss(), order.GetFees(), types.StatusPending, order.GetUserId(), order.GetAssigning(), order.GetValue()).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}
func (a *Service) writeAsset(symbol, _type string, userId int64, error bool) error {

	row, err := a.Context.Db.Query(`select id from balances where symbol = $1 and user_id = $2 and type = $3`, symbol, userId, _type)
	if err != nil {
		return err
	}
	defer row.Close()

	if !row.Next() {

		if _, err = a.Context.Db.Exec("insert into balances (user_id, symbol, type) values ($1, $2, $3)", userId, symbol, _type); err != nil {
			return err
		}

		return nil
	}

	if error {
		return status.Error(700991, "the fiat asset has already been generated")
	}

	return nil
}
func (a *Service) WriteBalance(symbol, _type string, userId int64, quantity float64, cross string) error {

	switch cross {
	case types.BalancePlus:

		if _, err := a.Context.Db.Exec("update balances set value = value + $2 where symbol = $1 and user_id = $3 and type = $4;", symbol, quantity, userId, _type); err != nil {
			return err
		}
		break
	case types.BalanceMinus:

		if _, err := a.Context.Db.Exec("update balances set value = value - $2 where symbol = $1 and user_id = $3 and type = $4;", symbol, quantity, userId, _type); err != nil {
			return err
		}
		break
	}

	return nil
}

func (a *Service) queryRange(symbol string, value float64) (min, max float64, ok bool) {

	if err := a.Context.Db.QueryRow("select min_trade, max_trade from assets where symbol = $1", symbol).Scan(&min, &max); err != nil {
		return min, max, ok
	}

	if value >= min && value <= max {
		return min, max, true
	}

	return min, max, ok
}
func (a *Service) QueryBalance(symbol, _type string, userId int64) (balance float64) {

	// _ = a.Context.Db.QueryRow("select value as balance from balances where symbol = $1 and user_id = $2 and type = $3", symbol, userId, _type).Scan(&balance)
	_ = a.Context.Db.QueryRow("select value as balance from balances where symbol = $1 and user_id = $2", symbol, userId).Scan(&balance)

	return balance
}

func (a *Service) writeTrade(id int64, symbol string, value, price float64, convert bool) (float64, error) {

	order := a.queryOrder(id)
	// order.Value = value

	if convert {
		value = decimal.New(value).Mul(price).Float()
	}

	s, f, maker, err := a.querySum(id, symbol, value)
	if err != nil {
		return 0, err
	}

	if order.GetPosition() == types.PositionLong {
		order.Fees = decimal.New(f).Div(price).Float()
	} else {
		order.Fees = f
	}

	if _, err := a.Context.Db.Exec(`insert into trades (order_id, assigning, user_id, base_unit, quote_unit, quantity, fees, price, maker) values ($1, $2, $3, $4, $5, $6, $7, $8)`, order.GetId(), order.GetAssigning(), order.GetUserId(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetQuantity(), order.GetFees(), price, maker); err != nil {
		return 0, err
	}

	if f > 0 {

		if _, err := a.Context.Db.Exec("update assets set fees_charges = fees_charges + $2 where symbol = $1;", symbol, f); err != nil {
			return 0, err
		}
	}

	if err := a.Context.Publish(a.queryOrder(order.GetId()), "exchange", "future/status"); err != nil {
		return 0, err
	}

	return s, nil
}
func (a *Service) queryOrder(id int64) *types.Future {

	var (
		order types.Future
	)

	_ = a.Context.Db.QueryRow("select id, quantity, price, assigning, user_id, base_unit, quote_unit, status, create_at from futures where id = $1", id).Scan(&order.Id, &order.Quantity, &order.Price, &order.Assigning, &order.UserId, &order.BaseUnit, &order.QuoteUnit, &order.Status, &order.CreateAt)
	return &order
}
func (a *Service) queryOrders(userId int64, assigning string, position string, baseUnit string, quoteUnit string) ([]*types.Future, float64) {
	var (
		orders        []*types.Future
		totalQuantity float64 = 0.0
	)
	rows, err := a.Context.Db.Query("select id, quantity, price, assigning, user_id, base_unit, quote_unit, status, position, fees, create_at, leverage from futures where user_id = $1 and assigning = $2 and position = $3 and base_unit = $4 and quote_unit = $5", userId, assigning, position, baseUnit, quoteUnit)

	if err != nil {
		return nil, 0
	}
	defer rows.Close()
	for rows.Next() {
		var (
			order types.Future
		)
		if err := rows.Scan(&order.Id, &order.Quantity, &order.Price, &order.Assigning, &order.UserId, &order.BaseUnit, &order.QuoteUnit, &order.Status, &order.Position, &order.Fees, &order.CreateAt, &order.Leverage); err != nil {
			fmt.Println("Error occured", err)
			return nil, 0
		}
		totalQuantity += decimal.New(totalQuantity).Add(decimal.New(order.Quantity).Mul(order.Price).Div(order.Leverage).Float()).Float()
		// totalQuantity = totalQuantity + order.Quantity*order.Price/order.Leverage
		orders = append(orders, &order)
	}

	return orders, totalQuantity
}

func (a *Service) querySum(id int64, symbol string, value float64) (b, f float64, m bool, err error) {

	var (
		d float64
		s string
	)

	if err := a.Context.Db.QueryRow("select fees_trade, fees_discount from assets where symbol = $1", symbol).Scan(&f, &d); err != nil {
		return b, f, m, err
	}

	if err := a.Context.Db.QueryRow("select status from futures where id = $1;", id).Scan(&s); err != nil {
		return b, f, m, err
	}

	if s == types.StatusPending {
		m = true
	}

	if m {
		f = decimal.New(f).Sub(d).Float()
	}

	return decimal.New(value).Sub(decimal.New(decimal.New(value).Mul(f).Float()).Div(100).Float()).Float(), decimal.New(value).Sub(decimal.New(value).Sub(decimal.New(decimal.New(value).Mul(f).Float()).Div(100).Float()).Float()).Float(), m, nil
}

func (a *Service) queryQuantity(assigning string, position string, quantity, price float64, cross bool) float64 {

	if cross {

		switch assigning {
		case types.AssigningBuy:
			quantity = decimal.New(quantity).Div(price).Float()
		}

		return quantity

	} else {

		switch assigning {
		case types.AssigningSell:
			quantity = decimal.New(quantity).Mul(price).Float()
		}

		return quantity
	}
}
