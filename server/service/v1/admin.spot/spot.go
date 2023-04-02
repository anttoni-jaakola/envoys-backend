package admin_spot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/types"
	"os"
	"path/filepath"
	"strings"
)

// Service - The purpose of the Service struct is to store data related to a service, such as the Context, run and wait maps, and
// the block map. The Context is a pointer to an assets Context, which contains information about the service. The run
// and wait maps are booleans that indicate whether the service is running or waiting for an action. The block map is an
// integer that stores the block number associated with a particular service.
type Service struct {
	Context *assets.Context
}

// getAsset - This function is used to retrieve asset information from a database. It takes a currency symbol and a status
// boolean as arguments. It then queries the database to retrieve information about the currency and stores it in the
// 'response' variable. It then checks for the existence of the asset icon and stores the result in the 'icon' field
// of the 'response' variable. Finally, it returns the currency and an error value, if any.
func (e *Service) getAsset(symbol string, status bool) (*types.Asset, error) {

	var (
		response types.Asset
		maps     []string
		storage  []string
		chains   []byte
	)

	// The purpose of this code is to append an item to a list of maps if a certain condition is met. In this case, if the
	// "status" variable is true, a string will be appended to the list of maps.
	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	// This code is performing a query of a database table called "currencies" and scanning the results into a response
	// object. The query is using the symbol parameter to filter the results and strings.Join(maps, " ") to join any
	// additional parameters. If the query fails, an error is returned.
	if err := e.Context.Db.QueryRow(fmt.Sprintf(`select id, name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, "group", create_at, chains from assets where symbol = '%v' %s`, symbol, strings.Join(maps, " "))).Scan(
		&response.Id,
		&response.Name,
		&response.Symbol,
		&response.MinWithdraw,
		&response.MaxWithdraw,
		&response.MinTrade,
		&response.MaxTrade,
		&response.FeesTrade,
		&response.FeesDiscount,
		&response.FeesCharges,
		&response.FeesCosts,
		&response.Marker,
		&response.Status,
		&response.Group,
		&response.CreateAt,
		&chains,
	); err != nil {
		return &response, err
	}

	// The purpose of the code is to add a string to a storage slice. The string is made up of elements from the
	// e.Context.StoragePath, the word "static", the word "icon", and a string created from the response.GetSymbol()
	// function. The ... in the code indicates that the elements of the slice are being "unpacked" into to append() call.
	storage = append(storage, []string{e.Context.StoragePath, "static", "icon", fmt.Sprintf("%v.png", response.GetSymbol())}...)

	// This statement is checking to see if a file at the given filepath exists. If it does, then the response.Icon will be
	// set to true. This statement is used in order to prevent the creation of unnecessary files.
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		response.Icon = true
	}

	// The purpose of this code is to unmarshal a json object into a response object. This is done using the
	// json.Unmarshal() function. The function takes the json object (chains) and a reference to the response.ChainsIds
	// object. If there is an error, it will be returned with the error context.
	if err := json.Unmarshal(chains, &response.ChainsIds); err != nil {
		return &response, err
	}

	return &response, nil
}

// getUnit - This function is used to get a unit from a database based on a given symbol. It queries the database for a row that
// contains the given symbol as either the base_unit or the quote_unit, and scans the row for the id, price, base_unit,
// quote_unit, and status of the unit. If successful, it returns the response and nil for the error, otherwise it returns
// an empty response and the error.
func (e *Service) getUnit(symbol string) (*types.Pair, error) {

	var (
		response types.Pair
	)

	// This code is part of a function which queries a database for a row that matches the given symbol. The if statement is
	// used to scan the row for the requested values and return the response. If an error occurs during the scanning
	// process, the function returns the response and the error.
	if err := e.Context.Db.QueryRow(`select id, price, base_unit, quote_unit, status from pairs where base_unit = $1 or quote_unit = $1`, symbol).Scan(&response.Id, &response.Price, &response.BaseUnit, &response.QuoteUnit, &response.Status); err != nil {
		return &response, err
	}

	return &response, nil
}

// getChain - This function is used to query a row from a database table "chains" with the given id and status. It then scans the
// row into a types.Chain struct. If there is an error, it returns the struct with an error. Otherwise, it returns the
// struct with no error.
func (e *Service) getChain(id int64, status bool) (*types.Chain, error) {

	var (
		chain types.Chain
		maps  []string
	)

	// The purpose of this code is to add the string "and status = true" to the maps slice, if the status variable is set to true.
	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	// This code is used to query a database for a row of data which matches the given id. The query is built by joining the
	// strings in the maps array and is passed to the QueryRow method. The data is then scanned into the chain object and
	// returned. If there is an error, it will be returned instead.
	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, name, rpc, block, network, explorer_link, platform, confirmation, time_withdraw, fees, tag, parent_symbol, decimals, status from chains where id = %[1]d %[2]s", id, strings.Join(maps, " "))).Scan(
		&chain.Id,
		&chain.Name,
		&chain.Rpc,
		&chain.Block,
		&chain.Network,
		&chain.ExplorerLink,
		&chain.Platform,
		&chain.Confirmation,
		&chain.TimeWithdraw,
		&chain.Fees,
		&chain.Tag,
		&chain.ParentSymbol,
		&chain.Decimals,
		&chain.Status,
	); err != nil {
		return &chain, errors.New("chain not found or chain network off")
	}

	return &chain, nil
}

// getPair - This function is used to get a specific pair from the database, based on the id and status passed as arguments. The
// function returns a pointer to a 'types.Pair' struct and an error if any. It prepares a query to select the specified
// pair from the database, based on the given id and status. It then scans the results and stores them in the struct, and
// finally returns the struct and an error if any.
func (e *Service) getPair(id int64, status bool) (*types.Pair, error) {

	var (
		chain types.Pair
		maps  []string
	)

	// The purpose of this code is to append a string to a list of maps if a certain condition is true. In this case, if the
	// variable "status" is true, the string "and status = %v" with "true" as the placeholder value is added to the list of maps.
	if status {
		maps = append(maps, fmt.Sprintf("and status = %v", true))
	}

	// This code is used to query a database and retrieve information about a pair with a specified id. The query is formed
	// using the fmt.Sprintf() function, and it is a combination of a string and the id parameter. The retrieved information
	// is then assigned to the chain struct. Finally, the code returns the chain struct and an error if it fails.
	if err := e.Context.Db.QueryRow(fmt.Sprintf("select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs where id = %[1]d %[2]s", id, strings.Join(maps, " "))).Scan(
		&chain.Id,
		&chain.BaseUnit,
		&chain.QuoteUnit,
		&chain.Price,
		&chain.BaseDecimal,
		&chain.QuoteDecimal,
		&chain.Status,
	); err != nil {
		return &chain, err
	}

	return &chain, nil
}

// getContractById - This function is used to retrieve a contract from a database by its ID. It queries the database for the contract with
// the specified ID, and then scans the query result into the fields of the pbspot.Contract struct. The function then
// returns the contract and any errors that may have occurred.
func (e *Service) getContractById(id int64) (*types.Contract, error) {

	var (
		contract types.Contract
	)

	// This code is a SQL query used to retrieve data from a database. The purpose of the query is to select data from the
	// contracts and chains tables for a specific contract with a given ID. The query will return information such as the
	// symbol, chain ID, address, fees withdraw, protocol, decimals, and platform of the contract. The query uses the Scan()
	// method to store the retrieved data in the contract variable. The if statement is used to check for errors and return
	// the contract along with an error if one occurs.
	if err := e.Context.Db.QueryRow(`select c.id, c.symbol, c.chain_id, c.address, c.fees, c.protocol, c.decimals, n.platform from contracts c inner join chains n on n.id = c.chain_id where c.id = $1`, id).Scan(&contract.Id, &contract.Symbol, &contract.ChainId, &contract.Address, &contract.Fees, &contract.Protocol, &contract.Decimals, &contract.Platform); err != nil {
		return &contract, err
	}

	return &contract, nil
}
