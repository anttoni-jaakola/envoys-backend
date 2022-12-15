package spot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/query"
	"google.golang.org/grpc/status"
	"strings"
)

// GetMarketPriceRule - get market price.
func (e *Service) GetMarketPriceRule(ctx context.Context, req *pbspot.GetRequestPriceManual) (*pbspot.ResponsePrice, error) {

	var (
		response pbspot.ResponsePrice
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "pairs", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if price := marketplace.Price().Unit(req.GetBaseUnit(), req.GetQuoteUnit()); price > 0 {
		response.Price = price
	}

	return &response, nil
}

// SetCurrencyRule - create new insert/update currency.
func (e *Service) SetCurrencyRule(ctx context.Context, req *pbspot.SetRequestCurrencyRule) (*pbspot.ResponseCurrency, error) {

	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "currencies", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.Currency.GetName()) < 4 {
		return &response, e.Context.Error(status.Error(86618, "currency name must not be less than < 4 characters"))
	}

	if len(req.Currency.GetSymbol()) < 2 {
		return &response, e.Context.Error(status.Error(17078, "currency symbol must not be less than < 2 characters"))
	}

	serialize, err := json.Marshal(req.Currency.GetChainsIds())
	if err != nil {
		return &response, e.Context.Error(err)
	}

	req.Symbol = strings.ToLower(req.GetSymbol())
	req.Currency.Symbol = strings.ToLower(req.Currency.GetSymbol())

	if len(req.GetSymbol()) > 0 {

		if _, err := e.Context.Db.Exec("update spot_currencies set name = $1, symbol = $2, min_withdraw = $3, max_withdraw = $4, min_deposit = $5, min_trade = $6, max_trade = $7, fees_trade = $8, fees_discount = $9, marker = $10, status = $11, fin_type = $12, chains = $13 where symbol = $14;",
			req.Currency.GetName(),
			req.Currency.GetSymbol(),
			req.Currency.GetMinWithdraw(),
			req.Currency.GetMaxWithdraw(),
			req.Currency.GetMinDeposit(),
			req.Currency.GetMinTrade(),
			req.Currency.GetMaxTrade(),
			req.Currency.GetFeesTrade(),
			req.Currency.GetFeesDiscount(),
			req.Currency.GetMarker(),
			req.Currency.GetStatus(),
			req.Currency.GetFinType(),
			serialize,
			req.GetSymbol(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

		if req.GetSymbol() != req.Currency.GetSymbol() {
			_, _ = e.Context.Db.Exec("update spot_wallets set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update spot_assets set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update spot_trades set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2) where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update spot_transfers set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2)  where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update spot_orders set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2)  where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update spot_reserves set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update spot_currencies set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
		}

	} else {

		if _, err := e.Context.Db.Exec("insert into spot_currencies (name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, marker, fin_type, status, chains) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
			req.Currency.GetName(),
			req.Currency.GetSymbol(),
			req.Currency.GetMinWithdraw(),
			req.Currency.GetMaxWithdraw(),
			req.Currency.GetMinDeposit(),
			req.Currency.GetMinTrade(),
			req.Currency.GetMaxTrade(),
			req.Currency.GetFeesTrade(),
			req.Currency.GetFeesDiscount(),
			req.Currency.GetMarker(),
			req.Currency.GetFinType(),
			req.Currency.GetStatus(),
			serialize,
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}

	if len(req.GetImage()) > 0 {

		if len(req.GetSymbol()) > 0 {
			q.Name = req.GetSymbol()
		} else {
			q.Name = req.Currency.GetSymbol()
		}

		if err := migrate.Image(req.GetImage(), "icon", q.Name); err != nil {
			return &response, e.Context.Error(err)
		}
	} else {
		if req.GetSymbol() != req.Currency.GetSymbol() {
			if err := migrate.Rename("icon", req.GetSymbol(), req.Currency.GetSymbol()); err != nil {
				return &response, e.Context.Error(err)
			}
		}
	}

	return &response, nil
}

// GetCurrencyRule - currencies list.
func (e *Service) GetCurrencyRule(ctx context.Context, req *pbspot.GetRequestCurrencyRule) (*pbspot.ResponseCurrency, error) {

	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "currencies", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if currency, _ := e.getCurrency(req.GetSymbol(), false); currency.GetId() > 0 {
		response.Fields = append(response.Fields, currency)
	}

	return &response, nil
}

// GetCurrenciesRule - currencies list.
func (e *Service) GetCurrenciesRule(ctx context.Context, req *pbspot.GetRequestCurrenciesRule) (*pbspot.ResponseCurrency, error) {

	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "currencies", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where symbol like %[1]s or name like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from spot_currencies %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, fin_type, create_at from spot_currencies %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Currency
			)

			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.Symbol,
				&item.MinWithdraw,
				&item.MaxWithdraw,
				&item.MinDeposit,
				&item.MinTrade,
				&item.MaxTrade,
				&item.FeesTrade,
				&item.FeesDiscount,
				&item.FeesCharges,
				&item.FeesCosts,
				&item.Marker,
				&item.Status,
				&item.FinType,
				&item.CreateAt,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// DeleteCurrencyRule - delete currency by symbol.
func (e *Service) DeleteCurrencyRule(ctx context.Context, req *pbspot.DeleteRequestCurrencyRule) (*pbspot.ResponseCurrency, error) {

	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "currencies", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if row, _ := e.getUnit(req.GetSymbol()); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from spot_pairs where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
	}

	if row, _ := e.getCurrency(req.GetSymbol(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from spot_wallets where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from spot_assets where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from spot_trades where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from spot_transfers where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from spot_orders where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from spot_reserves where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from spot_currencies where symbol = $1", row.GetSymbol())
	}

	if err := migrate.Remove("icon", req.GetSymbol()); err != nil {
		return &response, e.Context.Error(err)
	}

	return &response, nil
}

// GetChainsRule - chains list.
func (e *Service) GetChainsRule(ctx context.Context, req *pbspot.GetRequestChainsRule) (*pbspot.ResponseChain, error) {

	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "chains", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	_ = e.Context.Db.QueryRow("select count(*) as count from spot_chains").Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(`select id, name, rpc, block, network, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, tag, status from spot_chains order by id desc limit $1 offset $2`, req.GetLimit(), offset)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Chain
			)

			if err = rows.Scan(&item.Id, &item.Name, &item.Rpc, &item.Block, &item.Network, &item.ExplorerLink, &item.Platform, &item.Confirmation, &item.TimeWithdraw, &item.FeesWithdraw, &item.Tag, &item.Status); err != nil {
				return &response, e.Context.Error(err)
			}

			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetChainRule - chain info.
func (e *Service) GetChainRule(ctx context.Context, req *pbspot.GetRequestChainRule) (*pbspot.ResponseChain, error) {

	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "chains", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if chain, _ := e.getChain(req.GetId(), false); chain.GetId() > 0 {
		response.Fields = append(response.Fields, chain)
	}

	return &response, nil
}

// SetChainRule - chain new insert/update info.
func (e *Service) SetChainRule(ctx context.Context, req *pbspot.SetRequestChainRule) (*pbspot.ResponseChain, error) {

	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "chains", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.Chain.GetName()) < 4 {
		return &response, e.Context.Error(status.Error(86611, "chain name must not be less than < 4 characters"))
	}

	if len(req.Chain.GetRpc()) < 10 {
		return &response, e.Context.Error(status.Error(44511, "chain rpc address must be at least < 10 characters"))
	}

	if ok := help.Ping(req.Chain.GetRpc()); !ok {
		return &response, e.Context.Error(status.Error(45601, "chain server address not available"))
	}

	if req.GetId() > 0 {

		if _, err := e.Context.Db.Exec("update spot_chains set name = $1, rpc = $2, rpc_key = $3, rpc_user = $4, rpc_password = $5, network = $6, block = $7, explorer_link = $8, platform = $9, confirmation = $10, time_withdraw = $11, fees_withdraw = $12, address = $13, tag = $14, parent_symbol = $15, status = $16 where id = $17;",
			req.Chain.GetName(),
			req.Chain.GetRpc(),
			req.Chain.GetRpcKey(),
			req.Chain.GetRpcUser(),
			req.Chain.GetRpcPassword(),
			req.Chain.GetNetwork(),
			req.Chain.GetBlock(),
			req.Chain.GetExplorerLink(),
			req.Chain.GetPlatform(),
			req.Chain.GetConfirmation(),
			req.Chain.GetTimeWithdraw(),
			req.Chain.GetFeesWithdraw(),
			req.Chain.GetAddress(),
			req.Chain.GetTag(),
			req.Chain.GetParentSymbol(),
			req.Chain.GetStatus(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		if _, err := e.Context.Db.Exec("insert into spot_chains (name, rpc, rpc_key, rpc_user, rpc_password, network, block, explorer_link, platform, confirmation, time_withdraw, fees_withdraw, address, tag, parent_symbol, status) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)",
			req.Chain.GetName(),
			req.Chain.GetRpc(),
			req.Chain.GetRpcKey(),
			req.Chain.GetRpcUser(),
			req.Chain.GetRpcPassword(),
			req.Chain.GetNetwork(),
			req.Chain.GetBlock(),
			req.Chain.GetExplorerLink(),
			req.Chain.GetPlatform(),
			req.Chain.GetConfirmation(),
			req.Chain.GetTimeWithdraw(),
			req.Chain.GetFeesWithdraw(),
			req.Chain.GetAddress(),
			req.Chain.GetTag(),
			req.Chain.GetParentSymbol(),
			req.Chain.GetStatus(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// DeleteChainRule - delete chain by id.
func (e *Service) DeleteChainRule(ctx context.Context, req *pbspot.DeleteRequestChainRule) (*pbspot.ResponseChain, error) {

	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "currencies", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if row, _ := e.getChain(req.GetId(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec(fmt.Sprintf(`update spot_currencies set chains = jsonb_path_query_array(chains, '$[*] ? (@ != %[1]d)') where chains @> '%[1]d'`, row.GetId()))
		_, _ = e.Context.Db.Exec("delete from spot_chains where id = $1", row.GetId())
	}
	response.Success = true

	return &response, nil
}

// GetPairsRule - get all pairs.
func (e *Service) GetPairsRule(ctx context.Context, req *pbspot.GetRequestPairsRule) (*pbspot.ResponsePair, error) {

	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "pairs", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where base_unit like %[1]s or quote_unit like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from spot_pairs %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from spot_pairs %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Pair
			)

			if err = rows.Scan(
				&item.Id,
				&item.BaseUnit,
				&item.QuoteUnit,
				&item.Price,
				&item.BaseDecimal,
				&item.QuoteDecimal,
				&item.Status,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetPairRule - get pair by id.
func (e *Service) GetPairRule(ctx context.Context, req *pbspot.GetRequestPairRule) (*pbspot.ResponsePair, error) {

	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "pairs", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if pair, _ := e.getPair(req.GetId(), false); pair.GetId() > 0 {
		response.Fields = append(response.Fields, pair)
	}

	return &response, nil
}

// SetPairRule - set or update pair by id.
func (e *Service) SetPairRule(ctx context.Context, req *pbspot.SetRequestPairRule) (*pbspot.ResponsePair, error) {

	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "pairs", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.Pair.GetBaseUnit()) == 0 && len(req.Pair.GetQuoteUnit()) == 0 {
		return &response, e.Context.Error(status.Error(55615, "base currency and quote currency must be set"))
	}

	if req.Pair.GetPrice() == 0 {
		return &response, e.Context.Error(status.Error(46517, "the price must be set"))
	}

	if req.GetId() > 0 {

		if req.Pair.GetGraphClear() {
			_, _ = e.Context.Db.Exec("delete from spot_trades where base_unit = $1 and quote_unit = $2", req.Pair.GetBaseUnit(), req.Pair.GetQuoteUnit())
		}

		if _, err := e.Context.Db.Exec("update spot_pairs set base_unit = $1, quote_unit = $2, price = $3, base_decimal = $4, quote_decimal = $5, status = $6 where id = $7;",
			req.Pair.GetBaseUnit(),
			req.Pair.GetQuoteUnit(),
			req.Pair.GetPrice(),
			req.Pair.GetBaseDecimal(),
			req.Pair.GetQuoteDecimal(),
			req.Pair.GetStatus(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		if _ = e.Context.Db.QueryRow("select id from spot_pairs where base_unit = $1 and quote_unit = $2 or base_unit = $2 and quote_unit = $1", req.Pair.GetBaseUnit(), req.Pair.GetQuoteUnit()).Scan(&q.Id); q.Id > 0 {
			return &response, e.Context.Error(status.Error(50605, "the pair you are trying to create is already in the list of pairs"))
		}

		if _, err := e.Context.Db.Exec("insert into spot_pairs (base_unit, quote_unit, price, base_decimal, quote_decimal, status) values ($1, $2, $3, $4, $5, $6)",
			req.Pair.GetBaseUnit(),
			req.Pair.GetQuoteUnit(),
			req.Pair.GetPrice(),
			req.Pair.GetBaseDecimal(),
			req.Pair.GetQuoteDecimal(),
			req.Pair.GetStatus(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// DeletePairRule - delete pair by id.
func (e *Service) DeletePairRule(ctx context.Context, req *pbspot.DeleteRequestPairRule) (*pbspot.ResponsePair, error) {

	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "pairs", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if row, _ := e.getPair(req.GetId(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from spot_pairs where id = $1", row.GetId())
		_, _ = e.Context.Db.Exec("delete from spot_trades where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
		_, _ = e.Context.Db.Exec("delete from spot_transfers where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
		_, _ = e.Context.Db.Exec("delete from spot_orders where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
	}
	response.Success = true

	return &response, nil
}

// GetContractsRule - get all contract.
func (e *Service) GetContractsRule(ctx context.Context, req *pbspot.GetRequestContractsRule) (*pbspot.ResponseContract, error) {

	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "contracts", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where c.address like %[1]s or c.symbol like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from spot_contracts c %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf("select c.id, c.symbol, c.chain_id, c.address, c.fees_withdraw, c.protocol, n.platform, n.parent_symbol from spot_contracts c inner join spot_chains n on n.id = c.chain_id %s order by c.id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Contract
			)

			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.ChainId,
				&item.Address,
				&item.FeesWithdraw,
				&item.Protocol,
				&item.Platform,
				&item.ParentSymbol,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			if chain, _ := e.getChain(item.GetChainId(), false); chain.GetId() > 0 {
				item.ChainName = chain.GetName()
			}

			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetContractRule - get contract by id.
func (e *Service) GetContractRule(ctx context.Context, req *pbspot.GetRequestContractRule) (*pbspot.ResponseContract, error) {

	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "contracts", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if row, _ := e.getContractById(req.GetId()); row.GetId() > 0 {
		response.Fields = append(response.Fields, row)
	}

	return &response, nil
}

// SetContractRule - set new contract.
func (e *Service) SetContractRule(ctx context.Context, req *pbspot.SetRequestContractRule) (*pbspot.ResponseContract, error) {

	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "contracts", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.Contract.GetSymbol()) == 0 {
		return &response, e.Context.Error(status.Error(56616, "contract/currency symbol required"))
	}

	if err := e.helperAddress(req.Contract.GetAddress(), req.Contract.GetPlatform()); err != nil {
		return &response, e.Context.Error(err)
	}

	if req.GetId() > 0 {

		if _, err := e.Context.Db.Exec("update spot_contracts set symbol = $1, chain_id = $2, address = $3, fees_withdraw = $4, protocol = $5, decimals = $6 where id = $7;",
			req.Contract.GetSymbol(),
			req.Contract.GetChainId(),
			req.Contract.GetAddress(),
			req.Contract.GetFeesWithdraw(),
			req.Contract.GetProtocol(),
			req.Contract.GetDecimals(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		if _, err := e.Context.Db.Exec("insert into spot_contracts (symbol, chain_id, address, fees_withdraw, protocol, decimals) values ($1, $2, $3, $4, $5, $6)",
			req.Contract.GetSymbol(),
			req.Contract.GetChainId(),
			req.Contract.GetAddress(),
			req.Contract.GetFeesWithdraw(),
			req.Contract.GetProtocol(),
			req.Contract.GetDecimals(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// DeleteContractRule - delete contract by id.
func (e *Service) DeleteContractRule(ctx context.Context, req *pbspot.DeleteRequestContractRule) (*pbspot.ResponseContract, error) {

	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "contracts", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if row, _ := e.getContractById(req.GetId()); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from spot_contracts where id = $1", row.GetId())
		_, _ = e.Context.Db.Exec("delete from spot_wallets where symbol = $1 and protocol = $2", row.GetSymbol(), row.GetProtocol())
		_, _ = e.Context.Db.Exec("delete from spot_transactions where symbol = $1 and protocol = $2", row.GetSymbol(), row.GetProtocol())
		_, _ = e.Context.Db.Exec("delete from spot_reserves where symbol = $1 and protocol = $2", row.GetSymbol(), row.GetProtocol())
	}
	response.Success = true

	return &response, nil
}

// GetTransactionsRule - get transactions by user id.
func (e *Service) GetTransactionsRule(ctx context.Context, req *pbspot.GetRequestTransactionsManual) (*pbspot.ResponseTransaction, error) {

	var (
		response pbspot.ResponseTransaction
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts", query.RoleSpot) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	switch req.GetTxType() {
	case pbspot.TxType_DEPOSIT:
		maps = append(maps, fmt.Sprintf("where tx_type = %d", pbspot.TxType_DEPOSIT))
	case pbspot.TxType_WITHDRAWS:
		maps = append(maps, fmt.Sprintf("where tx_type = %d", pbspot.TxType_WITHDRAWS))
	default:
		maps = append(maps, fmt.Sprintf("where (tx_type = %d or tx_type = %d)", pbspot.TxType_WITHDRAWS, pbspot.TxType_DEPOSIT))
	}

	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("and (symbol like %[1]s or id::text like %[1]s or hash like %[1]s)", "'%"+req.GetSearch()+"%'"))
	}
	maps = append(maps, fmt.Sprintf("and user_id = %v", req.GetId()))

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from spot_transactions %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, symbol, hash, value, price, fees, confirmation, "to", chain_id, user_id, tx_type, fin_type, platform, protocol, status, create_at from spot_transactions %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Transaction
			)

			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.Hash,
				&item.Value,
				&item.Price,
				&item.Fees,
				&item.Confirmation,
				&item.To,
				&item.ChainId,
				&item.UserId,
				&item.TxType,
				&item.FinType,
				&item.Platform,
				&item.Protocol,
				&item.Status,
				&item.CreateAt,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			item.Chain, err = e.getChain(item.GetChainId(), false)
			if err != nil {
				return nil, err
			}
			item.ChainId = 0

			if item.GetProtocol() != pbspot.Protocol_MAINNET {
				item.Fees = decimal.FromFloat(item.GetFees()).Mul(decimal.FromFloat(item.GetPrice())).Float64()
			}

			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}

	}

	return &response, nil
}

// GetOrdersRule - get orders by user id.
func (e *Service) GetOrdersRule(ctx context.Context, req *pbspot.GetRequestOrdersManual) (*pbspot.ResponseOrder, error) {
	//TODO implement me
	panic("implement me")
}

// GetAssetsRule - get assets by user id.
func (e *Service) GetAssetsRule(ctx context.Context, req *pbspot.GetRequestAssetsManual) (*pbspot.ResponseAsset, error) {
	//TODO implement me
	panic("implement me")
}
