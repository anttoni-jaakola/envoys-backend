package ohlcv

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbohlcv"
	"github.com/cryptogateway/backend-envoys/server/types"
	"google.golang.org/grpc/status"
	"strings"
)

// GetTicker - The purpose of this code is to create a service that retrieves OHLC (open-high-low-close) data from a database and
// returns it in a response. It is used to set limits on the number of results returned, filter the results based on a
// time range, perform calculations on the data, and store the results in an array.
func (s *Service) GetTicker(_ context.Context, req *pbohlcv.GetRequest) (*pbohlcv.Response, error) {

	// The purpose of this code is to create three variables with zero values: response, limit and maps. The response
	// variable is of type pbohlcv.Response, the limit variable is of type string, and the maps variable is of type
	// slice of strings.
	var (
		response pbohlcv.Response
		limit    string
		maps     []string
	)

	// This code checks if the limit of the request is set to 0. If it is, then it sets the limit to 30. This is likely done
	// so that a request has a sensible limit, even if one wasn't specified.
	if req.GetLimit() == 0 {
		req.Limit = 500
	}

	// This code is used to set a limit to the request. It checks if req.GetLimit() is greater than 0. If so, it sets the
	// limit variable to a string with the limit set to that amount. This is likely used to set a limit on the amount of
	// data that will be returned in the response.
	if req.GetLimit() > 0 {
		limit = fmt.Sprintf("limit %d", req.GetLimit())
	}

	// This code is checking to see if the "From" and "To" values in the request are greater than 0. If they are, a
	// formatted string will be appended to the "maps" array containing a timestamp that is less than the "To" value in the
	// request. This code is likely used to filter a query based on a time range.
	if req.GetTo() > 0 {
		maps = append(maps, fmt.Sprintf(`and to_char(o.create_at::timestamp, 'yyyy-mm-dd hh24:mi:ss') < to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss')`, req.GetTo()))
	}

	// This code is used to query the database to return OHLC (open-high-low-close) data. The SQL query is using the
	// fmt.Sprintf function to substitute the variables (req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "),
	// help.Resolution(req.GetResolution()), limit) into the query. The query is then executed, and the results are stored
	// in the rows variable. Finally, the rows variable is closed at the end of the code.
	rows, err := s.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', o.create_at))::integer buckettime, first(o.price, o.create_at) as open, last(o.price, o.create_at) as close, first(o.price, o.price) as low, last(o.price, o.price) as high, sum(o.quantity) as volume, avg(o.price) as avg_price, o.base_unit, o.quote_unit from ohlcv as o where o.base_unit = '%[1]s' and o.quote_unit = '%[2]s' %[3]s group by buckettime, o.base_unit, o.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The purpose of the for rows.Next() loop is to iterate through the rows in a database table. It is used to perform
	// some action on each row of the table. This could include retrieving data from the row, updating data in the row, or
	// deleting the row.
	for rows.Next() {

		// The purpose of the variable "item" is to store data of type types.Ticker. This could be used to store an array of
		// candles or other data related to types.Ticker.
		var (
			item types.Ticker
		)

		// This code is checking for errors while scanning a row of data from a database. It is assigning the values of the row
		// to the variables item.Time, item.Open, item.Close, item.Low, item.High, item.Volume, item.Price, item.BaseUnit, and
		// item.QuoteUnit. If an error occurs during the scan, the code will return an error response.
		if err = rows.Scan(&item.Time, &item.Open, &item.Close, &item.Low, &item.High, &item.Volume, &item.Price, &item.BaseUnit, &item.QuoteUnit); err != nil {
			return &response, err
		}

		// This code is likely appending an item to a response.Fields array. It is likely used to add an item to the array and
		// modify the array.
		response.Fields = append(response.Fields, &item)
	}

	// The purpose of the following code is to declare a variable called stats of the type pbohlcv.Stats. This variable will
	// be used to store information related to the pbohlcv.Stats data type.
	var (
		stats types.Stats
	)

	// This code is used to fetch and analyze data from a database. It uses the QueryRow() method to retrieve data from the
	// database and then scan it into the stats variable. The code is specifically used to get the count, volume, low, high,
	// first and last values from the trades table for a given base unit and quote unit.
	_ = s.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from ohlcv as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

	// This code checks if the length of the 'response.Fields' array is greater than 1. If so, it assigns the 'Close' value
	// of the second element in the 'response.Fields' array to the 'Previous' field of the 'stats' object.
	if len(response.Fields) > 1 {
		stats.Previous = response.Fields[1].Close
	}

	//The purpose of this statement is to assign the pointer stats to the Stats field of the response object. This allows
	//the response object to access the data stored in the stats variable.
	response.Stats = &stats

	return &response, nil
}

// SetTicker - The purpose of this code is to retrieve two candles with a given resolution from a spot exchange, add a new row to a
// database table, publish a message to an exchange on a specific topic, and append the returned values to a response array.
func (s *Service) SetTicker(_ context.Context, req *pbohlcv.SetRequest) (*pbohlcv.Response, error) {

	// The purpose of this code is to declare a variable called 'response' of type 'pbohlcv.Response'. This variable will be
	// used to store data that is returned from a function call involving the 'pbohlcv' library.
	var (
		response pbohlcv.Response
	)

	// This code is checking if the key provided in the request (req.GetKey()) matches the secret stored in the context
	// (s.Context.Secrets[2]). If the keys don't match, the code is returning an error with the code 654333 and the message
	// "the access key is incorrect". This is likely part of an authorization process to make sure only authorized users are
	// able to access a certain resource.
	if req.GetKey() != s.Context.Secrets[2] {
		return &response, status.Error(654333, "the access key is incorrect")
	}

	// This piece of code is inserting data into a database table. The purpose of this code is to add a new row to the
	// "ohlcv" table, based on the values stored in the params array. The five columns in the table are assigning,
	// base_unit, quote_unit, price, and quantity, and each of these is being populated with the corresponding value from
	// the params array. The code then checks for any errors that may have occurred while executing the query and returns if any are found.
	if _, err := s.Context.Db.Exec(`insert into ohlcv (assigning, base_unit, quote_unit, price, quantity) values ($1, $2, $3, $4, $5)`, req.GetAssigning(), req.GetBaseUnit(), req.GetQuoteUnit(), req.GetPrice(), req.GetValue()); s.Context.Debug(err) {
		return &response, err
	}

	// The for loop is used to iterate through each element in the Depth() array. The underscore is used to assign the index
	// number to a variable that is not used in the loop. The interval variable is used to access the contents of each
	// element in the Depth() array.
	for _, interval := range help.Depth() {

		// This code is used to retrieve two candles with a given resolution from a spot exchange. The purpose of the migrate,
		// err := e.GetTicker() line is to make a request to the spot exchange using the BaseUnit, QuoteUnit, Limit, and
		// Resolution parameters provided. The if err != nil { return err } line is used to check if there was an error with
		// the request and return that error if necessary.
		migrate, err := s.GetTicker(context.Background(), &pbohlcv.GetRequest{BaseUnit: req.GetBaseUnit(), QuoteUnit: req.GetQuoteUnit(), Limit: 2, Resolution: interval})
		if err != nil {
			return &response, err
		}

		// This code is used to publish a message to an exchange on a specific topic. The message is "migrate" and the topic is
		// "trade/ticker:interval". The purpose of this code is to send a message to the exchange,
		// action based on the message. The if statement is used to check for any errors that may occur during the publishing
		// of the message. If an error is encountered, it will be returned.
		if err := s.Context.Publish(migrate, "exchange", fmt.Sprintf("trade/ticker:%v", interval)); err != nil {
			return &response, err
		}

		// The purpose of this statement is to append the values of the migrate.Fields array to the response.Fields array. This
		// statement essentially adds the values of migrate.Fields array to the existing values in response.Fields array.
		response.Fields = append(response.Fields, migrate.Fields...)
	}

	return &response, nil
}
