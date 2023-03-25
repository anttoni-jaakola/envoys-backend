package stock

import (
	"context"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/decimal"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto"

	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"github.com/cryptogateway/backend-envoys/server/service/account"
	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/status"
	"strings"
	"time"
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
		agent.Status = proto.Status_PENDING
	case pbstock.Type_BROKER:
		agent.Status = proto.Status_ACCESS
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
	if _ = s.Context.Db.QueryRow("select count(*) as count from agents where status = $1 and type = $2 and broker_id = $3", proto.Status_PENDING, pbstock.Type_AGENT, agent.GetId()).Scan(&response.Count); response.GetCount() > 0 {

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
		rows, err := s.Context.Db.Query(`select a.id, a.user_id, a.broker_id, a.type, a.status, a.create_at, b.secret from agents a left join kyc b on b.user_id = a.user_id  where a.status = $1 and a.type = $2 and a.broker_id = $3 order by a.id desc limit $4 offset $5`, proto.Status_PENDING, pbstock.Type_AGENT, agent.GetId(), req.GetLimit(), offset)
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
	// proto.Status, respectively. These variables can then be used to store values of that type.
	var (
		response pbstock.ResponseBlocked
		status   proto.Status
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
	case proto.Status_BLOCKED:
		_, _ = s.Context.Db.Exec("update agents set status = $2 where id = $1", req.GetId(), proto.Status_ACCESS)
		response.Success = true
	case proto.Status_ACCESS:
		_, _ = s.Context.Db.Exec("update agents set status = $2 where id = $1", req.GetId(), proto.Status_BLOCKED)
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
			_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Balance)

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
			_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Balance)

			response.Fields = append(response.Fields, &item)
		}
	}

	return &response, nil
}

// GetAsset - The purpose of this code is to create a service function called GetAsset which receives a context and a
// GetRequestAsset and returns a ResponseAsset and an error. The code is used to query a database to get information
// about a specified asset (req.GetSymbol()) and calculate its volume in orders with status 'PENDING' and type 'STOCK'.
// The code also checks for a valid authentication token in the context to ensure that only authorized users can access
// certain resources. The code also looks up the balance of the specified asset in the database and stores it in the
// item.Balance variable. Finally, the code returns the ResponseAsset and any errors that may have occurred.
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

	// This code snippet is used to get an agent using the given authentication credentials. If there is an error when
	// trying to get the agent, the code snippet will return an error to the user.
	agent, err := s.getAgent(auth)
	if err != nil {
		return &response, err
	}

	// This if statement is checking if the broker ID of the agent is 0. If it is, then it sets the Type of the item to
	// BROKER, which is a member of the pbstock type. This is likely done to ensure that the item has the correct type for the given broker.
	if agent.GetBrokerId() == 0 {
		item.Type = pbstock.Type_BROKER
	}
	item.AgentStatus = agent.GetStatus()

	// This query is used to calculate the volume of the specified asset (req.GetSymbol()) in orders with status 'PENDING'
	// and type 'STOCK' for the user specified by 'auth'. The result of the query is then stored in the variable
	// 'item.Volume'.
	_ = s.Context.Db.QueryRow(`select coalesce(sum(case when base_unit = $1 then value when quote_unit = $1 then value * price end), 0.00) as volume from orders where base_unit = $1 and status = $2 and type = $3 and user_id = $4 or quote_unit = $1 and status = $2 and type = $3 and user_id = $4`, req.GetSymbol(), proto.Status_PENDING, proto.Type_STOCK, auth).Scan(&item.Volume)

	// The purpose of the following code is to check if a given asset exists in the database.
	// The code is querying the database with a specific symbol, user_id and asset type and then scanning the result of the query into a boolean value which is stored in the item.Exist variable.
	_ = s.Context.Db.QueryRow("select exists(select id from assets where symbol = $1 and user_id = $2 and type = $3)::bool", req.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Exist)

	// This code is used to query a database. The purpose of this code is to query a database and select the id and name
	// from the stocks table where the symbol is equal to the value stored in the req.GetSymbol() variable. If an error
	// occurs, it will return an error. The row.Close() statement is used to ensure that the database connection is closed
	// when the query is finished.
	row, err := s.Context.Db.Query(`select id, name, symbol, status from stocks where symbol = $1`, req.GetSymbol())
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The code 'if row.Next()' is used to check if there are any more rows left in a result set from a database query. It
	// advances the row pointer to the next row and returns true if there is a row, or false if there are no more rows.
	if row.Next() {

		// This code is checking for an error when scanning the row of a database. If an error is found, the code will return
		// an error response and the error itself.
		if err := row.Scan(&item.Id, &item.Name, &item.Symbol, &item.Status); err != nil {
			return &response, err
		}

		// The purpose of this code is to query the database for a user's balance on a certain stock or other asset, and then
		// store the retrieved balance in the item.Balance variable.
		_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Balance)

	} else {

		// This code retrieves data from a database table called "currencies". The code is using a parameterized query to
		// protect against SQL injection attacks. The parameters are taken from the req.GetSymbol() and pbspot.Type_FIAT
		// variables. The row.Close() statement is used to close the open database connection when the query is finished,
		// preventing potential connection leaks.
		row, err = s.Context.Db.Query(`select id, name, symbol, status from currencies where symbol = $1 and type = $2`, req.GetSymbol(), pbspot.Type_FIAT)
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The code 'if row.Next()' is used to check if there are any more rows left in a result set from a database query. It
		// advances the row pointer to the next row and returns true if there is a row, or false if there are no more rows.
		if row.Next() {

			// This code is checking for an error when scanning the row of a database. If an error is found, the code will return
			// an error response and the error itself.
			if err := row.Scan(&item.Id, &item.Name, &item.Symbol, &item.Status); err != nil {
				return &response, err
			}

			spew.Dump(req.GetSymbol(), item.GetSymbol(), auth)

			// The purpose of this code is to query the database for a user's balance on a certain stock or other asset, and then
			// store the retrieved balance in the item.Balance variable.
			_ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, item.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Balance)

			item.Tag = pbstock.Tag_FIAT
		}
	}
	response.Fields = append(response.Fields, &item)

	return &response, nil
}

// SetAsset - The purpose of this code is to set an asset in the database with a given symbol, user_id, and asset type. It checks to
// make sure that the given authentication token is valid and that the asset does not already exist in the database. If
// the asset does not exist, the code inserts the values into the "assets" table in the database. If the asset does
// exist, it returns an error. The code also returns the response data of the request in the form of a pbstock.ResponseAsset variable.
func (s *Service) SetAsset(ctx context.Context, req *pbstock.SetRequestAsset) (*pbstock.ResponseAsset, error) {

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

	// The purpose of the following code is to check if a given asset exists in the database.
	// The code is querying the database with a specific symbol, user_id and asset type and then scanning the result of the query into a boolean value which is stored in the item.Exist variable.
	if _ = s.Context.Db.QueryRow("select exists(select id from assets where symbol = $1 and user_id = $2 and type = $3)::bool", req.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Exist); !item.Exist {

		// This is an if statement to insert values into the database table "assets". The statement sets the variables,
		// user_id, symbol, and type, to the values of the auth, req.GetSymbol(), and proto.Type_STOCK variables,
		// respectively. If an error occurs, the statement will return an error.
		if _, err = s.Context.Db.Exec("insert into assets (user_id, symbol, type) values ($1, $2, $3);", auth, req.GetSymbol(), proto.Type_STOCK); err != nil {
			return &response, err
		}

		item.Exist = true
	} else {
		return &response, status.Error(778543, "your asset has already been activated before")
	}
	response.Fields = append(response.Fields, &item)

	return &response, nil
}

// SetWithdraw - The purpose of this code is to create a service that allows users to withdraw assets from their account. The code
// checks for valid authentication, retrieves an agent using the given authentication credentials, checks the status of
// the agent and returns an error if it is blocked, gets the value of an item from the database, checks to make sure
// there are enough funds to withdraw the requested amount, updates the user's balance, and inserts a new record into the
// withdrawals table in the database. Finally, the code returns a response indicating whether the withdrawal was successful.
func (s *Service) SetWithdraw(ctx context.Context, req *pbstock.SetRequestWithdraw) (*pbstock.ResponseWithdraw, error) {

	// The purpose of the following code is to declare two variables, response and item, of type pbstock.ResponseWithdraw
	// and pbstock.Withdraw respectively. These variables are used to store data related to a ResponseWithdraw and a
	// withdrawal respectively, which are both structs from the pbstock package.
	var (
		response pbstock.ResponseWithdraw
		item     pbstock.Withdraw
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
	agent, err := s.getAgent(auth)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to check the status of the agent and if it is "BLOCKED", then it will return an error
	// message with a status code of 523217. This can help the application determine if the agent should be allowed to
	// continue with its current task.
	if agent.GetStatus() == proto.Status_BLOCKED {
		return &response, status.Error(523217, "your asset blocked")
	}

	// This code is checking the value of the agent's broker ID. If it is set to 0, then the item's status is set to FILLED,
	// and the item's ID is set to the agent's ID. If the agent's broker ID is not set to 0, then the item's status is set
	// to PENDING, and the item's ID is set to the agent's broker ID.
	if agent.GetBrokerId() == 0 {
		item.Status = proto.Status_FILLED
		item.Id = agent.GetId()
	} else {
		item.Status = proto.Status_PENDING
		item.Id = agent.GetBrokerId()
	}

	// This code is querying a database for records that match certain criteria. The Query function takes a SQL statement
	// and two arguments, req.GetSymbol() and proto.Type_STOCK, which represent the criteria for the query. The results of
	// the query are stored in a "row" object and can be accessed using the "row.Close()" function. If an error occurs while
	// executing the query, the "err" variable is used to return an error message. The "defer row.Close()" statement ensures
	// that the row object is closed and the connection to the database is properly terminated when the function ends.
	row, err := s.Context.Db.Query(`select balance, symbol from assets where symbol = $1 and type = $2 and user_id = $3`, req.GetSymbol(), proto.Type_STOCK, auth)
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
			if _, err := s.Context.Db.Exec("update assets set balance = balance - $2 where symbol = $1 and user_id = $3 and type = $4;", req.GetSymbol(), req.GetQuantity(), auth, proto.Type_STOCK); err != nil {
				return &response, err
			}

			// This line of code is used to insert data into a table called "withdraws" in a database. The four values being
			// inserted are: symbol, quantity, status, broker_id, and user_id. These values are being taken from the request (req) and the
			// item (item). The line also checks for any errors that might occur during the insertion process, and if an error is found it returns an error message.
			if err = s.Context.Db.QueryRow("insert into withdraws (symbol, quantity, status, broker_id, user_id) values ($1, $2, $3, $4, $5) returning id, symbol, quantity, status, broker_id, user_id, create_at", req.GetSymbol(), req.GetQuantity(), item.GetStatus(), item.GetId(), auth).Scan(
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

// GetWithdraws - The code snippet above is a function that is used to retrieve information about stock withdrawals from a database. It
// takes a context, request and response variables as arguments. It checks that an authentication token is present in the
// context, gets an agent using the authentication credentials, then checks the request's ID value. It then generates a
// query to select the count of records from the withdrawals table and stores it in the response.Count variable. It also
// sets an offset for a paginated request and runs a query to retrieve data from two tables (withdraws and accounts)
// based on the parameters and conditions given. Finally, it loops over each row of the result set, assigning each
// column's value to the appropriate item.field, and appends the item to the response.Fields array.
func (s *Service) GetWithdraws(ctx context.Context, req *pbstock.GetRequestWithdraws) (*pbstock.ResponseWithdraw, error) {

	// The purpose of the code snippet above is to declare two variables, response and maps. The variable response is of
	// type pbstock.ResponseWithdraw, while the variable maps is of type string slice.
	var (
		response pbstock.ResponseWithdraw
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
	agent, err := s.getAgent(auth)
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
	if _ = s.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count from withdraws %s`, strings.Join(maps, " "))).Scan(&response.Count); response.GetCount() > 0 {

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
		rows, err := s.Context.Db.Query(fmt.Sprintf(`select a.id, a.symbol, b.name, a.user_id, a.quantity, a.broker_id, a.status, a.create_at from withdraws a inner join accounts b on b.id = a.user_id %[1]s order by a.id desc limit %[2]d offset %[3]d`, strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for rows.Next() loop is used to iterate through the result set of a SQL query. It will loop over each row of the
		// result set, allowing the user to access the data from each row.
		for rows.Next() {

			// The variable 'item' is being declared as a type of 'pbstock.Withdraw', which is a type of struct in the pbstock
			// package. This variable is being declared so that it can be used in a program to store information about a stock withdrawal.
			var (
				item pbstock.Withdraw
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

// CancelWithdraw - This code is a function in a stock service that is used to cancel a withdraw request. The code is responsible for
// retrieving the value and symbol from the withdrawals table, updating the user's balance, and updating the status of
// the withdraw to CANCEL. The code is also responsible for validating the user's authentication token and checking for errors.
func (s *Service) CancelWithdraw(ctx context.Context, req *pbstock.CancelRequestWithdraw) (*pbstock.ResponseWithdraw, error) {

	// The purpose of the following code is to declare two variables: response and item. The first variable, response, is of
	// type pbstock.ResponseWithdraw and the second variable, item, is of type pbstock.Withdraw. This code is used to create
	// variables to store data in a program.
	var (
		response pbstock.ResponseWithdraw
		item     pbstock.Withdraw
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
		agent, err := s.getAgent(auth)
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
	row, err := s.Context.Db.Query(fmt.Sprintf(`select quantity, symbol, user_id from withdraws where id = %[1]d %[2]s`, req.GetId(), strings.Join(maps, " ")))
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
		if _, err := s.Context.Db.Exec("update assets set balance = balance + $2 where symbol = $1 and user_id = $3 and type = $4;", item.GetSymbol(), item.GetValue(), item.GetUserId(), proto.Type_STOCK); err != nil {
			return &response, err
		}

		// This code is executing an SQL query to update the status of a withdraw with the given ID and user ID to CANCEL. The
		// "_" is being used as a placeholder for the result of s.Context.Db.Exec, which is not being used. The "err" is the
		// error that is returned if the query fails. If the query fails, the code is returning an error.
		if _, err := s.Context.Db.Exec("update withdraws set status = $3 where id = $1 and user_id = $2;", req.GetId(), item.GetUserId(), proto.Status_CANCEL); err != nil {
			return &response, err
		}
		response.Success = true

		// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
		// item to the Fields array.
		response.Fields = append(response.Fields, &item)
	}

	return &response, nil
}

// SetBrokerAsset - The purpose of this code is to set the broker asset by querying a database for a user's balance on a certain stock or
// other asset, updating the balance in the database, and then returning a response. The code also checks for valid
// authentication tokens and ensures that the user is a broker before allowing them to add stock security turnover.
func (s *Service) SetBrokerAsset(ctx context.Context, req *pbstock.SetRequestBrokerAsset) (*pbstock.ResponseBrokerAsset, error) {

	// The purpose of the above code is to declare two variables, response and item, of type pbstock.ResponseBrokerAsset and
	// pbstock.Asset, respectively. These variables are used to store data responses from the stockbroker and asset details for the stock.
	var (
		response pbstock.ResponseBrokerAsset
		item     pbstock.Asset
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

	// This code is checking if the broker ID is greater than 0. If it is, the code returns an error message indicating that
	// the user is not a broker and is therefore not able to add stock security turnover.
	if agent.GetBrokerId() > 0 {
		return &response, status.Error(568904, "you are not a broker to add in stock security turnover")
	}

	// This code is used to query a database. The purpose of this code is to query a database and select the id and name
	// from the stocks table where the symbol is equal to the value stored in the req.GetSymbol() variable. If an error
	// occurs, it will return an error. The row.Close() statement is used to ensure that the database connection is closed
	// when the query is finished.
	row, err := s.Context.Db.Query(`select id from stocks where symbol = $1 and status = $2`, req.GetSymbol(), true)
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
			if _ = s.Context.Db.QueryRow(`select balance from assets where symbol = $1 and user_id = $2 and type = $3`, req.GetSymbol(), auth, proto.Type_STOCK).Scan(&item.Balance); item.GetBalance() == 0 {
				return &response, status.Error(796743, "your asset balance is zero, you cannot withdraw the asset from circulation")
			}

			// This code is an example of an SQL query that is used to update the balance of a particular asset. The purpose of
			// this code is to subtract a certain quantity from the balance of an asset with a given symbol, user ID, and type.
			// This code checks for an error after executing the query and returns an error if there is one.
			if _, err := s.Context.Db.Exec("update assets set balance = balance - $2 where symbol = $1 and user_id = $3 and type = $4;", req.GetSymbol(), req.GetQuantity(), auth, proto.Type_STOCK); err != nil {
				return &response, err
			}

		} else {

			// This code is used to update the balance of a user's assets in a database. The code updates the user's balance by
			// subtracting the quantity given. The values being used to update the balance are stored in variables, and are passed
			// into the code as parameters ($1, $2, and $3). The code also checks for errors and returns an error if one is found.
			if _, err := s.Context.Db.Exec("update assets set balance = balance + $2 where symbol = $1 and user_id = $3 and type = $4;", req.GetSymbol(), req.GetQuantity(), auth, proto.Type_STOCK); err != nil {
				return &response, err
			}
		}

		response.Success = true
	} else {
		return &response, status.Error(854333, "the asset is not a stock security, or the asset is temporarily disabled by the administration")
	}

	return &response, nil
}

// GetCandles - This function is a method of the service struct and serves to provide a response of candles for a given request. It is
// responsible for querying the database for the candle data, and then formatting it into a response struct. It also
// calculates some statistics based on the requested data and adds them to the response struct.
func (s *Service) GetCandles(_ context.Context, req *pbstock.GetRequestCandles) (*pbstock.ResponseCandles, error) {

	// The purpose of this code is to create three variables with zero values: response, limit and maps. The response
	// variable is of type pbstock.ResponseCandles, the limit variable is of type string, and the maps variable is of type
	// slice of strings.
	var (
		response pbstock.ResponseCandles
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
		maps = append(maps, fmt.Sprintf(`and to_char(ohlc.create_at::timestamp, 'yyyy-mm-dd hh24:mi:ss') < to_char(to_timestamp(%[1]d), 'yyyy-mm-dd hh24:mi:ss')`, req.GetTo()))
	}

	// This code is used to query the database to return OHLC (open-high-low-close) data. The SQL query is using the
	// fmt.Sprintf function to substitute the variables (req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "),
	// help.Resolution(req.GetResolution()), limit) into the query. The query is then executed, and the results are stored
	// in the rows variable. Finally, the rows variable is closed at the end of the code.
	rows, err := s.Context.Db.Query(fmt.Sprintf("select extract(epoch from time_bucket('%[4]s', ohlc.create_at))::integer buckettime, first(ohlc.price, ohlc.create_at) as open, last(ohlc.price, ohlc.create_at) as close, first(ohlc.price, ohlc.price) as low, last(ohlc.price, ohlc.price) as high, sum(ohlc.quantity) as volume, avg(ohlc.price) as avg_price, ohlc.base_unit, ohlc.quote_unit from trades as ohlc where ohlc.base_unit = '%[1]s' and ohlc.quote_unit = '%[2]s' %[3]s group by buckettime, ohlc.base_unit, ohlc.quote_unit order by buckettime desc %[5]s", req.GetBaseUnit(), req.GetQuoteUnit(), strings.Join(maps, " "), help.Resolution(req.GetResolution()), limit))
	if err != nil {
		return &response, err
	}
	defer rows.Close()

	// The purpose of the for rows.Next() loop is to iterate through the rows in a database table. It is used to perform
	// some action on each row of the table. This could include retrieving data from the row, updating data in the row, or
	// deleting the row.
	for rows.Next() {

		// The purpose of the variable "item" is to store data of type pbstock.Candles. This could be used to store an array of
		// candles or other data related to pbstock.Candles.
		var (
			item pbstock.Candles
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

	// The purpose of the following code is to declare a variable called stats of the type pbspot.Stats. This variable will
	// be used to store information related to the pbspot.Stats data type.
	var (
		stats pbstock.Stats
	)

	// This code is used to fetch and analyze data from a database. It uses the QueryRow() method to retrieve data from the
	// database and then scan it into the stats variable. The code is specifically used to get the count, volume, low, high,
	// first and last values from the trades table for a given base unit and quote unit.
	_ = s.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count, sum(h24.quantity) as volume, first(h24.price, h24.price) as low, last(h24.price, h24.price) as high, first(h24.price, h24.create_at) as first, last(h24.price, h24.create_at) as last from trades as h24 where h24.create_at > now()::timestamp - '24 hours'::interval and h24.base_unit = '%[1]s' and h24.quote_unit = '%[2]s'`, req.GetBaseUnit(), req.GetQuoteUnit())).Scan(&stats.Count, &stats.Volume, &stats.Low, &stats.High, &stats.First, &stats.Last)

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

// GetPair - The purpose of this code is to query a database for a specific pair of stocks, scan the row of the database, and add
// the item to an array in the response object. It checks for an error when scanning, and returns the error if one is
// found.
func (s *Service) GetPair(_ context.Context, req *pbstock.GetRequestPair) (*pbstock.ResponsePair, error) {

	var (
		response pbstock.ResponsePair
		item     pbstock.Pair
	)

	row, err := s.Context.Db.Query(`select id, name, symbol, zone, base_decimal, quote_decimal, status from stocks where symbol = $1 and zone = $2`, req.GetBaseUnit(), req.GetQuoteUnit())
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The code 'if row.Next()' is used to check if there are any more rows left in a result set from a database query. It
	// advances the row pointer to the next row and returns true if there is a row, or false if there are no more rows.
	if row.Next() {

		// This code is checking for an error when scanning the row of a database. If an error is found, the code will return
		// an error response and the error itself.
		if err := row.Scan(&item.Id, &item.Name, &item.BaseUnit, &item.QuoteUnit, &item.BaseDecimal, &item.QuoteDecimal, &item.Status); err != nil {
			return &response, err
		}
		item.Symbol = fmt.Sprintf("%v/%v", item.BaseUnit, &item.QuoteUnit)
	}

	// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
	// item to the Fields array.
	response.Fields = append(response.Fields, &item)

	return &response, nil
}

// GetPairs - This code is a function that is used to get pairs from a database. It takes in a context and a GetRequestPairs
// request, and returns a ResponsePair and an error. The code scans the row of the database and checks for any errors. If
// an error is found, the code will return an error response and the error itself. It then formats the symbol before
// appending the item to the Fields array in the response object. Finally, it returns the response and a nil error.
func (s *Service) GetPairs(_ context.Context, _ *pbstock.GetRequestPairs) (*pbstock.ResponsePair, error) {

	var (
		response pbstock.ResponsePair
	)

	row, err := s.Context.Db.Query(`select id, name, price, symbol, zone, base_decimal, quote_decimal, status from stocks`)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	for row.Next() {

		var (
			item pbstock.Pair
		)

		// This code is checking for an error when scanning the row of a database. If an error is found, the code will return
		// an error response and the error itself.
		if err := row.Scan(&item.Id, &item.Name, &item.Price, &item.BaseUnit, &item.QuoteUnit, &item.BaseDecimal, &item.QuoteDecimal, &item.Status); err != nil {
			return &response, err
		}
		item.Symbol = fmt.Sprintf("%v/%v", item.BaseUnit, &item.QuoteUnit)

		// The purpose of this code snippet is to check if the exchange (e) has the ratio of the given pair (pair). If so, the
		// ratio is assigned to the pair. The if statement checks is the ratio is returned by the getRatio() function, and if
		// it is, the ok variable will be true, and the ratio will be assigned to the pair.
		if ratio, ok := s.getRatio(item.GetBaseUnit(), item.GetQuoteUnit()); ok {
			item.Ratio = ratio
		}

		// This code is appending an item to the Fields array in the response object. The purpose of this code is to add an
		// item to the Fields array.
		response.Fields = append(response.Fields, &item)
	}

	return &response, nil
}

// GetPrice - This function is used to get the price of a stock from a database. It takes in two parameters (the base unit and quote
// unit) and returns the price of the stock as a response. The function makes use of the context and database to query
// the database for the price of the stock.
func (s *Service) GetPrice(_ context.Context, req *pbstock.GetRequestPrice) (*pbstock.ResponsePrice, error) {

	var (
		response pbstock.ResponsePrice
	)

	_ = s.Context.Db.QueryRow(`select price from stocks where symbol = $1 and zone = $2`, req.GetBaseUnit(), req.GetQuoteUnit()).Scan(&response.Price)

	return &response, nil
}

func (s *Service) SetOrder(ctx context.Context, req *pbstock.SetRequestOrder) (*pbstock.ResponseOrder, error) {

	// The purpose of this code is to declare two variables of type pbstock.ResponseOrder and pbstock.Order respectively.
	// Declaring the variables allows them to be used in the code.
	var (
		response pbstock.ResponseOrder
		order    pbstock.Order
	)

	// This code snippet checks if the request is authenticated by calling the Auth() method on the Context object. If the
	// authentication fails, the code returns an error.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to create a Service object that uses the context stored in the variable e. The Service
	// object is then assigned to the variable migrate.
	migrate := account.Service{
		Context: s.Context,
	}

	// This code is attempting to query a user from migrate using the provided authentication credentials (auth). If the
	// query fails, an error is returned.
	user, err := migrate.QueryUser(auth)
	if err != nil {
		return nil, err
	}

	// This code is checking the user's status. If the user's status is not valid (GetStatus() returns false), it returns an
	// error message informing the user that their account and assets have been blocked and instructing them to contact
	// technical support for any questions.
	if !user.GetStatus() {
		return &response, status.Error(748990, "your account and assets have been blocked, please contact technical support for any questions")
	}

	// This is setting the order quantity and value based on the request quantity and price.
	// The request quantity is used to set the order quantity, and the order value is calculated by multiplying the request quantity by the request price.
	order.Quantity = req.GetQuantity()
	order.Value = req.GetQuantity()

	// This is a switch statement that is used to evaluate the trade type of the request object. Depending on the trade
	// type, different actions can be taken. For example, if the trade type is "buy", the code may execute a certain set of
	// instructions to purchase the item, and if the trade type is "sell", the code may execute a different set of instructions to sell the item.
	switch req.GetTrading() {
	case proto.Trading_MARKET:

		// The purpose of this code is to set the price of the order (order.Price) to the market price of the requested base
		// and quote units, assigning, and price, which is retrieved from the "e.getMarket" function.
		order.Price = s.getMarket(req.GetBaseUnit(), req.GetQuoteUnit(), req.GetAssigning(), req.GetPrice())

		// This if statement is checking to see if the request is to buy something. If it is, it is calculating the quantity
		// and value of the order by dividing the quantity by the price.
		if req.GetAssigning() == proto.Assigning_BUY {
			order.Quantity, order.Value = decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float(), decimal.New(req.GetQuantity()).Div(order.GetPrice()).Float()
		}

	case proto.Trading_LIMIT:

		// The purpose of this code is to set the value of the order.Price variable to the value returned by the GetPrice()
		// method of the req object.
		order.Price = req.GetPrice()
	default:
		return &response, status.Error(82284, "invalid type trade position")
	}

	// The purpose of these lines of code is to assign the values of certain variables to the corresponding values from a
	// request object.  This is typically done when creating an order object from the request information.  In this case,
	// the values of the order object are set to the UserId, BaseUnit, QuoteUnit, Assigning, Status, and CreateAt variables
	// in the request object.  The Status is set to PENDING and the CreateAt is set to the current time.
	order.UserId = user.GetId()
	order.BaseUnit = req.GetBaseUnit()
	order.QuoteUnit = req.GetQuoteUnit()
	order.Assigning = req.GetAssigning()
	order.Status = proto.Status_PENDING
	order.Trading = req.GetTrading()
	order.CreateAt = time.Now().UTC().Format(time.RFC3339)

	// This code is checking for an error in the helperOrder() function and if one is found, it returns an error response
	// and calls the Context.Error() method with the error. The quantity variable is used to store the result of helperOrder(), which is used to complete the order.
	quantity, err := s.helperOrder(&order)
	if err != nil {
		return &response, err
	}

	// This is a conditional statement used to set a new order and check for any errors that might occur. If an error is
	// encountered, the statement will return a response and an Error context to indicate that an error has occurred.
	if order.Id, err = s.setOrder(&order); err != nil {
		return &response, err
	}

	// The switch statement is used to evaluate the value of the expression "order.GetAssigning()" and execute the
	// corresponding case statement. It is a type of conditional statement that allows a program to make decisions based on different conditions.
	switch order.GetAssigning() {
	case proto.Assigning_BUY:

		// This code snippet is likely a part of a function that processes an order. The purpose of the code is to use the
		// function "setAsset()" to set the base unit and user ID of the order to false. If an error occurs during the process,
		// the code will return the response and an error message.
		if err := s.setAsset(order.GetBaseUnit(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		// This code is checking the balance of a user and attempting to subtract the specified quantity from it. If the
		// operation is successful, it will continue with the program. If an error occurs, it will return an error response.
		if err := s.setBalance(order.GetQuoteUnit(), order.GetUserId(), quantity, proto.Balance_MINUS); err != nil {
			return &response, err
		}

		// The purpose of e.trade(&order, pbstock.Side_BID) is to replay a trade initiation with the given order and
		// side (BID). This is typically used when the trade is being initiated manually by an operator or trader. It allows
		// the trade to be replayed with the same parameters for accuracy and consistency.
		s.trade(&order, proto.Side_BID)

		break
	case proto.Assigning_SELL:

		// This code snippet is likely a part of a function that processes an order. The purpose of the code is to use the
		// function "setAsset()" to set the base unit and user ID of the order to false. If an error occurs during the process,
		// the code will return the response and an error message.
		if err := s.setAsset(order.GetQuoteUnit(), order.GetUserId(), false); err != nil {
			return &response, err
		}

		// This code is checking the balance of a user and attempting to subtract the specified quantity from it. If the
		// operation is successful, it will continue with the program. If an error occurs, it will return an error response.
		if err := s.setBalance(order.GetBaseUnit(), order.GetUserId(), quantity, proto.Balance_MINUS); err != nil {
			return &response, err
		}

		// The purpose of e.trade(&order, pbstock.Side_ASK) is to replay a trade initiation with the given order and
		// side (ASK). This is typically used when the trade is being initiated manually by an operator or trader. It allows
		// the trade to be replayed with the same parameters for accuracy and consistency.
		s.trade(&order, proto.Side_ASK)

		break
	default:
		return &response, status.Error(11588, "invalid assigning trade position")
	}

	// This statement is used to append an element to the "Fields" slice of the "response" struct. The element being
	// appended is the "order" struct.
	response.Fields = append(response.Fields, &order)

	return &response, nil
}

func (s *Service) GetOrders(ctx context.Context, req *pbstock.GetRequestOrders) (*pbstock.ResponseOrder, error) {

	// The purpose of this is to declare two variables: response and maps. The variable response is of type
	// pbstock.ResponseOrder, while the variable maps is of type string array.
	var (
		response pbstock.ResponseOrder
		maps     []string
	)

	// This code checks if the limit of the request is set to 0. If it is, then it sets the limit to 30. This is likely done
	// so that a request has a sensible limit, even if one wasn't specified.
	if req.GetLimit() == 0 {
		req.Limit = 30
	}

	// The purpose of this switch statement is to generate a SQL query with the correct assignment clause. Depending on
	// the value of req.GetAssigning(), the maps slice will be appended with the corresponding formatted string.
	switch req.GetAssigning() {
	case proto.Assigning_BUY:
		maps = append(maps, fmt.Sprintf("where assigning = %[1]d and type = %[2]d", proto.Assigning_BUY, proto.Type_STOCK))
	case proto.Assigning_SELL:
		maps = append(maps, fmt.Sprintf("where assigning = %[1]d and type = %[2]d", proto.Assigning_SELL, proto.Type_STOCK))
	default:
		maps = append(maps, fmt.Sprintf("where (assigning = %[1]d and type = %[3]d or assigning = %[2]d and type = %[3]d)", proto.Assigning_BUY, proto.Assigning_SELL, proto.Type_STOCK))
	}

	// This checks to see if the request (req) has an owner. If it does, the code after this statement will be executed.
	if req.GetOwner() {

		// This code is used to check the authentication of the user. The auth variable is used to store the authentication
		// credentials of the user, and the err variable is used to store any errors that might occur during the authentication
		// process. If an error occurs, the response and error is returned.
		auth, err := s.Context.Auth(ctx)
		if err != nil {
			return &response, err
		}

		// The purpose of this code is to append a formatted string to a slice of strings (maps). The string will include the
		// value of the auth variable and will be of the format "and user_id = '%v'", where %v is a placeholder for the value of auth.
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", auth))

		//	The code snippet is most likely within an if statement, and the purpose of the else if statement is to check if the
		//	user ID of the request is greater than 0. This could be used to check if the user is logged in or has an active
		//	session before performing a certain action.
	} else if req.GetUserId() > 0 {

		//This code is appending a string to a slice of strings (maps) which includes a formatted string containing the user
		//ID from a request object (req). This is likely part of an SQL query being built, with the user ID being used to filter the results.
		maps = append(maps, fmt.Sprintf("and user_id = '%v'", req.GetUserId()))
	}

	// The purpose of this switch statement is to add a condition to a query string based on the status of the request
	// (req.GetStatus()). Depending on the value of the status, a string is added to the maps slice using the fmt.Sprintf()
	// function. This string contains a condition that will be used in the query string.
	switch req.GetStatus() {
	case proto.Status_FILLED:
		maps = append(maps, fmt.Sprintf("and status = %d", proto.Status_FILLED))
	case proto.Status_PENDING:
		maps = append(maps, fmt.Sprintf("and status = %d", proto.Status_PENDING))
	case proto.Status_CANCEL:
		maps = append(maps, fmt.Sprintf("and status = %d", proto.Status_CANCEL))
	}

	// This code checks if the length of the base unit and the quote unit in the request are greater than 0. If they are, it
	// appends a string to the maps variable which includes a formatted SQL query containing the base and quote unit. This
	// is likely part of a larger SQL query used to search for data in a database.
	if len(req.GetBaseUnit()) > 0 && len(req.GetQuoteUnit()) > 0 {
		maps = append(maps, fmt.Sprintf("and base_unit = '%v' and quote_unit = '%v'", req.GetBaseUnit(), req.GetQuoteUnit()))
	}

	// The purpose of this code is to query the database to count the number of orders and total value of the orders in the
	// database. It then stores the count and volume in the response variable.
	_ = s.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count, sum(value) as volume from orders %s", strings.Join(maps, " "))).Scan(&response.Count, &response.Volume)

	// This statement is testing if the response from a user has a count that is greater than 0. If the response has a count
	// greater than 0, then something else will occur.
	if response.GetCount() > 0 {

		// This code is used to calculate the offset for a page of results in a request. It calculates the offset by
		// multiplying the limit (number of results per page) by the page number. If the page number is greater than 0, then
		// the offset is recalculated by multiplying the limit by one minus the page number.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to perform a SQL query on a database. It is used to select certain columns from the orders table
		// and to order them by the id in descending order. The limit and offset parameters are used to limit the number of
		// rows returned and to specify where in the result set to start returning rows from. The strings.Join function is used to join the "maps" parameter which is an array of strings.
		rows, err := s.Context.Db.Query(fmt.Sprintf("select id, assigning, price, value, quantity, base_unit, quote_unit, user_id, create_at, status from orders %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through the rows of the result set. The rows.Next() command will return true if the
		// iteration is successful and false if the iteration has reached the end of the result set. The loop will continue to
		// execute until the rows.Next() returns false.
		for rows.Next() {

			// The purpose of the above code is to declare a variable called item with the type pbstock.Order. This allows the
			// program to create an object of type pbstock.Order and assign it to the item variable.
			var (
				item pbstock.Order
			)

			// This code is scanning the rows returned from a database query and assigning the values to the variables in the item
			// struct. If an error is encountered during the scanning process, an error is returned.
			if err = rows.Scan(&item.Id, &item.Assigning, &item.Price, &item.Value, &item.Quantity, &item.BaseUnit, &item.QuoteUnit, &item.UserId, &item.CreateAt, &item.Status); err != nil {
				return &response, err
			}

			// The purpose of this statement is to add an item to the existing list of fields in a response object. The
			// response.Fields list is appended with the item, which is passed as an argument to the append function.
			response.Fields = append(response.Fields, &item)
		}

		// This code checks for an error in the rows object. If an error is found, the function will return a response and an error message.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

func (s *Service) CancelOrder(ctx context.Context, req *pbstock.CancelRequestOrder) (*pbstock.ResponseOrder, error) {
	//TODO implement me
	panic("implement me")
}
