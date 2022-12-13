package spot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/query"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	Context *assets.Context

	run, wait map[int64]bool
	block     map[int64]int64
}

// Initialization - perform actions.
func (e *Service) Initialization() {
	go e.replayPriceScale()
	go e.replayMarket()
	go e.replayChainStatus()
	go e.replayDeposit()
	go e.replayWithdraw()
}

// getSecure - check secure code.
func (e *Service) getSecure(ctx context.Context) (secure string, err error) {

	_account, err := e.Context.Auth(ctx)
	if err != nil {
		return secure, err
	}

	if err := e.Context.Db.QueryRow("select secure from accounts where id = $1", _account).Scan(&secure); err != nil {
		return secure, err
	}

	return secure, nil
}

// setSecure - write new secure code.
func (e *Service) setSecure(ctx context.Context, cleaning bool) error {

	_account, err := e.Context.Auth(ctx)
	if err != nil {
		return err
	}

	code, err := help.KeyCode(6)
	if err != nil {
		return err
	}

	if !cleaning {

		var (
			migrate = query.Migrate{
				Context: e.Context,
			}
		)

		go migrate.SamplePosts(_account, "secure", code)

	} else {
		code = ""
	}

	if _, err = e.Context.Db.Exec("update accounts set secure = $2 where id = $1;", _account, code); err != nil {
		return err
	}

	return nil
}

// setTrade - creating a new trade.
func (e *Service) setTrade(param ...*pbspot.Order) error {

	if param[0].GetValue() == 0 {
		return nil
	}

	if _, err := e.Context.Db.Exec(`insert into spot_trades (assigning, base_unit, quote_unit, price, quantity) values ($1, $2, $3, $4, $5)`, param[0].GetAssigning(), param[0].GetBaseUnit(), param[0].GetQuoteUnit(), param[0].GetPrice(), param[0].GetValue()); err != nil {
		return err
	}

	for i := 0; i < 2; i++ {

		param[i].Quantity = param[0].GetValue()
		if param[i].Param.GetEqual() {
			param[i].Quantity = param[1].GetValue()
		}

		param[i].Price = param[i].GetPrice()
		if param[i].Param.GetMaker() {
			param[i].Price = param[0].GetPrice()
		}

		if _, err := e.Context.Db.Exec(`insert into spot_transfers (order_id, assigning, user_id, base_unit, quote_unit, price, quantity, fees) values ($1, $2, $3, $4, $5, $6, $7, $8)`, param[i].GetId(), param[i].GetAssigning(), param[i].GetUserId(), param[i].GetBaseUnit(), param[i].GetQuoteUnit(), param[i].GetPrice(), param[i].GetQuantity(), e.getFees(param[i].GetQuoteUnit(), param[i].Param.GetMaker())); err != nil {
			return err
		}

		if param[i].Param.GetFees() > 0 {

			symbol := param[0].GetQuoteUnit()
			if param[i].Param.GetTurn() {
				symbol = param[0].GetBaseUnit()
			}

			if _, err := e.Context.Db.Exec("update spot_currencies set fees_charges = fees_charges + $2 where symbol = $1;", symbol, param[i].Param.GetFees()); err != nil {
				return err
			}
		}

		if err := e.Context.Publish(e.getOrder(param[i].GetId()), "exchange", "order/status"); err != nil {
			return err
		}
	}

	for _, interval := range help.Depth() {

		migrate, err := e.GetGraph(context.Background(), &pbspot.GetRequestGraph{BaseUnit: param[0].GetBaseUnit(), QuoteUnit: param[1].GetQuoteUnit(), Limit: 2, Resolution: interval})
		if err != nil {
			return err
		}

		if err := e.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/graph:%v", interval)); err != nil {
			return err
		}
	}

	return nil
}

// setOrder - create new order.
func (e *Service) setOrder(order *pbspot.Order) (id int64, err error) {

	if err := e.Context.Db.QueryRow("insert into spot_orders (assigning, base_unit, quote_unit, price, value, quantity, user_id) values ($1, $2, $3, $4, $5, $6, $7) returning id", order.GetAssigning(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetPrice(), order.GetQuantity(), order.GetValue(), order.GetUserId()).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

// setAsset - create new asset.
func (e *Service) setAsset(symbol string, userId int64, error bool) error {

	row, err := e.Context.Db.Query(`select id from spot_assets where symbol = $1 and user_id = $2`, symbol, userId)
	if err != nil {
		return err
	}
	defer row.Close()

	if !row.Next() {
		if _, err = e.Context.Db.Exec("insert into spot_assets (user_id, symbol) values ($1, $2)", userId, symbol); err != nil {
			return err
		}

		return nil
	}

	if error {
		return status.Error(700991, "the fiat asset has already been generated")
	}

	return nil
}

// getAsset - check asset by id and symbol..
func (e *Service) getAsset(symbol string, userId int64) (exist bool) {
	_ = e.Context.Db.QueryRow("select exists(select balance as balance from spot_assets where symbol = $1 and user_id = $2)::bool", symbol, userId).Scan(&exist)
	return exist
}

// setBalance - update user asset balance.
func (e *Service) setBalance(symbol string, userId int64, quantity float64, cross pbspot.Balance) error {

	switch cross {
	case pbspot.Balance_PLUS:
		if _, err := e.Context.Db.Exec("update spot_assets set balance = balance + $2 where symbol = $1 and user_id = $3;", symbol, quantity, userId); err != nil {
			return err
		}
		break
	case pbspot.Balance_MINUS:
		if _, err := e.Context.Db.Exec("update spot_assets set balance = balance - $2 where symbol = $1 and user_id = $3;", symbol, quantity, userId); err != nil {
			return err
		}
		break
	}

	return nil
}

// setTransaction - записываем новую транзакцию.
func (e *Service) setTransaction(transaction *pbspot.Transaction) (*pbspot.Transaction, error) {

	row, err := e.Context.Db.Query("select id from spot_transactions where hash = $1", transaction.GetHash())
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {

		if err := e.Context.Db.QueryRow(`insert into spot_transactions (symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning id, create_at, status`,
			transaction.GetSymbol(),
			transaction.GetHash(),
			transaction.GetValue(),
			transaction.GetFees(),
			transaction.GetConfirmation(),
			transaction.GetTo(),
			transaction.GetBlock(),
			transaction.GetChainId(),
			transaction.GetUserId(),
			transaction.GetTxType(),
			transaction.GetFinType(),
			transaction.GetPlatform(),
			transaction.GetProtocol(),
		).Scan(&transaction.Id, &transaction.CreateAt, &transaction.Status); err != nil {
			return nil, err
		}

		transaction.Chain, err = e.getChain(transaction.GetChainId(), false)
		if err != nil {
			return nil, err
		}
		transaction.Chain.Rpc, transaction.Chain.RpcPassword, transaction.Chain.RpcUser, transaction.Chain.RpcKey, transaction.ChainId = "", "", "", "", 0

		return transaction, nil
	}

	return nil, nil
}

// getFees - fees cryptocurrency
func (e *Service) getFees(symbol string, maker bool) (fees float64) {

	var (
		discount float64
	)

	if err := e.Context.Db.QueryRow("select fees_trade, fees_discount from spot_currencies where symbol = $1", symbol).Scan(&fees, &discount); err != nil {
		return fees
	}

	if maker {
		fees = decimal.FromFloat(fees).Sub(decimal.FromFloat(discount)).Float64()
	}

	return fees
}

// getSum - sum fees trade.
func (e *Service) getSum(symbol string, value float64, maker bool) (balance, fees float64) {

	var (
		discount float64
	)

	if err := e.Context.Db.QueryRow("select fees_trade, fees_discount from spot_currencies where symbol = $1", symbol).Scan(&fees, &discount); err != nil {
		return balance, fees
	}

	if maker {
		fees = decimal.FromFloat(fees).Sub(decimal.FromFloat(discount)).Float64()
	}

	// Calculate the fees from the current amount.
	// New balance including fees.
	return value - (value - (value - decimal.FromFloat(value).Mul(decimal.FromFloat(fees)).Float64()/100)), value - (value - decimal.FromFloat(value).Mul(decimal.FromFloat(fees)).Float64()/100)
}

// getAddress - user asset wallet address.
func (e *Service) getAddress(userId int64, symbol string, platform pbspot.Platform, protocol pbspot.Protocol) (address string) {
	_ = e.Context.Db.QueryRow("select coalesce(w.address, '') from spot_assets a inner join spot_wallets w on w.platform = $3 and w.protocol = $4 and w.symbol = a.symbol and w.user_id = a.user_id where a.symbol = $1 and a.user_id = $2", symbol, userId, platform, protocol).Scan(&address)
	return address
}

// getEntropy - entropy account.
func (e *Service) getEntropy(userId int64) (entropy []byte, err error) {

	if err := e.Context.Db.QueryRow("select entropy from accounts where id = $1 and status = $2", userId, true).Scan(&entropy); err != nil {
		return entropy, err
	}

	return entropy, nil
}

// getQuantity - quantity fixed.
func (e *Service) getQuantity(assigning pbspot.Assigning, quantity, price float64, cross bool) float64 {

	if cross {

		switch assigning {
		case pbspot.Assigning_BUY:
			quantity = decimal.FromFloat(quantity).Div(decimal.FromFloat(price)).Float64()
		}

		return quantity

	} else {

		switch assigning {
		case pbspot.Assigning_BUY:
			quantity = decimal.FromFloat(quantity).Mul(decimal.FromFloat(price)).Float64()
		}

		return quantity
	}
}

// getVolume - total volume pending orders.
func (e *Service) getVolume(base, quote string, assign pbspot.Assigning) (volume float64) {
	_ = e.Context.Db.QueryRow("select coalesce(sum(value), 0.00) from spot_orders where base_unit = $1 and quote_unit = $2 and assigning = $3 and status = $4", base, quote, assign, pbspot.Status_PENDING).Scan(&volume)
	return volume
}

// getOrder - order info.
func (e *Service) getOrder(id int64) *pbspot.Order {

	var (
		order pbspot.Order
	)

	_ = e.Context.Db.QueryRow("select id, value, quantity, price, assigning, user_id, base_unit, quote_unit, status, create_at from spot_orders where id = $1", id).Scan(&order.Id, &order.Value, &order.Quantity, &order.Price, &order.Assigning, &order.UserId, &order.BaseUnit, &order.QuoteUnit, &order.Status, &order.CreateAt)
	return &order
}

// getBalance - internal asset sum balance.
func (e *Service) getBalance(symbol string, userId int64) (balance float64) {
	_ = e.Context.Db.QueryRow("select balance as balance from spot_assets where symbol = $1 and user_id = $2", symbol, userId).Scan(&balance)
	return balance
}

// getRange - ranges min and max trade value fixed.
func (e *Service) getRange(symbol string, value float64) (min, max float64, ok bool) {

	if err := e.Context.Db.QueryRow("select min_trade, max_trade from spot_currencies where symbol = $1", symbol).Scan(&min, &max); err != nil {
		return min, max, ok
	}

	if value >= min && value <= max {
		return min, max, true
	}

	return min, max, ok
}

// getUnit - pair unit.
func (e *Service) getUnit(symbol string) (*pbspot.Pair, error) {

	var (
		response pbspot.Pair
	)

	if err := e.Context.Db.QueryRow(`select id, price, base_unit, quote_unit, status from spot_pairs where base_unit = $1 or quote_unit = $1`, symbol).Scan(&response.Id, &response.Price, &response.BaseUnit, &response.QuoteUnit, &response.Status); err != nil {
		return &response, err
	}

	return &response, nil
}

// getCurrency - getting information about the currency.
func (e *Service) getCurrency(symbol string, status bool) (*pbspot.Currency, error) {

	var (
		response pbspot.Currency
		maps     []string
		storage  []string
		chains   []byte
	)

	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, fin_type, create_at, chains from spot_currencies where symbol = '%v' %s", symbol, strings.Join(maps, " "))).Scan(
		&response.Id,
		&response.Name,
		&response.Symbol,
		&response.MinWithdraw,
		&response.MaxWithdraw,
		&response.MinDeposit,
		&response.MinTrade,
		&response.MaxTrade,
		&response.FeesTrade,
		&response.FeesDiscount,
		&response.FeesCharges,
		&response.FeesCosts,
		&response.Marker,
		&response.Status,
		&response.FinType,
		&response.CreateAt,
		&chains,
	); err != nil {
		return &response, err
	}

	storage = append(storage, []string{e.Context.StoragePath, "static", "icon", fmt.Sprintf("%v.png", response.GetSymbol())}...)
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		response.Icon = true
	}

	if err := json.Unmarshal(chains, &response.ChainsIds); err != nil {
		return &response, e.Context.Error(err)
	}

	return &response, nil
}

// getContract - getting information about the contract.
func (e *Service) getContract(symbol string, chainId int64) (*pbspot.Contract, error) {

	var (
		contract pbspot.Contract
	)

	if err := e.Context.Db.QueryRow(`select id, address, fees_withdraw, protocol, decimals from spot_contracts where symbol = $1 and chain_id = $2`, symbol, chainId).Scan(&contract.Id, &contract.Address, &contract.FeesWithdraw, &contract.Protocol, &contract.Decimals); err != nil {
		return &contract, err
	}

	return &contract, nil
}

// getContractById - getting information about a contract by id.
func (e *Service) getContractById(id int64) (*pbspot.Contract, error) {

	var (
		contract pbspot.Contract
	)

	if err := e.Context.Db.QueryRow(`select c.id, c.symbol, c.chain_id, c.address, c.fees_withdraw, c.protocol, c.decimals, n.platform from spot_contracts c inner join spot_chains n on n.id = c.chain_id where c.id = $1`, id).Scan(&contract.Id, &contract.Symbol, &contract.ChainId, &contract.Address, &contract.FeesWithdraw, &contract.Protocol, &contract.Decimals, &contract.Platform); err != nil {
		return &contract, err
	}

	return &contract, nil
}

// getChain - getting information about the chain.
func (e *Service) getChain(id int64, status bool) (*pbspot.Chain, error) {

	var (
		chain pbspot.Chain
		maps  []string
	)

	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status from spot_chains where id = %[1]d %[2]s", id, strings.Join(maps, " "))).Scan(
		&chain.Id,
		&chain.Name,
		&chain.Rpc,
		&chain.RpcKey,
		&chain.RpcUser,
		&chain.RpcPassword,
		&chain.Block,
		&chain.Network,
		&chain.ExplorerLink,
		&chain.Platform,
		&chain.Confirmation,
		&chain.TimeWithdraw,
		&chain.FeesWithdraw,
		&chain.Address,
		&chain.Tag,
		&chain.ParentSymbol,
		&chain.Status,
	); err != nil {
		return &chain, err
	}

	return &chain, nil
}

// getPair - getting information about a pair.
func (e *Service) getPair(id int64, status bool) (*pbspot.Pair, error) {

	var (
		chain pbspot.Pair
		maps  []string
	)

	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from spot_pairs where id = %[1]d %[2]s", id, strings.Join(maps, " "))).Scan(
		&chain.Id,
		&chain.BaseUnit,
		&chain.QuoteUnit,
		&chain.Price,
		&chain.BaseDecimal,
		&chain.QuoteDecimal,
		&chain.Status,
	); err != nil {
		return &chain, err
	}

	return &chain, nil
}

// getMarket - buy at market price.
func (e *Service) getMarket(base, quote string, assigning pbspot.Assigning, price float64) float64 {

	var (
		ok bool
	)

	if price, ok = e.getPrice(base, quote); !ok {
		return price
	}

	switch assigning {
	case pbspot.Assigning_BUY:
		_ = e.Context.Db.QueryRow("select min(price) as price from spot_orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and price >= $4 and status = $5", pbspot.Assigning_SELL, base, quote, price, pbspot.Status_PENDING).Scan(&price)
	case pbspot.Assigning_SELL:
		_ = e.Context.Db.QueryRow("select max(price) as price from spot_orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and price <= $4 and status = $5", pbspot.Assigning_BUY, base, quote, price, pbspot.Status_PENDING).Scan(&price)
	}

	return price
}

// getPrice - price tickers.
func (e *Service) getPrice(base, quote string) (price float64, ok bool) {

	if err := e.Context.Db.QueryRow("select price from spot_pairs where base_unit = $1 and quote_unit = $2", base, quote).Scan(&price); err != nil {
		return price, ok
	}

	return price, true
}

// getRatio - pair ratio price.
func (e *Service) getRatio(base, quote string) (ratio float64, ok bool) {

	migrate, err := e.GetGraph(context.Background(), &pbspot.GetRequestGraph{BaseUnit: base, QuoteUnit: quote, Limit: 2})
	if err != nil {
		return ratio, ok
	}

	if len(migrate.Fields) == 2 {
		ratio = ((migrate.Fields[0].Close - migrate.Fields[1].Close) / migrate.Fields[1].Close) * 100
	}

	return ratio, true
}

// getReserve - unit reserve coins.
func (e *Service) getReserve(symbol string, platform pbspot.Platform, protocol pbspot.Protocol) (reserve float64) {
	_ = e.Context.Db.QueryRow(`select sum(value) from spot_reserves where symbol = $1 and platform = $2 and protocol = $3`, symbol, platform, protocol).Scan(&reserve)
	return reserve
}

// setReserve - insert/update reserve.
func (e *Service) setReserve(userId int64, address, symbol string, value float64, platform pbspot.Platform, protocol pbspot.Protocol, cross pbspot.Balance) error {

	row, err := e.Context.Db.Query("select id from spot_reserves where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5", userId, symbol, platform, protocol, address)
	if err != nil {
		return err
	}
	defer row.Close()

	if row.Next() {

		switch cross {
		case pbspot.Balance_PLUS:
			if _, err := e.Context.Db.Exec("update spot_reserves set value = value + $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, protocol, address, value); err != nil {
				return err
			}
			break
		case pbspot.Balance_MINUS:
			if _, err := e.Context.Db.Exec("update spot_reserves set value = value - $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, protocol, address, value); err != nil {
				return err
			}
			break
		}

		return nil
	}

	if _, err = e.Context.Db.Exec("insert into spot_reserves (user_id, symbol, platform, protocol, address, value) values ($1, $2, $3, $4, $5, $6)", userId, symbol, platform, protocol, address, value); err != nil {
		return err
	}

	return nil
}

// setReserveLock - lock reserve asset.
func (e *Service) setReserveLock(userId int64, symbol string, platform pbspot.Platform, protocol pbspot.Protocol) error {
	if _, err := e.Context.Db.Exec("update spot_reserves set lock = $5 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4;", userId, symbol, platform, protocol, true); err != nil {
		return err
	}
	return nil
}

// setReserveLock - unlock reserve asset.
func (e *Service) setReserveUnlock(userId int64, symbol string, platform pbspot.Platform, protocol pbspot.Protocol) error {
	if _, err := e.Context.Db.Exec("update spot_reserves set lock = $5 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4;", userId, symbol, platform, protocol, false); err != nil {
		return err
	}
	return nil
}

// getStatus - base and quote status.
func (e *Service) getStatus(base, quote string) bool {

	if _, err := e.getCurrency(base, true); err != nil {
		return false
	}

	if _, err := e.getCurrency(quote, true); err != nil {
		return false
	}

	return true
}

// done - done group by chain id.
func (e *Service) done(id int64) {
	e.wait[id] = true
}
