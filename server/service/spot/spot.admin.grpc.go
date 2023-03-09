package spot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/keypair"
	"github.com/cryptogateway/backend-envoys/assets/common/marketplace"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"google.golang.org/grpc/status"
	"strings"
)

// GetMarketPriceRule - This function is used to get the market price rule from the context. It first checks the authentication of the context
// and then checks whether the user has the rules for writing and editing data. If so, it gets the market price from the
// request and returns it in the response.
func (e *Service) GetMarketPriceRule(ctx context.Context, req *pbspot.GetRequestPriceManual) (*pbspot.ResponsePrice, error) {

	// This code creates two variables, one called response and one called migrate. The response variable is of type
	// pbspot.ResponsePrice and the migrate variable is of type query.Migrate. The migrate variable also has a field called
	// Context, which is set to the value of the e.Context variable. This code is used to create two variables that can be used in a program.
	var (
		response pbspot.ResponsePrice
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking if the user has the necessary permissions to make changes to the data. If the user does not
	// have the proper permissions, an error message is returned. The "migrate.Rules" function is used to check the user's
	// authorization and the "query.RoleSpot" argument is used to specify the type of action (e.g. writing or editing). If
	// the user is not authorized to make changes to the data, the code returns an error message.
	if !migrate.Rules(auth, "pairs", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// The purpose of this code is to check the price of a product in a marketplace. It uses the Price() function of the
	// marketplace to get the price, and then it checks if the price is greater than 0. If it is, the code assigns the price
	// to a response variable.
	if price := marketplace.Price().Unit(req.GetBaseUnit(), req.GetQuoteUnit()); price > 0 {
		response.Price = price
	}

	return &response, nil
}

// SetCurrencyRule - This function is used to set currency rules in a context. It checks the authorization of the user, the length of the
// currency name and symbol, and sets the currency by either updating existing data or inserting new data into the
// database. It also handles renaming and creating images related to the currency.
func (e *Service) SetCurrencyRule(ctx context.Context, req *pbspot.SetRequestCurrencyRule) (*pbspot.ResponseCurrency, error) {

	// The purpose of this code is to create variables which can be used in order to make queries and handle responses from
	// the PBSpot API. The variables are of type 'pbspot.ResponseCurrency', 'query.Migrate', and 'query.Query'. These
	// variables can then be used to make calls to the API, handle responses, and perform migrations as needed.
	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking if the user has the proper permissions to access and edit data. If they do not have the
	// required authorization, it returns an error message letting them know they are not allowed to access the data.
	if !migrate.Rules(auth, "currencies", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code checks if the length of the currency name is less than 4 characters. If it is less than 4 characters, it
	// returns an error message indicating that the currency name must not be less than 4 characters.
	if len(req.Currency.GetName()) < 4 {
		return &response, e.Context.Error(status.Error(86618, "currency name must not be less than < 4 characters"))
	}

	// This code checks the length of the currency symbol. If the length of the currency symbol is less than 2 characters,
	// it returns an error message indicating that the currency symbol must not be less than 2 characters. This is to ensure that the currency symbol is valid.
	if len(req.Currency.GetSymbol()) < 2 {
		return &response, e.Context.Error(status.Error(17078, "currency symbol must not be less than < 2 characters"))
	}

	// This code is using the json.Marshal function to convert a Go data structure req.Currency.GetChainsIds() into JSON. If
	// an error occurs, the error is returned with the Context.Error function.
	serialize, err := json.Marshal(req.Currency.GetChainsIds())
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This line of code converts the symbol (which is a string) to lowercase letters. This is often used when doing string
	// comparisons and searches, as it makes the comparison easier and more accurate.
	req.Symbol = strings.ToLower(req.GetSymbol())

	// This code is setting the Currency Symbol field of the req object to the lowercase version of the symbol that is
	// returned from the GetSymbol() method. This is likely being done to ensure that symbols are consistently stored in the same format.
	req.Currency.Symbol = strings.ToLower(req.Currency.GetSymbol())

	// This statement is checking if the length of the req.GetSymbol() is greater than 0. If it is, then the code following
	// this statement will be executed. This statement is used to verify that the req.GetSymbol() is not empty.
	if len(req.GetSymbol()) > 0 {

		// This code is part of an update statement in which the purpose is to update the currency's information in the
		// database. This statement is written in the Go programming language, and it uses the Exec method to execute a SQL
		// query that updates the currency's name, symbol, min/max withdraw/deposit/trade, fees, marker, status, type, and
		// chains based on the parameters passed in through the req object. The last parameter, req.GetSymbol(), is used to identify which record should be updated.
		if _, err := e.Context.Db.Exec("update currencies set name = $1, symbol = $2, min_withdraw = $3, max_withdraw = $4, min_trade = $5, max_trade = $6, fees_trade = $7, fees_discount = $8, marker = $9, status = $10, type = $11, chains = $12 where symbol = $13;",
			req.Currency.GetName(),
			req.Currency.GetSymbol(),
			req.Currency.GetMinWithdraw(),
			req.Currency.GetMaxWithdraw(),
			req.Currency.GetMinTrade(),
			req.Currency.GetMaxTrade(),
			req.Currency.GetFeesTrade(),
			req.Currency.GetFeesDiscount(),
			req.Currency.GetMarker(),
			req.Currency.GetStatus(),
			req.Currency.GetType(),
			serialize,
			req.GetSymbol(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

		// The purpose of this code is to update all the relevant tables in a database when the currency symbol changes. It
		// does this by checking if the currency symbol from the request is different from the existing currency symbol, and if
		// so, it updates the associated tables with the new symbol.
		if req.GetSymbol() != req.Currency.GetSymbol() {
			_, _ = e.Context.Db.Exec("update wallets set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update assets set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update trades set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2) where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update transfers set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2)  where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update orders set base_unit = coalesce(nullif(base_unit, $1), $2), quote_unit = coalesce(nullif(quote_unit, $1), $2)  where base_unit = $1 or quote_unit = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update reserves set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
			_, _ = e.Context.Db.Exec("update currencies set symbol = $2 where symbol = $1", req.GetSymbol(), req.Currency.GetSymbol())
		}

	} else {

		// This code is inserting new information into a table called currencies. The information being inserted is coming from
		// the req.Currency object. The information is being inserted into a specific order, corresponding to the columns of
		// the table. The purpose is to store the information about a currency in the currencies table.
		if _, err := e.Context.Db.Exec("insert into currencies (name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, marker, type, status, chains) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			req.Currency.GetName(),
			req.Currency.GetSymbol(),
			req.Currency.GetMinWithdraw(),
			req.Currency.GetMaxWithdraw(),
			req.Currency.GetMinTrade(),
			req.Currency.GetMaxTrade(),
			req.Currency.GetFeesTrade(),
			req.Currency.GetFeesDiscount(),
			req.Currency.GetMarker(),
			req.Currency.GetType(),
			req.Currency.GetStatus(),
			serialize,
		); err != nil {
			return &response, e.Context.Error(err)
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
			q.Name = req.Currency.GetSymbol()
		}

		// This code is used to perform an image migration. It takes the image from the request, saves it as "icon" with a
		// given name, and resizes it to a resolution of 300x300. If the migration fails for any reason, it will return an
		// error in the response.
		if err := migrate.Image(req.GetImage(), "icon", q.Name, 300, 300); err != nil {
			return &response, e.Context.Error(err)
		}
	} else {

		// This if statement is checking if the symbol of the request (req.GetSymbol()) is not equal to the currency symbol of
		// the request (req.Currency.GetSymbol()). This could be used to check if the currency being requested is valid for the given symbol.
		if req.GetSymbol() != req.Currency.GetSymbol() {

			// This code is part of a function that is likely updating a database and renaming a currency's symbol. The purpose of
			// this code is to attempt to rename the currency's symbol using the migrate.Rename() function, and if an error
			// occurs, the error is returned using the e.Context.Error() function.
			if err := migrate.Rename("icon", req.GetSymbol(), req.Currency.GetSymbol()); err != nil {
				return &response, e.Context.Error(err)
			}
		}
	}

	return &response, nil
}

// GetCurrencyRule - This code is part of a service that is used to get currency rules. It is responsible for authenticating the user and
// retrieving the authentication data. It then checks to see if the user has the correct permissions to access the
// currency data. If so, it retrieves the requested currency information and returns it. If there is an error, it is returned to the caller.
func (e *Service) GetCurrencyRule(ctx context.Context, req *pbspot.GetRequestCurrencyRule) (*pbspot.ResponseCurrency, error) {

	// This code creates two variables, response and migrate. The variable response is of type pbspot.ResponseCurrency and
	// the variable migrate is of type query.Migrate. The migrate variable contains a Context field and it is assigned the
	// e.Context value. The purpose of this code is to create two variables and assign values to them.
	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking if the user has the correct permissions to access the data. The `migrate.Rules` function takes
	// in the user's authorization, the data they are trying to access (in this case "currencies"), and their role (in this
	// case "RoleSpot"). If the user does not have the correct permissions, the code returns an error with status code 12011.
	if !migrate.Rules(auth, "currencies", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking if the currency ID of the requested symbol is greater than 0. If it is, it will add the
	// currency to a list of fields in the response.
	if currency, _ := e.getCurrency(req.GetSymbol(), false); currency.GetId() > 0 {
		response.Fields = append(response.Fields, currency)
	}

	return &response, nil
}

// GetCurrenciesRule - This code is part of a service for retrieving currency rules. It is used to authenticate the user, retrieve the
// authentication data, and query the database for the currency rule data requested by the user. It will return any
// errors encountered during the authentication and query processes.
func (e *Service) GetCurrenciesRule(ctx context.Context, req *pbspot.GetRequestCurrenciesRule) (*pbspot.ResponseCurrency, error) {

	// The purpose of this code is to declare three variables: response, migrate, and maps. The variable response is of the
	// type pbspot.ResponseCurrency. The variable migrate is of the type query.Migrate, and has a field of type context. The
	// variable maps is of the type string array.
	var (
		response pbspot.ResponseCurrency
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
		return &response, e.Context.Error(err)
	}

	// The purpose of this if statement is to check if the user has the necessary permissions to perform the desired action.
	// If the user does not have the correct permissions, an error is returned to the user.
	if !migrate.Rules(auth, "currencies", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is used to add a WHERE clause to a SQL query. If req.GetSearch() returns a non-zero length string, then a
	// WHERE clause is appended to the maps array, checking if either the symbol or name field matches the search string.
	// The %[1]s syntax is used to insert the value of req.GetSearch() in the same place twice in the string.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where symbol like %[1]s or name like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	// This code is used to count the number of rows in the 'currencies' table in the database. The code uses the QueryRow()
	// function to execute a statement which retrieves the count of the number of rows in the 'currencies' table. The result
	// is then stored in the 'response.Count' variable. The 'strings.Join(maps, " ")' is used to join the elements of the
	// 'maps' array with a space in between.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from currencies %s", strings.Join(maps, " "))).Scan(&response.Count)

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
		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, status, type, create_at from currencies %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate over a set of rows returned from a database query. It is used to step
		// through each row in a result set one row at a time, retrieving data from each column and performing any necessary
		// operations or calculations.
		for rows.Next() {

			var (
				item pbspot.Currency
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
				&item.Type,
				&item.CreateAt,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			// This code is used to add an item to the response Fields array. The response.Fields array is a slice of pointers
			// that stores data. The code uses the append function to add the item to the end of the array.
			response.Fields = append(response.Fields, &item)
		}

		// The purpose of this code is to check for errors when reading data from a database. If an error is encountered, it
		// will return an error response and the corresponding error message.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// DeleteCurrencyRule - This code is part of a delete currency rule function. The purpose of the code is to delete a currency rule, including
// the associated data in the wallets, assets, trades, transfers, orders, reserves and currencies tables. The code also
// attempts to authenticate the user before deleting the data and checks their authorization to delete the data. If there
// is an error, it is returned to the caller.
func (e *Service) DeleteCurrencyRule(ctx context.Context, req *pbspot.DeleteRequestCurrencyRule) (*pbspot.ResponseCurrency, error) {

	// The purpose of this code is to create two variables, response and migrate, which are of the types
	// pbspot.ResponseCurrency and query.Migrate, respectively. The migrate variable is also given a Context property which
	// is equal to the context stored in the e variable.
	var (
		response pbspot.ResponseCurrency
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking whether the user has the correct permissions for writing and editing data. If the user does not
	// have the necessary permissions, it will return an error message.
	if !migrate.Rules(auth, "currencies", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking if the row in the database corresponding to the requested symbol has an ID greater than zero.
	// If it does, it will execute a delete statement from the table 'pairs' to remove the row.
	if row, _ := e.getUnit(req.GetSymbol()); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from pairs where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
	}

	// This code is part of a function used to delete a currency from the database. The first part of the code checks if the
	// currency exists and if it does, the code will then execute a series of SQL queries to delete the currency and all
	// related data in other related tables.
	if row, _ := e.getCurrency(req.GetSymbol(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from wallets where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from assets where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from trades where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from transfers where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from orders where base_unit = $1 or quote_unit = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from reserves where symbol = $1", row.GetSymbol())
		_, _ = e.Context.Db.Exec("delete from currencies where symbol = $1", row.GetSymbol())
	}

	// This code is used to remove a symbol from the "icon" database. It is checking for any errors that may occur while
	// performing the removal, and if an error is encountered, it will return the response and the error context.
	if err := migrate.RemoveFiles("icon", req.GetSymbol()); err != nil {
		return &response, e.Context.Error(err)
	}

	return &response, nil
}

// GetChainsRule - This code is a function to get the chains rule from the database. It authenticates the user and checks if they have
// the correct privileges to access the data. The function then queries the database and returns the chains rule data in
// a ResponseChain object. It also handles pagination and limits the number of results.
func (e *Service) GetChainsRule(ctx context.Context, req *pbspot.GetRequestChainsRule) (*pbspot.ResponseChain, error) {

	// The purpose of this code is to create two variables, response and migrate, that can be used in the rest of the
	// program. The response variable is of type pbspot.ResponseChain, and the migrate variable is of type query.Migrate,
	// with the context set to the value of the e.Context variable.
	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// The purpose of this code is to check if the limit of the request (req) is set to 0. If it is, then the code sets the limit to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This if statement is used to check if the given user has the necessary permissions to perform a certain operation. If
	// the user does not have the permission, an error message is returned.
	if !migrate.Rules(auth, "chains", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is used to query a database and store the result in the response.Count variable. It is typically used to
	// get a count of the number of entries in a table.
	_ = e.Context.Db.QueryRow("select count(*) as count from chains").Scan(&response.Count)

	// This is a conditional statement that checks if the value of response.GetCount() is greater than 0. If the condition
	// is true, then the code inside the block will be executed. If not, the code will be skipped.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for a paginated query. It takes into account the limit (the number of
		// items per page) and the page number to determine the offset (the number of items to skip). If the page number is
		// greater than 0, the offset is calculated by subtracting 1 from the page number before multiplying by the limit.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query a database and fetch data from the database. The query is selecting certain columns from
		// the table "chains" and ordering them in descending order of id, with a limit and an offset set by the request. If
		// there is an error, the error is returned. Finally, the rows object is closed.
		rows, err := e.Context.Db.Query(`select id, name, rpc, block, network, explorer_link, platform, confirmation, time_withdraw, fees, tag, decimals, status from chains order by id desc limit $1 offset $2`, req.GetLimit(), offset)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for loop with rows.Next() is used to iterate through the rows of a result set returned from a SQL query. The
		// rows.Next() function will return true if there is a row to read and false when it reaches the end of the result set.
		// The loop will execute the code inside the loop for every row in the result set until it reaches the end.
		for rows.Next() {

			var (
				item pbspot.Chain
			)

			// This code is used to scan through a row of data and assign each column value to a variable. The variables are
			// item.Id, item.Name, item.Rpc, etc. The if statement checks for any errors while scanning the row and returns an
			// error if any occur.
			if err = rows.Scan(&item.Id, &item.Name, &item.Rpc, &item.Block, &item.Network, &item.ExplorerLink, &item.Platform, &item.Confirmation, &item.TimeWithdraw, &item.Fees, &item.Tag, &item.Decimals, &item.Status); err != nil {
				return &response, e.Context.Error(err)
			}

			// This code is adding the item to the response.Fields array. The purpose of this line of code is to append the item
			// to the existing array of response.Fields.
			response.Fields = append(response.Fields, &item)
		}

		// This if statement is used to check for any errors when dealing with the rows of a database. If an error is
		// encountered, the statement returns the response and an error message that can be used for debugging.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetChainRule - This code is part of an authentication process. Its purpose is to attempt to authenticate the user and retrieve the
// authentication data. If there is an error, it is returned to the caller. It also checks the user's permissions to
// ensure they have the appropriate rules for writing and editing data. Finally, if the request is valid, it retrieves
// the relevant data from the database and returns it to the caller.
func (e *Service) GetChainRule(ctx context.Context, req *pbspot.GetRequestChainRule) (*pbspot.ResponseChain, error) {

	// The purpose of this code is to declare two variables: response, which is of type pbspot.ResponseChain, and migrate,
	// which is of type query.Migrate. To migrate variable is also assigned a Context property with the value of e.Context.
	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking to see if the user has certain permissions (rules) to perform an action on a particular set of
	// data (chains). If the user does not have the required permissions, an error message is returned.
	if !migrate.Rules(auth, "chains", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking if the chain with the given ID exists and then adding it to the response if it does. The "_" is
	// used to ignore the second return value from the getChain() function.
	if chain, _ := e.getChain(req.GetId(), false); chain.GetId() > 0 {
		response.Fields = append(response.Fields, chain)
	}

	return &response, nil
}

// SetChainRule - This code is part of a service that sets up a chain rule. The purpose of this code is to authenticate the user,
// validate the data provided, and then insert or update the data in the chain database. The code also performs error
// checking to ensure that the user has the correct permissions and that the data provided is valid. Additionally, the
// code also performs a ping test to make sure that the chain server address is available.
func (e *Service) SetChainRule(ctx context.Context, req *pbspot.SetRequestChainRule) (*pbspot.ResponseChain, error) {

	// The purpose of this code is to declare two variables, response and migrate. The variable response is of type
	// pbspot.ResponseChain, and the variable migrate is of type query.Migrate. The variable migrate also has the Context
	// field set to e.Context. This code is likely part of a larger program that uses these variables to do something.
	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking the user's authorization for writing and editing data. The first part of the if statement
	// checks is the user has the necessary rules to write and edit data in the chains. The second part of the if statement
	// checks is the user has been explicitly denied the right to make changes to the data. If either of these checks fail,
	// the code returns an error message indicating that the user does not have the necessary rules to write and edit the data.
	if !migrate.Rules(auth, "chains", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This statement checks the length of the chain name and returns an error if it is less than 4 characters. This is
	// likely done to ensure that the chain name is long enough to be a valid name.
	if len(req.Chain.GetName()) < 4 {
		return &response, e.Context.Error(status.Error(86611, "chain name must not be less than < 4 characters"))
	}

	// This code is used to check if the length of the RPC address in the request is at least 10 characters. If it is not,
	// then the code will return an error message of 44511, "chain rpc address must be at least < 10 characters". This is
	// done to ensure that the RPC address is valid and will meet the minimum requirements for use.
	if len(req.Chain.GetRpc()) < 10 {
		return &response, e.Context.Error(status.Error(44511, "chain rpc address must be at least < 10 characters"))
	}

	// This code is checking if the chain server address is available by pinging it with the help.Ping function. If the ping
	// fails, an error is returned and the response is not sent. This is likely to alert the user that their request failed
	// because the chain server is unavailable.
	if ok := help.Ping(req.Chain.GetRpc()); !ok {
		return &response, e.Context.Error(status.Error(45601, "chain server address not available"))
	}

	// This is a conditional statement that checks if the value of the req.GetId() function is greater than 0. If it is,
	// then the code in the code block that follows will be executed. If it is not, then the code will be skipped.
	if req.GetId() > 0 {

		// This code is an SQL statement that updates the values of a database entry in the "chains" table. It sets the values
		// of the database fields (name, rpc, network, block, explorer_link, platform, confirmation, time_withdraw,
		// fees_withdraw, tag, parent_symbol, and status) to values passed in the request (req). The id of the entry
		// to be updated is also passed in the request. The purpose of this code is to update the values of a particular database entry in the "chains" table.
		if _, err := e.Context.Db.Exec("update chains set name = $1, rpc = $2, network = $3, block = $4, explorer_link = $5, platform = $6, confirmation = $7, time_withdraw = $8, fees = $9, tag = $10, parent_symbol = $11, decimals = $12, status = $13 where id = $14;",
			req.Chain.GetName(),
			req.Chain.GetRpc(),
			req.Chain.GetNetwork(),
			req.Chain.GetBlock(),
			req.Chain.GetExplorerLink(),
			req.Chain.GetPlatform(),
			req.Chain.GetConfirmation(),
			req.Chain.GetTimeWithdraw(),
			req.Chain.GetFees(),
			req.Chain.GetTag(),
			req.Chain.GetParentSymbol(),
			req.Chain.GetDecimals(),
			req.Chain.GetStatus(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		// This code is used to insert data into a database table called 'chains'. The purpose of this code is to insert the
		// values of the 'req.Chain' object into the specified fields of the 'chains' table. The variables that are being
		// inserted are the name, RPC, network, block, explorer link, platform, confirmation, time withdraw, fees withdraw,
		// tag, parent symbol, and status of the chain object.
		if _, err := e.Context.Db.Exec("insert into chains (name, rpc, network, block, explorer_link, platform, confirmation, time_withdraw, fees, tag, parent_symbol, status) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
			req.Chain.GetName(),
			req.Chain.GetRpc(),
			req.Chain.GetNetwork(),
			req.Chain.GetBlock(),
			req.Chain.GetExplorerLink(),
			req.Chain.GetPlatform(),
			req.Chain.GetConfirmation(),
			req.Chain.GetTimeWithdraw(),
			req.Chain.GetFees(),
			req.Chain.GetTag(),
			req.Chain.GetParentSymbol(),
			req.Chain.GetStatus(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// DeleteChainRule - This code is part of a function used to delete a chain rule from a database. It performs authentication, checks for
// authorization, and then deletes the requested chain rule from the database. It then returns a response indicating that
// the operation was successful.
func (e *Service) DeleteChainRule(ctx context.Context, req *pbspot.DeleteRequestChainRule) (*pbspot.ResponseChain, error) {

	// The code above is declaring two variables: response, of type pbspot.ResponseChain, and migrate, of type
	// query.Migrate. The variable migrate is being initialized with a context from the variable e. The purpose of this code
	// is to declare and initialize two variables that will be used in the code following it.
	var (
		response pbspot.ResponseChain
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking the user's access privileges to determine if they are allowed to perform the requested
	// operation. It checks to see if the user has the 'currencies' rule and if they have the 'deny-record' rule. If they do
	// not have the 'currencies' rule, or if they do have the 'deny-record' rule, the code returns an error message.
	if !migrate.Rules(auth, "currencies", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking if a chain exists with the given ID, and then deleting it from the database if it does exist.
	// The first line is getting the chain from the database, and the second line is removing it from the database. The
	// third line is deleting the chain from the chains table.
	if row, _ := e.getChain(req.GetId(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec(fmt.Sprintf(`update currencies set chains = jsonb_path_query_array(chains, '$[*] ? (@ != %[1]d)') where chains @> '%[1]d'`, row.GetId()))
		_, _ = e.Context.Db.Exec("delete from chains where id = $1", row.GetId())
	}
	response.Success = true

	return &response, nil
}

// GetPairsRule - This code is part of a service that is responsible for retrieving data about pairs (e.g. currency pairs) from a
// database. The code attempts to authenticate the user and retrieve the authentication data. It then checks the user has
// the appropriate permissions to access the data. If so, it queries the database for data and returns the retrieved data
// to the caller. If any errors occur, they are returned to the caller.
func (e *Service) GetPairsRule(ctx context.Context, req *pbspot.GetRequestPairsRule) (*pbspot.ResponsePair, error) {

	// The code above creates three variables: response, migrate, and maps. The variable response is of type
	// pbspot.ResponsePair and the variable migrate is of type query.Migrate. The variable maps is of type string and is
	// initialized to an empty slice. The variable migrate is initialized to a query.Migrate object with a context field set
	// to the value of e.Context.
	var (
		response pbspot.ResponsePair
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
		return &response, e.Context.Error(err)
	}

	// This code is checking if a user has permissions to perform a certain action. The migrate.Rules() function is used to
	// check if the user has the necessary authorization to write and edit data, and if they do not, an error is returned
	// indicating that they do not have the necessary permissions.
	if !migrate.Rules(auth, "pairs", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking to see if the request includes a search term. If a search term is present, it appends a string
	// to the maps variable that includes a "where" clause in a SQL query to search for the search term in the base_unit and
	// quote_unit fields.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where base_unit like %[1]s or quote_unit like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	// The purpose of this code is to query the database for the total number of entries in the "pairs" table, and store the
	// result in the response struct. The strings.Join() function is used to create a single string from the maps slice,
	// which is then appended to the SQL query string to create the full request. Finally, the result of the query is stored
	// in the "Count" field of the response struct using the Scan() function.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from pairs %s", strings.Join(maps, " "))).Scan(&response.Count)

	// This if statement is used to check if the response (which is assumed to be a collection of some sort) contains at
	// least one element. If it does, the code inside the if statement will be executed.
	if response.GetCount() > 0 {

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
		rows, err := e.Context.Db.Query(fmt.Sprintf("select id, base_unit, quote_unit, price, base_decimal, quote_decimal, status from pairs %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate over the rows of a database query result. It is used to loop through the
		// rows of a query result and process each row one at a time.
		for rows.Next() {

			var (
				item pbspot.Pair
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
				&item.Status,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			// This appends the item to the end of the response.Fields slice. It adds the item to the existing response.Fields
			// slice, allowing for the creation of a larger slice with the new item included.
			response.Fields = append(response.Fields, &item)
		}

		// The purpose of this code is to check for any errors that occurred while operating on the rows, and to return an
		// error response if there is an issue. It is a safety measure to help make sure that any errors are caught and handled properly.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetPairRule - This code is part of a function that is used to get a pair rule from a database. It first attempts to authenticate the
// user before retrieving the data. It then checks if the user has the necessary permissions to access the data, and if
// so, retrieves the pair rule from the database. Finally, it returns the retrieved pair rule as part of a response along
// with any errors that may have occurred during the process.
func (e *Service) GetPairRule(ctx context.Context, req *pbspot.GetRequestPairRule) (*pbspot.ResponsePair, error) {

	// The purpose of this code is to declare a variable response of type pbspot.ResponsePair and a variable migrate of type
	// query.Migrate. The variable migrate is also initialized with a Context field set to the value of the variable e.Context.
	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// The purpose of the if statement is to check whether the user has the necessary permissions (defined by to
	// migrate.Rules function) to write and edit data. If the user does not have the necessary permissions, an error message is returned.
	if !migrate.Rules(auth, "pairs", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This statement is checking if the ID of the pair (obtained with the getPair method) is greater than 0. If it is, then
	// the pair is appended to the response.Fields array.
	if pair, _ := e.getPair(req.GetId(), false); pair.GetId() > 0 {
		response.Fields = append(response.Fields, pair)
	}

	return &response, nil
}

// SetPairRule - This code is part of a function which is used to set a pair rule. The code is responsible for authenticating the user,
// validating the request data, and either creating a new pair or updating an existing one. It also checks for any errors
// that may occur and returns them to the caller.
func (e *Service) SetPairRule(ctx context.Context, req *pbspot.SetRequestPairRule) (*pbspot.ResponsePair, error) {

	// The purpose of this code is to declare three variables - response, migrate, and q. The variable response is declared
	// as a type pbspot.ResponsePair, migrate is declared as a type query.Migrate, and q is declared as a type query.Query.
	// The variable migrate is also initialized with the value e.Context.
	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
		q query.Query
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking if the user has the appropriate permissions to write and edit data. If the user does not have
	// the rules for writing and editing data, then the code will return an error message.
	if !migrate.Rules(auth, "pairs", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code checks to make sure that the base and quote currencies for a given request (req) have been set. If either
	// of them have not been set, then an error is returned indicating that both must be set.
	if len(req.Pair.GetBaseUnit()) == 0 && len(req.Pair.GetQuoteUnit()) == 0 {
		return &response, e.Context.Error(status.Error(55615, "base currency and quote currency must be set"))
	}

	// This if statement is used to check if the price of a given pair is set to 0. If it is, the statement will return an
	// error with the status code 46517 and the message "the price must be set". This is likely in place to ensure that the
	// user is not attempting to purchase a product at an incorrect price.
	if req.Pair.GetPrice() == 0 {
		return &response, e.Context.Error(status.Error(46517, "the price must be set"))
	}

	// This is a conditional statement that checks if the value of req.GetId() is greater than 0. If the condition is true,
	// then the code inside the curly braces will be executed. Otherwise, the code will be skipped. This conditional
	// statement is usually used to determine if a certain condition is met before executing certain code.
	if req.GetId() > 0 {

		// This code is part of a larger program and its purpose is to delete trades from the database that have a base unit
		// and quote unit that match the values provided in the request. The if statement is checking the value of
		// GetGraphClear() before executing the delete statement.
		if req.Pair.GetGraphClear() {
			_, _ = e.Context.Db.Exec("delete from trades where base_unit = $1 and quote_unit = $2", req.Pair.GetBaseUnit(), req.Pair.GetQuoteUnit())
		}

		// This code is used to update an entry in the database table 'pairs', using the values in the 'req' struct. It updates
		// the 'base_unit', 'quote_unit', 'price', 'base_decimal', 'quote_decimal' and 'status' fields of the database table,
		// where the value of the 'id' field of the database table is equal to the value of the 'Id' field in the 'req' struct.
		// The code also includes an if statement to check for any errors in the process.
		if _, err := e.Context.Db.Exec("update pairs set base_unit = $1, quote_unit = $2, price = $3, base_decimal = $4, quote_decimal = $5, status = $6 where id = $7;",
			req.Pair.GetBaseUnit(),
			req.Pair.GetQuoteUnit(),
			req.Pair.GetPrice(),
			req.Pair.GetBaseDecimal(),
			req.Pair.GetQuoteDecimal(),
			req.Pair.GetStatus(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		// This code is checking if the requested pair (req.Pair) already exists in the list of pairs stored in the database.
		// If it does exist, an error message is returned with a status code of 50605.
		if _ = e.Context.Db.QueryRow("select id from pairs where base_unit = $1 and quote_unit = $2 or base_unit = $2 and quote_unit = $1", req.Pair.GetBaseUnit(), req.Pair.GetQuoteUnit()).Scan(&q.Id); q.Id > 0 {
			return &response, e.Context.Error(status.Error(50605, "the pair you are trying to create is already in the list of pairs"))
		}

		// This code is used to insert a record into a database table called 'pairs' using the values from a request object. It
		// is using the 'Exec' function from the database context to execute an SQL statement for inserting the values into the
		// table. The 'if _, err' statement is checking for any errors that may have occurred from the execution of the
		// statement. If an error is detected, the code will return an error response.
		if _, err := e.Context.Db.Exec("insert into pairs (base_unit, quote_unit, price, base_decimal, quote_decimal, status) values ($1, $2, $3, $4, $5, $6)",
			req.Pair.GetBaseUnit(),
			req.Pair.GetQuoteUnit(),
			req.Pair.GetPrice(),
			req.Pair.GetBaseDecimal(),
			req.Pair.GetQuoteDecimal(),
			req.Pair.GetStatus(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// DeletePairRule - The purpose of this code is to delete a given Pair Rule from a database. It checks the authentication of the user and
// checks the rules for writing and editing data. If the user is authorized, the code will delete the Pair Rule from the
// database and return a response indicating success.
func (e *Service) DeletePairRule(ctx context.Context, req *pbspot.DeleteRequestPairRule) (*pbspot.ResponsePair, error) {

	// The purpose of the above code is to create two variables, response and migrate. response is of type
	// pbspot.ResponsePair, while migrate is of type query.Migrate and has a field named Context which is initialized to e.Context.
	var (
		response pbspot.ResponsePair
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking to see if the user has the necessary authorization to perform a write or edit operation on the
	// data. If the user does not have the correct authorization to do so, an error is returned with a status code of 12011.
	if !migrate.Rules(auth, "pairs", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code snippet is attempting to delete a pair from the database. The first line is checking if the pair exists in
	// the database, based on the request's ID. If the ID is greater than zero, then the code proceeds to delete the pair
	// from the database, as well as all associated trades, transfers, and orders.
	if row, _ := e.getPair(req.GetId(), false); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from pairs where id = $1", row.GetId())
		_, _ = e.Context.Db.Exec("delete from trades where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
		_, _ = e.Context.Db.Exec("delete from transfers where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
		_, _ = e.Context.Db.Exec("delete from orders where base_unit = $1 and quote_unit = $2", row.GetBaseUnit(), row.GetQuoteUnit())
	}
	response.Success = true

	return &response, nil
}

// GetContractsRule - This code is part of a service that is responsible for retrieving contracts rules from a database. It takes in a
// GetRequestContractsRule and a context, authenticates the user using the context, and then queries the database for the
// requested information. The response is returned as a pbspot.ResponseContract, which contains the requested
// information. If there is an error, it is returned to the caller.
func (e *Service) GetContractsRule(ctx context.Context, req *pbspot.GetRequestContractsRule) (*pbspot.ResponseContract, error) {

	// The code snippet above is declaring three variables: response, migrate, and maps. The variable response is declared
	// as type pbspot.ResponseContract, migrate is declared as type query.Migrate and maps is declared as an array of
	// strings. The purpose of this is to create three variables with the necessary types, so they can be used in the code. These variables can be used to store data and manipulate it, depending on the purpose of the code.
	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	// The purpose of this code is to check if the request limit is 0, and if so, set a default limit of 30. This ensures
	// that a limit is always provided in the request.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This if statement checks is the user has the correct authorization for writing and editing data for the contracts
	// query. If the user does not have the correct authorization, the statement returns an error message indicating the
	// user does not have the correct rules for writing and editing data.
	if !migrate.Rules(auth, "contracts", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// The purpose of this code is to search for either the address or symbol of a customer using the GetSearch() method.
	// The "%" symbol is used to indicate that the search should match any characters before and after the specified text in
	// the req.GetSearch() method. The code will append the search query to the maps variable, which will then be used to
	// search the customer database.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where c.address like %[1]s or c.symbol like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	// The purpose of this code is to query a database and scan the result into the response.Count field. The query is
	// constructed using a printf-style string formatting, as well as a strings.Join() function to combine elements of the maps variable.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from contracts c %s", strings.Join(maps, " "))).Scan(&response.Count)

	// This code checks if the response has a count greater than 0. If it does, then some code is executed. This is used to
	// make sure that the response contains data before the code is executed.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for a paginated request. The GetPage() and GetLimit() methods retrieve the
		// desired page number and the number of records to be displayed, respectively. If the desired page number is greater
		// than 0, the offset is equal to the limit multiplied by the page number minus one. This is to ensure that the first
		// page will display the expected number of records. If the desired page number is 0, then the offset is equal to the
		// limit multiplied by the page number.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query the database. Specifically, it is used to query the contracts table and the chains table,
		// joining the two tables via the chain_id column. The query includes a limit and offset, which are specified in the
		// 'req' object, as well as any additional conditions specified in the 'maps' object. The data returned is stored in
		// the 'rows' object and is then used to construct a response. If an error occurs, it is logged and the response is returned.
		rows, err := e.Context.Db.Query(fmt.Sprintf("select c.id, c.symbol, c.chain_id, c.address, c.fees, c.decimals, c.protocol, n.platform, n.parent_symbol from contracts c inner join chains n on n.id = c.chain_id %s order by c.id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate over the rows of a database query result. The loop will execute once for
		// each row, and the row data will be available for use within the loop.
		for rows.Next() {

			var (
				item pbspot.Contract
			)

			// This code is used to scan a result set from a SQL query and assign the data to the variables in the item struct.
			// To err variable checks if there was an error when scanning the result set and if there was, the error is returned.
			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.ChainId,
				&item.Address,
				&item.Fees,
				&item.Decimals,
				&item.Protocol,
				&item.Platform,
				&item.ParentSymbol,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			// This code checks if the chain ID is valid and if it is, it sets the item's chain name to the chain's name.
			if chain, _ := e.getChain(item.GetChainId(), false); chain.GetId() > 0 {
				item.ChainName = chain.GetName()
			}

			// The purpose of the above statement is to add the item to the end of the existing Fields array in the response
			// object. This is done by using the append function to add the item to the end of the response.Fields array.
			response.Fields = append(response.Fields, &item)
		}

		// The purpose of this code is to check for errors after running a query on a database. If an error is found, it will
		// return an error message and the response variable.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetContractRule - The purpose of this code is to authenticate a user and retrieve their authentication data, as well as to check if the
// user has the correct permissions to modify data in the "contracts" table. It also checks if the ID of the contract
// retrieved from the function e.getContractById is greater than 0 and if it is, then it is appending the row to the
// response.Fields array.
func (e *Service) GetContractRule(ctx context.Context, req *pbspot.GetRequestContractRule) (*pbspot.ResponseContract, error) {

	// The purpose of this code is to create two variables, response and migrate. The response variable is a type of
	// pbspot.ResponseContract, while the migrate variable is a type of query.Migrate that includes the context of e.Context.
	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking to see if the user has the correct permissions to modify data in the "contracts" table. If the
	// user does not have the correct permissions (represented by the query.RoleSpot parameter), then an error is returned.
	if !migrate.Rules(auth, "contracts", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking if the ID of the contract retrieved from the function e.getContractById is greater than 0 and
	// if it is, then it is appending the row to the response.Fields array. This is likely done to ensure that the contract
	// exists before it is added to the response.Fields array.
	if row, _ := e.getContractById(req.GetId()); row.GetId() > 0 {
		response.Fields = append(response.Fields, row)
	}

	return &response, nil
}

// SetContractRule - This code is used to set a contract rule for a user. It first attempts to authenticate the user and retrieve the
// authentication data. It then checks if the user has the correct permissions. It also validates the contract data that
// is being set, such as the symbol, chain ID, address, fees withdraw, protocol, and decimals. If the ID is greater than
// 0, it updates the existing contract rule, otherwise it inserts a new one. Finally, it returns a success message to the caller.
func (e *Service) SetContractRule(ctx context.Context, req *pbspot.SetRequestContractRule) (*pbspot.ResponseContract, error) {

	// The purpose of the code is to create two variables, response and migrate. The response variable is of type
	// pbspot.ResponseContract and the migrate variable is of type query.Migrate. To migrate variable also has an
	// additional field of Context, which is set to the Context field of the e variable.
	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking the authorization of a user based on their role. The first expression uses the migrate.Rules()
	// function to check if the user has permission to write or edit data in the contracts table. If the user does not have
	// permission, the second expression uses migrate.Rules() to check if the user has permission to deny-record in the
	// query.RoleDefault table. If neither of these conditions are true, the code returns an error message indicating that
	// the user does not have permission to write and edit data.
	if !migrate.Rules(auth, "contracts", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This piece of code checks if the length of the "req.Contract.GetSymbol()" is equal to 0. If it is, then it will
	// return an error indicating that a contract or currency symbol is required. This is to ensure that the request has a
	// valid contract or currency symbol before it is processed.
	if len(req.Contract.GetSymbol()) == 0 {
		return &response, e.Context.Error(status.Error(56616, "contract/currency symbol required"))
	}

	// This code checks to see if the address provided for the contract is valid for the given platform. If it is not valid,
	// an error is returned.
	if err := keypair.ValidateCryptoAddress(req.Contract.GetAddress(), req.Contract.GetPlatform()); err != nil {
		return &response, e.Context.Error(err)
	}

	// The purpose of this code is to get the chain specified in the request, using the getChain method from the e object,
	// and save it in a variable called chain. If there is an error, the code will return an error.
	chain, err := e.getChain(req.Contract.GetChainId(), false)
	if err != nil {
		return nil, err
	}

	// This code checks to make sure that the fee of the contract is not less than the fee of the network of the parent. If
	// the fee of the contract is less than the fee of the network, an error message is returned.
	if req.Contract.GetFees() < chain.GetFees() {
		return &response, e.Context.Error(status.Errorf(32798, "the fee of the contract must not be less than the fee of the network of the parent %v face value", chain.GetParentSymbol()))
	}

	// This code is checking to see if the request ID is greater than 0. If the ID is greater than 0, then the code will
	// execute whatever follows the if statement.
	if req.GetId() > 0 {

		// This code is used to update existing contracts in the database. It takes the updated information from the request
		// (req) and assigns it to the corresponding fields in the database. It also checks for any errors and returns the
		// response accordingly.
		if _, err := e.Context.Db.Exec("update contracts set symbol = $1, chain_id = $2, address = $3, fees = $4, protocol = $5, decimals = $6 where id = $7;",
			req.Contract.GetSymbol(),
			req.Contract.GetChainId(),
			req.Contract.GetAddress(),
			req.Contract.GetFees(),
			req.Contract.GetProtocol(),
			req.Contract.GetDecimals(),
			req.GetId(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	} else {

		// This code is used to insert data into a contracts table in a database. The six variables in the parameter list
		// correspond to the columns of the table. The if statement checks for any errors that occur when executing the query
		// and returns an error if one is found.
		if _, err := e.Context.Db.Exec("insert into contracts (symbol, chain_id, address, fees, protocol, decimals) values ($1, $2, $3, $4, $5, $6)",
			req.Contract.GetSymbol(),
			req.Contract.GetChainId(),
			req.Contract.GetAddress(),
			req.Contract.GetFees(),
			req.Contract.GetProtocol(),
			req.Contract.GetDecimals(),
		); err != nil {
			return &response, e.Context.Error(err)
		}

	}
	response.Success = true

	return &response, nil
}

// DeleteContractRule - This code is part of a service which is used to delete a contract rule. The code first authenticates the user to check
// if they have permission to delete a contract rule. If the user is authenticated, the code then checks to see if the
// contract rule exists. If it does, the code then deletes the related data from various tables in the database. Finally,
// a response is returned indicating the success of the operation.
func (e *Service) DeleteContractRule(ctx context.Context, req *pbspot.DeleteRequestContractRule) (*pbspot.ResponseContract, error) {

	// The purpose of this code is to create two variables, response and migrate. The response variable is of type
	// pbspot.ResponseContract, and the migrate variable is of type query.Migrate and has an attribute of Context set to the
	// value of e.Context.
	var (
		response pbspot.ResponseContract
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code checks the user's authentication (auth) to see if they have the appropriate rules ("contracts" and
	// "deny-record") to write and edit data. If they do not have the necessary rules, it returns an error message.
	if !migrate.Rules(auth, "contracts", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is part of a function that deletes a contract from a database. The if statement is used to check whether
	// the contract exists in the database by checking if the ID of the contract is greater than 0. If it is, the code
	// executes a series of SQL queries to delete the contract, its wallets, transactions, and reserves from the database.
	if row, _ := e.getContractById(req.GetId()); row.GetId() > 0 {
		_, _ = e.Context.Db.Exec("delete from contracts where id = $1", row.GetId())
		_, _ = e.Context.Db.Exec("delete from wallets where symbol = $1 and protocol = $2", row.GetSymbol(), row.GetProtocol())
		_, _ = e.Context.Db.Exec("delete from transactions where symbol = $1 and protocol = $2", row.GetSymbol(), row.GetProtocol())
		_, _ = e.Context.Db.Exec("delete from reserves where symbol = $1 and protocol = $2", row.GetSymbol(), row.GetProtocol())
	}
	response.Success = true

	return &response, nil
}

// GetTransactionsRule - This code is part of a function that retrieves transactions from a database. The purpose of the code is to
// authenticate the user, retrieve the authentication data, check the user's permission to access the data, and then
// query the database for the transactions based on the given parameters. It then returns the retrieved data and any
// potential errors to the caller.
func (e *Service) GetTransactionsRule(ctx context.Context, req *pbspot.GetRequestTransactionsManual) (*pbspot.ResponseTransaction, error) {

	// The purpose of this code is to declare three variables. The first variable, response, is of type
	// pbspot.ResponseTransaction. The second variable, migrate, is of type query.Migrate, with a Context field set to the
	// value of e.Context. The third variable, maps, is of type string slice.
	var (
		response pbspot.ResponseTransaction
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	// The purpose of this code is to set a limit on the request if no limit is specified. If req.GetLimit() returns 0, then
	// the req.Limit will be set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking whether a user has the necessary permissions to write and edit data in the "accounts" table. If
	// the user does not have the permissions, an error is returned with a status code and message.
	if !migrate.Rules(auth, "accounts", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This switch statement is used to create an SQL query depending on the transaction type requested. Depending on the
	// request, it will append different strings to the "maps" array, which can then be used in an SQL query. In the default
	// case, it will append a string with both transaction types.
	switch req.GetAssignment() {
	case pbspot.Assignment_DEPOSIT:
		maps = append(maps, fmt.Sprintf("where assignment = %d", pbspot.Assignment_DEPOSIT))
	case pbspot.Assignment_WITHDRAWS:
		maps = append(maps, fmt.Sprintf("where assignment = %d", pbspot.Assignment_WITHDRAWS))
	default:
		maps = append(maps, fmt.Sprintf("where (assignment = %d or assignment = %d)", pbspot.Assignment_WITHDRAWS, pbspot.Assignment_DEPOSIT))
	}

	// This code checks if the request (req) contains a search term (GetSearch()) that is longer than 0. If it does, it
	// appends the search term to a maps slice, with the search term formatted as a string in the SQL 'like' syntax. This
	// allows for a search query to be performed with the search term.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("and (symbol like %[1]s or id::text like %[1]s or hash like %[1]s)", "'%"+req.GetSearch()+"%'"))
	}

	// The purpose of this code is to add a condition to a query. The additional condition is to check the
	// user_id from the request (req.GetId()) and append it to the existing map of conditions (maps).
	maps = append(maps, fmt.Sprintf("and user_id = %v", req.GetId()))

	// This code is used to query a database for the number of transactions and store the result in the response.Count
	// variable. The fmt.Sprintf function is used to construct the query from strings and maps, which are then used to
	// execute the query.
	_ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from transactions %s", strings.Join(maps, " "))).Scan(&response.Count)

	// This if statement is used to check if the GetCount() function returns a value greater than 0. If it does, then the
	// code inside the statement will be executed.
	if response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to retrieve data from a database. It is making a query to the database using the fmt.Sprintf()
		// function, which allows for string formatting. The query is selecting all columns from the transactions table, and
		// ordering them by the 'id' column in descending order. The limit and offset parameters are supplied from the
		// req.GetLimit() and offset variables. If an error occurs when running the query, it will return an error message. The
		// rows.Close() function is being used to close the query and free up any resources used by it.
		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, symbol, hash, value, price, fees, chain_id, confirmation, "to", user_id, assignment, type, platform, protocol, status, create_at from transactions %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for rows.Next() statement is a loop used in SQL to iterate through the result set of a query. It is used to
		// check if there is another row of data available to be read from the result set. If there is another row, then the
		// code inside the loop is executed. If there are no more rows, then the loop ends.
		for rows.Next() {

			var (
				item pbspot.Transaction
			)

			// The purpose of this code is to scan the rows of the database for fields related to a particular item. This code is
			// assigning the values in the database to the fields of the item, such as Id, Symbol, Hash, Value, Price, Fees,
			// Confirmation, To, UserId, TxType, FinType, Platform, Protocol, Status, and CreateAt. If an error occurs
			// while scanning the rows, the code will return an error.
			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.Hash,
				&item.Value,
				&item.Price,
				&item.Fees,
				&item.ChainId,
				&item.Confirmation,
				&item.To,
				&item.UserId,
				&item.Assignment,
				&item.Type,
				&item.Platform,
				&item.Protocol,
				&item.Status,
				&item.CreateAt,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			// This code is part of a function that is attempting to retrieve a chain from a database. The code is defining a
			// local variable, "item.Chain", and attempting to assign the result of a function, "e.getChain(item.GetChainId(),
			// false)", to this variable. If this function returns an error, the function will return a nil (empty chain) and the error.
			item.Chain, err = e.getChain(item.GetChainId(), false)
			if err != nil {
				return nil, err
			}

			// The purpose of this code is to adjust the fees of an item based on the protocol it is using. If the item is not
			// using the mainnet protocol, the fees will be multiplied by the price of the item.
			if item.GetProtocol() != pbspot.Protocol_MAINNET {
				item.Fees = decimal.New(item.GetFees()).Mul(item.GetPrice()).Float()
			}

			// This statement is appending the item to the response.Fields slice. The purpose of this statement is to add the item
			// to the existing list of items stored in the response.Fields slice.
			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}

	}

	return &response, nil
}

// GetReservesRule - This code is part of a GoLang function which is used to retrieve a list of reservations from a database. It is
// responsible for authenticating the user and checking the user's permissions, setting a limit on the request if no
// limit is specified, searching the database for a given keyword, setting an offset for paginated requests, querying the
// database to select certain data, iterating over the query results, and appending items to the response object. It
// returns the response object along with an error, if applicable.
func (e *Service) GetReservesRule(ctx context.Context, req *pbspot.GetRequestReservesManual) (*pbspot.ResponseReserve, error) {

	// The purpose of this code is to declare three variables: response, migrate, and maps. The response variable is of type
	// pbspot.ResponseReserve, to migrate variable is of type query.Migrate, and the maps variable is of type string slice.
	var (
		response pbspot.ResponseReserve
		migrate  = query.Migrate{
			Context: e.Context,
		}
		maps []string
	)

	// The purpose of this code is to set a limit on the request if no limit is specified. If req.GetLimit() returns 0, then
	// the req.Limit will be set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code is checking whether a user has the necessary permissions to write and edit data in the "accounts" table. If
	// the user does not have the permissions, an error is returned with a status code and message.
	if !migrate.Rules(auth, "reserves", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is checking the length of the request's search query. If it is greater than 0, it appends a formatted
	// string to the maps variable. The formatted string is a SQL statement that will search for a given keyword in four
	// different columns of a database (symbol, user_id, address, and symbol). The '%' signs are used as wildcards, so the
	// search query will match any record that contains the keyword in any of these columns.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where (symbol like %[1]s or user_id::text like %[1]s or address like %[1]s or symbol like %[1]s)", "'%"+req.GetSearch()+"%'"))
	}

	// This code is used to check if a query result contains at least one row. The QueryRow function is used to query the
	// database, and the Scan function is used to store the result in the variable response.GetCount(). The if statement
	// checks if the result contains at least one row by comparing the value of response.GetCount() to 0.
	if _ = e.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from reserves %s", strings.Join(maps, " "))).Scan(&response.Count); response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query a database to select certain data. The fmt.Sprintf function is used to build a query
		// string based on the passed in arguments. The strings.Join function is used to create a comma-separated list of items
		// from a slice of strings (maps). The query string is then used to query the database using the Context.Db.Query
		// function. The query result is then stored in the rows variable. The rows.Close function is used to close the
		// database connection when the query is finished.
		rows, err := e.Context.Db.Query(fmt.Sprintf(`select id, symbol, user_id, value, reverse, address, platform, protocol, lock from reserves %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The above code is used in a loop known as a for-range loop. This loop is used to iterate over a range of values, in
		// this case, the rows of a database. With each iteration of the loop, the rows.Next() function is called, which
		// returns a boolean value indicating whether there are more rows to iterate over. If true, the loop will
		// continue, and if false, it will terminate.
		for rows.Next() {

			// The variable "item" is being declared as a type of pbspot.Reserve. This is used to refer to an object that contains
			// information related to a reservation, such as a customer's name, the date of the reservation, and the items they
			// have reserved.
			var (
				item pbspot.Reserve
			)

			// This code is part of a database query. It is scanning the columns of the query's results and assigning each
			// column's value to the appropriate item.field. If an error occurs during the scan, to err variable will not be nil
			// and the error will be logged.
			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.UserId,
				&item.Value,
				&item.Reverse,
				&item.Address,
				&item.Platform,
				&item.Protocol,
				&item.Lock,
			); err != nil {
				return &response, e.Context.Error(err)
			}

			// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
			// item to the Fields array.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// SetReserveUnlockRule - This code is setting up the functions necessary to set a reserve unlock rule in a database. It is checking the user's
// authentication to see if they have the appropriate rules to write and edit data, and if so, it updates the reserves
// table to set the lock to true for the requested ID. If the user does not have the necessary rules, it returns an error
// message. Once the rule is set, it returns a response and no error.
func (e *Service) SetReserveUnlockRule(ctx context.Context, req *pbspot.SetRequestReserveUnlockManual) (*pbspot.ResponseReserve, error) {

	// The code above is setting up two variables, "response" and "migrate". The variable "response" is of type
	// pbspot.ResponseReserve, while the variable "migrate" is of type query.Migrate. The variable "migrate" is being given
	// a value of type query.Migrate, with the context set to the value of the "e" variable. This is likely being used to
	// set up the variables to be used in a function.
	var (
		response pbspot.ResponseReserve
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code checks the user's authentication (auth) to see if they have the appropriate rules ("contracts" and
	// "deny-record") to write and edit data. If they do not have the necessary rules, it returns an error message.
	if !migrate.Rules(auth, "reserves", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if _, err := e.Context.Db.Exec("update reserves set lock = $1 where id = $2;", false, req.GetId()); err != nil {
		return &response, e.Context.Error(err)
	}

	return &response, nil
}

// GetAssetsRule - This code is a function used to retrieve a list of assets from a database. It sets up a limit on the request if no
// limit is specified, authenticates the user, checks their permissions, and retrieves the asset data from the database.
// It also sets up an offset for a paginated request and appends the asset data to the response. It returns the response
// and any errors that occur to the caller.
func (e *Service) GetAssetsRule(ctx context.Context, req *pbspot.GetRequestAssetsManual) (*pbspot.ResponseAsset, error) {

	// The purpose of this code is to define two variables, response and migrate. The first variable, response, is of type
	// pbspot.ResponseAsset, while the second variable, migrate, is of type query.Migrate. The query.Migrate variable is
	// initialized with an e.Context value.
	var (
		response pbspot.ResponseAsset
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// The purpose of this code is to set a limit on the request if no limit is specified. If req.GetLimit() returns 0, then
	// the req.Limit will be set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	if !migrate.Rules(auth, "accounts", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if _ = e.Context.Db.QueryRow("select count(*) as count from assets where user_id = $1", req.GetId()).Scan(&response.Count); response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := e.Context.Db.Query("select id, balance, symbol from assets where user_id = $1 order by id desc limit $2 offset $3", req.GetId(), req.GetLimit(), offset)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for rows.Next() statement is used in SQL queries to loop through the results of a query. It retrieves the next
		// row from the result set, and assigns the values of the row to variables specified in the query. This allows the
		// programmer to iterate through the result set, one row at a time, and process the data as needed.
		for rows.Next() {

			// The purpose of this code is to declare a variable called asset of the type pbspot.Currency. This allows the code to
			// reference this type of asset later in the code.
			var (
				asset pbspot.Currency
			)

			// This code is part of a larger program, and its purpose is to scan the rows in a database for a particular asset and
			// assign the corresponding id, balance, and symbol to the asset variable. If an error occurs at any point during the
			// rows.Scan, the code returns an error response and passes the error to the context.
			if err := rows.Scan(&asset.Id, &asset.Balance, &asset.Symbol); err != nil {
				return &response, e.Context.Error(err)
			}

			// This statement is used to append a field to the response.Fields array. It is used to add a new element to an array.
			// The element being added is the asset variable.
			response.Fields = append(response.Fields, &asset)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// GetRepaymentsRule - This code is part of a function that implements a service for retrieving repayment rules. The purpose of the code is
// to query a database for repayment rules, authenticate the user, and limit the request if no limit is specified. It
// also calculates the offset for paginated requests. Finally, it appends the retrieved data to a response object and
// returns the response.
func (e *Service) GetRepaymentsRule(ctx context.Context, req *pbspot.GetRequestRepaymentsRule) (*pbspot.ResponseRepayment, error) {

	// The code above defines two variables, response and migrate. The response variable is of type
	// pbspot.ResponseRepayment, while the migrate variable is of type query.Migrate. To migrate variable is initialized
	// with the context from the e.Context variable. The purpose of this code is to define two variables for use in the program.
	var (
		response pbspot.ResponseRepayment
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// The purpose of this code is to set a limit on the request if no limit is specified. If req.GetLimit() returns 0, then
	// the req.Limit will be set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// The purpose of this code is to check if the user has the necessary authorization (i.e. migrate.Rules) to perform a
	// certain action (i.e. writing and editing data) related to a specific resource (i.e. repayments). If the user does not
	// have the necessary authorization, an error message is returned.
	if !migrate.Rules(auth, "repayments", query.RoleSpot) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This if statement is used to query a database to check if there are any transactions that meet certain criteria. The
	// criteria are that the allocation must be either INTERNAL or EXTERNAL, the status must be either RESERVE or FILLED,
	// and the protocol must be MAINNET or greater. Additionally, the fees must be greater than 0. If the query returns a
	// count of transactions that meet these criteria, the statement will return true.
	if _ = e.Context.Db.QueryRow("select count(*) as count from transactions where (allocation = $1 and status = $2 and protocol = $3 or allocation = $4 and status = $5 and protocol > $6) and fees > 0", pbspot.Allocation_INTERNAL, pbspot.Status_RESERVE, pbspot.Protocol_MAINNET, pbspot.Allocation_EXTERNAL, pbspot.Status_FILLED, pbspot.Protocol_MAINNET).Scan(&response.Count); response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is responsible for querying a database for a set of records that meet certain conditions. Specifically, it
		// is looking for records from the "transactions" table where the allocation is either 'internal' or 'external', the
		// status is either 'reserve' or 'filled', and the protocol is either 'mainnet' or 'testnet'. It is also limiting the
		// results to records with a fee greater than 0, and ordering the results by the ID field in descending order. Finally,
		// it is limiting the results to a certain number and specifying an offset.
		rows, err := e.Context.Db.Query("select id, value, fees, symbol, chain_id, protocol, platform, status, allocation, repayment, create_at from transactions where (allocation = $1 and status = $2 and protocol = $3 or allocation = $4 and status = $5 and protocol > $6) and fees > 0 order by id desc limit $7 offset $8", pbspot.Allocation_INTERNAL, pbspot.Status_RESERVE, pbspot.Protocol_MAINNET, pbspot.Allocation_EXTERNAL, pbspot.Status_FILLED, pbspot.Protocol_MAINNET, req.GetLimit(), offset)
		if err != nil {
			return &response, e.Context.Error(err)
		}
		defer rows.Close()

		// The for rows.Next() statement is used in SQL queries to loop through the results of a query. It retrieves the next
		// row from the result set, and assigns the values of the row to variables specified in the query. This allows the
		// programmer to iterate through the result set, one row at a time, and process the data as needed.
		for rows.Next() {

			// The purpose of this code is to declare a variable called item of the type pbspot.Repayment. This allows the code to
			// reference this type of item later in the code.
			var (
				item pbspot.Repayment
			)

			// The purpose of this code is to scan the rows of a table and assign the data to variables. The variables are of type
			// item.Id, item.Value, item.Fees, item.Symbol, item.Protocol, item.Platform, item.Status, item.Allocation,
			// item.Repayment, and item.CreateAt. If there is an error, then the function will return the response and an error message.
			if err := rows.Scan(&item.Id, &item.Value, &item.Fees, &item.Symbol, &item.ChainId, &item.Protocol, &item.Platform, &item.Status, &item.Allocation, &item.Repayment, &item.CreateAt); err != nil {
				return &response, e.Context.Error(err)
			}

			// This code is retrieving a chain from the e (environment) object, based on the chain ID in the item object, and then
			// setting the ParentSymbol field of the item object to the chain's ParentSymbol. The purpose is to set the
			// ParentSymbol field of the item object.
			chain, err := e.getChain(item.GetChainId(), false)
			if err != nil {
				return nil, err
			}
			item.ParentSymbol = chain.GetParentSymbol()

			// This statement is used to append a field to the response.Fields array. It is used to add a new element to an array.
			// The element being added is the item variable.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, e.Context.Error(err)
		}
	}

	return &response, nil
}

// SetRepaymentsRule - The purpose of the code above is to set a repayment rule for a transaction made with the pbspot service. The code
// first authenticates the user and checks to see if they have the necessary rules to write and edit data. It then
// queries the database for the fees and chain_id associated with the transaction. The code then checks to see if the
// exchange fees are sufficient to cover the deficit and, if so, updates the fees_charges and fees_costs of the currency
// in the database. Finally, if there is an error, an error response is returned.
func (e *Service) SetRepaymentsRule(ctx context.Context, req *pbspot.SetRequestRepaymentRule) (*pbspot.ResponseRepayment, error) {

	// The code above defines two variables, response and migrate. The response variable is of type
	// pbspot.ResponseRepayment, while the migrate variable is of type query.Migrate. To migrate variable is initialized
	// with the context from the e.Context variable. The purpose of this code is to define two variables for use in the program.
	var (
		response pbspot.ResponseRepayment
		migrate  = query.Migrate{
			Context: e.Context,
		}
	)

	// This code is part of an authentication process. The purpose of this code is to attempt to authenticate the user and
	// retrieve the authentication data. If there is an error, it is returned to the caller.
	auth, err := e.Context.Auth(ctx)
	if err != nil {
		return &response, e.Context.Error(err)
	}

	// This code checks the user's authentication (auth) to see if they have the appropriate rules ("contracts" and
	// "deny-record") to write and edit data. If they do not have the necessary rules, it returns an error message.
	if !migrate.Rules(auth, "repayments", query.RoleSpot) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, e.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	// This code is used to query a database for information. Specifically, it is querying for the fees and chain_id
	// associated with a transaction with an id equal to the id stored in the req variable. The row variable is used to
	// store the result of the query. The if statement is used to check for any errors that may have occurred during the
	// query. To defer row.Close() statement is used to close the query after it is finished, to avoid any memory leaks.
	row, err := e.Context.Db.Query(`select fees, chain_id from transactions where id = $2`, req.GetId())
	if err != nil {
		return &response, e.Context.Error(err)
	}
	defer row.Close()

	// The if statement is used to test a condition and execute a block of code if the condition is true. In this case, the
	// if statement is used to check if the row has a next value. If there is a next value, then the code in the block will be executed.
	if row.Next() {

		// The line of code above declares a variable, item, of type pbspot.Transaction. This is a type of data structure that
		// is used to store information about a transaction made using the pbspot service, such as the amount, date, and other details.
		var (
			item pbspot.Transaction
		)

		// The purpose of this if statement is to scan each row of the database table and store the values from the row into
		// the item.Fees and item.ChainId variables. If the scan is unsuccessful, the error is returned in the response.
		if err := row.Scan(&item.Fees, &item.ChainId); err != nil {
			return &response, e.Context.Error(err)
		}

		// The purpose of this code is to get the chain specified in the request, using the getChain method from the e object,
		// and save it in a variable called chain. If there is an error, the code will return an error.
		chain, err := e.getChain(item.GetChainId(), false)
		if err != nil {
			return nil, err
		}

		// This code checks if the exchange fees are sufficient to cover the deficit. If not, it returns an error.
		if _ = e.Context.Db.QueryRow("select fees_charges from currencies where symbol = $1", chain.GetParentSymbol()).Scan(&item.Value); item.GetFees() > item.GetValue() {
			return &response, e.Context.Error(status.Error(521233, "exchange fees are insufficient to cover the deficit"))
		}

		// This code is updating the fees_charges and fees_costs of a currency in a database using the Exec() method. The code
		// is using $1 and $2, which are placeholder variables for the first and second parameters given to the Exec() method.
		// The first parameter is the symbol of the parent chain and the second parameter is the fees associated with the item. If an error occurs, the code will return an error response.
		if _, err := e.Context.Db.Exec("update currencies set fees_charges = fees_charges - $2, fees_costs = fees_costs + $2 where symbol = $1;", chain.GetParentSymbol(), item.GetFees()); err != nil {
			return &response, e.Context.Error(err)
		}
		response.Success = true

		return &response, nil
	}

	return &response, e.Context.Error(status.Error(865456, "no such transaction exists"))
}
