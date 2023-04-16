package admin_market

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	admin_pbmarket "github.com/cryptogateway/backend-envoys/server/proto/v1/admin.pbmarket"
	"github.com/cryptogateway/backend-envoys/server/service/v2/provider"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
	"strings"
)

// GetPrice - This function is used to get the market price rule from the context. It first checks the authentication of the context
// and then checks whether the user has the rules for writing and editing data. If so, it gets the market price from the
// request and returns it in the response.
func (e *Service) GetPrice(ctx context.Context, req *admin_pbmarket.GetRequestPrice) (*admin_pbmarket.ResponsePrice, error) {

	// This code creates two variables, one called response and one called migrate. The response variable is of type
	// admin_pbmarket.ResponsePrice and the migrate variable is of type query.Migrate. To migrate variable also has a field called
	// Context, which is set to the value of the e.Context variable. This code is used to create two variables that can be used in a program.
	var (
		response admin_pbmarket.ResponsePrice
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking if the user has the necessary permissions to make changes to the data. If the user does not
	// have the proper permissions, an error message is returned. The "migrate.Rules" function is used to check the user's
	// authorization and the "query.RoleMarket" argument is used to specify the type of action (e.g. writing or editing). If
	// the user is not authorized to make changes to the data, the code returns an error message.
	if !migrate.Rules(auth, "pairs", query.RoleMarket) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// The purpose of this code is to check the price of a product in a marketplace. It uses the Price() function of the
	// marketplace to get the price, and then it checks if the price is greater than 0. If it is, the code assigns the price
	// to a response variable.
	if price := marketplace.Price().Unit(req.GetBaseUnit(), req.GetQuoteUnit()); price > 0 {
		response.Price = price
	}

	return &response, nil
}

// SetAsset - This function is used to set currency rules in a context. It checks the authorization of the user, the length of the
// currency name and symbol, and sets the currency by either updating existing data or inserting new data into the
// database. It also handles renaming and creating images related to the currency.
func (e *Service) SetAsset(ctx context.Context, req *admin_pbmarket.SetRequestAsset) (*admin_pbmarket.ResponseAsset, error) {

	// The purpose of this code is to create variables which can be used in order to make queries and handle responses from
	// the admin_pbmarket API. The variables are of type 'admin_pbmarket.ResponseAsset', 'query.Migrate', and 'query.Query'. These
	// variables can then be used to make calls to the API, handle responses, and perform migrations as needed.
	var (
		response admin_pbmarket.ResponseAsset
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking if the user has the proper permissions to access and edit data. If they do not have the
	// required authorization, it returns an error message letting them know they are not allowed to access the data.
	if !migrate.Rules(auth, "assets", query.RoleMarket) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// This code checks if the length of the currency name is less than 4 characters. If it is less than 4 characters, it
	// returns an error message indicating that the asset name must not be less than 4 characters.
	if len(req.Asset.GetName()) < 4 {
		return &response, status.Error(86618, "asset name must not be less than < 4 characters")
	}

	// This code checks the length of the asset symbol. If the length of the currency symbol is less than 2 characters,
	// it returns an error message indicating that the asset symbol must not be less than 2 characters. This is to ensure that the asset symbol is valid.
	if len(req.Asset.GetSymbol()) < 2 {
		return &response, status.Error(17078, "asset symbol must not be less than < 2 characters")
	}

	// This code is using the json.Marshal function to convert a Go data structure req.Asset.GetFields() into JSON. If
	// an error occurs, the error is returned with the Context.Error function.
	serialize, err := json.Marshal(req.Asset.GetFields())
	if err != nil {
		return &response, err
	}

	// This line of code converts the symbol (which is a string) to lowercase letters. This is often used when doing string
	// comparisons and searches, as it makes the comparison easier and more accurate.
	req.Symbol = strings.ToLower(req.GetSymbol())

	// This code is setting the asset symbol field of the req object to the lowercase version of the symbol that is
	// returned from the GetSymbol() method. This is likely being done to ensure that symbols are consistently stored in the same format.
	req.Asset.Symbol = strings.ToLower(req.Asset.GetSymbol())

	// This statement is checking if the length of the req.GetSymbol() is greater than 0. If it is, then the code following
	// this statement will be executed. This statement is used to verify that the req.GetSymbol() is not empty.
	if len(req.GetSymbol()) > 0 {

		// Provider is used to create a Service instance with the given context.
		_provider := provider.Service{
			Context: e.Context,
		}

		// Query the asset with the given symbol from the provider and store the result in asset. Return an error if there is one.
		asset, err := _provider.QueryAsset(req.GetSymbol(), false)
		if err != nil {
			return &response, err
		}

		// This code is part of an update statement in which the purpose is to update the asset's information in the
		// database. This statement is written in the Go programming language, and it uses the Exec method to execute a SQL
		// query that updates the asset's name, symbol, min/max withdraw/deposit/trade, fees, marker, status, type, and
		// chains based on the parameters passed in through the req object. The last parameter, req.GetSymbol(), is used to identify which record should be updated.
		if _, err := e.Context.Db.Exec(`update assets set name = $1, symbol = $2, min_withdraw = $3, max_withdraw = $4, min_trade = $5, max_trade = $6, fees_trade = $7, fees_discount = $8, marker = $9, status = $10, "group" = $11, chains = $12 where symbol = $13;`,
			req.Asset.GetName(),
			req.Asset.GetSymbol(),
			req.Asset.GetMinWithdraw(),
			req.Asset.GetMaxWithdraw(),
			req.Asset.GetMinTrade(),
			req.Asset.GetMaxTrade(),
			req.Asset.GetFeesTrade(),
			req.Asset.GetFeesDiscount(),
			req.Asset.GetMarker(),
			req.Asset.GetStatus(),
			req.Asset.GetGroup(),
			serialize,
			req.GetSymbol(),
		); err != nil {
			return &response, err
		}

		// The purpose of this code is to update all the relevant tables in a database when the asset symbol changes. It
		// does this by checking if the currency symbol from the request is different from the existing currency symbol, and if
		// so, it updates the associated tables with the new symbol.
		if req.GetSymbol() != req.Asset.GetSymbol() {
			_, _ = e.Context.Db.Exec("update balances set symbol = $2 where symbol = $1 and type = $3", req.GetSymbol(), req.Asset.GetSymbol(), asset.GetType())
			_, _ = e.Context.Db.Exec("update ohlcv set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2) where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Asset.GetSymbol())
			_, _ = e.Context.Db.Exec("update trades set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2) where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Asset.GetSymbol())
			_, _ = e.Context.Db.Exec("update orders set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2) where base_unit = $1 and type = $3 or quote_unit = $1 and type = $3", req.GetSymbol(), req.Asset.GetSymbol(), asset.GetType())
			_, _ = e.Context.Db.Exec("update reserves set symbol = $2 where symbol = $1", req.GetSymbol(), req.Asset.GetSymbol())
			_, _ = e.Context.Db.Exec("update assets set symbol = $2 where symbol = $1", req.GetSymbol(), req.Asset.GetSymbol())
		}

	} else {

		// This code is inserting new information into a table called assets. The information being inserted is coming from
		// the req.Asset object. The information is being inserted into a specific order, corresponding to the columns of
		// the table. The purpose is to store the information about a currency in the currencies table.
		if _, err := e.Context.Db.Exec(`insert into assets (name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, marker, "group", status, type, chains) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			req.Asset.GetName(),
			req.Asset.GetSymbol(),
			req.Asset.GetMinWithdraw(),
			req.Asset.GetMaxWithdraw(),
			req.Asset.GetMinTrade(),
			req.Asset.GetMaxTrade(),
			req.Asset.GetFeesTrade(),
			req.Asset.GetFeesDiscount(),
			req.Asset.GetMarker(),
			req.Asset.GetGroup(),
			req.Asset.GetStatus(),
			req.Asset.GetType(),
			serialize,
		); err != nil {
			return &response, err
		}

	}

	// This if statement is checking to see if the length of the "req.GetImage()" is greater than 0. If it is, then the code
	// within the statement will execute. This could be used to check if the "req.GetImage()" contains any data before
	// attempting to do something with it.
	if len(req.GetImage()) > 0 {

		// This code is checking if the GetSymbol() method of req is greater than 0, and if so, it assigns the value of
		// GetSymbol() to the variable q.Name. If GetSymbol() is not greater than 0, it assigns the value of the
		// Currency.GetSymbol() method of req to the variable q.Name.
		if len(req.GetSymbol()) > 0 {
			q.Name = req.GetSymbol()
		} else {
			q.Name = req.Asset.GetSymbol()
		}

		// This code is used to perform an image migration. It takes the image from the request, saves it as "icon" with a
		// given name, and resizes it to a resolution of 300x300. If the migration fails for any reason, it will return an
		// error in the response.
		if err := migrate.Image(req.GetImage(), "icon", q.Name, 300, 300); err != nil {
			return &response, err
		}
	} else {

		// This if statement is checking if the symbol of the request (req.GetSymbol()) is not equal to the currency symbol of
		// the request (req.Asset.GetSymbol()). This could be used to check if the currency being requested is valid for the given symbol.
		if req.GetSymbol() != req.Asset.GetSymbol() {

			// This code is part of a function that is likely updating a database and renaming a currency's symbol. The purpose of
			// this code is to attempt to rename the currency's symbol using to migrate.Rename() function, and if an error
			// occurs, the error is returned using to err.
			if err := migrate.Rename("icon", req.GetSymbol(), req.Asset.GetSymbol()); err != nil {
				return &response, err
			}
		}
	}

	return &response, nil
}

// GetAsset - This code is part of a service that is used to get asset rules. It is responsible for authenticating the user and
// retrieving the authentication data. It then checks to see if the user has the correct permissions to access the
// currency data. If so, it retrieves the requested currency information and returns it. If there is an error, it is returned to the caller.
func (e *Service) GetAsset(ctx context.Context, req *admin_pbmarket.GetRequestAsset) (*admin_pbmarket.ResponseAsset, error) {

	// This code creates two variables, response and migrate. The variable response is of type admin_pbmarket.ResponseAsset and
	// the variable migrate is of type query.Migrate. To migrate variable contains a Context field, and it is assigned the
	// e.Context value. The purpose of this code is to create two variables and assign values to them.
	var (
		response admin_pbmarket.ResponseAsset
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking if the user has the correct permissions to access the data. The `migrate.Rules` function takes
	// in the user's authorization, the data they are trying to access (in this case "currencies"), and their role (in this
	// case "RoleSpot"). If the user does not have the correct permissions, the code returns an error with status code 12011.
	if !migrate.Rules(auth, "assets", query.RoleMarket) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// Provider is used to create a Service instance with the given context.
	_provider := provider.Service{
		Context: e.Context,
	}

	// This code is checking if the currency ID of the requested symbol is greater than 0. If it is, it will add the
	// asset to a list of fields in the response.
	if currency, _ := _provider.QueryAsset(req.GetSymbol(), false); currency.GetId() > 0 {
		response.Fields = append(response.Fields, currency)
	}

	return &response, nil
}

// GetAssets - This code is part of a service for retrieving currency rules. It is used to authenticate the user, retrieve the
// authentication data, and query the database for the currency rule data requested by the user. It will return any
// errors encountered during the authentication and query processes.
func (e *Service) GetAssets(ctx context.Context, req *admin_pbmarket.GetRequestAssets) (*admin_pbmarket.ResponseAsset, error) {

	// The purpose of this code is to declare three variables: response, migrate, and maps. The variable response is of the
	// type admin_pbmarket.ResponseAsset. The variable migrate is of the type query.Migrate, and has a field of type context. The
	// variable maps is of the type string array.
	var (
		response admin_pbmarket.ResponseAsset
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	// The purpose of the above code is to set a default limit for the request if the limit is not specified by the user.
	// This prevents the request from having an undefined limit and ensures that the request is processed with a valid limit value.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of this if statement is to check if the user has the necessary permissions to perform the desired action.
	// If the user does not have the correct permissions, an error is returned to the user.
	if !migrate.Rules(auth, "assets", query.RoleMarket) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// This code is used to add a WHERE clause to a SQL query. If req.GetSearch() returns a non-zero length string, then a
	// WHERE clause is appended to the maps array, checking if either the symbol or name field matches the search string.
	// The %[1]s syntax is used to insert the value of req.GetSearch() in the same place twice in the string.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where case when '%[1]v' is not null and '%[1]v' <> '' then type = '%[1]v' else true end and (symbol like %[2]s or name like %[2]s)", req.GetType(), "'%"+req.GetSearch()+"%'"))
	} else if len(req.GetType()) > 0 {
		maps = append(maps, fmt.Sprintf("where type = '%v'", req.GetType()))
	}

	// This code is used to count the number of rows in the 'currencies' table in the database. The code uses the QueryRow()
	// function to execute a statement which retrieves the count of the number of rows in the 'currencies' table. The result
	// is then stored in the 'response.Count' variable. The 'strings.Join(maps, " ")' is used to join the elements of the
	// 'maps' array with a space in between.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from assets %s", strings.Join(maps, " "))).Scan(&response.Count)

	// The purpose of this code is to check if the response object has a count value greater than 0. If it has a value
	// greater than 0, it means that the response has been successfully processed and can be used.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for pagination in a request. It takes the limit (how many results per
		// page) and page (which page to display) from the request and calculates how many results to skip over (offset) before
		// displaying the page. The second line adjusts the offset if the page is greater than 0. If the page is 0, then the
		// offset should remain 0 so that the first page of results is displayed.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query the database. It builds a query with the given parameters (maps, req.GetLimit(), offset).
		// It then attempts to execute it and, if an error is encountered, the function returns the error. Finally, it closes the rows.
		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, "group", type, create_at from assets %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate over a set of rows returned from a database query. It is used to step
		// through each row in a result set one row at a time, retrieving data from each column and performing any necessary
		// operations or calculations.
		for rows.Next() {

			var (
				item types.Asset
			)

			// This code is used to scan through a row from a database and assign each column value to the corresponding item
			// field. This allows the item to be populated with the values from the row. The if statement is used to check if any
			// errors occurred while scanning the row, and if so, return an error.
			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.Symbol,
				&item.MinWithdraw,
				&item.MaxWithdraw,
				&item.MinTrade,
				&item.MaxTrade,
				&item.FeesTrade,
				&item.FeesDiscount,
				&item.FeesCharges,
				&item.FeesCosts,
				&item.Marker,
				&item.Status,
				&item.Group,
				&item.Type,
				&item.CreateAt,
			); err != nil {
				return &response, err
			}

			// This code is used to add an item to the response Fields array. The response.Fields array is a slice of pointers
			// that stores data. The code uses the append function to add the item to the end of the array.
			response.Fields = append(response.Fields, &item)
		}

		// The purpose of this code is to check for errors when reading data from a database. If an error is encountered, it
		// will return an error response and the corresponding error message.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// DeleteAsset - This code is part of a delete currency rule function. The purpose of the code is to delete a currency rule, including
// the associated data in the wallets, assets, trades, transfers, orders, reserves and currencies tables. The code also
// attempts to authenticate the user before deleting the data and checks their authorization to delete the data. If there
// is an error, it is returned to the caller.
func (e *Service) DeleteAsset(ctx context.Context, req *admin_pbmarket.DeleteRequestAsset) (*admin_pbmarket.ResponseAsset, error) {

	// The purpose of this code is to create two variables, response and migrate, which are of the types
	// admin_pbmarket.ResponseAsset and query.Migrate, respectively. To migrate variable is also given a Context property which
	// is equal to the context stored in the e variable.
	var (
		response admin_pbmarket.ResponseAsset
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking whether the user has the correct permissions for writing and editing data. If the user does not
	// have the necessary permissions, it will return an error message.
	if !migrate.Rules(auth, "assets", query.RoleMarket) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// Provider is used to create a Service instance with the given context.
	_provider := provider.Service{
		Context: e.Context,
	}

	// This code is part of a function used to delete a currency from the database. The first part of the code checks if the
	// currency exists and if it does, the code will then execute a series of SQL queries to delete the currency and all
	// related data in other related tables.
	if row, _ := _provider.QueryAsset(req.GetSymbol(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from pairs where base_unit = $1 or quote_unit = $1", req.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from balances where symbol = $1 and type = $2", row.GetSymbol(), row.GetType())
		_, _ = e.Context.Db.Exec("delete from ohlcv where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from trades where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from orders where base_unit = $1 and type = $2 or quote_unit = $1 and type = $2", row.GetSymbol(), row.GetType())
		_, _ = e.Context.Db.Exec("delete from reserves where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from assets where symbol = $1", row.GetSymbol())
	}

	// This code is used to remove a symbol from the "icon" database. It is checking for any errors that may occur while
	// performing the removal, and if an error is encountered, it will return the response and the error context.
	if err := migrate.RemoveFiles("icon", req.GetSymbol()); err != nil {
		return &response, err
	}

	return &response, nil
}

// GetPairs - This code is part of a service that is responsible for retrieving data about pairs (e.g. currency pairs) from a
// database. The code attempts to authenticate the user and retrieve the authentication data. It then checks the user has
// the appropriate permissions to access the data. If so, it queries the database for data and returns the retrieved data
// to the caller. If any errors occur, they are returned to the caller.
func (e *Service) GetPairs(ctx context.Context, req *admin_pbmarket.GetRequestPairs) (*admin_pbmarket.ResponsePair, error) {

	// The code above creates three variables: response, migrate, and maps. The variable response is of type
	// admin_pbmarket.ResponsePair and the variable migrate is of type query.Migrate. The variable maps is of type string and is
	// initialized to an empty slice. The variable migrate is initialized to a query.Migrate object with a context field set
	// to the value of e.Context.
	var (
		response admin_pbmarket.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	// The purpose of this code is to check if the value of the variable req.GetLimit() is equal to 0, and if it is, set the
	// value of the variable req.Limit to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking if a user has permissions to perform a certain action. The migrate.Rules() function is used to
	// check if the user has the necessary authorization to write and edit data, and if they do not, an error is returned
	// indicating that they do not have the necessary permissions.
	if !migrate.Rules(auth, "pairs", query.RoleMarket) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// This code is checking to see if the request includes a search term. If a search term is present, it appends a string
	// to the maps variable that includes a "where" clause in a SQL query to search for the search term in the base_unit and
	// quote_unit fields.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where case when '%[1]v' is not null and '%[1]v' <> '' then type = '%[1]v' else true end and (base_unit like %[2]s or quote_unit like %[2]s)", req.GetType(), "'%"+req.GetSearch()+"%'"))
	} else if len(req.GetType()) > 0 {
		maps = append(maps, fmt.Sprintf("where type = '%v'", req.GetType()))
	}

	// The purpose of this code is to query the database for the total number of entries in the "pairs" table, and store the
	// result in the response struct. The strings.Join() function is used to create a single string from the maps slice,
	// which is then appended to the SQL query string to create the full request. Finally, the result of the query is stored
	// in the "Count" field of the response struct using the Scan() function.
	if _ = e.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count from pairs %s`, strings.Join(maps, " "))).Scan(&response.Count); response.GetCount() > 0 {

		// The purpose of this code is to calculate the offset of results based on the limit and page that the user has
		// requested. This offset is used in pagination to determine which set of results to retrieve from a database. The code
		// uses the limit and page parameters to calculate the offset. If the page is greater than 0, the offset is calculated
		// by multiplying the limit by the page minus one. Otherwise, the offset is calculated by multiplying the limit by the page.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query the database for data. The specific query is retrieving rows from the table "pairs" with
		// parameters specified by the variables strings.Join(maps, " "), req.GetLimit(), and offset. The query returns the
		// data in columns named id, base_unit, quote_unit, price, base_decimal, quote_decimal, and status. The query is also
		// ordered by the id column in descending order and limited to the req.GetLimit() number of rows with an offset of
		// offset. If an error occurs, the code returns the response variable and an error. Finally, the rows.Close() statement
		// is used to close the connection to the database when the query is complete.
		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, base_unit, quote_unit, price, base_decimal, quote_decimal, type, status from pairs %[1]s order by id desc limit %[2]d offset %[3]d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate over the rows of a database query result. It is used to loop through the
		// rows of a query result and process each row one at a time.
		for rows.Next() {

			var (
				item types.Pair
			)

			// This code is used to scan a row from a database query and assign the values to variables. The if statement checks
			// for any errors that may arise from the scan and returns an error if one occurs. The variables in the scan statement
			// are all parts of an item and are being assigned the values from the row.
			if err = rows.Scan(
				&item.Id,
				&item.BaseUnit,
				&item.QuoteUnit,
				&item.Price,
				&item.BaseDecimal,
				&item.QuoteDecimal,
				&item.Type,
				&item.Status,
			); err != nil {
				return &response, err
			}

			// This appends the item to the end of the response.Fields slice. It adds the item to the existing response.Fields
			// slice, allowing for the creation of a larger slice with the new item included.
			response.Fields = append(response.Fields, &item)
		}

		// The purpose of this code is to check for any errors that occurred while operating on the rows, and to return an
		// error response if there is an issue. It is a safety measure to help make sure that any errors are caught and handled properly.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// GetPair - This code is part of a function that is used to get a pair rule from a database. It first attempts to authenticate the
// user before retrieving the data. It then checks if the user has the necessary permissions to access the data, and if
// so, retrieves the pair rule from the database. Finally, it returns the retrieved pair rule as part of a response along
// with any errors that may have occurred during the process.
func (e *Service) GetPair(ctx context.Context, req *admin_pbmarket.GetRequestPair) (*admin_pbmarket.ResponsePair, error) {

	// The purpose of this code is to declare a variable response of type admin_pbmarket.ResponsePair and a variable migrate of type
	// query.Migrate. The variable migrate is also initialized with a Context field set to the value of the variable e.Context.
	var (
		response admin_pbmarket.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of the if statement is to check whether the user has the necessary permissions (defined by to
	// migrate.Rules function) to write and edit data. If the user does not have the necessary permissions, an error message is returned.
	if !migrate.Rules(auth, "pairs", query.RoleMarket) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// Provider is used to create a Service instance with the given context.
	_provider := provider.Service{
		Context: e.Context,
	}

	// This statement is checking if the ID of the pair (obtained with the getPair method) is greater than 0. If it is, then
	// the pair is appended to the response.Fields array.
	if pair, _ := _provider.QueryPair(req.GetId(), types.TypeZero, false); pair.GetId() > 0 {
		response.Fields = append(response.Fields, pair)
	}

	return &response, nil
}

// SetPair - This code is part of a function which is used to set a pair rule. The code is responsible for authenticating the user,
// validating the request data, and either creating a new pair or updating an existing one. It also checks for any errors
// that may occur and returns them to the caller.
func (e *Service) SetPair(ctx context.Context, req *admin_pbmarket.SetRequestPair) (*admin_pbmarket.ResponsePair, error) {

	// The purpose of this code is to declare three variables - response, migrate, and q. The variable response is declared
	// as a type admin_pbmarket.ResponsePair, migrate is declared as a type query.Migrate, and q is declared as a type query.Query.
	// The variable migrate is also initialized with the value e.Context.
	var (
		response admin_pbmarket.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking if the user has the appropriate permissions to write and edit data. If the user does not have
	// the rules for writing and editing data, then the code will return an error message.
	if !migrate.Rules(auth, "pairs", query.RoleMarket) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// This code checks to make sure that the base and quote currencies for a given request (req) have been set. If either
	// of them have not been set, then an error is returned indicating that both must be set.
	if len(req.Pair.GetBaseUnit()) == 0 && len(req.Pair.GetQuoteUnit()) == 0 {
		return &response, status.Error(55615, "base currency and quote currency must be set")
	}

	// This if statement is used to check if the price of a given pair is set to 0. If it is, the statement will return an
	// error with the status code 46517 and the message "the price must be set". This is likely in place to ensure that the
	// user is not attempting to purchase a product at an incorrect price.
	if req.Pair.GetPrice() == 0 {
		return &response, status.Error(46517, "the price must be set")
	}

	// This is a conditional statement that checks if the value of req.GetId() is greater than 0. If the condition is true,
	// then the code inside the curly braces will be executed. Otherwise, the code will be skipped. This conditional
	// statement is usually used to determine if a certain condition is met before executing certain code.
	if req.GetId() > 0 {

		// This code is part of a larger program and its purpose is to delete trades from the database that have a base unit
		// and quote unit that match the values provided in the request. The if statement is checking the value of
		// GetGraphClear() before executing the delete statement.
		if req.Pair.GetGraphClear() {
			_, _ = e.Context.Db.Exec("delete from ohlcv where base_unit = $1 and quote_unit = $2", req.Pair.GetBaseUnit(), req.Pair.GetQuoteUnit())
		}

		// This code is used to update an entry in the database table 'pairs', using the values in the 'req' struct. It updates
		// the 'base_unit', 'quote_unit', 'price', 'base_decimal', 'quote_decimal' and 'status' fields of the database table,
		// where the value of the 'id' field of the database table is equal to the value of the 'Id' field in the 'req' struct.
		// The code also includes an if statement to check for any errors in the process.
		if _, err := e.Context.Db.Exec("update pairs set base_unit = $1, quote_unit = $2, price = $3, base_decimal = $4, quote_decimal = $5, type = $6, status = $7 where id = $8;",
			req.Pair.GetBaseUnit(),
			req.Pair.GetQuoteUnit(),
			req.Pair.GetPrice(),
			req.Pair.GetBaseDecimal(),
			req.Pair.GetQuoteDecimal(),
			req.Pair.GetType(),
			req.Pair.GetStatus(),
			req.GetId(),
		); err != nil {
			return &response, err
		}

	} else {

		// This code is checking if the requested pair (req.Pair) already exists in the list of pairs stored in the database.
		// If it does exist, an error message is returned with a status code of 50605.
		if _ = e.Context.Db.QueryRow("select id from pairs where base_unit = $1 and quote_unit = $2 or base_unit = $2 and quote_unit = $1", req.Pair.GetBaseUnit(), req.Pair.GetQuoteUnit()).Scan(&q.Id); q.Id > 0 {
			return &response, status.Error(50605, "the pair you are trying to create is already in the list of pairs")
		}

		// This code is used to insert a record into a database table called 'pairs' using the values from a request object. It
		// is using the 'Exec' function from the database context to execute an SQL statement for inserting the values into the
		// table. The 'if _, err' statement is checking for any errors that may have occurred from the execution of the
		// statement. If an error is detected, the code will return an error response.
		if _, err := e.Context.Db.Exec("insert into pairs (base_unit, quote_unit, price, base_decimal, quote_decimal, type, status) values ($1, $2, $3, $4, $5, $6, $7)",
			req.Pair.GetBaseUnit(),
			req.Pair.GetQuoteUnit(),
			req.Pair.GetPrice(),
			req.Pair.GetBaseDecimal(),
			req.Pair.GetQuoteDecimal(),
			req.Pair.GetType(),
			req.Pair.GetStatus(),
		); err != nil {
			return &response, err
		}

	}
	response.Success = true

	return &response, nil
}

// DeletePair - The purpose of this code is to delete a given Pair Rule from a database. It checks the authentication of the user and
// checks the rules for writing and editing data. If the user is authorized, the code will delete the Pair Rule from the
// database and return a response indicating success.
func (e *Service) DeletePair(ctx context.Context, req *admin_pbmarket.DeleteRequestPair) (*admin_pbmarket.ResponsePair, error) {

	// The purpose of the above code is to create two variables, response and migrate. response is of type
	// admin_pbmarket.ResponsePair, while migrate is of type query.Migrate and has a field named Context which is initialized to e.Context.
	var (
		response admin_pbmarket.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking to see if the user has the necessary authorization to perform a write or edit operation on the
	// data. If the user does not have the correct authorization to do so, an error is returned with a status code of 12011.
	if !migrate.Rules(auth, "pairs", query.RoleMarket) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// Provider is used to create a Service instance with the given context.
	_provider := provider.Service{
		Context: e.Context,
	}

	// This code snippet is attempting to delete a pair from the database. The first line is checking if the pair exists in
	// the database, based on the request's ID. If the ID is greater than zero, then the code proceeds to delete the pair
	// from the database, as well as all associated trades, transfers, and orders.
	if row, _ := _provider.QueryPair(req.GetId(), types.TypeZero, false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from pairs where id = $1", row.GetId())
		_, _ = e.Context.Db.Exec("delete from ohlcv where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
		_, _ = e.Context.Db.Exec("delete from trades where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
		_, _ = e.Context.Db.Exec("delete from orders where base_unit = $1 and quote_unit = $2 and type = $3", row.GetBaseUnit(), row.GetQuoteUnit(), row.GetType())
	}
	response.Success = true

	return &response, nil
}
