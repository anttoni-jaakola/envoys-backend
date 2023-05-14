package future

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
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
func (a *Service) GetOrders(ctx context.Context, req *pbfuture.GetRequestOrders) (*pbfuture.ResponseOrder, error) {
	var (
		response pbfuture.ResponseOrder
		maps     []string
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	switch req.GetAssigning() {
	case types.AssigningOpen:
		maps = append(maps, fmt.Sprintf("where assigning = '%v'", types.AssigningOpen))
	case types.AssigningClose:
		maps = append(maps, fmt.Sprintf("where assigning = '%v'", types.AssigningClose))
	default:
		maps = append(maps, fmt.Sprintf("where assigning = '%v' or assigning = '%v'", types.AssigningClose, types.AssigningOpen))
	}

	if len(req.GetPosition()) > 0 {
		if err := types.Position(req.GetPosition()); err != nil {
			return &response, err
		}
		maps = append(maps, fmt.Sprintf("and position = '%v'", req.GetPosition()))
	}
	// check order type

	if req.GetOwner() {

		auth, err := a.Context.Auth(ctx)
		if err != nil {
			return &response, err
		}

		maps = append(maps, fmt.Sprintf("and user_id = '%v'", auth))

	} else if req.GetUserId() > 0 {

		maps = append(maps, fmt.Sprintf("and user_id = '%v'", req.GetUserId()))
	}
	switch req.GetStatus() {
	case types.StatusFilled:
		maps = append(maps, fmt.Sprintf("and status = '%v'", types.StatusFilled))
	case types.StatusPending:
		maps = append(maps, fmt.Sprintf("and status = '%v'", types.StatusPending))
	case types.StatusCancel:
		maps = append(maps, fmt.Sprintf("and status = '%v'", types.StatusCancel))
	}
	if len(req.GetBaseUnit()) > 0 && len(req.GetQuoteUnit()) > 0 {
		maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))
	}
	_ = a.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count, sum(value) as volume from futures %s", strings.Join(maps, " "))).Scan(&response.Count, &response.Volume)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := a.Context.Db.Query(fmt.Sprintf("select id, assigning, price, value, quantity, base_unit, quote_unit, user_id, create_at, Position, status from futures %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item types.Future
			)

			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Value, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt, &item.Position, &item.Status); err != nil {
				return &response, err
			}

			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

func (a *Service) SetOrder(ctx context.Context, req *pbfuture.SetRequestOrder) (*pbfuture.ResponseOrder, error) {

	var (
		response pbfuture.ResponseOrder
		order    types.Future
	)

	// auth, err := a.Context.Auth(ctx)
	// if err != nil {
	// 	return &response, err
	// }
	var auth int64 = 6

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
	order.OrderType = req.GetOrderType()

	switch req.GetOrderType() {
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
	order.OrderType = req.GetOrderType()
	order.Leverage = req.GetLeverage()
	order.Status = types.StatusPending
	order.CreateAt = time.Now().UTC().Format(time.RFC3339)

	margin, err := a.queryValidateOrder(&order)
	if err != nil {
		return &response, err
	}
	order.Value = decimal.New(req.GetQuantity()).Mul(req.GetPrice()).Float()

	if order.Id, err = a.writeOrder(&order); err != nil {
		return &response, err
	}
	fmt.Println("order saved id ", order.Id)

	switch order.GetAssigning() {
	case types.AssigningOpen:

		if err := a.WriteBalance(order.GetQuoteUnit(), order.GetAssigning(), order.GetUserId(), margin, types.BalanceMinus); err != nil {
			return &response, err
		}

		break
	case types.AssigningClose:

		if err := a.WriteBalance(order.GetBaseUnit(), order.GetAssigning(), order.GetUserId(), margin, types.BalancePlus); err != nil {
			return &response, err
		}

		a.closePosition(&order)

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
