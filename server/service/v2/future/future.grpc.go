package future

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
	"github.com/cryptogateway/backend-envoys/server/service/v2/account"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
)

func (a *Service) GetFutures(_ context.Context, req *pbfuture.GetRequestFutures) (*pbfuture.ResponseFutures, error) {
	var (
		response pbfuture.ResponseFutures
		// exist    bool
	)
	return &response, nil
}

func (a *Service) SetOrder(ctx context.Context, req *pbfuture.SetRequestOrder) (*pbfuture.ResponseOrder, error) {

	var (
		response pbfuture.ResponseOrder
		order    types.FutureOrder
	)
	// if err := types.Type(req.GetType()); err != nil {
	// 	return &response, err
	// }

	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	if err := a.queryValidatePair(req.GetBaseUnit(), req.GetQuoteUnit(), "future"); err != nil {
		return &response, err
	}

	_account := account.Service{
		Context: a.Context,
	}

	user, err := _account.QueryUser(auth)
	if err != nil {
		return nil, err
	}

	if !user.GetStatus() {
		return &response, status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions")
	}

	order.Quantity = req.GetQuantity()
	order.Position = req.GetPosition()
	order.Trading = req.GetTrading()

	switch req.GetTrading() {
	case types.TradingMarket:

		// order.Price = a.queryMarket(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetType(), req.GetAssigning(), req.GetPrice())

		// if req.GetAssigning() == types.AssigningBuy {
		// 	order.Quantity, order.Value = decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float(), decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float()
		// }
		order.Price = req.GetPrice()

	case types.TradingLimit:

		order.Price = req.GetPrice()
	default:
		return &response, status.Error(82284, "invalid type trade position")
	}

	order.UserId = user.GetId()
	order.BaseUnit = req.GetBaseUnit()
	order.QuoteUnit = req.GetQuoteUnit()
	order.Assigning = req.GetAssigning()
	order.Trading = req.GetTrading()
	order.Leverage = req.GetLeverage()
	order.Status = types.StatusPending
	order.CreateAt = time.Now().UTC().Format(time.RFC3339)

	quantity, err := a.queryValidateOrder(&order)
	if err != nil {
		return &response, err
	}

	if order.Id, err = a.writeOrder(&order); err != nil {
		return &response, err
	}

	switch order.GetAssigning() {
	case types.AssigningOpen:

		if err := a.writeAsset(order.GetBaseUnit(), order.GetAssigning(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		if err := a.WriteBalance(order.GetQuoteUnit(), order.GetAssigning(), order.GetUserId(), quantity, types.BalanceMinus); err != nil {
			return &response, err
		}

		a.trade(&order, types.AssigningSell)

		break
	case types.AssigningClose:

		if err := a.writeAsset(order.GetQuoteUnit(), order.GetAssigning(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		if err := a.WriteBalance(order.GetBaseUnit(), order.GetAssigning(), order.GetUserId(), quantity, types.BalanceMinus); err != nil {
			return &response, err
		}

		a.trade(&order, types.AssigningBuy)

		break
	default:
		return &response, status.Error(11588, "invalid assigning trade position")
	}

	response.Fields = append(response.Fields, &order)
	return &response, nil

}
func (a *Service) SetTicker(_ context.Context, req *pbfuture.SetRequestTicker) (*pbfuture.ResponseTicker, error) {

	var (
		response pbfuture.ResponseTicker
	)

	if req.GetKey() != a.Context.Secrets[2] {
		return &response, status.Error(654333, "the access key is incorrect")
	}

	if _, err := a.Context.Db.Exec(`insert into ohlcv (assigning, base_unit, quote_unit, price, quantity) values ($1, $2, $3, $4, $5)`, req.GetAssigning(), req.GetBaseUnit(), req.GetQuoteUnit(), req.GetPrice(), req.GetValue()); a.Context.Debug(err) {
		return &response, err
	}

	for _, interval := range help.Depth() {

		migrate, err := a.GetTicker(context.Background(), &pbfuture.GetRequestTicker{BaseUnit: req.GetBaseUnit(), QuoteUnit: req.GetQuoteUnit(), Limit: 2, Resolution: interval})
		if err != nil {
			return &response, err
		}

		if err := a.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/ticker:%v", interval)); err != nil {
			return &response, err
		}

		response.Fields = append(response.Fields, migrate.Fields...)
	}

	return &response, nil
}
func (a *Service) GetTicker(_ context.Context, req *pbfuture.GetRequestTicker) (*pbfuture.ResponseTicker, error) {

	var (
		response pbfuture.ResponseTicker
		limit    string
		maps     []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 500
	}

	if req.GetLimit() > 0 {
		limit = fmt.Sprintf("limit %d", req.GetLimit())
	}

	if req.GetTo() > 0 {
		maps = append(maps, fmt.Sprintf(`and to_char(o.create_at::timestamp, 'yyyy-mm-dd hh24:mi:ss') < to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss')`, req.GetTo()))
	}

	rows, err := a.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', o.create_at))::integer buckettime, first(o.price, o.create_at) as open, last(o.price, o.create_at) as close, first(o.price, o.price) as low, last(o.price, o.price) as high, sum(o.quantity) as volume, avg(o.price) as avg_price, o.base_unit, o.quote_unit from ohlcv as o where o.base_unit = '%[1]s' and o.quote_unit = '%[2]s' %[3]s group by buckettime, o.base_unit, o.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	for rows.Next() {

		var (
			item types.Ticker
		)

		if err = rows.Scan(&item.Time, &item.Open, &item.Close, &item.Low, &item.High, &item.Volume, &item.Price, &item.BaseUnit, &item.QuoteUnit); err != nil {
			return &response, err
		}

		response.Fields = append(response.Fields, &item)
	}

	var (
		stats types.Stats
	)

	_ = a.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from ohlcv as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

	if len(response.Fields) > 1 {
		stats.Previous = response.Fields[1].Close
	}

	response.Stats = &stats

	return &response, nil
}
