package service

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"google.golang.org/grpc/status"
	"strings"
)

// GetSymbol - check if exchange unit exists or does not exist.
func (e *ExchangeService) GetSymbol(_ context.Context, req *proto.GetExchangeRequestSymbol) (*proto.ResponseSymbol, error) {

	var (
		response proto.ResponseSymbol
	)

	if row, err := e.getCurrency(req.GetBaseUnit(), false); err != nil {
		return &response, e.Context.Error(status.Errorf(11584, "this base currency does not exist, %v", row.GetSymbol()))
	}

	if row, err := e.getCurrency(req.GetQuoteUnit(), false); err != nil {
		return &response, e.Context.Error(status.Errorf(11582, "this quote currency does not exist, %v", row.GetSymbol()))
	}
	response.Success = true

	return &response, nil
}

// GetAnalysis - analysis list.
func (e *ExchangeService) GetAnalysis(ctx context.Context, req *proto.GetExchangeRequestAnalysis) (*proto.ResponseAnalysis, error) {

	var (
		response proto.ResponseAnalysis
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	_ = e.Context.Db.QueryRow(`select count(*) as count from pairs where status = $1`, true).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(`select id, base_unit, quote_unit, price from pairs where status = $3 order by id desc limit $1 offset $2`, req.GetLimit(), offset, true)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				analysis proto.Analysis
			)

			if err := rows.Scan(&analysis.Id, &analysis.BaseUnit, &analysis.QuoteUnit, &analysis.Price); err != nil {
				return &response, e.Context.Error(err)
			}

			migrate, err := e.GetGraph(context.Background(), &proto.GetExchangeRequestGraph{BaseUnit: analysis.GetBaseUnit(), QuoteUnit: analysis.GetQuoteUnit(), Limit: 40})
			if err != nil {
				return &response, e.Context.Error(err)
			}

			for i := 0; i < len(migrate.Fields); i++ {
				analysis.Chart = append(analysis.Chart, migrate.Fields[i].GetPrice())
			}

			for i := 0; i < 2; i++ {

				var (
					assigning proto.Assigning
				)

				switch i {
				case 0:
					assigning = proto.Assigning_BUY
				case 1:
					assigning = proto.Assigning_SELL
				}

				migrate, err := e.GetOrders(context.Background(), &proto.GetExchangeRequestOrdersManual{
					BaseUnit:  analysis.GetBaseUnit(),
					QuoteUnit: analysis.GetQuoteUnit(),
					Assigning: assigning,
					Status:    proto.Status_FILLED,
					UserId:    account,
					Limit:     2,
				})
				if err != nil {
					return &response, e.Context.Error(err)
				}

				if len(migrate.GetFields()) == 2 {
					switch i {
					case 0:
						analysis.BuyRatio = decimal.FromFloat(100 * (analysis.GetPrice() - migrate.Fields[0].GetPrice()) / migrate.Fields[0].GetPrice()).Round(2).Float64()
					case 1:
						analysis.SelRatio = decimal.FromFloat(100 * (analysis.GetPrice() - migrate.Fields[0].GetPrice()) / migrate.Fields[0].GetPrice()).Round(2).Float64()
					}
				}
			}

			response.Fields = append(response.Fields, &analysis)
		}
	}

	return &response, nil
}

// GetMarkers - show marker zone.
func (e *ExchangeService) GetMarkers(_ context.Context, _ *proto.GetExchangeRequestMarkers) (*proto.ResponseMarker, error) {

	var (
		response proto.ResponseMarker
	)

	rows, err := e.Context.Db.Query("select symbol from currencies where marker = $1", true)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			symbol string
		)

		if err := rows.Scan(&symbol); err != nil {
			return &response, e.Context.Error(err)
		}

		response.Fields = append(response.Fields, symbol)
	}

	return &response, nil
}

// GetPairs - show all asset pairs with a positive status.
func (e *ExchangeService) GetPairs(_ context.Context, req *proto.GetExchangeRequestPairs) (*proto.ResponsePair, error) {

	var (
		response proto.ResponsePair
	)

	rows, err := e.Context.Db.Query("select id, base_unit, quote_unit, base_decimal, quote_decimal, status from pairs where base_unit = $1 or quote_unit = $1", req.GetSymbol())
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			pair proto.Pair
		)

		if err := rows.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit, &pair.BaseDecimal, &pair.QuoteDecimal, &pair.Status); err != nil {
			return &response, e.Context.Error(err)
		}

		if req.GetSymbol() == pair.GetQuoteUnit() {
			pair.Symbol = pair.GetBaseUnit()
		} else {
			pair.Symbol = pair.GetQuoteUnit()
		}

		if ratio, ok := e.getRatio(pair.GetBaseUnit(), pair.GetQuoteUnit()); ok {
			pair.Ratio = ratio
		}

		if price, ok := e.getPrice(pair.GetBaseUnit(), pair.GetQuoteUnit()); ok {
			pair.Price = price
		}

		if ok := e.getStatus(pair.GetBaseUnit(), pair.GetQuoteUnit()); !ok {
			pair.Status = false
		}

		response.Fields = append(response.Fields, &pair)
	}

	return &response, nil
}

// GetPair - show pair details.
func (e *ExchangeService) GetPair(_ context.Context, req *proto.GetExchangeRequestPair) (*proto.ResponsePair, error) {

	var (
		response proto.ResponsePair
	)

	row, err := e.Context.Db.Query(`select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs where base_unit = $1 and quote_unit = $2`, req.GetBaseUnit(), req.GetQuoteUnit())
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {

		var (
			pair proto.Pair
		)

		if err := row.Scan(&pair.Id, &pair.BaseUnit, &pair.QuoteUnit, &pair.Price, &pair.BaseDecimal, &pair.QuoteDecimal, &pair.Status); err != nil {
			return &response, e.Context.Error(err)
		}

		if ok := e.getStatus(req.GetBaseUnit(), req.GetQuoteUnit()); !ok {
			pair.Status = false
		}

		response.Fields = append(response.Fields, &pair)
	}

	return &response, nil
}

// SetOrder - create a new order, search for new orders to make a deal.
func (e *ExchangeService) SetOrder(ctx context.Context, req *proto.SetExchangeRequestOrder) (*proto.ResponseOrder, error) {

	var (
		response proto.ResponseOrder
		order    proto.Order
		err      error
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	user, err := e.getUser(account)
	if err != nil {
		return nil, err
	}
	if !user.Status {
		return &response, e.Context.Error(status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions"))
	}

	switch req.GetTradeType() {
	case proto.TradeType_MARKET:
		order.Price = e.getMarket(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetAssigning(), req.GetPrice())
	case proto.TradeType_LIMIT:
		order.Price = req.GetPrice()
	default:
		return &response, e.Context.Error(status.Error(82284, "invalid type trade position"))
	}

	if err := e.helperSymbol(req.GetBaseUnit()); err != nil {
		return &response, e.Context.Error(err)
	}

	if err := e.helperSymbol(req.GetQuoteUnit()); err != nil {
		return &response, e.Context.Error(err)
	}

	if err := e.helperPair(req.GetBaseUnit(), req.GetQuoteUnit()); err != nil {
		return &response, e.Context.Error(err)
	}

	order.UserId = user.GetId()
	order.BaseUnit = req.GetBaseUnit()
	order.QuoteUnit = req.GetQuoteUnit()
	order.Quantity = req.GetQuantity()
	order.Value = req.GetQuantity()
	order.Assigning = req.GetAssigning()
	order.Status = proto.Status_PENDING

	quantity, err := e.helperOrder(&order)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if order.Id, err = e.setOrder(&order); err != nil {
		return &response, e.Context.Error(err)
	}

	switch req.GetAssigning() {
	case proto.Assigning_BUY:

		if err := e.setAsset(order.GetBaseUnit(), order.GetUserId(), false); err != nil {
			return &response, e.Context.Error(err)
		}

		if err := e.setBalance(order.GetQuoteUnit(), order.GetUserId(), quantity, proto.Balance_MINUS); err != nil {
			return &response, e.Context.Error(err)
		}

		e.replayTradeInit(&order, proto.Spread_BID)

		break
	case proto.Assigning_SELL:

		if err := e.setAsset(order.GetQuoteUnit(), order.GetUserId(), false); err != nil {
			return &response, e.Context.Error(err)
		}

		if err := e.setBalance(order.GetBaseUnit(), order.GetUserId(), quantity, proto.Balance_MINUS); err != nil {
			return &response, e.Context.Error(err)
		}

		e.replayTradeInit(&order, proto.Spread_ASK)

		break
	default:
		return &response, status.Error(11588, "invalid assigning trade position")
	}

	response.Fields = append(response.Fields, &order)

	return &response, nil
}

// GetOrders - show all active orders.
func (e *ExchangeService) GetOrders(ctx context.Context, req *proto.GetExchangeRequestOrdersManual) (*proto.ResponseOrder, error) {

	var (
		response proto.ResponseOrder
		query    []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case proto.Assigning_BUY:
		query = append(query, fmt.Sprintf("where assigning = %d", proto.Assigning_BUY))
	case proto.Assigning_SELL:
		query = append(query, fmt.Sprintf("where assigning = %d", proto.Assigning_SELL))
	default:
		query = append(query, fmt.Sprintf("where (assigning = %d or assigning = %d)", proto.Assigning_BUY, proto.Assigning_SELL))
	}

	if req.GetOwner() {
		account, err := e.Context.Auth(ctx)
		if err != nil {
			return &response, e.Context.Error(err)
		}

		query = append(query, fmt.Sprintf("and user_id = '%v'", account))
	} else if req.GetUserId() > 0 {
		query = append(query, fmt.Sprintf("and user_id = '%v'", req.GetUserId()))
	}

	switch req.GetStatus() {
	case proto.Status_FILLED:
		query = append(query, fmt.Sprintf("and status = %d", proto.Status_FILLED))
	case proto.Status_PENDING:
		query = append(query, fmt.Sprintf("and status = %d", proto.Status_PENDING))
	}

	if req.GetDecimal() > 0 {
		query = append(query, fmt.Sprintf("and (value between %[2]v and %[1]v)", req.GetDecimal(), 0))
	}

	// Get request to display pair information.
	if len(req.GetBaseUnit()) > 0 && len(req.GetQuoteUnit()) > 0 {
		query = append(query, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))
	}

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count, sum(value) as volume from orders %s", strings.Join(query, " "))).Scan(&response.Count, &response.Volume)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, assigning, price, value, quantity, base_unit, quote_unit, user_id, create_at, status from orders %s order by id desc limit %d offset %d", strings.Join(query, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item proto.Order
			)

			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Value, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt, &item.Status); err != nil {
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

// GetAssets - show all assets with a positive status.
func (e *ExchangeService) GetAssets(ctx context.Context, _ *proto.GetExchangeRequestAssetsManual) (*proto.ResponseAsset, error) {

	var (
		response proto.ResponseAsset
	)

	rows, err := e.Context.Db.Query("select id, name, symbol, status from currencies")
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			asset proto.Currency
		)

		if err := rows.Scan(&asset.Id, &asset.Name, &asset.Symbol, &asset.Status); err != nil {
			return &response, e.Context.Error(err)
		}

		if account, err := e.Context.Auth(ctx); err == nil {
			if balance := e.getBalance(asset.GetSymbol(), account); balance > 0 {
				asset.Balance = balance
			}
		}

		response.Fields = append(response.Fields, &asset)
	}

	return &response, nil
}

// SetAsset - new or update asset address.
func (e *ExchangeService) SetAsset(ctx context.Context, req *proto.SetExchangeRequestAsset) (*proto.ResponseAsset, error) {

	var (
		response proto.ResponseAsset
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	switch req.GetFinType() {
	case proto.FinType_CRYPTO:

		var (
			cross keypair.CrossChain
		)

		// Create a new record with asset wallet data, address and entropy.
		entropy, err := e.getEntropy(account)
		if err != nil {
			return &response, e.Context.Error(err)
		}

		if response.Address, _, err = cross.New(fmt.Sprintf("%v-&*39~763@)", e.Context.Secrets[1]), entropy, req.GetPlatform()); err != nil {
			return &response, e.Context.Error(err)
		}

		row, err := e.Context.Db.Query(`select id from assets where symbol = $1 and user_id = $2`, req.GetSymbol(), account)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer row.Close()

		if row.Next() {

			if address := e.getAddress(account, req.GetSymbol(), req.GetPlatform(), req.GetProtocol()); len(address) == 0 {
				if _, err = e.Context.Db.Exec("insert into wallets (address, symbol, platform, protocol, user_id) values ($1, $2, $3, $4, $5)", response.GetAddress(), req.GetSymbol(), req.GetPlatform(), req.GetProtocol(), account); err != nil {
					return &response, e.Context.Error(err)
				}

				return &response, nil
			}

			return &response, e.Context.Error(status.Error(700990, "the asset address has already been generated"))
		}

		if _, err = e.Context.Db.Exec("insert into assets (user_id, symbol) values ($1, $2);", account, req.GetSymbol()); err != nil {
			return &response, e.Context.Error(err)
		}

		if _, err = e.Context.Db.Exec("insert into wallets (address, symbol, platform, protocol, user_id) values ($1, $2, $3, $4, $5)", response.GetAddress(), req.GetSymbol(), req.GetPlatform(), req.GetProtocol(), account); err != nil {
			return &response, e.Context.Error(err)
		}

	case proto.FinType_FIAT:

		if err := e.setAsset(req.GetSymbol(), account, true); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// GetAsset - show asset information.
func (e *ExchangeService) GetAsset(ctx context.Context, req *proto.GetExchangeRequestAsset) (*proto.ResponseAsset, error) {

	var (
		response proto.ResponseAsset
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	row, err := e.getCurrency(req.GetSymbol(), false)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	row.Balance = e.getBalance(req.GetSymbol(), account)

	_ = e.Context.Db.QueryRow(`select coalesce(sum(case when base_unit = $1 then value when quote_unit = $1 then value * price end), 0.00) as volume from orders where base_unit = $1 and assigning = $2 and status = $4 and user_id = $5 or quote_unit = $1 and assigning = $3 and status = $4 and user_id = $5`, req.GetSymbol(), proto.Assigning_SELL, proto.Assigning_BUY, proto.Status_PENDING, account).Scan(&row.Volume)

	for i := 0; i < len(row.GetChainsIds()); i++ {

		if chain, _ := e.getChain(row.GetChainsIds()[i], false); chain.GetId() > 0 {
			chain.Rpc, chain.RpcKey, chain.RpcUser, chain.RpcPassword, chain.Address = "", "", "", "", ""

			if contract, _ := e.getContract(row.GetSymbol(), row.GetChainsIds()[i]); contract.GetId() > 0 {
				chain.FeesWithdraw, contract.FeesWithdraw = contract.GetFeesWithdraw(), 0

				price, err := e.GetPrice(context.Background(), &proto.GetExchangeRequestPriceManual{
					BaseUnit:  chain.GetParentSymbol(),
					QuoteUnit: req.GetSymbol(),
				})
				if err != nil {
					return &response, e.Context.Error(err)
				}

				chain.FeesWithdraw = decimal.FromFloat(chain.GetFeesWithdraw()).Mul(decimal.FromFloat(price.GetPrice())).Float64()
				chain.Contract = contract
			}

			chain.Reserve = e.getReserve(req.GetSymbol(), chain.GetPlatform(), chain.Contract.GetProtocol())
			chain.Address = e.getAddress(account, req.GetSymbol(), chain.GetPlatform(), chain.Contract.GetProtocol())
			chain.Exist = e.getAsset(req.GetSymbol(), account)

			row.Chains = append(row.Chains, chain)
		}

	}
	row.ChainsIds = make([]int64, 0)

	response.Fields = append(response.Fields, row)

	return &response, nil
}

// GetGraph - show statistics of pair trades.
func (e *ExchangeService) GetGraph(_ context.Context, req *proto.GetExchangeRequestGraph) (*proto.ResponseGraph, error) {

	var (
		response proto.ResponseGraph
		query    []string
		limit    string
	)

	if req.GetLimit() > 0 {
		limit = fmt.Sprintf("limit %d", req.GetLimit())
	}

	if req.GetFrom() > 0 && req.GetTo() > 0 {
		query = append(query, fmt.Sprintf(`and to_char(ohlc.create_at, 'yyyy-mm-dd hh24:mi:ss.ff6 +00:00')::timestamptz between to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss.ff6 +00:00')::timestamptz and to_char(to_timestamp(%[2]d), 'yyyy-mm-dd hh24:mi:ss.ff6 +00:00')::timestamptz`, req.GetFrom(), req.GetTo()))
	}

	/**
	 *  The time slicing queries in various databases are done differently.
	 *  Postgres supports series() mysql does not, timescale has buckets, the others don't etc.
	 *  having support for timescaledb is important for the scale of the project.
	 */
	rows, err := e.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', ohlc.create_at))::integer buckettime, first(ohlc.price, ohlc.create_at) as open, last(ohlc.price, ohlc.create_at) as close, first(ohlc.price, ohlc.price) as low, last(ohlc.price, ohlc.price) as high, sum(ohlc.quantity) as volume, avg(ohlc.price) as avg_price, ohlc.base_unit, ohlc.quote_unit from trades as ohlc where ohlc.base_unit = '%[1]s' and ohlc.quote_unit = '%[2]s' %[3]s group by buckettime, ohlc.base_unit, ohlc.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(query, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, e.Context.Error(err)
	}

	defer rows.Close()

	for rows.Next() {

		var (
			item proto.Graph
		)

		if err = rows.Scan(&item.Time, &item.Open, &item.Close, &item.Low, &item.High, &item.Volume, &item.Price, &item.BaseUnit, &item.QuoteUnit); err != nil {
			return &response, e.Context.Error(err)
		}

		response.Fields = append(response.Fields, &item)
	}

	var (
		stats proto.Stats
	)

	_ = e.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from trades as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

	if len(response.Fields) > 1 {
		stats.Previous = response.Fields[1].Close
	}
	response.Stats = &stats

	return &response, nil
}

// GetTransfers - order stats list.
func (e *ExchangeService) GetTransfers(ctx context.Context, req *proto.GetExchangeRequestTransfers) (*proto.ResponseTransfer, error) {

	var (
		response proto.ResponseTransfer
		query    []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case proto.Assigning_BUY:
		query = append(query, fmt.Sprintf("where assigning = %d", proto.Assigning_BUY))
	case proto.Assigning_SELL:
		query = append(query, fmt.Sprintf("where assigning = %d", proto.Assigning_SELL))
	default:
		query = append(query, fmt.Sprintf("where (assigning = %d or assigning = %d)", proto.Assigning_BUY, proto.Assigning_SELL))
	}

	query = append(query, fmt.Sprintf("and order_id = '%v'", req.GetOrderId()))

	if req.GetOwner() {
		account, err := e.Context.Auth(ctx)
		if err != nil {
			return &response, e.Context.Error(err)
		}

		query = append(query, fmt.Sprintf("and user_id = '%v'", account))
	}

	rows, err := e.Context.Db.Query(fmt.Sprintf("select id, user_id, base_unit, quote_unit, price, quantity, assigning, fees, create_at from transfers %s order by id desc limit %d", strings.Join(query, " "), req.GetLimit()))
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item proto.Transfer
		)

		if err = rows.Scan(&item.Id, &item.UserId, &item.BaseUnit, &item.QuoteUnit, &item.Price, &item.Quantity, &item.Assigning, &item.Fees, &item.CreateAt); err != nil {
			return &response, e.Context.Error(err)
		}

		response.Fields = append(response.Fields, &item)
	}

	if err = rows.Err(); err != nil {
		return &response, e.Context.Error(err)
	}

	return &response, nil
}

// GetTrades - show all trades.
func (e *ExchangeService) GetTrades(_ context.Context, req *proto.GetExchangeRequestTrades) (*proto.ResponseTrades, error) {

	var (
		response proto.ResponseTrades
		query    []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case proto.Assigning_BUY:
		query = append(query, fmt.Sprintf("where assigning = %d", proto.Assigning_BUY))
	case proto.Assigning_SELL:
		query = append(query, fmt.Sprintf("where assigning = %d", proto.Assigning_SELL))
	default:
		query = append(query, fmt.Sprintf("where (assigning = %d or assigning = %d)", proto.Assigning_BUY, proto.Assigning_SELL))
	}

	// Get request to display pair information.
	query = append(query, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from trades %s", strings.Join(query, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, assigning, price, quantity, base_unit, quote_unit, create_at from trades %s order by id desc limit %d offset %d`, strings.Join(query, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item proto.Trade
			)

			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.CreateAt); err != nil {
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

// GetPrice - Конвертация валюты.
func (e *ExchangeService) GetPrice(_ context.Context, req *proto.GetExchangeRequestPriceManual) (*proto.ResponsePrice, error) {

	var (
		response proto.ResponsePrice
		ok       bool
	)

	if response.Price, ok = e.getPrice(req.GetBaseUnit(), req.GetQuoteUnit()); ok {
		return &response, nil
	}

	if response.Price, ok = e.getPrice(req.GetQuoteUnit(), req.GetBaseUnit()); ok {
		response.Price = decimal.FromFloat(decimal.FromFloat(1).Div(decimal.FromFloat(response.Price)).Float64()).Round(8).Float64()
	}

	return &response, nil
}

// GetTransactions - вывод входящих и исходящих транзакций.
func (e *ExchangeService) GetTransactions(ctx context.Context, req *proto.GetExchangeRequestTransactionsManual) (*proto.ResponseTransaction, error) {

	var (
		response proto.ResponseTransaction
		query    []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetTxType() {
	case proto.TxType_DEPOSIT:
		query = append(query, fmt.Sprintf("where tx_type = %d", proto.TxType_DEPOSIT))
	case proto.TxType_WITHDRAWS:
		query = append(query, fmt.Sprintf("where tx_type = %d", proto.TxType_WITHDRAWS))
	default:
		query = append(query, fmt.Sprintf("where (tx_type = %d or tx_type = %d)", proto.TxType_WITHDRAWS, proto.TxType_DEPOSIT))
	}

	if len(req.GetSymbol()) > 0 {
		query = append(query, fmt.Sprintf("and symbol = '%v'", req.GetSymbol()))
	}

	query = append(query, fmt.Sprintf("and status != %d", proto.Status_RESERVE))

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	query = append(query, fmt.Sprintf("and user_id = %v", account))

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from transactions %s", strings.Join(query, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, symbol, hash, value, price, fees, confirmation, "to", chain_id, user_id, tx_type, fin_type, platform, protocol, status, create_at from transactions %s order by id desc limit %d offset %d`, strings.Join(query, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item proto.Transaction
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

			if item.GetProtocol() != proto.Protocol_MAINNET {
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

// SetWithdraw - Записываем запросы на вывод средств.
func (e *ExchangeService) SetWithdraw(ctx context.Context, req *proto.SetExchangeRequestWithdraw) (*proto.ResponseWithdraw, error) {

	var (
		response proto.ResponseWithdraw
		fees     float64
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if req.GetRefresh() {
		if err := e.setSecure(ctx, false); err != nil {
			return &response, e.Context.Error(err)
		}

		return &response, nil
	}

	// Проверяем пользователя на блокировку активов, если пользователь блокирован то не пропускаем его дальше.
	user, err := e.getUser(account)
	if err != nil {
		return nil, err
	}
	if !user.Status {
		return &response, e.Context.Error(status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions"))
	}

	// Проверяем что бы пользователь не выводил средства на локальный адрес, то есть адрес который был сгенерирован в бирже.
	if err := e.helperInternalAsset(req.GetAddress()); err != nil {
		return &response, e.Context.Error(err)
	}

	chain, err := e.getChain(req.GetId(), true)
	if err != nil {
		return &response, e.Context.Error(status.Errorf(11584, "the chain array by id %v is currently unavailable", req.GetId()))
	}

	currency, err := e.getCurrency(req.GetSymbol(), false)
	if err != nil {
		return &response, e.Context.Error(status.Errorf(10029, "the currency requested array by id %v is currently unavailable", req.GetSymbol()))
	}

	contract, _ := e.getContract(req.GetSymbol(), chain.GetId())
	if contract.GetProtocol() != proto.Protocol_MAINNET {
		price, err := e.GetPrice(context.Background(), &proto.GetExchangeRequestPriceManual{
			BaseUnit:  chain.GetParentSymbol(),
			QuoteUnit: req.GetSymbol(),
		})
		if err != nil {
			return &response, e.Context.Error(err)
		}

		req.Price = price.GetPrice()
		chain.FeesWithdraw = contract.GetFeesWithdraw()

		fees = decimal.FromFloat(contract.GetFeesWithdraw()).Mul(decimal.FromFloat(req.GetPrice())).Float64()
	} else {
		fees = chain.GetFeesWithdraw()
	}

	// Предупреждаем пользователя, о том что вывод на тот же адрес, с которого производится вывод, недопустимо.
	if address := e.getAddress(account, req.GetSymbol(), req.GetPlatform(), contract.GetProtocol()); address == strings.ToLower(req.GetAddress()) {
		return &response, e.Context.Error(status.Error(758690, "your cannot send from an address to the same address"))
	}

	secure, err := e.getSecure(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if len(req.GetSecure()) != 6 {
		return &response, e.Context.Error(status.Error(16763, "the code must be 6 numbers"))
	}

	if secure != req.GetSecure() || secure == "" {
		return &response, e.Context.Error(status.Errorf(58990, "security code %v is incorrect", req.GetSecure()))
	}

	// Проверяем корректность адреса.
	if err := e.helperAddress(req.GetAddress(), req.GetPlatform()); err != nil {
		return &response, e.Context.Error(err)
	}

	if err := e.helperWithdraw(req.GetQuantity(), e.getReserve(req.GetSymbol(), req.GetPlatform(), contract.GetProtocol()), e.getBalance(req.GetSymbol(), account), currency.GetMaxWithdraw(), currency.GetMinWithdraw(), fees); err != nil {
		return &response, e.Context.Error(err)
	}

	if err := e.setBalance(req.GetSymbol(), account, req.GetQuantity(), proto.Balance_MINUS); e.Context.Debug(err) {
		return &response, e.Context.Error(err)
	}

	if _, err := e.Context.Db.Exec(`insert into transactions (symbol, value, price, "to", chain_id, platform, protocol, fees, user_id, tx_type, fin_type) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		req.GetSymbol(),
		req.GetQuantity(),
		req.GetPrice(),
		req.GetAddress(),
		req.GetId(),
		req.GetPlatform(),
		contract.GetProtocol(),
		chain.GetFeesWithdraw(),
		account,
		proto.TxType_WITHDRAWS,
		currency.GetFinType(),
	); err != nil {
		return &response, e.Context.Error(err)
	}

	if err := e.setSecure(ctx, true); err != nil {
		return &response, e.Context.Error(err)
	}
	response.Success = true

	return &response, nil
}

// CancelWithdraw - Отменяем запрос на вывод, если статус в ожидании.
func (e *ExchangeService) CancelWithdraw(ctx context.Context, req *proto.CancelExchangeRequestWithdraw) (*proto.ResponseWithdraw, error) {

	var (
		response proto.ResponseWithdraw
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	row, err := e.Context.Db.Query(`select id, user_id, symbol, value from transactions where id = $1 and status = $2 and user_id = $3 order by id`, req.GetId(), proto.Status_PENDING, account)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {

		var (
			item proto.Transaction
		)

		if err = row.Scan(&item.Id, &item.UserId, &item.Symbol, &item.Value); err != nil {
			return &response, e.Context.Error(err)
		}

		if _, err := e.Context.Db.Exec("update transactions set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), proto.Status_CANCEL); err != nil {
			return &response, e.Context.Error(err)
		}

		if err := e.setBalance(item.GetSymbol(), item.GetUserId(), item.GetValue(), proto.Balance_PLUS); e.Context.Debug(err) {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// CancelOrder - delete an order according to the specified parameters.
func (e *ExchangeService) CancelOrder(ctx context.Context, req *proto.CancelExchangeRequestOrder) (*proto.ResponseOrder, error) {

	var (
		response proto.ResponseOrder
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	row, err := e.Context.Db.Query(`select id, value, quantity, price, assigning, base_unit, quote_unit, user_id, create_at from orders where status = $1 and id = $2 and user_id = $3 order by id`, proto.Status_PENDING, req.GetId(), account)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {

		var (
			item proto.Order
		)

		if err = row.Scan(&item.Id, &item.Value, &item.Quantity, &item.Price, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt); err != nil {
			return &response, e.Context.Error(err)
		}

		// If the order has been partially used and both values do not have equal shares,
		// then close the order and assign the status completed,
		// with updating the quantity.

		if _, err := e.Context.Db.Exec("update orders set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), proto.Status_CANCEL); err != nil {
			return &response, e.Context.Error(err)
		}

		switch item.Assigning {
		case proto.Assigning_BUY:

			// internal.FromFloat(float64).Mul(float64).Float64() - Фиксируем выходную сумму после калькуляции, так как после точки плавящая цена бывает не точной, а именно ошибочной,
			// например если взять число float64 0.1273806776220894 и умножить её на сумму float64 3140.19368924 то результат будет
			// float64 400.00000000000006, но правильный вариант это 400.
			if err := e.setBalance(item.GetQuoteUnit(), item.GetUserId(), decimal.FromFloat(item.GetValue()).Mul(decimal.FromFloat(item.GetPrice())).Float64(), proto.Balance_PLUS); err != nil {
				return &response, e.Context.Error(err)
			}

			break
		case proto.Assigning_SELL:

			if err := e.setBalance(item.GetBaseUnit(), item.GetUserId(), item.GetValue(), proto.Balance_PLUS); err != nil {
				return &response, e.Context.Error(err)
			}

			break
		}

		if err := e.Context.Publish(&item, "exchange", "order/cancel"); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {
		return &response, status.Error(11538, "the requested order does not exist")
	}
	response.Success = true

	return &response, nil
}
