package index

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto/pbindex"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"strings"
)

// GetStatistic - get statistic.
func (i *Service) GetStatistic(_ context.Context, _ *pbindex.GetRequestStatistic) (*pbindex.ResponseStatistic, error) {

	var (
		response    pbindex.ResponseStatistic
		statistic   pbindex.Statistic
		account     pbindex.Statistic_Account
		pair        pbindex.Statistic_Pair
		chain       pbindex.Statistic_Chain
		currency    pbindex.Statistic_Currency
		transaction pbindex.Statistic_Transaction
		order       pbindex.Statistic_Order
	)

	_ = i.Context.Db.QueryRow("select count(*) from accounts where status = $1", true).Scan(&account.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from accounts where status = $1", false).Scan(&account.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from spot_pairs where status = $1", true).Scan(&pair.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from spot_pairs where status = $1", false).Scan(&pair.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from spot_chains where status = $1", true).Scan(&chain.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from spot_chains where status = $1", false).Scan(&chain.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from spot_currencies where status = $1", true).Scan(&currency.Enable)
	_ = i.Context.Db.QueryRow("select count(*) from spot_currencies where status = $1", false).Scan(&currency.Disable)

	_ = i.Context.Db.QueryRow("select count(*) from spot_transactions where status = $1", pbspot.Status_FILLED).Scan(&transaction.Filled)
	_ = i.Context.Db.QueryRow("select count(*) from spot_transactions where status = $1", pbspot.Status_PENDING).Scan(&transaction.Pending)

	_ = i.Context.Db.QueryRow("select count(*) from spot_orders where assigning = $1 and status = $2", pbspot.Assigning_SELL, pbspot.Status_PENDING).Scan(&order.Sell)
	_ = i.Context.Db.QueryRow("select count(*) from spot_orders where assigning = $1 and status = $2", pbspot.Assigning_BUY, pbspot.Status_PENDING).Scan(&order.Buy)

	statistic.Accounts = &account
	statistic.Pairs = &pair
	statistic.Chains = &chain
	statistic.Currencies = &currency
	statistic.Transactions = &transaction
	statistic.Orders = &order

	rows, err := i.Context.Db.Query("select symbol, fees_charges, fees_costs from spot_currencies")
	if err != nil {
		return &response, i.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			reserve pbindex.Statistic_Reserve
			price   float64
		)

		if err := rows.Scan(&reserve.Symbol, &reserve.ValueCharged, &reserve.ValueCosts); err != nil {
			return &response, i.Context.Error(err)
		}

		_ = i.Context.Db.QueryRow("select sum(value) from spot_reserves where symbol = $1", reserve.Symbol).Scan(&reserve.Value)
		if reserve.Value > 0 {

			if reserve.Symbol != "usdt" {
				if err := i.Context.Db.QueryRow("select price from spot_pairs where base_unit = $1 and quote_unit = $2", reserve.Symbol, "usd").Scan(&price); err != nil {
					return &response, i.Context.Error(err)
				}

				reserve.ValueChargedConvert = decimal.New(reserve.ValueCharged).Mul(price).Float()
			}

		}

		statistic.Reserves = append(statistic.Reserves, &reserve)
	}
	response.Fields = &statistic

	return &response, nil
}

func (i *Service) GetCurrencies(_ context.Context, req *pbindex.GetRequestCurrencies) (*pbindex.ResponseCurrency, error) {

	var (
		response pbindex.ResponseCurrency
		query    []string
	)

	if req.GetId() > 0 {
		query = append(query, fmt.Sprintf("where id < %v", req.GetId()))
	}

	if err := i.Context.Db.QueryRow(fmt.Sprintf("select count(*) from spot_currencies %s", strings.Join(query, " "))).Scan(&response.Count); err != nil && err == sql.ErrNoRows {
		return &response, i.Context.Error(err)
	}

	rows, err := i.Context.Db.Query(fmt.Sprintf("select id, name, symbol, fees_trade, fees_discount from spot_currencies %s order by id desc limit %d", strings.Join(query, " "), 5))
	if err != nil {
		return &response, i.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			currency pbindex.Currency
		)

		if err := rows.Scan(&currency.Id, &currency.Name, &currency.Symbol, &currency.FeesTrade, &currency.FeesDiscount); err != nil {
			return &response, i.Context.Error(err)
		}

		rows, err := i.Context.Db.Query("select p.id, c.symbol from spot_pairs p inner join spot_currencies c on (c.symbol = p.base_unit or c.symbol = p.quote_unit) and c.symbol != $1 where p.base_unit = $1 or p.quote_unit = $1", currency.GetSymbol())
		if err != nil {
			return &response, i.Context.Error(err)
		}

		for rows.Next() {

			var (
				pair pbindex.Pair
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

func (i *Service) GetPairs(_ context.Context, _ *pbindex.GetRequestPairs) (*pbindex.ResponsePair, error) {

	var (
		response pbindex.ResponsePair
	)

	rows, err := i.Context.Db.Query("select id, base_unit, quote_unit from spot_pairs")
	if err != nil {
		return &response, i.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			pair pbindex.Pair
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
