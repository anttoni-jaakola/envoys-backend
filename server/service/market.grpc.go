package service

import (
	"errors"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"golang.org/x/net/context"
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

			columns := fields[0].([]interface{})[i].([]interface{})
			instrument.Currencies = append(instrument.Currencies, &proto.Instrument_Currency{
				Name:  columns[0].(string),
				Id:    int64(columns[1].(float64)),
				Size:  columns[2].(float64),
				Price: columns[3].(float64) / 100000000,
			})

		}

		for i := 0; i < len(fields[1].([]interface{})); i++ {

			columns := fields[1].([]interface{})[i].([]interface{})
			instrument.Pairs = append(instrument.Pairs, &proto.Instrument_Pair{
				Symbol:    columns[0].(string),
				Id:        int64(columns[1].(float64)),
				BaseName:  columns[2].(string),
				QuoteName: columns[3].(string),
			})

		}

		response.Fields = &instrument
	}

	return &response, nil
}

// GetPositions - get positions.
func (m *MarketService) GetPositions(_ context.Context, _ *proto.GetMarketRequestPosition) (*proto.ResponsePosition, error) {

	var (
		response proto.ResponsePosition
		position proto.Position
	)

	request, err := m.request("positions", nil)
	if err != nil {
		return &response, err
	}

	if fields, ok := request.([]interface{}); ok {
		position.Id = int64(request.([]interface{})[0].(float64))

		for i := 0; i < len(fields[1].([]interface{})); i++ {

			columns := fields[1].([]interface{})[i].([]interface{})
			position.Positions = append(position.Positions, &proto.Position_Position{
				Name:  columns[0].(string),
				Cid:   int64(columns[2].(float64)),
				Value: columns[1].(float64) / 100000000,
			})

		}

		for i := 0; i < len(fields[2].([]interface{})); i++ {

			columns := fields[2].([]interface{})[i].([]interface{})
			position.Orders = append(position.Orders, &proto.Position_Order{
				Id:        int64(columns[4].(float64)),
				Symbol:    columns[0].(string),
				Type:      int32(columns[1].(float64)),
				Side:      int32(columns[2].(float64)),
				Status:    int32(columns[3].(float64)),
				Cid:       int64(columns[5].(float64)),
				Price:     columns[6].(float64) / 100000000,
				Volume:    columns[7].(float64),
				Size:      columns[8].(float64),
				CreatedBy: int32(columns[10].(float64)),
				Timestamp: int64(columns[9].(float64)),
			})

		}

		for i := 0; i < len(fields[3].([]interface{})); i++ {

			columns := fields[3].([]interface{})[i].([]interface{})
			position.Settlements = append(position.Settlements, &proto.Position_Settlement{
				Id:        int64(columns[0].(float64)),
				BaseName:  columns[1].(string),
				QuoteName: columns[2].(string),
				BaseSize:  columns[3].(float64),
				QuoteSize: columns[4].(float64),
				Cid:       int64(columns[6].(float64)),
				Timestamp: int64(columns[5].(float64)),
			})

		}

		response.Fields = &position
	}

	return &response, nil
}

// GetLimits - get limits.
func (m *MarketService) GetLimits(_ context.Context, _ *proto.GetMarketRequestLimit) (*proto.ResponseLimit, error) {

	var (
		response proto.ResponseLimit
	)

	request, err := m.request("limits", nil)
	if err != nil {
		return &response, err
	}

	if fields, ok := request.([]interface{}); ok {

		for i := 0; i < len(fields); i++ {

			var (
				item proto.Limit
			)

			columns := fields[i].([]interface{})
			if column, ok := columns[0].(string); ok {
				item.Name = column
			}

			if column, ok := columns[1].(float64); ok {
				item.NetLimit = column / 100000000
			}

			if column, ok := columns[2].(float64); ok {
				item.GrossLimit = column / 100000000
			}

			response.Fields = append(response.Fields, &item)
		}

	}

	return &response, nil
}

// GetCounterpartyLimits - get counterparty limits.
func (m *MarketService) GetCounterpartyLimits(_ context.Context, _ *proto.GetMarketRequestCounterpartyLimit) (*proto.ResponseCounterpartyLimit, error) {

	var (
		response proto.ResponseCounterpartyLimit
	)

	request, err := m.request("climits", nil)
	if err != nil {
		return &response, err
	}

	if fields, ok := request.([]interface{}); ok {

		for i := 0; i < len(fields); i++ {

			columns := fields[i].([]interface{})
			response.Fields = append(response.Fields, &proto.CounterpartyLimit{
				Cid:              int64(columns[6].(float64)),
				Name:             columns[0].(string),
				NetLimit:         columns[1].(float64) / 100000000,
				GrossLimit:       columns[2].(float64) / 100000000,
				NetUtilisation:   columns[3].(float64) / 100000000,
				GrossUtilisation: columns[4].(float64) / 100000000,
				Reserved:         int32(columns[5].(float64)),
				TakerMarkup:      int32(columns[7].(float64)),
			})

		}

	}

	return &response, nil
}

// GetSettlementRequests - get settlement requests.
func (m *MarketService) GetSettlementRequests(_ context.Context, _ *proto.GetMarketRequestSettlementRequest) (*proto.ResponseSettlementRequest, error) {

	var (
		response proto.ResponseSettlementRequest
		item     proto.SettlementRequest
	)

	request, err := m.request("settlementRequests", nil)
	if err != nil {
		return &response, err
	}

	if fields, ok := request.([]interface{}); ok {

		for i := 0; i < len(fields[0].([]interface{})); i++ {

			columns := fields[0].([]interface{})[i].([]interface{})
			item.Incoming = append(item.Incoming, &proto.SettlementRequest_Item{
				Cid:       int64(columns[0].(float64)),
				Name:      columns[1].(string),
				Flags:     int32(columns[2].(float64)),
				Size:      columns[3].(float64) / 100000000,
				Comment:   columns[4].(string),
				Timestamp: int64(columns[5].(float64)),
			})

		}

		for i := 0; i < len(fields[1].([]interface{})); i++ {

			columns := fields[1].([]interface{})[i].([]interface{})
			item.Incoming = append(item.Incoming, &proto.SettlementRequest_Item{
				Cid:       int64(columns[0].(float64)),
				Name:      columns[1].(string),
				Flags:     int32(columns[2].(float64)),
				Size:      columns[3].(float64) / 100000000,
				Comment:   columns[4].(string),
				Timestamp: int64(columns[5].(float64)),
			})

		}

		response.Fields = &item
	}

	return &response, nil
}

// GetSettlementTransactions - get settlement transactions.
func (m *MarketService) GetSettlementTransactions(_ context.Context, _ *proto.GetMarketRequestSettlementTransaction) (*proto.ResponseSettlementTransaction, error) {

	var (
		response proto.ResponseSettlementTransaction
		item     proto.SettlementTransaction
	)

	request, err := m.request("settlementTransactions", nil)
	if err != nil {
		return &response, err
	}

	if fields, ok := request.([]interface{}); ok {

		for i := 0; i < len(fields[0].([]interface{})); i++ {

			columns := fields[0].([]interface{})[i].([]interface{})
			item.Incoming = append(item.Incoming, &proto.SettlementTransaction_Item{
				Cid:       int64(columns[0].(float64)),
				Name:      columns[1].(string),
				Size:      columns[2].(float64) / 100000000,
				Oid:       int64(columns[3].(float64)),
				Comment:   columns[4].(string),
				CreateAt:  int64(columns[5].(float64)),
				TxId:      columns[6].(string),
				SentAt:    int64(columns[7].(float64)),
				Flags:     int32(columns[8].(float64)),
				Timestamp: int64(columns[9].(float64)),
				DealId:    int64(columns[10].(float64)),
				Fee:       columns[11].(float64) / 100000000,
			})

		}

		for i := 0; i < len(fields[1].([]interface{})); i++ {

			columns := fields[1].([]interface{})[i].([]interface{})
			item.Incoming = append(item.Incoming, &proto.SettlementTransaction_Item{
				Cid:       int64(columns[0].(float64)),
				Name:      columns[1].(string),
				Size:      columns[2].(float64) / 100000000,
				Oid:       int64(columns[3].(float64)),
				Comment:   columns[4].(string),
				CreateAt:  int64(columns[5].(float64)),
				TxId:      columns[6].(string),
				SentAt:    int64(columns[7].(float64)),
				Flags:     int32(columns[8].(float64)),
				Timestamp: int64(columns[9].(float64)),
				DealId:    int64(columns[10].(float64)),
				Fee:       columns[11].(float64) / 100000000,
			})

		}

		response.Fields = &item
	}

	return &response, nil
}

// GetBook - get book list.
func (m *MarketService) GetBook(_ context.Context, req *proto.GetMarketRequestBook) (*proto.ResponseBook, error) {

	var (
		response proto.ResponseBook
	)

	/*var (
		response proto.ResponseBook
		item     proto.MarketBook
	)

	params := map[string]interface{}{
		"instrument": req.GetInstrument(),
		"tradable":   req.GetTradable(),
	}

	request, err := m.request("book", params)
	if err != nil {
		return &response, m.Context.Error(err)
	}

	if fields, ok := request.([]interface{}); ok {

		for i := 0; i < len(fields[0].([]interface{})); i++ {

			columns := fields[0].([]interface{})[i].([]interface{})
			item.Bid = append(item.Bid, &proto.MarketBook_Book{
				Price: columns[0].(float64) / 100000000,
				Size:  columns[1].(float64) / 100000000,
			})

		}

		for i := 0; i < len(fields[1].([]interface{})); i++ {

			columns := fields[1].([]interface{})[i].([]interface{})
			item.Ask = append(item.Ask, &proto.MarketBook_Book{
				Price: columns[0].(float64) / 100000000,
				Size:  columns[1].(float64) / 100000000,
			})

		}

		response.Fields = &item
	}*/

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
