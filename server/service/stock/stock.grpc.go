package stock

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/pbasset"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"github.com/cryptogateway/backend-envoys/server/service/account"
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

	// This code checks to see if an agent has been verified. If the agent has not been verified, it will return an error
	// message to indicate that the agent has not been KYC verified.
	if !agent.Success {
		return &response, status.Error(53678, "you have not been verified KYC")
	}

	// The purpose of this switch statement is to assign a value to the "_status" variable based on the type of request
	// (req.GetType()) that is received. If the request type is "AGENT", then the "_status" variable is set to "PENDING". If
	// the request type is "BROKER", then the "_status" variable is set to "ACCESS".
	switch req.GetType() {
	case pbstock.Type_AGENT:
		agent.Status = pbstock.Status_PENDING
	case pbstock.Type_BROKER:
		agent.Status = pbstock.Status_ACCESS
	}

	// This code is used to insert data into the 'agents' table in a database. The code is also checking for any errors that
	// may occur during the insertion process, and if an error occurs, it will return an error message.
	if _, err := s.Context.Db.Exec(`insert into agents (type, name, broker_id, status, user_id) values ($1, $2, $3, $4, $5)`,
		req.GetType(),
		req.GetName(),
		req.GetBrokerId(),
		agent.GetStatus(),
		auth,
	); err != nil {
		return &response, status.Error(646788, "you have already created an agent and broker account")
	}

	// This code is checking to see if an Agent object was returned when calling the getAgent() function with the auth
	// parameter. If the Id property of the Agent object is greater than 0, then the Agent object is appended to the
	// response.Fields array.
	if agent, _ := s.getAgent(auth); agent.Id > 0 {

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

	// This code is checking to see if an Agent object was returned when calling the getAgent() function with the auth
	// parameter. If the Id property of the Agent object is greater than 0, then the Agent object is appended to the
	// response.Fields array.
	if agent, _ := s.getAgent(auth); agent.Id > 0 {
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
		maps = append(maps, fmt.Sprintf("where type = %[2]d and (name like %[1]s or id::text like %[1]s)", "'%"+req.GetSearch()+"%'", pbstock.Type_BROKER))
	} else {
		maps = append(maps, fmt.Sprintf("where type = %[1]d", pbstock.Type_BROKER))
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
	agent, _ := s.getAgent(auth)

	// The purpose of this code is to query a database and check if the count of agents with a certain status, type, and
	// broker ID is greater than 0. If the count is greater than 0, then the code will execute the following code block.
	if _ = s.Context.Db.QueryRow("select count(*) as count from agents where status = $1 and type = $2 and broker_id = $3", pbstock.Status_PENDING, pbstock.Type_AGENT, agent.GetId()).Scan(&response.Count); response.GetCount() > 0 {

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
		rows, err := s.Context.Db.Query(`select a.id, a.user_id, a.broker_id, a.type, a.status, a.create_at, b.secret from agents a left join kyc b on b.user_id = a.user_id  where a.status = $1 and a.type = $2 and a.broker_id = $3 order by a.id desc limit $4 offset $5`, pbstock.Status_PENDING, pbstock.Type_AGENT, agent.GetId(), req.GetLimit(), offset)
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
	agent, err := s.getAgent(auth)
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
	agent, err := s.getAgent(auth)
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
	agent, err := s.getAgent(auth)
	if err != nil {
		return &response, err
	}

	// This if statement is used to query an SQL database to check if a certain agent is associated with any accounts. The
	// query is using the ID of the agent and the type of account, and is checking if there is a count of any results
	// greater than 0. If the count is greater than 0, the code within the if statement will execute.
	if _ = s.Context.Db.QueryRow("select count(*) as count from agents a left join accounts b on b.id = a.user_id where a.broker_id = $1 and a.type = $2", agent.GetId(), pbstock.Type_AGENT).Scan(&response.Count); response.GetCount() > 0 {

		// This code is setting an offset for a Paginated request. The offset is used to determine the index of the first item
		// that should be returned. This code is calculating the offset by multiplying the limit (the number of items per page)
		// with the page number. If the page number is greater than 0, the offset is calculated with the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query the database for information from the tables agents and accounts. The query is looking
		// for information on agents, with the broker_id, type, limit, and offset specified by the parameters agent.GetId(),
		// pbstock.Type_AGENT, req.GetLimit(), offset. The query will return the id, name, email, user_id, broker_id, type,
		// status, and create_at associated with the specified parameters. The result of the query will be stored in the rows
		// variable, and the query will be closed when the defer rows.Close() line is executed.
		rows, err := s.Context.Db.Query(`select a.id, b.name, b.email, a.user_id, a.broker_id, a.type, a.status, a.create_at from agents a inner join accounts b on b.id = a.user_id where a.broker_id = $1 and a.type = $2 order by a.id desc limit $3 offset $4`, agent.GetId(), pbstock.Type_AGENT, req.GetLimit(), offset)
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
	// pbstock.Status, respectively. These variables can then be used to store values of that type.
	var (
		response pbstock.ResponseBlocked
		status   pbstock.Status
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.getAgent(auth)
	if err != nil {
		return &response, err
	}

	// This code is used to query a database for information about a specific agent. The variables req, agent, and pbstock
	// are passed in to this function as parameters. The code queries the database for the status of the agent from the
	// given id, broker_id, and type, and stores it in the variable "status". If there is an error with the query, the
	// function returns an error.
	if err = s.Context.Db.QueryRow("select status as count from agents where id = $1 and broker_id = $2 and type = $3", req.GetId(), agent.GetId(), pbstock.Type_AGENT).Scan(&status); err != nil {
		return &response, err
	}

	// This code is used to update the status of an agent when given a request. Depending on the initial status, the code
	// will either block the agent or give them access.
	switch status {
	case pbstock.Status_BLOCKED:
		_, _ = s.Context.Db.Exec("update agents set status = $2 where id = $1", req.GetId(), pbstock.Status_ACCESS)
	case pbstock.Status_ACCESS:
		_, _ = s.Context.Db.Exec("update agents set status = $2 where id = $1", req.GetId(), pbstock.Status_BLOCKED)
	}

	return &response, nil
}

// GetAssets - The purpose of this code is to query a database for assets such as stocks and currencies, and store the retrieved data
// in a ResponseAsset object. The code also checks for the presence of an authentication token in the context, and
// queries the database for a user's balance on a certain asset. The code is used to retrieve data from a database and
// store it in a ResponseAsset object.
func (s *Service) GetAssets(ctx context.Context, req *pbstock.GetRequestAssets) (*pbstock.ResponseAsset, error) {

	// The purpose of the code is to declare a variable called "response" of type "pbstock.ResponseAsset". This variable
	// will be used to store the response data of a request to the pbstock service.
	var (
		response pbstock.ResponseAsset
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The "if req.GetFiat()" statement is a conditional statement that checks if the GetFiat() method returns a true value.
	// If the method returns true, the code within the statement block will be executed. This statement is used to determine
	// what code should be executed depending on the result of the GetFiat() method.
	if req.GetFiat() {

		// This code is querying a database table for all the currencies of type "FIAT" (e.g. US Dollar). The purpose of the
		// code is to retrieve the ID, name, and symbol of each currency so they can be used in further processing. The row
		// variable is used to store the query results, err is used to store any errors that may occur during the query, and
		// defer row.Close() is used to ensure that the query results are properly closed.
		row, err := s.Context.Db.Query(`select id, name, symbol, status from currencies where type = $1`, pbspot.Type_FIAT)
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The purpose of the statement "for row.Next()" is to loop through each row of a database query result set and perform
		// some action on it. It is used to iterate through a result set of a query in order to process each row. It will
		// execute the code block each time it iterates over a row in the result set.
		for row.Next() {

			// The variable "item" is being declared as a pbstock.Asset type. This is a way of telling the compiler that the
			// variable "item" is going to hold a value of type pbstock.Asset.
			var (
				item pbstock.Asset
			)

			// This code is used to scan the row and assign the values of the columns to the item struct.  The if statement checks
			// whether the row scan was successful.  If it was not, the response and an error are returned.
			if err := row.Scan(&item.Id, &item.Name, &item.Symbol, &item.Status); err != nil {
				return &response, err
			}

			// The purpose of this code is to query the database for a user's balance on a certain stock or other asset, and then
			// store the retrieved balance in the item.Balance variable.
			_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, pbasset.Type_STOCK).Scan(&item.Balance)

			response.Fields = append(response.Fields, &item)
		}

	} else {

		// This code is used to query the stocks table in a database. The row variable is used to store the results of the
		// query, while the err variable is used to store any errors that occur while running the query. The defer statement is
		// used to ensure that the row object is closed when the function exits, regardless of whether an error occurs or not.
		row, err := s.Context.Db.Query(`select id, name, symbol, tag, zone, price, status from stocks`)
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The purpose of the statement "for row.Next()" is to loop through each row of a database query result set and perform
		// some action on it. It is used to iterate through a result set of a query in order to process each row. It will
		// execute the code block each time it iterates over a row in the result set.
		for row.Next() {

			// The variable "item" is being declared as a pbstock.Asset type. This is a way of telling the compiler that the
			// variable "item" is going to hold a value of type pbstock.Asset.
			var (
				item pbstock.Asset
			)

			// This code is used to scan a row of data from a database and collect the values into variables. The err variable is
			// used to check if there was any error while scanning the row. If there is an error, the function returns an error
			// response and the error.
			if err := row.Scan(&item.Id, &item.Name, &item.Symbol, &item.Tag, &item.Zone, &item.Price, &item.Status); err != nil {
				return &response, err
			}

			// The purpose of this code is to query the database for a user's balance on a certain stock or other asset, and then
			// store the retrieved balance in the item.Balance variable.
			_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, pbasset.Type_STOCK).Scan(&item.Balance)

			response.Fields = append(response.Fields, &item)
		}
	}

	return &response, nil
}

func (s *Service) GetAsset(ctx context.Context, req *pbstock.GetRequestAsset) (*pbstock.ResponseAsset, error) {

	// The purpose of the code is to declare a variable called "response" of type "pbstock.ResponseAsset". This variable
	// will be used to store the response data of a request to the pbstock service.
	var (
		response pbstock.ResponseAsset
		item     pbstock.Asset
	)

	// This code is checking to make sure a valid authentication token is present in the context. If it is not, it returns
	// an error. This is necessary to ensure that only authorized users are accessing certain resources.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	row, err := s.Context.Db.Query(`select id from stocks where symbol = $1`, req.GetSymbol())
	if err != nil {
		return &response, err
	}
	defer row.Close()

	if row.Next() {

		// The purpose of this code is to query the database for a user's balance on a certain stock or other asset, and then
		// store the retrieved balance in the item.Balance variable.
		_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, pbasset.Type_STOCK).Scan(&item.Balance)

	} else {

	}

	return &response, nil
}

func (s *Service) SetAsset(ctx context.Context, req *pbstock.SetRequestAsset) (*pbstock.ResponseAsset, error) {
	//TODO implement me
	panic("implement me")
}
