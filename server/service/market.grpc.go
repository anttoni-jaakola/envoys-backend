package service

import (
	"errors"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"golang.org/x/net/context"
	"strings"
	"time"
)

// GetInstruments - get instruments.
func (m *MarketService) GetInstruments(_ context.Context, _ *proto.GetMarketRequestInstrument) (*proto.ResponseInstrument, error) {

	var (
		response   proto.ResponseInstrument
		instrument proto.Instrument
	)

	request, err := m.request("instruments", nil)
	if err != nil {
		return &response, err
	}

	if fields, ok := request.([]interface{}); ok {

		for i := 0; i < len(fields[0].([]interface{})); i++ {

			var (
				disable bool
			)

			columns := fields[0].([]interface{})[i].([]interface{})
			for i := 0; i < len(m.Context.MarketPairs); i++ {
				if strings.HasPrefix(m.Context.MarketPairs[i], columns[0].(string)) || strings.HasSuffix(m.Context.MarketPairs[i], columns[0].(string)) {
					disable = true
				}
			}

			if disable {
				instrument.Currencies = append(instrument.Currencies, &proto.Instrument_Currency{
					Name:  columns[0].(string),
					Id:    int64(columns[1].(float64)),
					Size:  columns[2].(float64),
					Price: columns[3].(float64) / 100000000,
				})
			}
		}

		for i := 0; i < len(fields[1].([]interface{})); i++ {

			var (
				disable bool
			)

			columns := fields[1].([]interface{})[i].([]interface{})
			for i := 0; i < len(m.Context.MarketPairs); i++ {
				if strings.HasPrefix(m.Context.MarketPairs[i], columns[0].(string)) || strings.HasSuffix(m.Context.MarketPairs[i], columns[0].(string)) {
					disable = true
				}
			}

			if disable {
				instrument.Pairs = append(instrument.Pairs, &proto.Instrument_Pair{
					Symbol:    columns[0].(string),
					Id:        int64(columns[1].(float64)),
					BaseName:  columns[2].(string),
					QuoteName: columns[3].(string),
				})
			}
		}

		response.Fields = &instrument
	}

	return &response, nil
}

// GetBook - get book list.
func (m *MarketService) GetBook(_ context.Context, req *proto.GetMarketRequestBook) (*proto.ResponseBook, error) {

	var (
		response proto.ResponseBook
	)

	m.connect.Send(map[string]interface{}{
		"event":  "unbind",
		"feed":   "F",
		"feedId": req.GetInstrument(),
	})

	time.Sleep(time.Second)

	m.connect.Send(map[string]interface{}{
		"event":  "bind",
		"feed":   "F",
		"feedId": req.GetInstrument(),
	})

	return &response, nil
}

// SetOrder - set new order.
func (m *MarketService) SetOrder(ctx context.Context, req *proto.SetMarketRequestOrder) (*proto.ResponseMarketOrder, error) {

	var (
		response proto.ResponseMarketOrder
	)

	account, err := m.Context.Auth(ctx)
	if err != nil {
		return &response, m.Context.Error(err)
	}

	params := map[string]interface{}{
		"instrument":    req.GetSymbol(),
		"clientOrderId": account,
		"price":         req.GetPrice() * 1e8,
		"size":          req.GetSize() * 1e8,
		"side":          req.GetSide(),
		"type":          req.GetType(),
		"cod":           false,
	}

	request, err := m.request("add", params)
	if err != nil {
		return &response, m.Context.Error(err)
	}

	if fields, ok := request.(map[string]interface{}); ok {

		if number, ok := fields["error"]; ok {
			var (
				err error
			)

			switch number.(float64) {
			case 70:
				err = errors.New("invalid order size")
			case 71:
				err = errors.New("invalid order price")
			case 72:
				err = errors.New("invalid order flags")
			case 73:
				err = errors.New("order type not allowed")
			case 74:
				err = errors.New("client order id already in use")
			case 75:
				err = errors.New("add failed - Post-Only")
			case 76:
				err = errors.New("add failed - IOC: no orders to match")
			case 77:
				err = errors.New("add failed - FOK: not enough liquidity")
			case 78:
				err = errors.New("qdd failed - SMP (self-trade prevention)")
			case 79:
				err = errors.New("add failed - limits")
			}

			return &response, m.Context.Error(err)
		}

		if req.GetPrice() == 0 {
			maps := fields["deals"].([]interface{})[len(fields["deals"].([]interface{}))-1]

			req.Price = maps.(map[string]interface{})["price"].(float64) / 100000000
			req.Volume = maps.(map[string]interface{})["volume"].(float64) / 100000000
		}

		if _, err = m.Context.Db.Exec("insert into market_orders (id, uid, symbol, price, volume, size, side, type, cid) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)", decimal.FromFloat(fields["id"].(float64)).Int64(), account, req.GetSymbol(), req.GetPrice(), req.GetVolume(), req.GetSize(), req.GetSide(), req.GetType(), 87); err != nil {
			return &response, m.Context.Error(err)
		}

		migrate, err := m.GetOrders(ctx, &proto.GetMarketRequestOrder{Symbol: req.GetSymbol(), Limit: 1})
		if m.Context.Debug(err) {
			return &response, m.Context.Error(err)
		}
		response.Fields = append(response.Fields, migrate.Fields[0])

		return &response, nil
	}

	return &response, errors.New("missing order")
}

// GetOrders - get private order book.
func (m *MarketService) GetOrders(ctx context.Context, req *proto.GetMarketRequestOrder) (*proto.ResponseMarketOrder, error) {

	var (
		response proto.ResponseMarketOrder
	)

	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	account, err := m.Context.Auth(ctx)
	if err != nil {
		return &response, m.Context.Error(err)
	}

	_ = m.Context.Db.QueryRow("select count(*) from market_orders where uid = $1 and symbol = $2", account, req.GetSymbol()).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := m.Context.Db.Query("select id, uid, symbol, price, volume, size, side, type, cid, create_at from market_orders where uid = $1 and symbol = $2 order by id desc limit $3 offset $4", account, req.GetSymbol(), req.GetLimit(), offset)
		if err != nil {
			return &response, m.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				order proto.MarketOrder
			)

			if err := rows.Scan(&order.Id, &order.Uid, &order.Symbol, &order.Price, &order.Volume, &order.Size, &order.Side, &order.Type, &order.Cid, &order.CreateAt); err != nil {
				return &response, m.Context.Error(err)
			}

			response.Fields = append(response.Fields, &order)
		}

	}

	return &response, nil
}
