package stock

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"github.com/cryptogateway/backend-envoys/server/query"
	"google.golang.org/grpc/status"
	"strings"
)

// SetMarketRule - set new stock item to markets.
func (e *Service) SetMarketRule(ctx context.Context, req *pbstock.SetRequestMarketRule) (*pbstock.ResponseMarket, error) {

	var (
		response pbstock.ResponseMarket
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "markets", query.RoleStock) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.Market.GetName()) < 4 {
		return &response, e.Context.Error(status.Error(86618, "market name must not be less than < 4 characters"))
	}

	if len(req.Market.GetSymbol()) < 2 {
		return &response, e.Context.Error(status.Error(17078, "market symbol must not be less than < 2 characters"))
	}

	req.Symbol = strings.ToLower(req.GetSymbol())
	req.Market.Symbol = strings.ToLower(req.Market.GetSymbol())

	if len(req.GetSymbol()) > 0 {

		if _, err := e.Context.Db.Exec("update stock_markets set name = $1, symbol = $2, unit = $3, code = $4, qty_shares = $5, address = $6, price_buy = $7, price_sell = $8, price_market = $9, exchange = $10, sector = $11, method = $12, kind = $13, type = $14, start_at = $15, stop_at = $16, status = $17 where symbol = $18;",
			req.Market.GetName(),
			req.Market.GetSymbol(),
			req.Market.GetUnit(),
			req.Market.GetCode(),
			req.Market.GetQtyShares(),
			req.Market.GetAddress(),
			req.Market.GetPriceBuy(),
			req.Market.GetPriceSell(),
			req.Market.GetPriceMarket(),
			req.Market.GetExchange(),
			req.Market.GetSector(),
			req.Market.GetMethod(),
			req.Market.GetKind(),
			req.Market.GetType(),
			req.Market.GetStartAt(),
			req.Market.GetStopAt(),
			req.Market.GetStatus(),
			req.GetSymbol(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		if _, err := e.Context.Db.Exec("insert into stock_markets (name, symbol, unit, code, qty_shares, address, price_buy, price_sell, price_market, exchange, sector, method, kind, type, start_at, stop_at, status) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)",
			req.Market.GetName(),
			req.Market.GetSymbol(),
			req.Market.GetUnit(),
			req.Market.GetCode(),
			req.Market.GetQtyShares(),
			req.Market.GetAddress(),
			req.Market.GetPriceBuy(),
			req.Market.GetPriceSell(),
			req.Market.GetPriceMarket(),
			req.Market.GetExchange(),
			req.Market.GetSector(),
			req.Market.GetMethod(),
			req.Market.GetKind(),
			req.Market.GetType(),
			req.Market.GetStartAt(),
			req.Market.GetStopAt(),
			req.Market.GetStatus(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}

	if len(req.GetImage()) > 0 {

		if len(req.GetSymbol()) > 0 {
			q.Name = req.GetSymbol()
		} else {
			q.Name = req.Market.GetSymbol()
		}

		if err := migrate.Image(req.GetImage(), "icon", fmt.Sprintf("stock-%v", q.Name)); err != nil {
			return &response, e.Context.Error(err)
		}
	} else {
		if req.GetSymbol() != req.Market.GetSymbol() {
			if err := migrate.Rename("icon", fmt.Sprintf("stock-%v", req.GetSymbol()), fmt.Sprintf("stock-%v", req.Market.GetSymbol())); err != nil {
				return &response, e.Context.Error(err)
			}
		}
	}

	return &response, nil
}

// GetMarketRule - get market by symbol.
func (e *Service) GetMarketRule(ctx context.Context, req *pbstock.GetRequestMarketRule) (*pbstock.ResponseMarket, error) {

	var (
		response pbstock.ResponseMarket
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "markets", query.RoleStock) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if market, _ := e.getMarket(req.GetSymbol(), false); market.GetId() > 0 {
		response.Fields = append(response.Fields, market)
	}

	return &response, nil
}

// GetMarketsRule - get markets.
func (e *Service) GetMarketsRule(ctx context.Context, req *pbstock.GetRequestMarketsRule) (*pbstock.ResponseMarket, error) {

	var (
		response pbstock.ResponseMarket
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

	if !migrate.Rules(account, "markets", query.RoleStock) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where symbol like %[1]s or name like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from stock_markets %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, name, symbol, code, qty_shares, address, price_buy, price_sell, price_market, exchange, sector, unit, method, start_at, stop_at, kind, type, status, create_at from stock_markets %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbstock.Market
			)

			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.Symbol,
				&item.Code,
				&item.QtyShares,
				&item.Address,
				&item.PriceBuy,
				&item.PriceSell,
				&item.PriceMarket,
				&item.Exchange,
				&item.Sector,
				&item.Unit,
				&item.Method,
				&item.StartAt,
				&item.StopAt,
				&item.Kind,
				&item.Type,
				&item.Status,
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

// DeleteMarketRule - delete stock item by id.
func (e *Service) DeleteMarketRule(ctx context.Context, req *pbstock.DeleteRequestMarketRule) (*pbstock.ResponseMarket, error) {

	var (
		response pbstock.ResponseMarket
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "markets", query.RoleStock) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	_, _ = e.Context.Db.Exec("delete from stock_markets where id = $1", req.GetId())

	return &response, nil
}

// SetSectorRule - set new sector.
func (e *Service) SetSectorRule(ctx context.Context, req *pbstock.SetRequestSectorRule) (*pbstock.ResponseSector, error) {

	var (
		response pbstock.ResponseSector
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "sectors", query.RoleStock) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.Sector.GetName()) < 4 {
		return &response, e.Context.Error(status.Error(86618, "sector name must not be less than < 4 characters"))
	}

	if len(req.Sector.GetSymbol()) < 2 {
		return &response, e.Context.Error(status.Error(17078, "sector symbol must not be less than < 2 characters"))
	}

	req.Sector.Symbol = strings.ToLower(req.Sector.GetSymbol())
	if req.GetId() > 0 {

		var (
			item pbstock.Sector
		)

		if err := e.Context.Db.QueryRow("select name, symbol from stock_sectors where id = $1", req.GetId()).Scan(
			&item.Name,
			&item.Symbol,
		); err != nil {
			return &response, err
		}

		_, _ = e.Context.Db.Exec("update stock_markets set unit = $3, sector = $4 where unit = $1 and sector = $2", item.GetSymbol(), item.GetName(), req.Sector.GetSymbol(), req.Sector.GetName())

		if _, err := e.Context.Db.Exec("update stock_sectors set name = $1, symbol = $2, status = $3 where id = $4;",
			req.Sector.GetName(),
			req.Sector.GetSymbol(),
			req.Sector.GetStatus(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		if _, err := e.Context.Db.Exec("insert into stock_sectors (name, symbol, status) values ($1, $2, $3)",
			req.Sector.GetName(),
			req.Sector.GetSymbol(),
			req.Sector.GetStatus(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}

	return &response, nil
}

// GetSectorRule - get sector.
func (e *Service) GetSectorRule(ctx context.Context, req *pbstock.GetRequestSectorRule) (*pbstock.ResponseSector, error) {

	var (
		response pbstock.ResponseSector
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "sectors", query.RoleStock) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if sector, _ := e.getSector(req.GetId()); sector.GetId() > 0 {
		response.Fields = append(response.Fields, sector)
	}

	return &response, nil
}

// GetSectorsRule - get sectors.
func (e *Service) GetSectorsRule(ctx context.Context, req *pbstock.GetRequestSectorsRule) (*pbstock.ResponseSector, error) {

	var (
		response pbstock.ResponseSector
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "sectors", query.RoleStock) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	_ = e.Context.Db.QueryRow("select count(*) as count from stock_sectors").Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query("select id, name, symbol, status from stock_sectors order by id desc limit $1 offset $2", req.GetLimit(), offset)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item pbstock.Sector
			)

			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.Symbol,
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

// DeleteSectorRule - delete sector by id.
func (e *Service) DeleteSectorRule(ctx context.Context, req *pbstock.DeleteRequestSectorRule) (*pbstock.ResponseSector, error) {

	var (
		response pbstock.ResponseSector
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	account, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(account, "sectors", query.RoleStock) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	_, _ = e.Context.Db.Exec("delete from stock_sectors where id = $1", req.GetId())

	return &response, nil
}

// SetRegistrarRule - set a new registrar.
func (e *Service) SetRegistrarRule(ctx context.Context, rule *pbstock.SetRequestRegistrarRule) (*pbstock.ResponseRegistrar, error) {
	//TODO implement me
	panic("implement me")
}

// GetRegistrarRule - get registrar by id.
func (e *Service) GetRegistrarRule(ctx context.Context, rule *pbstock.GetRequestRegistrarRule) (*pbstock.ResponseRegistrar, error) {
	//TODO implement me
	panic("implement me")
}

// GetRegistrarsRule - get registrars.
func (e *Service) GetRegistrarsRule(ctx context.Context, rule *pbstock.GetRequestRegistrarsRule) (*pbstock.ResponseRegistrar, error) {
	//TODO implement me
	panic("implement me")
}

// DeleteRegistrarRule - delete registrar by id.
func (e *Service) DeleteRegistrarRule(ctx context.Context, rule *pbstock.DeleteRequestRegistrarRule) (*pbstock.ResponseRegistrar, error) {
	//TODO implement me
	panic("implement me")
}

// SetDepositaryRule - set new depositary.
func (e *Service) SetDepositaryRule(ctx context.Context, rule *pbstock.SetRequestDepositaryRule) (*pbstock.ResponseDepositary, error) {
	//TODO implement me
	panic("implement me")
}

// GetDepositaryRule - get depositary by id.
func (e *Service) GetDepositaryRule(ctx context.Context, rule *pbstock.GetRequestDepositaryRule) (*pbstock.ResponseDepositary, error) {
	//TODO implement me
	panic("implement me")
}

// GetDepositariesRule - get all list depositaries.
func (e *Service) GetDepositariesRule(ctx context.Context, rule *pbstock.GetRequestDepositariesRule) (*pbstock.ResponseDepositary, error) {
	//TODO implement me
	panic("implement me")
}

// DeleteDepositaryRule - delete depositary by id.
func (e *Service) DeleteDepositaryRule(ctx context.Context, rule *pbstock.DeleteRequestDepositaryRule) (*pbstock.ResponseDepositary, error) {
	//TODO implement me
	panic("implement me")
}
