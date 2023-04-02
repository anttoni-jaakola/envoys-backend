package ads

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbads"
	"github.com/cryptogateway/backend-envoys/server/types"
)

// GetAdvertisements - This function is used to retrieve a list of advertisements from a database based on the specified parameters. It will
// return a response containing the total number of advertisements that match the type specified in the request, and a
// list of advertisement objects which contain the id, title, text, link, and type for each advertisement. It will also
// limit the number of advertisements returned based on the specified limit and page number. If the random flag is set, the advertisements will be sorted in a random order.
func (s *Service) GetAdvertisements(_ context.Context, req *pbads.GetRequestAdvertisements) (*pbads.ResponseAdvertising, error) {

	// The line of code is declaring two variables, response and by.  The variable response is of type
	// pbads.ResponseAdvertising, which is likely a custom type from a library. The variable by is of type string, which is
	// a built-in type in most languages. This code is declaring two variables, which will be used later in the program.
	var (
		response pbads.ResponseAdvertising
		by       string
	)

	// The purpose of this code is to set a default value of 30 for the req.Limit variable if the value of req.GetLimit() is
	// 0. This way, if the req.GetLimit() is 0, the req.Limit variable will be set to 30, ensuring that it will not be left uninitialized.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is checking if there are any records in the advertising table with the type specified in the request. If
	// there are records with the specified type, the response.Count will be set to the number of records, and the if
	// statement will evaluate to true.
	if _ = s.Context.Db.QueryRow("select count(*) as count from advertising where pattern = $1", req.GetPattern()).Scan(&response.Count); response.GetCount() > 0 {

		if req.GetRandom() {
			by = "random()"
		} else {
			by = "id"
		}

		// This code is calculating the offset to be used when paginating a set of data. The offset is calculated by
		// multiplying the limit (number of items to be returned) by the page number. If the page number is greater than 0, the
		// offset is calculated by multiplying the limit by page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query a database. The fmt.Sprintf statement is used to create a SQL query string based on the
		// request parameters (req) passed to it. The query is then executed using the Db.Query method, and the results are
		// returned in the form of a rows object. The rows.Close() method is used to close the rows object, and is important in
		// order to ensure all resources associated with the operation are properly released.
		rows, err := s.Context.Db.Query(fmt.Sprintf("select id, title, text, link, pattern from advertising where pattern = '%v' order by %v desc limit %d offset %d", req.GetPattern(), by, req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop with the rows.Next() statement is used to iterate over the result set of a query. It is used to access
		// each row in the result set one at a time. This allows the user to access and process individual records from the database.
		for rows.Next() {

			// The purpose of the var statement is to declare a new variable called "item" of type "pbads.Advertising". This
			// variable can then be used to store a reference to an Advertising object that can be used for various operations within the program.
			var (
				item types.Advertising
			)

			// This code is part of a function that is used to retrieve an item from a database. The purpose of this specific code
			// is to assign the values retrieved from the database to the various fields of the item object. The if statement is
			// used to check for any errors while attempting to scan the database. If an error occurs, the function returns an error response.
			if err = rows.Scan(
				&item.Id,
				&item.Title,
				&item.Text,
				&item.Link,
				&item.Pattern,
			); err != nil {
				return &response, err
			}

			// This code adds an item to the Fields slice in the response struct. The purpose of adding an item is to store or
			// keep track of the data in the response struct.
			response.Fields = append(response.Fields, &item)
		}

		// This code is testing for any errors that occur when retrieving rows from a database. If an error is found, the code
		// returns an error response and an error object. This helps to ensure that any errors are handled properly and the
		// application is able to continue functioning.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// GetAdvertising - This function is part of a service used to retrieve information about an advertising item from a database. It takes a
// context and a request object containing an ID as parameters. It then queries the database for the advertising item
// with the given ID and scans the results into a struct. This struct is then appended to a response struct which is returned along with a nil error.
func (s *Service) GetAdvertising(_ context.Context, req *pbads.GetRequestAdvertising) (*pbads.ResponseAdvertising, error) {

	// The purpose of the following statement is to declare two variables, response and item, which both have the type
	// pbads.ResponseAdvertising and types.Advertising respectively. These variables will be used in the code that follows.
	var (
		response pbads.ResponseAdvertising
		item     types.Advertising
	)

	// This code is querying a database for a specific row of information associated with a given ID. The "_" is a blank
	// identifier, which is used to discard the row value since it is not needed. The Scan method is taking the five data
	// points (id, title, text, link, pattern) associated with the given ID and assigning them to the item object. This allows
	// the item object to store the data points associated with the given ID.
	_ = s.Context.Db.QueryRow("select id, title, text, link, pattern from advertising where id = $1", req.GetId()).Scan(&item.Id, &item.Title, &item.Text, &item.Link, &item.Pattern)

	// This statement appends the item to the response.Fields slice. The purpose of this statement is to add the item to the
	// existing list of items in the response.Fields slice.
	response.Fields = append(response.Fields, &item)

	return &response, nil
}
