package stock

import (
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	Context *assets.Context
}

// Initialization - perform actions.
func (e *Service) Initialization() {}

// getMarket - getting information about the market.
func (e *Service) getMarket(symbol string, status bool) (*pbstock.Market, error) {

	var (
		response pbstock.Market
		maps     []string
		storage  []string
	)

	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, name, symbol, unit, code, qty_shares, address, price_buy, price_sell, price_market, exchange, sector, method, kind, type, start_at, stop_at, status, create_at from stock_markets where symbol = '%v' %s", symbol, strings.Join(maps, " "))).Scan(
		&response.Id,
		&response.Name,
		&response.Symbol,
		&response.Unit,
		&response.Code,
		&response.QtyShares,
		&response.Address,
		&response.PriceBuy,
		&response.PriceSell,
		&response.PriceMarket,
		&response.Exchange,
		&response.Sector,
		&response.Method,
		&response.Kind,
		&response.Type,
		&response.StartAt,
		&response.StopAt,
		&response.Status,
		&response.CreateAt,
	); err != nil {
		return &response, err
	}

	storage = append(storage, []string{e.Context.StoragePath, "static", "icon", fmt.Sprintf("%v.png", fmt.Sprintf("stock-%v", response.GetSymbol()))}...)
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		response.Icon = true
	}

	return &response, nil
}

// getSector - getting information about the sector.
func (e *Service) getSector(id int64) (*pbstock.Sector, error) {

	var (
		response pbstock.Sector
	)

	if err := e.Context.Db.QueryRow("select id, name, symbol, status from stock_sectors where id = $1", id).Scan(
		&response.Id,
		&response.Name,
		&response.Symbol,
		&response.Status,
	); err != nil {
		return &response, err
	}

	return &response, nil
}
