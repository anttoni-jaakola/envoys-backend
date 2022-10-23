package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"os"
	"path/filepath"
	"strings"
)

type ExchangeService struct {
	Context *assets.Context

	run, wait map[int64]bool
	block     map[int64]int64
}

// Initialization - perform actions.
func (e *ExchangeService) Initialization() {
	go e.replayPriceScale()
	go e.replayMarket()
	go e.replayChainStatus()
	go e.replayDeposit()
	go e.replayWithdraw()
}

// getSecure - check secure code.
func (e *ExchangeService) getSecure(ctx context.Context) (secure string, err error) {

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return secure, err
	}

	if err := e.Context.Db.QueryRow("select secure from accounts where id = $1", account).Scan(&secure); err != nil {
		return secure, err
	}

	return secure, nil
}

// setSecure - write new secure code.
func (e *ExchangeService) setSecure(ctx context.Context, cleaning bool) error {

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return err
	}

	code, err := help.KeyCode(6)
	if err != nil {
		return err
	}

	if !cleaning {

		var (
			migrate = Query{
				Context: e.Context,
			}
		)

		go migrate.SamplePosts(account, "secure", code)

	} else {
		code = ""
	}

	if _, err = e.Context.Db.Exec("update accounts set secure = $2 where id = $1;", account, code); err != nil {
		return err
	}

	return nil
}

// setTradeStorage - new trade stats.
func (e *ExchangeService) setInternalTransfer(id, userId int64, assign proto.Assigning, base, quote string, price, value, fees float64, maker, turn bool) error {

	if _, err := e.Context.Db.Exec(`insert into transfers (assigning, order_id, user_id, base_unit, quote_unit, price, quantity, fees) values ($1, $2, $3, $4, $5, $6, $7, $8)`, assign, id, userId, base, quote, price, value, e.getFees(quote, maker)); err != nil {
		return err
	}

	if fees > 0 {

		if turn {
			quote = base
		}

		if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges + $2 where symbol = $1;", quote, fees); err != nil {
			return err
		}
	}

	return nil
}

// setTrade - creating a new trade.
func (e *ExchangeService) setTrade(param ...*proto.Order) error {

	if param[0].GetValue() == 0 {
		return nil
	}

	if _, err := e.Context.Db.Exec(`insert into trades (assigning, base_unit, quote_unit, price, quantity, ask_price, bid_price) values ($1, $2, $3, $4, $5, $6, $7)`, param[0].GetAssigning(), param[0].GetBaseUnit(), param[0].GetQuoteUnit(), param[0].GetPrice(), param[0].GetValue(), param[0].GetPrice(), param[1].GetPrice()); err != nil {
		return err
	}
	return nil
}

// setOrder - create new order.
func (e *ExchangeService) setOrder(order *proto.Order) (id int64, err error) {

	if err := e.Context.Db.QueryRow("insert into orders (assigning, base_unit, quote_unit, price, value, quantity, user_id) values ($1, $2, $3, $4, $5, $6, $7) returning id", order.GetAssigning(), order.GetBaseUnit(), order.GetQuoteUnit(), order.GetPrice(), order.GetQuantity(), order.GetValue(), order.GetUserId()).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

// setAsset - create new asset.
func (e *ExchangeService) setAsset(symbol string, userId int64, error bool) error {

	row, err := e.Context.Db.Query(`select id from assets where symbol = $1 and user_id = $2`, symbol, userId)
	if err != nil {
		return err
	}
	defer row.Close()

	if !row.Next() {
		if _, err = e.Context.Db.Exec("insert into assets (user_id, symbol) values ($1, $2)", userId, symbol); err != nil {
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
func (e *ExchangeService) getAsset(symbol string, userId int64) (exist bool) {
	_ = e.Context.Db.QueryRow("select exists(select balance as balance from assets where symbol = $1 and user_id = $2)::bool", symbol, userId).Scan(&exist)
	return exist
}

// setBalance - update user asset balance.
func (e *ExchangeService) setBalance(symbol string, userId int64, quantity float64, cross proto.Balance) error {

	switch cross {
	case proto.Balance_PLUS:
		if _, err := e.Context.Db.Exec("update assets set balance = balance + $2 where symbol = $1 and user_id = $3;", symbol, quantity, userId); err != nil {
			return err
		}
		break
	case proto.Balance_MINUS:
		if _, err := e.Context.Db.Exec("update assets set balance = balance - $2 where symbol = $1 and user_id = $3;", symbol, quantity, userId); err != nil {
			return err
		}
		break
	}

	return nil
}

// setTransaction - записываем новую транзакцию.
func (e *ExchangeService) setTransaction(transaction *proto.Transaction) (*proto.Transaction, error) {

	row, err := e.Context.Db.Query("select id from transactions where hash = $1", transaction.GetHash())
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {

		if err := e.Context.Db.QueryRow(`insert into transactions (symbol, hash, value, fees, confirmation, "to", block, chain_id, user_id, tx_type, fin_type, platform, protocol) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) returning id, create_at, status`,
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
func (e *ExchangeService) getFees(symbol string, maker bool) (fees float64) {

	var (
		discount float64
	)

	if err := e.Context.Db.QueryRow("select fees_trade, fees_discount from currencies where symbol = $1", symbol).Scan(&fees, &discount); err != nil {
		return fees
	}

	if maker {
		fees = decimal.FromFloat(fees).Sub(decimal.FromFloat(discount)).Float64()
	}

	return fees
}

// getSum - sum fees trade.
func (e *ExchangeService) getSum(symbol string, value float64, maker bool) (balance, fees float64) {

	var (
		discount float64
	)

	if err := e.Context.Db.QueryRow("select fees_trade, fees_discount from currencies where symbol = $1", symbol).Scan(&fees, &discount); err != nil {
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
func (e *ExchangeService) getAddress(userId int64, symbol string, platform proto.Platform, protocol proto.Protocol) (address string) {
	_ = e.Context.Db.QueryRow("select coalesce(w.address, '') from assets a inner join wallets w on w.platform = $3 and w.protocol = $4 and w.symbol = a.symbol and w.user_id = a.user_id where a.symbol = $1 and a.user_id = $2", symbol, userId, platform, protocol).Scan(&address)
	return address
}

// getEntropy - entropy account.
func (e *ExchangeService) getEntropy(userId int64) (entropy []byte, err error) {

	if err := e.Context.Db.QueryRow("select entropy from accounts where id = $1 and status = $2", userId, true).Scan(&entropy); err != nil {
		return entropy, err
	}

	return entropy, nil
}

// getQuantity - quantity fixed.
func (e *ExchangeService) getQuantity(assigning proto.Assigning, quantity, price float64, cross bool) float64 {

	if cross {

		switch assigning {
		case proto.Assigning_BUY:
			quantity = decimal.FromFloat(quantity).Div(decimal.FromFloat(price)).Float64()
		}

		return quantity

	} else {

		switch assigning {
		case proto.Assigning_BUY:
			quantity = decimal.FromFloat(quantity).Mul(decimal.FromFloat(price)).Float64()
		}

		return quantity
	}
}

// getVolume - total volume pending orders.
func (e *ExchangeService) getVolume(base, quote string, assign proto.Assigning) (volume float64) {
	_ = e.Context.Db.QueryRow("select coalesce(sum(value), 0.00) from orders where base_unit = $1 and quote_unit = $2 and assigning = $3 and status = $4", base, quote, assign, proto.Status_PENDING).Scan(&volume)
	return volume
}

// getOrder - order info.
func (e *ExchangeService) getOrder(id int64) *proto.Order {

	var (
		order proto.Order
	)

	_ = e.Context.Db.QueryRow("select id, value, quantity, price, assigning, user_id, base_unit, quote_unit, status, create_at from orders where id = $1", id).Scan(&order.Id, &order.Value, &order.Quantity, &order.Price, &order.Assigning, &order.UserId, &order.BaseUnit, &order.QuoteUnit, &order.Status, &order.CreateAt)
	return &order
}

// getBalance - internal asset sum balance.
func (e *ExchangeService) getBalance(symbol string, userId int64) (balance float64) {
	_ = e.Context.Db.QueryRow("select balance as balance from assets where symbol = $1 and user_id = $2", symbol, userId).Scan(&balance)
	return balance
}

// getRange - ranges min and max trade value fixed.
func (e *ExchangeService) getRange(symbol string, value float64) (min, max float64, ok bool) {

	if err := e.Context.Db.QueryRow("select min_trade, max_trade from currencies where symbol = $1", symbol).Scan(&min, &max); err != nil {
		return min, max, ok
	}

	if value >= min && value <= max {
		return min, max, true
	}

	return min, max, ok
}

// getUnit - pair unit.
func (e *ExchangeService) getUnit(symbol string) (*proto.Pair, error) {

	var (
		response proto.Pair
	)

	if err := e.Context.Db.QueryRow(`select id, price, base_unit, quote_unit, status from pairs where base_unit = $1 or quote_unit = $1`, symbol).Scan(&response.Id, &response.Price, &response.BaseUnit, &response.QuoteUnit, &response.Status); err != nil {
		return &response, err
	}

	return &response, nil
}

// getCurrency - getting information about the currency.
func (e *ExchangeService) getCurrency(symbol string, status bool) (*proto.Currency, error) {

	var (
		response proto.Currency
		query    []string
		storage  []string
		chains   []byte
	)

	if status {
		query = append(query, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, fin_type, create_at, chains from currencies where symbol = '%v' %s", symbol, strings.Join(query, " "))).Scan(
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
func (e *ExchangeService) getContract(symbol string, chainId int64) (*proto.Contract, error) {

	var (
		contract proto.Contract
	)

	if err := e.Context.Db.QueryRow(`select id, address, fees_withdraw, protocol, decimals from contracts where symbol = $1 and chain_id = $2`, symbol, chainId).Scan(&contract.Id, &contract.Address, &contract.FeesWithdraw, &contract.Protocol, &contract.Decimals); err != nil {
		return &contract, err
	}

	return &contract, nil
}

// getContractById - getting information about a contract by id.
func (e *ExchangeService) getContractById(id int64) (*proto.Contract, error) {

	var (
		contract proto.Contract
	)

	if err := e.Context.Db.QueryRow(`select c.id, c.symbol, c.chain_id, c.address, c.fees_withdraw, c.protocol, c.decimals, n.platform from contracts c inner join chains n on n.id = c.chain_id where c.id = $1`, id).Scan(&contract.Id, &contract.Symbol, &contract.ChainId, &contract.Address, &contract.FeesWithdraw, &contract.Protocol, &contract.Decimals, &contract.Platform); err != nil {
		return &contract, err
	}

	return &contract, nil
}

// getChain - getting information about the chain.
func (e *ExchangeService) getChain(id int64, status bool) (*proto.Chain, error) {

	var (
		chain proto.Chain
		query []string
	)

	if status {
		query = append(query, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, name, rpc, rpc_key, rpc_user, rpc_password, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status from chains where id = %[1]d %[2]s", id, strings.Join(query, " "))).Scan(
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
func (e *ExchangeService) getPair(id int64, status bool) (*proto.Pair, error) {

	var (
		chain proto.Pair
		query []string
	)

	if status {
		query = append(query, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs where id = %[1]d %[2]s", id, strings.Join(query, " "))).Scan(
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
func (e *ExchangeService) getMarket(base, quote string, assigning proto.Assigning, price float64) float64 {

	var (
		ok bool
	)

	if price, ok = e.getPrice(base, quote); !ok {
		return price
	}

	switch assigning {
	case proto.Assigning_BUY:
		_ = e.Context.Db.QueryRow("select min(price) as price from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and price >= $4 and status = $5", proto.Assigning_SELL, base, quote, price, proto.Status_PENDING).Scan(&price)
	case proto.Assigning_SELL:
		_ = e.Context.Db.QueryRow("select max(price) as price from orders where assigning = $1 and base_unit = $2 and quote_unit = $3 and price <= $4 and status = $5", proto.Assigning_BUY, base, quote, price, proto.Status_PENDING).Scan(&price)
	}

	return price
}

// getPrice - price tickers.
func (e *ExchangeService) getPrice(base, quote string) (price float64, ok bool) {

	if err := e.Context.Db.QueryRow("select price from pairs where base_unit = $1 and quote_unit = $2", base, quote).Scan(&price); err != nil {
		return price, ok
	}

	return price, true
}

// getRatio - pair ratio price.
func (e *ExchangeService) getRatio(base, quote string) (ratio float64, ok bool) {

	migrate, err := e.GetGraph(context.Background(), &proto.GetExchangeRequestGraph{BaseUnit: base, QuoteUnit: quote, Limit: 2})
	if err != nil {
		return ratio, ok
	}

	if len(migrate.Fields) == 2 {
		ratio = ((migrate.Fields[0].Close - migrate.Fields[1].Close) / migrate.Fields[1].Close) * 100
	}

	return ratio, true
}

// getReserve - unit reserve coins.
func (e *ExchangeService) getReserve(symbol string, platform proto.Platform, protocol proto.Protocol) (reserve float64) {
	_ = e.Context.Db.QueryRow(`select sum(value) from reserves where symbol = $1 and platform = $2 and protocol = $3`, symbol, platform, protocol).Scan(&reserve)
	return reserve
}

// setReserve - insert/update reserve.
func (e *ExchangeService) setReserve(userId int64, address, symbol string, value float64, platform proto.Platform, protocol proto.Protocol, cross proto.Balance) error {

	row, err := e.Context.Db.Query("select id from reserves where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5", userId, symbol, platform, protocol, address)
	if err != nil {
		return err
	}
	defer row.Close()

	if row.Next() {

		switch cross {
		case proto.Balance_PLUS:
			if _, err := e.Context.Db.Exec("update reserves set value = value + $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, protocol, address, value); err != nil {
				return err
			}
			break
		case proto.Balance_MINUS:
			if _, err := e.Context.Db.Exec("update reserves set value = value - $6 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4 and address = $5;", userId, symbol, platform, protocol, address, value); err != nil {
				return err
			}
			break
		}

		return nil
	}

	if _, err = e.Context.Db.Exec("insert into reserves (user_id, symbol, platform, protocol, address, value) values ($1, $2, $3, $4, $5, $6)", userId, symbol, platform, protocol, address, value); err != nil {
		return err
	}

	return nil
}

// setReserveLock - lock reserve asset.
func (e *ExchangeService) setReserveLock(userId int64, symbol string, platform proto.Platform, protocol proto.Protocol) error {
	if _, err := e.Context.Db.Exec("update reserves set lock = $5 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4;", userId, symbol, platform, protocol, true); err != nil {
		return err
	}
	return nil
}

// setReserveLock - unlock reserve asset.
func (e *ExchangeService) setReserveUnlock(userId int64, symbol string, platform proto.Platform, protocol proto.Protocol) error {
	if _, err := e.Context.Db.Exec("update reserves set lock = $5 where user_id = $1 and symbol = $2 and platform = $3 and protocol = $4;", userId, symbol, platform, protocol, false); err != nil {
		return err
	}
	return nil
}

// getStatus - base and quote status.
func (e *ExchangeService) getStatus(base, quote string) bool {

	if _, err := e.getCurrency(base, true); err != nil {
		return false
	}

	if _, err := e.getCurrency(quote, true); err != nil {
		return false
	}

	return true
}

// getUser - get user.
func (e *ExchangeService) getUser(id int64) (*proto.ResponseAccount, error) {
	migrate := AccountService{
		Context: e.Context,
	}

	return migrate.getUser(id)
}

// done - done group by chain id.
func (e *ExchangeService) done(id int64) {
	e.wait[id] = true
}
