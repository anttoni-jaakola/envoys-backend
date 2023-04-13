package stock

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbstock"
	"github.com/cryptogateway/backend-envoys/server/service/v2/account"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"strings"
)

// SetAgent - The purpose of this code is to create an agent or broker account in a database. It does this by taking in a request
// for a new agent, validating the authentication token in the context, and inserting the new agent into the database. It
// then returns a response and any errors that occurred during the process.
func (s *Service) SetAgent(ctx context.Context, req *pbstock.SetRequestAgent) (*pbstock.ResponseAgent, error) {

	var (
		response pbstock.ResponseAgent
		agent    pbstock.Agent
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to query a database for a particular user's secure information and store it in the
	// variable agent.Success. If the query returns an error, the error is returned in response and the context is sent an error message.
	if err := s.Context.Db.QueryRow(`select secure from kyc where user_id = $1`, auth).Scan(&agent.Success); err != nil {
		return &response, err
	}

	spew.Dump(req)

	// This code checks to see if an agent has been verified. If the agent has not been verified, it will return an error
	// message to indicate that the agent has not been KYC verified.
	if !agent.Success {
		return &response, status.Error(53678, "you have not been verified KYC")
	}

	// The purpose of this switch statement is to assign a value to the "_status" variable based on the type of request
	// (req.GetType()) that is received. If the request type is "AGENT", then the "_status" variable is set to "PENDING". If
	// the request type is "BROKER", then the "_status" variable is set to "ACCESS".
	switch req.GetType() {
	case types.UserTypeAgent:
		agent.Status = types.StatusPending
		agent.Type = types.UserTypeAgent
	case types.UserTypeBroker:
		agent.Status = types.StatusAccess
		agent.Type = types.UserTypeBroker
	default:
		return &response, status.Error(678543, "not found type")
	}

	// This code is used to insert data into the 'agents' table in a database. The code is also checking for any errors that
	// may occur during the insertion process, and if an error occurs, it will return an error message.
	if _, err := s.Context.Db.Exec(`insert into agents (name, broker_id, type, status, user_id) values ($1, $2, $3, $4, $5)`,
		req.GetName(),
		req.GetBrokerId(),
		agent.GetType(),
		agent.GetStatus(),
		auth,
	); err != nil {
		return &response, status.Error(646788, "you have already created an agent and broker account")
	}

	// This code is checking to see if an Agent object was returned when calling the queryAgent() function with the auth
	// parameter. If the Id property of the Agent object is greater than 0, then the Agent object is appended to the
	// response.Fields array.
	if agent, _ := s.queryAgent(auth); agent.Id > 0 {

		// This code checks for any errors when publishing the message to the exchange, and if there is an error, it will return
		// an error response and log the error.
		if err := s.Context.Publish(&agent, "exchange", "create/agent"); err != nil {
			return &response, err
		}

		response.Fields = append(response.Fields, agent)
	}

	return &response, nil
}

// GetAgent - The purpose of the above code is to get an agent from the context, validate the authentication token, and then append
// the agent to the response. This allows the code to securely access the agent information and return it to the caller.
func (s *Service) GetAgent(ctx context.Context, _ *pbstock.GetRequestAgent) (*pbstock.ResponseAgent, error) {

	// The purpose of the above code is to declare a variable named 'response' of type 'pbstock.ResponseAgent'. This allows
	// the code to store a value of type 'pbstock.ResponseAgent' in the 'response' variable.
	var (
		response pbstock.ResponseAgent
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is checking to see if an Agent object was returned when calling the queryAgent() function with the auth
	// parameter. If the Id property of the Agent object is greater than 0, then the Agent object is appended to the
	// response.Fields array.
	if agent, _ := s.queryAgent(auth); agent.Id > 0 {
		response.Fields = append(response.Fields, agent)
	}

	return &response, nil
}

// GetBrokers - The purpose of the above code is to query a database for brokers and construct a response object containing the
// brokers that match the criteria specified in the request. It is also responsible for setting a limit on the amount of
// data that is retrieved from the database, as well as calculating the offset for a paginated request. Finally, it
// checks for errors with the rows object and returns the response object along with an error if there is an error.
func (s *Service) GetBrokers(_ context.Context, req *pbstock.GetRequestBrokers) (*pbstock.ResponseBroker, error) {

	// The purpose of the above code is to declare a variable named 'response' of type 'pbstock.ResponseBroker'. This allows
	// the code to store a value of type 'pbstock.ResponseBroker' in the 'response' variable.
	var (
		response pbstock.ResponseBroker
		maps     []string
	)

	// This code checks if the request's limit is 0. If it is, it sets the request's limit to 30. This is likely done to
	// ensure that the request is not given an unlimited amount of data, which could cause performance issues.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code snippet is checking the length of the "req.GetSearch()" variable. If the length is greater than 0, then it
	// appends a string to the "maps" variable that performs a search for a broker with a given name or ID. If the length of
	// "req.GetSearch()" is 0, then it appends a string to the "maps" variable that searches for a broker without any search parameters.
	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where type = '%[2]v' and (name like %[1]s or id::text like %[1]s)", "'%"+req.GetSearch()+"%'", types.UserTypeBroker))
	} else {
		maps = append(maps, fmt.Sprintf("where type = '%[1]v'", types.UserTypeBroker))
	}

	// The purpose of this code is to query a database for the number of records in a table that match certain criteria
	// indicated by the 'maps' variable. It then scans the result into a variable called 'response' and checks if the count
	// is greater than 0. If it is, it will execute some code.
	if _ = s.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from agents %s", strings.Join(maps, " "))).Scan(&response.Count); response.GetCount() > 0 {

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
		rows, err := s.Context.Db.Query(fmt.Sprintf(`select id, name, user_id, broker_id, type, status from agents %s order by id desc limit %d offset %d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The above code is used in a loop known as a for-range loop. This loop is used to iterate over a range of values, in
		// this case, the rows of a database. With each iteration of the loop, the rows.Next() function is called, which
		// returns a boolean value indicating whether there are more rows to iterate over. If true, the loop will
		// continue, and if false, it will terminate.
		for rows.Next() {

			var (
				item pbstock.Agent
			)

			// This code is part of a database query. It is scanning the columns of the query's results and assigning each
			// column's value to the appropriate item.field. If an error occurs during the scan, to err variable will not be nil
			// and the error will be logged.
			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.UserId,
				&item.BrokerId,
				&item.Type,
				&item.Status,
			); err != nil {
				return &response, err
			}

			// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
			// item to the Fields array.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// GetRequests - The purpose of this code is to get requests from an agent and return them in the form of a pbstock.ResponseAgent
// object. The code begins by declaring a response variable of type pbstock.ResponseAgent. It then checks if the limit in
// the request is 0, and if it is, sets it to 30. It then checks to make sure the authentication token is valid and gets
// the agent using the authentication credentials. It then queries the database to check if the count of agents with a
// certain status, type, and broker ID is greater than 0. If the count is greater than 0, the code calculates an offset,
// queries the database for the data, stores the results in a row object, and iterates over the rows. In each iteration,
// it creates an item of type pbstock.Agent, scans the columns and assigns the values to the item's fields, queries a
// user using the authentication credentials, and appends the item to the response object's fields array. Finally, it
// checks if there is an error with the rows object and returns the response object along with an error.
func (s *Service) GetRequests(ctx context.Context, req *pbstock.GetRequestRequests) (*pbstock.ResponseAgent, error) {

	// The purpose of this code is to declare a variable named response of type pbstock.ResponseAgent. This variable will be
	// used to store the response from an agent when making a request.
	var (
		response pbstock.ResponseAgent
	)

	// This code checks if the request's limit is 0. If it is, it sets the request's limit to 30. This is likely done to
	// ensure that the request is not given an unlimited amount of data, which could cause performance issues.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, _ := s.queryAgent(auth)

	// The purpose of this code is to query a database and check if the count of agents with a certain status, type, and
	// broker ID is greater than 0. If the count is greater than 0, then the code will execute the following code block.
	if _ = s.Context.Db.QueryRow("select count(*) as count from agents where status = $1 and type = $2 and broker_id = $3", types.StatusPending, types.UserTypeAgent, agent.GetId()).Scan(&response.Count); response.GetCount() > 0 {

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
		rows, err := s.Context.Db.Query(`select a.id, a.user_id, a.broker_id, a.type, a.status, a.create_at, b.secret from agents a left join kyc b on b.user_id = a.user_id  where a.status = $1 and a.type = $2 and a.broker_id = $3 order by a.id desc limit $4 offset $5`, types.StatusPending, types.UserTypeAgent, agent.GetId(), req.GetLimit(), offset)
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The above code is used in a loop known as a for-range loop. This loop is used to iterate over a range of values, in
		// this case, the rows of a database. With each iteration of the loop, the rows.Next() function is called, which
		// returns a boolean value indicating whether there are more rows to iterate over. If true, the loop will
		// continue, and if false, it will terminate.
		for rows.Next() {

			// This variable declaration is creating a variable named item of type pbstock.Agent. The purpose of this variable is
			// to store a value of the Agent type, which is a type defined in the pbstock package.
			var (
				item pbstock.Agent
			)

			// This code is part of a database query. It is scanning the columns of the query's results and assigning each
			// column's value to the appropriate item.field. If an error occurs during the scan, to err variable will not be nil
			// and the error will be logged.
			if err = rows.Scan(
				&item.Id,
				&item.UserId,
				&item.BrokerId,
				&item.Type,
				&item.Status,
				&item.CreateAt,
				&item.Applicant,
			); err != nil {
				return &response, err
			}

			// The purpose of this code is to create a Service object that uses the context stored in the variable e. The Service
			// object is then assigned to the variable migrate.
			migrate := account.Service{
				Context: s.Context,
			}

			// This code is attempting to query a user from migrate using the provided authentication credentials (auth). If the
			// query fails, an error is returned.
			user, err := migrate.QueryUser(item.GetUserId())
			if err != nil {
				return &response, err
			}
			item.Name = user.GetName()
			item.Email = user.GetEmail()

			// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
			// item to the Fields array.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// SetSetting - This code is used to set the request settings for a stock service. It authenticates the user, retrieves the agent
// associated with the user, and then updates the status of the agent in the database. It also returns an error response
// if there is an issue with the user authentication or execution of the SQL statement.
func (s *Service) SetSetting(ctx context.Context, req *pbstock.GetRequestSetting) (*pbstock.ResponseSetting, error) {

	// The purpose of this code is to declare a variable called "response" of the type pbstock.ResponseSetting. This is a
	// variable that can be used to store data from the pbstock.ResponseSetting type.
	var (
		response pbstock.ResponseSetting
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	// This code is checking to see if the agent's ID is greater than 0, and if it is, then it is executing an SQL statement
	// to update the status of the agent in the database. The statement uses parameters for the ID, broker ID, and status of
	// the agent. If there is an error, it returns an error response.
	if agent.GetId() > 0 {

		var (
			item pbstock.Agent
		)

		// This if-statement is used to update the status of an agent in a database. It is using the QueryRow method to update
		// the status of the agent with the given user_id and broker_id. The Scan method is then used to assign each of the
		// returned values to the respective item fields. If an error occurs, the Error method is called and the response is returned.
		if err := s.Context.Db.QueryRow("update agents set status = $3 where user_id = $1 and broker_id = $2 returning id, user_id, broker_id, status;", req.GetUserId(), agent.GetId(), req.GetStatus()).Scan(&item.Id, &item.UserId, &item.BrokerId, &item.Status); err != nil {
			return &response, err
		}

		// This code checks for any errors when publishing the message to the exchange, and if there is an error, it will return
		// an error response and log the error.
		if err := s.Context.Publish(&item, "exchange", "status/agent"); err != nil {
			return &response, err
		}

		response.Success = true
	}

	return &response, nil
}

// DeleteAgent - The purpose of this code is to delete an agent from a database based on the ID and authentication credentials
// provided. It first checks to make sure a valid authentication token is present in the context, then retrieves the
// requested agent, and finally deletes the agent from the database using the given parameters. It also returns a
// response to the user if there was an error in the process.
func (s *Service) DeleteAgent(ctx context.Context, req *pbstock.GetRequestDeleteAgent) (*pbstock.ResponseAgent, error) {

	// The purpose of this code is to declare a variable named response of type pbstock.ResponseAgent. This variable will be
	// used to store the response from an agent when making a request.
	var (
		response pbstock.ResponseAgent
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	// This code is used to delete an agent in a database based on the ID and user_id provided. It checks that the ID is
	// greater than 0 to make sure that it is a valid agent and then proceeds to delete the agent from the database using
	// the given parameters.
	if agent.GetId() > 0 {
		_, _ = s.Context.Db.Exec("delete from agents where id = $1 and user_id = $2", req.GetId(), auth)
	}

	return &response, nil
}

// GetAgents - The purpose of this code is to query a database for information about agents and return a response containing the
// requested data. It sets a default limit value, checks for a valid authentication token, gets an agent using the given
// authentication credentials, and filters the results based on a search term if one is provided. It also calculates an
// offset for paginated requests and queries the database using the provided parameters. Finally, it returns the response and any errors that occurred.
func (s *Service) GetAgents(ctx context.Context, req *pbstock.GetRequestAgents) (*pbstock.ResponseAgent, error) {

	// The purpose of this code is to declare a variable named response of type pbstock.ResponseAgent. This variable will be
	// used to store the response from an agent when making a request.
	var (
		response pbstock.ResponseAgent
	)

	// The purpose of this code is to set a default limit value if the limit value requested (req.GetLimit()) is equal to
	// zero. In this case, the default limit value is set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	// This if statement is used to query an SQL database to check if a certain agent is associated with any accounts. The
	// query is using the ID of the agent and the type of account, and is checking if there is a count of any results
	// greater than 0. If the count is greater than 0, the code within the if statement will execute.
	if _ = s.Context.Db.QueryRow("select count(*) as count from agents a left join accounts b on b.id = a.user_id where a.broker_id = $1 and a.type = $2 and a.status = $3", agent.GetId(), types.UserTypeAgent, types.StatusAccess).Scan(&response.Count); response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query the database for information from the tables agents and accounts. The query is looking
		// for information on agents, with the broker_id, type, limit, and offset specified by the parameters agent.GetId(),
		// types.UserTypeAgent, req.GetLimit(), offset. The query will return the id, name, email, user_id, broker_id, type,
		// status, and create_at associated with the specified parameters. The result of the query will be stored in the rows
		// variable, and the query will be closed when the defer rows.Close() line is executed.
		rows, err := s.Context.Db.Query(`select a.id, b.name, b.email, a.user_id, a.broker_id, a.type, a.status, a.create_at from agents a inner join accounts b on b.id = a.user_id where a.broker_id = $1 and a.type = $2 and a.status = $3 order by a.id desc limit $4 offset $5`, agent.GetId(), types.UserTypeAgent, types.StatusAccess, req.GetLimit(), offset)
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The above code is used in a loop known as a for-range loop. This loop is used to iterate over a range of values, in
		// this case, the rows of a database. With each iteration of the loop, the rows.Next() function is called, which
		// returns a boolean value indicating whether there are more rows to iterate over. If true, the loop will
		// continue, and if false, it will terminate.
		for rows.Next() {

			var (
				item pbstock.Agent
			)

			// This code is part of a database query. It is scanning the columns of the query's results and assigning each
			// column's value to the appropriate item.field. If an error occurs during the scan, to err variable will not be nil
			// and the error will be logged.
			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.Email,
				&item.UserId,
				&item.BrokerId,
				&item.Type,
				&item.Status,
				&item.CreateAt,
			); err != nil {
				return &response, err
			}

			// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
			// item to the Fields array.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// SetBlocked - This code snippet is used to set the blocked status of an agent in a database. It checks for a valid authentication
// token in the context, gets the agent using the credentials, queries the database for the status of the agent from the
// given id, broker_id, and type, and then updates the status in the database accordingly.
func (s *Service) SetBlocked(ctx context.Context, req *pbstock.SetRequestAgentBlocked) (*pbstock.ResponseBlocked, error) {

	// The purpose of this code is to declare two variables, response and status, of type pbstock.ResponseBlocked and
	// types.Status, respectively. These variables can then be used to store values of that type.
	var (
		response pbstock.ResponseBlocked
		_status  string
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	// This code is used to query a database for information about a specific agent. The variables req, agent, and pbstock
	// are passed in to this function as parameters. The code queries the database for the status of the agent from the
	// given id, broker_id, and type, and stores it in the variable "status". If there is an error with the query, the
	// function returns an error.
	if err = s.Context.Db.QueryRow("select status as count from agents where id = $1 and broker_id = $2 and type = $3", req.GetId(), agent.GetId(), types.UserTypeAgent).Scan(&_status); err != nil {
		return &response, err
	}

	// This code is used to update the status of an agent when given a request. Depending on the initial _status, the code
	// will either block the agent or give them access.
	switch _status {
	case types.StatusBlocked:
		_, _ = s.Context.Db.Exec("update agents set status = $2 where id = $1", req.GetId(), types.StatusAccess)
		response.Success = true
	case types.StatusAccess:
		_, _ = s.Context.Db.Exec("update agents set status = $2 where id = $1", req.GetId(), types.StatusBlocked)
	}

	return &response, nil
}

// SetTransfer - The purpose of this code is to create a service that allows users to withdraw assets from their account. The code
// checks for valid authentication, retrieves an agent using the given authentication credentials, checks the status of
// the agent and returns an error if it is blocked, gets the value of an item from the database, checks to make sure
// there are enough funds to withdraw the requested amount, updates the user's balance, and inserts a new record into the
// withdrawals table in the database. Finally, the code returns a response indicating whether the withdrawal was successful.
func (s *Service) SetTransfer(ctx context.Context, req *pbstock.SetRequestTransfer) (*pbstock.ResponseTransfer, error) {

	// The purpose of the following code is to declare two variables, response and item, of type pbstock.ResponseTransfer
	// and pbstock.Transfer respectively. These variables are used to store data related to a ResponseTransfer and a
	// withdrawal respectively, which are both structs from the pbstock package.
	var (
		response pbstock.ResponseTransfer
		item     pbstock.Transfer
	)

	// This code checks if the quantity requested by the requester is equal to 0. If it is, it returns an error code
	// (844532) and a message ("value must not be null") to the requester. This code is useful for validating user input to
	// make sure it meets certain criteria.
	if req.GetQuantity() == 0 {
		return &response, status.Error(844532, "value must not be null")
	}

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to check the status of the agent and if it is "BLOCKED", then it will return an error
	// message with a status code of 523217. This can help the application determine if the agent should be allowed to
	// continue with its current task.
	if agent.GetStatus() == types.StatusBlocked {
		return &response, status.Error(523217, "your asset blocked")
	}

	// This code is checking the value of the agent's broker ID. If it is set to 0, then the item's status is set to FILLED,
	// and the item's ID is set to the agent's ID. If the agent's broker ID is not set to 0, then the item's status is set
	// to PENDING, and the item's ID is set to the agent's broker ID.
	if agent.GetBrokerId() == 0 {
		item.Status = types.StatusFilled
		item.Id = agent.GetId()
	} else {
		item.Status = types.StatusPending
		item.Id = agent.GetBrokerId()
	}

	// This code is querying a database for records that match certain criteria. The Query function takes a SQL statement
	// and two arguments, req.GetSymbol() and proto.Type_STOCK, which represent the criteria for the query. The results of
	// the query are stored in a "row" object and can be accessed using the "row.Close()" function. If an error occurs while
	// executing the query, the "err" variable is used to return an error message. The "defer row.Close()" statement ensures
	// that the row object is closed and the connection to the database is properly terminated when the function ends.
	row, err := s.Context.Db.Query(`select value, symbol from balances where symbol = $1 and type = $2 and user_id = $3`, req.GetSymbol(), types.TypeStock, auth)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The code 'if row.Next()' is used to check if there are any more rows left in a result set from a database query. It
	// advances the row pointer to the next row and returns true if there is a row, or false if there are no more rows.
	if row.Next() {

		// This code is checking for an error when scanning the row of a database. If an error is found, the code will return
		// an error response and the error itself.
		if err := row.Scan(&item.Value, &item.Symbol); err != nil {
			return &response, err
		}

		// This is a conditional statement that checks if the quantity requested is greater than or equal to the value of the
		// item. If the condition is true, a certain action will be taken; if it is false, a different action will be taken.
		if item.GetValue() >= req.GetQuantity() {

			// This code is used to update the balance of a user's assets in a database. The code updates the user's balance by
			// subtracting the quantity given. The values being used to update the balance are stored in variables, and are passed
			// into the code as parameters ($1, $2, and $3). The code also checks for errors and returns an error if one is found.
			if _, err := s.Context.Db.Exec("update balances set value = value - $2 where symbol = $1 and user_id = $3 and type = $4;", req.GetSymbol(), req.GetQuantity(), auth, types.TypeStock); err != nil {
				return &response, err
			}

			// This line of code is used to insert data into a table called "withdraws" in a database. The four values being
			// inserted are: symbol, quantity, status, broker_id, and user_id. These values are being taken from the request (req) and the
			// item (item). The line also checks for any errors that might occur during the insertion process, and if an error is found it returns an error message.
			if err = s.Context.Db.QueryRow("insert into transfer (symbol, quantity, status, broker_id, user_id) values ($1, $2, $3, $4, $5) returning id, symbol, quantity, status, broker_id, user_id, create_at", req.GetSymbol(), req.GetQuantity(), item.GetStatus(), item.GetId(), auth).Scan(
				&item.Id,
				&item.Symbol,
				&item.Value,
				&item.Status,
				&item.BrokerId,
				&item.UserId,
				&item.CreateAt,
			); err != nil {
				return &response, err
			}

			response.Success = true
		} else {
			return &response, status.Error(710076, "you do not have enough funds to withdraw the amount of the asset")
		}

		response.Fields = append(response.Fields, &item)
	}

	return &response, nil
}

// GetTransfers - The code snippet above is a function that is used to retrieve information about stock withdrawals from a database. It
// takes a context, request and response variables as arguments. It checks that an authentication token is present in the
// context, gets an agent using the authentication credentials, then checks the request's ID value. It then generates a
// query to select the count of records from the withdrawals table and stores it in the response.Count variable. It also
// sets an offset for a paginated request and runs a query to retrieve data from two tables (withdraws and accounts)
// based on the parameters and conditions given. Finally, it loops over each row of the result set, assigning each
// column's value to the appropriate item.field, and appends the item to the response.Fields array.
func (s *Service) GetTransfers(ctx context.Context, req *pbstock.GetRequestTransfers) (*pbstock.ResponseTransfer, error) {

	// The purpose of the code snippet above is to declare two variables, response and maps. The variable response is of
	// type pbstock.ResponseTransfer, while the variable maps is of type string slice.
	var (
		response pbstock.ResponseTransfer
		maps     []string
	)

	// The purpose of this code is to set a default limit value if the limit value requested (req.GetLimit()) is equal to
	// zero. In this case, the default limit value is set to 30.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	if req.GetUnshift() {
		maps = append(maps, fmt.Sprintf("where broker_id = %[1]d", agent.GetId()))
	} else {

		// This code is used to assign a value to the "Id" field in the "agent" object. If the "GetBrokerId()" method returns a
		// value of 0 then the "Id" field is assigned the value of the "GetId()" method. If the "GetBrokerId()" method returns
		// a value other than 0 then the "Id" field is assigned the value of the "GetBrokerId()" method.
		if agent.GetBrokerId() == 0 {
			agent.Id = agent.GetId()
		} else {
			agent.Id = agent.GetBrokerId()
		}

		maps = append(maps, fmt.Sprintf("where broker_id = %[1]d and user_id = %[2]d", agent.GetId(), auth))
	}

	// This code is checking if the length of the request's symbol is greater than 0. If it is, a new string is appended to
	// the maps variable with a formatted printf statement using the symbol from the request.
	if len(req.GetSymbol()) > 0 {
		maps = append(maps, fmt.Sprintf("and symbol = '%[1]s'", req.GetSymbol()))
	}

	// The code snippet is used to query a database table for a specific condition. The fmt.Sprintf() function is used to
	// construct a formatted string with the strings.Join() function used to join the maps argument. The query is used to
	// select the count of records from the withdrawals table, which is then stored in the response.Count variable. The
	// response.GetCount() function is then used to check if the count is greater than 0, which would indicate that the query was successful.
	if _ = s.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count from transfer %s`, strings.Join(maps, " "))).Scan(&response.Count); response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is running a SQL query to retrieve data from two tables, "withdraws" and "accounts", based on the
		// parameters and conditions given. The query is designed to return the "id", "name", "user_id", "quantity",
		// "broker_id", "status" and "create_at" fields from the "withdraws" and "accounts" tables, while filtering the results
		// based on the given parameters and conditions. It will also order the results by the "id" field, and limit the
		// results to the number of records given in the "req.GetLimit()" parameter. Finally, it will offset the results by the given "offset" parameter.
		rows, err := s.Context.Db.Query(fmt.Sprintf(`select a.id, a.symbol, b.name, a.user_id, a.quantity, a.broker_id, a.status, a.create_at from transfer a inner join accounts b on b.id = a.user_id %[1]s order by a.id desc limit %[2]d offset %[3]d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate through the result set of a SQL query. It will loop over each row of the
		// result set, allowing the user to access the data from each row.
		for rows.Next() {

			// The variable 'item' is being declared as a type of 'pbstock.Transfer', which is a type of struct in the pbstock
			// package. This variable is being declared so that it can be used in a program to store information about a stock transfer.
			var (
				item pbstock.Transfer
			)

			// This code is part of a database query. It is scanning the columns of the query's results and assigning each
			// column's value to the appropriate item.field. If an error occurs during the scan, to err variable will not be nil
			// and the error will be logged.
			if err = rows.Scan(
				&item.Id,
				&item.Symbol,
				&item.Name,
				&item.UserId,
				&item.Value,
				&item.BrokerId,
				&item.Status,
				&item.CreateAt,
			); err != nil {
				return &response, err
			}

			// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
			// item to the Fields array.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check if there is an error with the rows object. If there is an error, the code will return the
		// response object along with an error.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// CancelTransfer - This code is a function in a stock service that is used to cancel a transfer request. The code is responsible for
// retrieving the value and symbol from the withdrawals table, updating the user's balance, and updating the status of
// the withdrawal to CANCEL. The code is also responsible for validating the user's authentication token and checking for errors.
func (s *Service) CancelTransfer(ctx context.Context, req *pbstock.CancelRequestTransfer) (*pbstock.ResponseTransfer, error) {

	// The purpose of the following code is to declare two variables: response and item. The first variable, response, is of
	// type pbstock.ResponseTransfer and the second variable, item, is of type pbstock.Transfer. This code is used to create
	// variables to store data in a program.
	var (
		response pbstock.ResponseTransfer
		item     pbstock.Transfer
		maps     []string
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The if statement is used to check if the GetUnshift() function returns a value that evaluates to true. If it does,
	// then the code inside the if statement will be executed.
	if req.GetUnshift() {

		// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
		// trying to get the agent, the code snippet will return an error to the user.
		agent, err := s.queryAgent(auth)
		if err != nil {
			return &response, err
		}

		// This code is checking if the agent's broker ID is equal to 0. If it is, it is appending an additional parameter to a
		// list of parameters (maps) with the agent's ID. This means that the list of parameters will include the agent's ID if
		// their broker ID is 0.
		if agent.GetBrokerId() == 0 {
			maps = append(maps, fmt.Sprintf("and broker_id = %[1]d", agent.GetId()))
		}

	} else {
		maps = append(maps, fmt.Sprintf("and user_id = %[1]d", auth))
	}

	// This query is retrieving the value and symbol from the withdrawals table where the ID and user_id match the request ID
	// and auth variables, respectively. The row and err variables are used to store the results of the query and any
	// potential errors that may occur. The defer statement is used to ensure that the row.Close() method is called when the
	// function returns, which will close the row variable and free up any resources that were allocated for the query.
	row, err := s.Context.Db.Query(fmt.Sprintf(`select quantity, symbol, user_id from transfer where id = %[1]d %[2]s`, req.GetId(), strings.Join(maps, " ")))
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The code 'if row.Next()' is used to check if there are any more rows left in a result set from a database query. It
	// advances the row pointer to the next row and returns true if there is a row, or false if there are no more rows.
	if row.Next() {

		// This code is checking for an error when scanning the row of a database. If an error is found, the code will return
		// an error response and the error itself.
		if err := row.Scan(&item.Value, &item.Symbol, &item.UserId); err != nil {
			return &response, err
		}

		// This code is used to update the balance of a user's assets in a database. The code updates the user's balance by
		// subtracting the quantity given. The values being used to update the balance are stored in variables, and are passed
		// into the code as parameters ($1, $2, and $3). The code also checks for errors and returns an error if one is found.
		if _, err := s.Context.Db.Exec("update balances set value = value + $2 where symbol = $1 and user_id = $3 and type = $4;", item.GetSymbol(), item.GetValue(), item.GetUserId(), types.TypeStock); err != nil {
			return &response, err
		}

		// This code is executing an SQL query to update the status of a withdraw with the given ID and user ID to CANCEL. The
		// "_" is being used as a placeholder for the result of s.Context.Db.Exec, which is not being used. The "err" is the
		// error that is returned if the query fails. If the query fails, the code is returning an error.
		if _, err := s.Context.Db.Exec("update transfer set status = $3 where id = $1 and user_id = $2;", req.GetId(), item.GetUserId(), types.StatusCancel); err != nil {
			return &response, err
		}
		response.Success = true

		// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
		// item to the Fields array.
		response.Fields = append(response.Fields, &item)
	}

	return &response, nil
}

// SetAction - The purpose of this code is to set the broker asset by querying a database for a user's balance on a certain stock or
// other asset, updating the balance in the database, and then returning a response. The code also checks for valid
// authentication tokens and ensures that the user is a broker before allowing them to add stock security turnover.
func (s *Service) SetAction(ctx context.Context, req *pbstock.SetRequestAction) (*pbstock.ResponseAction, error) {

	// The purpose of the above code is to declare two variables, response and item, of type pbstock.ResponseAction and
	// pbstock.Asset, respectively. These variables are used to store data responses from the stockbroker and asset details for the stock.
	var (
		response pbstock.ResponseAction
		item     types.Asset
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.queryAgent(auth)
	if err != nil {
		return &response, err
	}

	// This code is checking if the broker ID is greater than 0. If it is, the code returns an error message indicating that
	// the user is not a broker and is therefore not able to add stock security turnover.
	if agent.GetBrokerId() > 0 {
		return &response, status.Error(568904, "you are not a broker to add in stock security turnover")
	}

	// This code is used to query a database. The purpose of this code is to query a database and select the id and name
	// from the stocks table where the symbol is equal to the value stored in the req.GetSymbol() variable. If an error
	// occurs, it will return an error. The row.Close() statement is used to ensure that the database connection is closed
	// when the query is finished.
	row, err := s.Context.Db.Query(`select id from assets where symbol = $1 and status = $2 and "group" = $3`, req.GetSymbol(), true, types.GroupAction)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The code 'if row.Next()' is used to check if there are any more rows left in a result set from a database query. It
	// advances the row pointer to the next row and returns true if there is a row, or false if there are no more rows.
	if row.Next() {

		// The purpose of this code is to update the balance of an asset in a database. The code checks if the request is an
		// "unshift" request and subtracts the given quantity from the balance if it is. If it is not an unshift request, the
		// code adds the given quantity to the balance. The code also checks for errors and returns an error if one is found.
		if req.GetUnshift() {

			// The purpose of this code is to query the database for a user's balance on a certain stock or other asset, and then
			// store the retrieved balance in the item.Balance variable.
			if _ = s.Context.Db.QueryRow(`select value from balances where symbol = $1 and user_id = $2 and type = $3`, req.GetSymbol(), auth, types.TypeStock).Scan(&item.Balance); item.GetBalance() == 0 {
				return &response, status.Error(796743, "your asset balance is zero, you cannot withdraw the asset from circulation")
			}

			// This code is an example of an SQL query that is used to update the balance of a particular asset. The purpose of
			// this code is to subtract a certain quantity from the balance of an asset with a given symbol, user ID, and type.
			// This code checks for an error after executing the query and returns an error if there is one.
			if _, err := s.Context.Db.Exec("update balances set value = value - $2 where symbol = $1 and user_id = $3 and type = $4;", req.GetSymbol(), req.GetQuantity(), auth, types.TypeStock); err != nil {
				return &response, err
			}

		} else {

			// This code is used to update the balance of a user's assets in a database. The code updates the user's balance by
			// subtracting the quantity given. The values being used to update the balance are stored in variables, and are passed
			// into the code as parameters ($1, $2, and $3). The code also checks for errors and returns an error if one is found.
			if _, err := s.Context.Db.Exec("update balances set value = value + $2 where symbol = $1 and user_id = $3 and type = $4;", req.GetSymbol(), req.GetQuantity(), auth, types.TypeStock); err != nil {
				return &response, err
			}
		}

		response.Success = true
	} else {
		return &response, status.Error(854333, "the asset is not a stock security, or the asset is temporarily disabled by the administration")
	}

	return &response, nil
}
