package spot

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/query"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

// GetSymbol - check if exchange unit exists or does not exist.
func (e *Service) GetSymbol(_ context.Context, req *pbspot.GetRequestSymbol) (*pbspot.ResponseSymbol, error) {

	var (
		response pbspot.ResponseSymbol
		exist    bool
	)

	if row, err := e.getCurrency(req.GetBaseUnit(), false); err != nil {
		return &response, e.Context.Error(status.Errorf(11584, "this base currency does not exist, %v", row.GetSymbol()))
	}

	if row, err := e.getCurrency(req.GetQuoteUnit(), false); err != nil {
		return &response, e.Context.Error(status.Errorf(11582, "this quote currency does not exist, %v", row.GetSymbol()))
	}

	if err := e.Context.Db.QueryRow("select exists(select id from spot_pairs where base_unit = $1 and quote_unit = $2)::bool", req.GetBaseUnit(), req.GetQuoteUnit()).Scan(&exist); err != nil || !exist {
		return &response, e.Context.Error(status.Errorf(11585, "this pair %v-%v does not exist", req.GetBaseUnit(), req.GetQuoteUnit()))
	}

	response.Success = true

	return &response, nil
}

// GetAnalysis - analysis list.
func (e *Service) GetAnalysis(ctx context.Context, req *pbspot.GetRequestAnalysis) (*pbspot.ResponseAnalysis, error) {

	var (
		response pbspot.ResponseAnalysis
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	_ = e.Context.Db.QueryRow(`select count(*) as count from spot_pairs where status = $1`, true).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(`select id, base_unit, quote_unit, price from spot_pairs where status = $3 order by id desc limit $1 offset $2`, req.GetLimit(), offset, true)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				analysis pbspot.Analysis
			)

			if err := rows.Scan(&analysis.Id, &analysis.BaseUnit, &analysis.QuoteUnit, &analysis.Price); err != nil {
				return &response, e.Context.Error(err)
			}

			migrate, err := e.GetGraph(context.Background(), &pbspot.GetRequestGraph{BaseUnit: analysis.GetBaseUnit(), QuoteUnit: analysis.GetQuoteUnit(), Limit: 40})
			if err != nil {
				return &response, e.Context.Error(err)
			}

			for i := 0; i < len(migrate.Fields); i++ {
				analysis.Chart = append(analysis.Chart, migrate.Fields[i].GetPrice())
			}

			for i := 0; i < 2; i++ {

				var (
					assigning pbspot.Assigning
				)

				switch i {
				case 0:
					assigning = pbspot.Assigning_BUY
				case 1:
					assigning = pbspot.Assigning_SELL
				}

				migrate, err := e.GetOrders(context.Background(), &pbspot.GetRequestOrdersManual{
					BaseUnit:  analysis.GetBaseUnit(),
					QuoteUnit: analysis.GetQuoteUnit(),
					Assigning: assigning,
					Status:    pbspot.Status_FILLED,
					UserId:    account,
					Limit:     2,
				})
				if err != nil {
					return &response, e.Context.Error(err)
				}

				if len(migrate.GetFields()) == 2 {
					switch i {
					case 0:
						analysis.BuyRatio = decimal.New(100 * (analysis.GetPrice() - migrate.Fields[0].GetPrice()) / migrate.Fields[0].GetPrice()).Round(2).Float()
					case 1:
						analysis.SelRatio = decimal.New(100 * (analysis.GetPrice() - migrate.Fields[0].GetPrice()) / migrate.Fields[0].GetPrice()).Round(2).Float()
					}
				}
			}

			response.Fields = append(response.Fields, &analysis)
		}
	}

	return &response, nil
}

// GetMarkers - show marker zone.
func (e *Service) GetMarkers(_ context.Context, _ *pbspot.GetRequestMarkers) (*pbspot.ResponseMarker, error) {

	var (
		response pbspot.ResponseMarker
	)

	rows, err := e.Context.Db.Query("select symbol from spot_currencies where marker = $1", true)
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
func (e *Service) GetPairs(_ context.Context, req *pbspot.GetRequestPairs) (*pbspot.ResponsePair, error) {

	var (
		response pbspot.ResponsePair
	)

	rows, err := e.Context.Db.Query("select id, base_unit, quote_unit, base_decimal, quote_decimal, status from spot_pairs where base_unit = $1 or quote_unit = $1", req.GetSymbol())
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			pair pbspot.Pair
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
func (e *Service) GetPair(_ context.Context, req *pbspot.GetRequestPair) (*pbspot.ResponsePair, error) {

	var (
		response pbspot.ResponsePair
	)

	row, err := e.Context.Db.Query(`select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from spot_pairs where base_unit = $1 and quote_unit = $2`, req.GetBaseUnit(), req.GetQuoteUnit())
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {

		var (
			pair pbspot.Pair
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
func (e *Service) SetOrder(ctx context.Context, req *pbspot.SetRequestOrder) (*pbspot.ResponseOrder, error) {

	var (
		response pbspot.ResponseOrder
		migrate  = query.Migrate{
			Context: e.Context,
		}
		order pbspot.Order
		err   error
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	user, err := migrate.User(account)
	if err != nil {
		return nil, err
	}
	if !user.Status {
		return &response, e.Context.Error(status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions"))
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

	order.Quantity = req.GetQuantity()
	order.Value = req.GetQuantity()

	switch req.GetTradeType() {
	case pbspot.TradeType_MARKET:
		order.Price = e.getMarket(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetAssigning(), req.GetPrice())

		if req.GetAssigning() == pbspot.Assigning_BUY {
			order.Quantity, order.Value = decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float(), decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float()
		}

	case pbspot.TradeType_LIMIT:
		order.Price = req.GetPrice()
	default:
		return &response, e.Context.Error(status.Error(82284, "invalid type trade position"))
	}

	order.UserId = user.GetId()
	order.BaseUnit = req.GetBaseUnit()
	order.QuoteUnit = req.GetQuoteUnit()
	order.Assigning = req.GetAssigning()
	order.Status = pbspot.Status_PENDING
	order.CreateAt = time.Now().UTC().Format(time.RFC3339)

	quantity, err := e.helperOrder(&order)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if order.Id, err = e.setOrder(&order); err != nil {
		return &response, e.Context.Error(err)
	}

	switch order.GetAssigning() {
	case pbspot.Assigning_BUY:

		if err := e.setAsset(order.GetBaseUnit(), order.GetUserId(), false); err != nil {
			return &response, e.Context.Error(err)
		}

		if err := e.setBalance(order.GetQuoteUnit(), order.GetUserId(), quantity, pbspot.Balance_MINUS); err != nil {
			return &response, e.Context.Error(err)
		}

		e.replayTradeInit(&order, pbspot.Side_BID)

		break
	case pbspot.Assigning_SELL:

		if err := e.setAsset(order.GetQuoteUnit(), order.GetUserId(), false); err != nil {
			return &response, e.Context.Error(err)
		}

		if err := e.setBalance(order.GetBaseUnit(), order.GetUserId(), quantity, pbspot.Balance_MINUS); err != nil {
			return &response, e.Context.Error(err)
		}

		e.replayTradeInit(&order, pbspot.Side_ASK)

		break
	default:
		return &response, status.Error(11588, "invalid assigning trade position")
	}

	response.Fields = append(response.Fields, &order)

	return &response, nil
}

// GetOrders - show all active orders.
func (e *Service) GetOrders(ctx context.Context, req *pbspot.GetRequestOrdersManual) (*pbspot.ResponseOrder, error) {

	var (
		response pbspot.ResponseOrder
		maps     []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case pbspot.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_BUY))
	case pbspot.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_SELL))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %d or assigning = %d)", pbspot.Assigning_BUY, pbspot.Assigning_SELL))
	}

	if req.GetOwner() {
		account, err := e.Context.Auth(ctx)
		if err != nil {
			return &response, e.Context.Error(err)
		}

		maps = append(maps, fmt.Sprintf("and user_id = '%v'", account))
	} else if req.GetUserId() > 0 {
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", req.GetUserId()))
	}

	switch req.GetStatus() {
	case pbspot.Status_FILLED:
		maps = append(maps, fmt.Sprintf("and status = %d", pbspot.Status_FILLED))
	case pbspot.Status_PENDING:
		maps = append(maps, fmt.Sprintf("and status = %d", pbspot.Status_PENDING))
	case pbspot.Status_CANCEL:
		maps = append(maps, fmt.Sprintf("and status = %d", pbspot.Status_CANCEL))
	}

	// Get request to display pair information.
	if len(req.GetBaseUnit()) > 0 && len(req.GetQuoteUnit()) > 0 {
		maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))
	}

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count, sum(value) as volume from spot_orders %s", strings.Join(maps, " "))).Scan(&response.Count, &response.Volume)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, assigning, price, value, quantity, base_unit, quote_unit, user_id, create_at, status from spot_orders %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Order
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
func (e *Service) GetAssets(ctx context.Context, _ *pbspot.GetRequestAssetsManual) (*pbspot.ResponseAsset, error) {

	var (
		response pbspot.ResponseAsset
	)

	rows, err := e.Context.Db.Query("select id, name, symbol, status from spot_currencies")
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			asset pbspot.Currency
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
func (e *Service) SetAsset(ctx context.Context, req *pbspot.SetRequestAsset) (*pbspot.ResponseAsset, error) {

	var (
		response pbspot.ResponseAsset
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	switch req.GetFinType() {
	case pbspot.FinType_CRYPTO:

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

		row, err := e.Context.Db.Query(`select id from spot_assets where symbol = $1 and user_id = $2`, req.GetSymbol(), account)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer row.Close()

		if row.Next() {

			if address := e.getAddress(account, req.GetSymbol(), req.GetPlatform(), req.GetProtocol()); len(address) == 0 {
				if _, err = e.Context.Db.Exec("insert into spot_wallets (address, symbol, platform, protocol, user_id) values ($1, $2, $3, $4, $5)", response.GetAddress(), req.GetSymbol(), req.GetPlatform(), req.GetProtocol(), account); err != nil {
					return &response, e.Context.Error(err)
				}

				return &response, nil
			}

			return &response, e.Context.Error(status.Error(700990, "the asset address has already been generated"))
		}

		if _, err = e.Context.Db.Exec("insert into spot_assets (user_id, symbol) values ($1, $2);", account, req.GetSymbol()); err != nil {
			return &response, e.Context.Error(err)
		}

		if _, err = e.Context.Db.Exec("insert into spot_wallets (address, symbol, platform, protocol, user_id) values ($1, $2, $3, $4, $5)", response.GetAddress(), req.GetSymbol(), req.GetPlatform(), req.GetProtocol(), account); err != nil {
			return &response, e.Context.Error(err)
		}

	case pbspot.FinType_FIAT:

		if err := e.setAsset(req.GetSymbol(), account, true); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// GetAsset - show asset information.
func (e *Service) GetAsset(ctx context.Context, req *pbspot.GetRequestAsset) (*pbspot.ResponseAsset, error) {

	var (
		response pbspot.ResponseAsset
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

	_ = e.Context.Db.QueryRow(`select coalesce(sum(case when base_unit = $1 then value when quote_unit = $1 then value * price end), 0.00) as volume from spot_orders where base_unit = $1 and assigning = $2 and status = $4 and user_id = $5 or quote_unit = $1 and assigning = $3 and status = $4 and user_id = $5`, req.GetSymbol(), pbspot.Assigning_SELL, pbspot.Assigning_BUY, pbspot.Status_PENDING, account).Scan(&row.Volume)

	for i := 0; i < len(row.GetChainsIds()); i++ {

		if chain, _ := e.getChain(row.GetChainsIds()[i], false); chain.GetId() > 0 {
			chain.Rpc, chain.RpcKey, chain.RpcUser, chain.RpcPassword, chain.Address = "", "", "", "", ""

			if contract, _ := e.getContract(row.GetSymbol(), row.GetChainsIds()[i]); contract.GetId() > 0 {
				chain.FeesWithdraw, contract.FeesWithdraw = contract.GetFeesWithdraw(), 0

				price, err := e.GetPrice(context.Background(), &pbspot.GetRequestPriceManual{
					BaseUnit:  chain.GetParentSymbol(),
					QuoteUnit: req.GetSymbol(),
				})
				if err != nil {
					return &response, e.Context.Error(err)
				}

				chain.FeesWithdraw = decimal.New(chain.GetFeesWithdraw()).Mul(price.GetPrice()).Float()
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
func (e *Service) GetGraph(_ context.Context, req *pbspot.GetRequestGraph) (*pbspot.ResponseGraph, error) {

	var (
		response pbspot.ResponseGraph
		limit    string
		maps     []string
	)

	if req.GetLimit() > 0 {
		limit = fmt.Sprintf("limit %d", req.GetLimit())
	}

	if req.GetFrom() > 0 && req.GetTo() > 0 {
		maps = append(maps, fmt.Sprintf(`and to_char(ohlc.create_at::timestamp, 'yyyy-mm-dd hh24:mi:ss') between to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss') and to_char(to_timestamp(%[2]d), 'yyyy-mm-dd hh24:mi:ss')`, req.GetFrom(), req.GetTo()))
	}

	rows, err := e.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', ohlc.create_at))::integer buckettime, first(ohlc.price, ohlc.create_at) as open, last(ohlc.price, ohlc.create_at) as close, first(ohlc.price, ohlc.price) as low, last(ohlc.price, ohlc.price) as high, sum(ohlc.quantity) as volume, avg(ohlc.price) as avg_price, ohlc.base_unit, ohlc.quote_unit from spot_trades as ohlc where ohlc.base_unit = '%[1]s' and ohlc.quote_unit = '%[2]s' %[3]s group by buckettime, ohlc.base_unit, ohlc.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item pbspot.Graph
		)

		if err = rows.Scan(&item.Time, &item.Open, &item.Close, &item.Low, &item.High, &item.Volume, &item.Price, &item.BaseUnit, &item.QuoteUnit); err != nil {
			return &response, e.Context.Error(err)
		}

		response.Fields = append(response.Fields, &item)
	}

	var (
		stats pbspot.Stats
	)

	_ = e.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from spot_trades as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

	if len(response.Fields) > 1 {
		stats.Previous = response.Fields[1].Close
	}
	response.Stats = &stats

	return &response, nil
}

// GetTransfers - order stats list.
func (e *Service) GetTransfers(ctx context.Context, req *pbspot.GetRequestTransfers) (*pbspot.ResponseTransfer, error) {

	var (
		response pbspot.ResponseTransfer
		maps     []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case pbspot.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_BUY))
	case pbspot.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_SELL))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %d or assigning = %d)", pbspot.Assigning_BUY, pbspot.Assigning_SELL))
	}

	maps = append(maps, fmt.Sprintf("and order_id = '%v'", req.GetOrderId()))

	if req.GetOwner() {
		account, err := e.Context.Auth(ctx)
		if err != nil {
			return &response, e.Context.Error(err)
		}

		maps = append(maps, fmt.Sprintf("and user_id = '%v'", account))
	}

	rows, err := e.Context.Db.Query(fmt.Sprintf("select id, user_id, base_unit, quote_unit, price, quantity, assigning, fees, create_at from spot_transfers %s order by id desc limit %d", strings.Join(maps, " "), req.GetLimit()))
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item pbspot.Transfer
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
func (e *Service) GetTrades(_ context.Context, req *pbspot.GetRequestTrades) (*pbspot.ResponseTrades, error) {

	var (
		response pbspot.ResponseTrades
		maps     []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case pbspot.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_BUY))
	case pbspot.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %d", pbspot.Assigning_SELL))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %d or assigning = %d)", pbspot.Assigning_BUY, pbspot.Assigning_SELL))
	}

	// Get request to display pair information.
	maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from spot_trades %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, assigning, price, quantity, base_unit, quote_unit, create_at from spot_trades %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbspot.Trade
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
func (e *Service) GetPrice(_ context.Context, req *pbspot.GetRequestPriceManual) (*pbspot.ResponsePrice, error) {

	var (
		response pbspot.ResponsePrice
		ok       bool
	)

	if response.Price, ok = e.getPrice(req.GetBaseUnit(), req.GetQuoteUnit()); ok {
		return &response, nil
	}

	if response.Price, ok = e.getPrice(req.GetQuoteUnit(), req.GetBaseUnit()); ok {
		response.Price = decimal.New(decimal.New(1).Div(response.Price).Float()).Round(8).Float()
	}

	return &response, nil
}

// GetTransactions - вывод входящих и исходящих транзакций.
func (e *Service) GetTransactions(ctx context.Context, req *pbspot.GetRequestTransactionsManual) (*pbspot.ResponseTransaction, error) {

	var (
		response pbspot.ResponseTransaction
		maps     []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetTxType() {
	case pbspot.TxType_DEPOSIT:
		maps = append(maps, fmt.Sprintf("where tx_type = %d", pbspot.TxType_DEPOSIT))
	case pbspot.TxType_WITHDRAWS:
		maps = append(maps, fmt.Sprintf("where tx_type = %d", pbspot.TxType_WITHDRAWS))
	default:
		maps = append(maps, fmt.Sprintf("where (tx_type = %d or tx_type = %d)", pbspot.TxType_WITHDRAWS, pbspot.TxType_DEPOSIT))
	}

	if len(req.GetSymbol()) > 0 {
		maps = append(maps, fmt.Sprintf("and symbol = '%v'", req.GetSymbol()))
	}

	maps = append(maps, fmt.Sprintf("and status != %d", pbspot.Status_RESERVE))

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	maps = append(maps, fmt.Sprintf("and user_id = %v", account))

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
				item.Fees = decimal.New(item.GetFees()).Mul(item.GetPrice()).Float()
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
func (e *Service) SetWithdraw(ctx context.Context, req *pbspot.SetRequestWithdraw) (*pbspot.ResponseWithdraw, error) {

	var (
		response pbspot.ResponseWithdraw
		migrate  = query.Migrate{
			Context: e.Context,
		}
		fees float64
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
	user, err := migrate.User(account)
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
	if contract.GetProtocol() != pbspot.Protocol_MAINNET {
		price, err := e.GetPrice(context.Background(), &pbspot.GetRequestPriceManual{
			BaseUnit:  chain.GetParentSymbol(),
			QuoteUnit: req.GetSymbol(),
		})
		if err != nil {
			return &response, e.Context.Error(err)
		}

		req.Price = price.GetPrice()
		chain.FeesWithdraw = contract.GetFeesWithdraw()

		fees = decimal.New(contract.GetFeesWithdraw()).Mul(req.GetPrice()).Float()
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

	if err := e.setBalance(req.GetSymbol(), account, req.GetQuantity(), pbspot.Balance_MINUS); e.Context.Debug(err) {
		return &response, e.Context.Error(err)
	}

	if _, err := e.Context.Db.Exec(`insert into spot_transactions (symbol, value, price, "to", chain_id, platform, protocol, fees, user_id, tx_type, fin_type) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		req.GetSymbol(),
		req.GetQuantity(),
		req.GetPrice(),
		req.GetAddress(),
		req.GetId(),
		req.GetPlatform(),
		contract.GetProtocol(),
		chain.GetFeesWithdraw(),
		account,
		pbspot.TxType_WITHDRAWS,
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
func (e *Service) CancelWithdraw(ctx context.Context, req *pbspot.CancelRequestWithdraw) (*pbspot.ResponseWithdraw, error) {

	var (
		response pbspot.ResponseWithdraw
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	row, err := e.Context.Db.Query(`select id, user_id, symbol, value from spot_transactions where id = $1 and status = $2 and user_id = $3 order by id`, req.GetId(), pbspot.Status_PENDING, account)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {

		var (
			item pbspot.Transaction
		)

		if err = row.Scan(&item.Id, &item.UserId, &item.Symbol, &item.Value); err != nil {
			return &response, e.Context.Error(err)
		}

		if _, err := e.Context.Db.Exec("update spot_transactions set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), pbspot.Status_CANCEL); err != nil {
			return &response, e.Context.Error(err)
		}

		if err := e.setBalance(item.GetSymbol(), item.GetUserId(), item.GetValue(), pbspot.Balance_PLUS); e.Context.Debug(err) {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// CancelOrder - delete an order according to the specified parameters.
func (e *Service) CancelOrder(ctx context.Context, req *pbspot.CancelRequestOrder) (*pbspot.ResponseOrder, error) {

	var (
		response pbspot.ResponseOrder
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	row, err := e.Context.Db.Query(`select id, value, quantity, price, assigning, base_unit, quote_unit, user_id, create_at from spot_orders where status = $1 and id = $2 and user_id = $3 order by id`, pbspot.Status_PENDING, req.GetId(), account)
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {

		var (
			item pbspot.Order
		)

		if err = row.Scan(&item.Id, &item.Value, &item.Quantity, &item.Price, &item.Assigning, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt); err != nil {
			return &response, e.Context.Error(err)
		}

		// If the order has been partially used and both values do not have equal shares,
		// then close the order and assign the status completed,
		// with updating the quantity.

		if _, err := e.Context.Db.Exec("update spot_orders set status = $3 where id = $1 and user_id = $2;", item.GetId(), item.GetUserId(), pbspot.Status_CANCEL); err != nil {
			return &response, e.Context.Error(err)
		}

		switch item.Assigning {
		case pbspot.Assigning_BUY:

			// internal.FromFloat(float64).Mul(float64).Float64() - Фиксируем выходную сумму после калькуляции, так как после точки плавящая цена бывает не точной, а именно ошибочной,
			// например если взять число float64 0.1273806776220894 и умножить её на сумму float64 3140.19368924 то результат будет
			// float64 400.00000000000006, но правильный вариант это 400.
			if err := e.setBalance(item.GetQuoteUnit(), item.GetUserId(), decimal.New(item.GetValue()).Mul(item.GetPrice()).Float(), pbspot.Balance_PLUS); err != nil {
				return &response, e.Context.Error(err)
			}

			break
		case pbspot.Assigning_SELL:

			if err := e.setBalance(item.GetBaseUnit(), item.GetUserId(), item.GetValue(), pbspot.Balance_PLUS); err != nil {
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
