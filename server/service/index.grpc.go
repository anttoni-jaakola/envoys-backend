package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"strings"
)

// GetStatistic - get statistic.
func (i *IndexService) GetStatistic(_ context.Context, _ *proto.GetIndexRequestStatistic) (*proto.ResponseStatistic, error) {

	var (
		response    proto.ResponseStatistic
		statistic   proto.Statistic
		account     proto.Statistic_Account
		pair        proto.Statistic_Pair
		chain       proto.Statistic_Chain
		currency    proto.Statistic_Currency
		transaction proto.Statistic_Transaction
		order       proto.Statistic_Order
	)

	_ = i.Context.Db.QueryRow("select count(*) from accounts where status = $1", true).Scan(&account.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from accounts where status = $1", false).Scan(&account.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from pairs where status = $1", true).Scan(&pair.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from pairs where status = $1", false).Scan(&pair.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from chains where status = $1", true).Scan(&chain.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from chains where status = $1", false).Scan(&chain.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from currencies where status = $1", true).Scan(&currency.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from currencies where status = $1", false).Scan(&currency.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from transactions where status = $1", proto.Status_FILLED).Scan(&transaction.Filled)
	_ = i.Context.Db.QueryRow("select count(*) from transactions where status = $1", proto.Status_PENDING).Scan(&transaction.Pending)

	_ = i.Context.Db.QueryRow("select count(*) from orders where assigning = $1 and status = $2", proto.Assigning_SELL, proto.Status_PENDING).Scan(&order.Sell)
	_ = i.Context.Db.QueryRow("select count(*) from orders where assigning = $1 and status = $2", proto.Assigning_BUY, proto.Status_PENDING).Scan(&order.Buy)

	statistic.Accounts = &account
	statistic.Pairs = &pair
	statistic.Chains = &chain
	statistic.Currencies = &currency
	statistic.Transactions = &transaction
	statistic.Orders = &order

	rows, err := i.Context.Db.Query("select symbol, fees_charges, fees_costs from currencies")
	if err != nil {
		return &response, i.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			reserve proto.Statistic_Reserve
		)

		if err := rows.Scan(&reserve.Symbol, &reserve.ValueCharged, &reserve.ValueCosts); err != nil {
			return &response, i.Context.Error(err)
		}

		_ = i.Context.Db.QueryRow("select sum(value) from reserves where symbol = $1", reserve.Symbol).Scan(&reserve.Value)
		if reserve.Value > 0 {
			migrate := ExchangeService{
				Context: i.Context,
			}
			if price, ok := migrate.getPrice(reserve.Symbol, "usd"); ok {
				reserve.ValueChargedConvert = decimal.FromFloat(reserve.ValueCharged).Mul(decimal.FromFloat(price)).Float64()
			}
		}

		statistic.Reserves = append(statistic.Reserves, &reserve)
	}
	response.Fields = &statistic

	return &response, nil
}

func (i *IndexService) GetCurrencies(_ context.Context, req *proto.GetIndexRequestCurrencies) (*proto.ResponseCurrency, error) {

	var (
		response proto.ResponseCurrency
		query    []string
	)

	if req.GetId() > 0 {
		query = append(query, fmt.Sprintf("where id < %v", req.GetId()))
	}

	if err := i.Context.Db.QueryRow(fmt.Sprintf("select count(*) from currencies %s", strings.Join(query, " "))).Scan(&response.Count); err != nil && err == sql.ErrNoRows {
		return &response, i.Context.Error(err)
	}

	rows, err := i.Context.Db.Query(fmt.Sprintf("select id, name, symbol, fees_trade, fees_discount from currencies %s order by id desc limit %d", strings.Join(query, " "), 5))
	if err != nil {
		return &response, i.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			currency proto.Currency
		)

		if err := rows.Scan(&currency.Id, &currency.Name, &currency.Symbol, &currency.FeesTrade, &currency.FeesDiscount); err != nil {
			return &response, i.Context.Error(err)
		}

		rows, err := i.Context.Db.Query("select p.id, c.symbol from pairs p inner join currencies c on (c.symbol = p.base_unit or c.symbol = p.quote_unit) and c.symbol != $1 where p.base_unit = $1 or p.quote_unit = $1", currency.GetSymbol())
		if err != nil {
			return &response, i.Context.Error(err)
		}

		for rows.Next() {

			var (
				pair proto.Pair
			)

			if err := rows.Scan(&pair.Id, &pair.Symbol); err != nil {
				return &response, i.Context.Error(err)
			}

			currency.Pairs = append(currency.Pairs, &pair)
		}
		rows.Close()

		response.Fields = append(response.Fields, &currency)
	}

	return &response, nil
}

func (i *IndexService) GetPairs(_ context.Context, _ *proto.GetIndexRequestPairs) (*proto.ResponsePair, error) {

	var (
		response proto.ResponsePair
	)

	rows, err := i.Context.Db.Query("select id, base_unit, quote_unit from pairs")
	if err != nil {
		return &response, i.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			pair proto.Pair
		)

		if err := rows.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit); err != nil {
			return &response, i.Context.Error(err)
		}

		price, ratio, err := i.getPrice(pair.GetBaseUnit(), pair.GetQuoteUnit())
		if err != nil {
			return &response, i.Context.Error(err)
		}

		if price, ok := price.(float64); ok {
			pair.Price = price
		}

		if ratio, ok := ratio.(float64); ok {
			pair.Ratio = ratio
		}

		response.Fields = append(response.Fields, &pair)
	}

	return &response, nil
}
