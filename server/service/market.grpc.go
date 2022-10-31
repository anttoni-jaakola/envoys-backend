package service

import (
	"github.com/cryptogateway/backend-envoys/server/proto"
	"golang.org/x/net/context"
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

			columns := fields[i].([]interface{})
			response.Fields = append(response.Fields, &proto.Limit{
				Name:             columns[0].(string),
				NetLimit:         columns[1].(float64) / 100000000,
				GrossLimit:       columns[2].(float64) / 100000000,
				NetUtilisation:   columns[3].(float64) / 100000000,
				GrossUtilisation: columns[4].(float64) / 100000000,
				Reserved:         int32(columns[5].(float64)),
			})

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
	//TODO implement me
	panic("implement me")
}
