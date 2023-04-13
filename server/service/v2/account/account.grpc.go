package account

import (
	"context"
	"encoding/json"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/pquerna/otp/totp"
	"google.golang.org/grpc/status"
)

// SetUser - This function is used to set a user's information manually. It takes in a context and a request containing the user's
// information as parameters. It then performs authentication to ensure the user is allowed to make the changes. It then
// sets the sample and password information specified in the request. Finally, it queries the user's information and
// returns it in the response.
func (a *Service) SetUser(ctx context.Context, req *pbaccount.SetRequestUser) (*pbaccount.ResponseUser, error) {

	// The variable 'response' is of type 'pbaccount.ResponseUser', which is likely a protocol buffer structure that is used
	// to store information about a user. This variable is used to store the user information that has been received from
	// the server.
	var (
		response pbaccount.ResponseUser
	)

	// This code is authenticating a user in order to access a resource. The 'a.Context.Auth(ctx)' function is used to
	// authenticate the user, and if there is an error, the return an error response.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This if statement is used to check if the length of the request's sample is greater than 0. If it is, it will attempt
	// to set the sample with the given authentication, and if an error occurs, it will return an error response.
	if len(req.GetSample()) > 0 {
		if err := a.writeSample(auth, req.GetSample()); err != nil {
			return &response, err
		}
	}

	// This is an if statement that is checking the length of two variables, req.GetOldPassword() and req.GetNewPassword().
	// If the length of both of these is greater than 0, then the code in the statement will execute. This if statement is
	// likely being used to ensure that the user has provided both an old and a new password before some action is taken.
	if len(req.GetOldPassword()) > 0 && len(req.GetNewPassword()) > 0 {

		// The purpose of this code is to set a new password for an individual using the old password as a verification. If
		// there is an error in setting the new password, the code will return an error.
		if err := a.writePassword(auth, req.GetOldPassword(), req.GetNewPassword()); err != nil {
			return &response, err
		}
	}

	// This code is querying a user using the auth parameter, and then checking for any errors that may be returned. If an
	// error is returned, it is returned with the response and an error is sent back through the context.
	user, err := a.QueryUser(auth)
	if err != nil {
		return &response, err
	}

	// This code is likely to be part of a function that is creating a response object. The purpose of this code is to add a
	// new element, user, to the existing list of Fields in the response object.
	response.Fields = append(response.Fields, user)

	return &response, nil
}

// GetUser - This function is a method of a Service struct, and it is used to get a User from an authentication context. It takes
// in a context.Context and a *pbaccount.GetRequestUser as arguments, and returns a *pbaccount.ResponseUser and an error.
// It authenticates a user in the context, queries the user, and then appends the user to the response fields, before
// returning it.
func (a *Service) GetUser(ctx context.Context, _ *pbaccount.GetRequestUser) (*pbaccount.ResponseUser, error) {

	// The purpose of this code is to declare two variables, response and err. The variable response is of type
	// pbaccount.ResponseUser and the variable err is of type error. These variables will be used in the code that follows.
	var (
		response pbaccount.ResponseUser
		err      error
	)

	// This code is used to authenticate a user within a given context (ctx). It creates a variable called auth which stores
	// the authentication information, and an err variable which stores any errors that occur. If an error occurs, the error
	// is returned and the function is exited.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is used to query a user based on an authentication parameter, and return the response to the user. If an
	// error occurs while querying the user, the code returns an error response to the user.
	user, err := a.QueryUser(auth)
	if err != nil {
		return &response, err
	}

	//This line of code sets the value of the FactorSecret property of the user object to an empty string. This property is
	//likely used to store a secret code or token that is used in authentication or security processes.
	user.FactorSecret = ""

	// This line of code is adding the user variable to the end of the response.Fields slice. The purpose of this is to add
	// the user to a list of fields that the response contains.
	response.Fields = append(response.Fields, user)

	return &response, nil
}

// GetActions - This function is a part of a service which is used to get a list of actions associated with a user. The context
// parameter is used to provide access to a user's authentication information and the req parameter is used to provide
// the page and limit information for the list. The function then queries the database to get the count of actions
// associated with the user and the list of actions. It scans the list of actions, unpacks the browser json data, and
// appends the action item to the response. Finally, it returns the response with the list of actions and the count of
// actions associated with the user.
func (a *Service) GetActions(ctx context.Context, req *pbaccount.GetRequestActions) (*pbaccount.ResponseActions, error) {

	// The purpose of the above code is to declare two variables: response and browser. The first variable, response, is of
	// the type pbaccount.ResponseActions, which is likely a custom type. The second variable, browser, is of the type byte
	// array, which can be used to store raw data.
	var (
		response pbaccount.ResponseActions
		browser  []byte
	)

	// This code is authenticating a user who is trying to access a resource. The variable 'auth' will contain the
	// authentication information for the user, and the variable 'err' will contain any error that occurs during the
	// authentication process. If an error occurs, the code will return an error response.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is used to query the database and check the number of actions associated with a given user. The QueryRow
	// function runs a query and scans the results into the variable "response.Count". If an error occurs, the code returns
	// the response and an error message.
	if _ = a.Context.Db.QueryRow("select count(*) from actions where user_id = $1", auth).Scan(&response.Count); response.Count > 0 {

		// The purpose of this code is to calculate the offset for a page of data from a request. The offset is used to
		// determine the starting point of a query when selecting data from a database. Specifically, it calculates the offset
		// based on the limit and page number provided in the request. If the page number is greater than 0, the offset is
		// calculated by multiplying the limit by the page number minus 1.
		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		// This code is used to query a database for data. The query is executed with the given parameters (auth,
		// req.GetLimit(), offset). If the query is successful, the rows are returned and stored for further use. If the query
		// fails, an error is returned and the code exits. The rows.Close() function is used to close the connection to the database.
		rows, err := a.Context.Db.Query("select id, os, device, ip, browser, create_at from actions where user_id = $1 order by id desc limit $2 offset $3", auth, req.GetLimit(), offset)
		if err != nil {
			return &response, err
		}
		defer rows.Close()

		// The for loop is used to iterate through a set of rows returned by a database query. The rows.Next() method is used
		// to move the cursor to the next row in the result set, and the loop will execute until there are no more rows to
		// iterate through.
		for rows.Next() {

			// The variable "item" is used to store a data type of types.Action. It is used to declare a variable of a
			// specific type so that it can be used in the code.
			var (
				item types.Action
			)

			// This code is used to scan the rows of a database query and store their values into the variables item.Id, item.Os,
			// item.Device, item.Ip, browser, and item.CreateAt. If the scan is successful, the function will return &response,
			// otherwise it will return an error.
			if err := rows.Scan(&item.Id, &item.Os, &item.Device, &item.Ip, &browser, &item.CreateAt); err != nil {
				return &response, err
			}

			// This code is used to unmarshal a JSON object into a struct. The if statement checks for an error when calling
			// json.Unmarshal and if there is an error it returns the response and an error message.
			if err := json.Unmarshal(browser, &item.Browser); err != nil {
				return &response, err
			}

			// The purpose of this code is to add a new item to the existing list of response fields. The 'append' function takes
			// in the current list of response fields and adds the new 'item' to the end of the list.
			response.Fields = append(response.Fields, &item)
		}

		// This code is used to check for errors with the rows object. The if statement will check if an error is present in
		// the rows object, and if there is an error, the code will return the response and an error with the context.
		if err = rows.Err(); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// SetFactor - This function is used to set a user's secure factor. It takes in a context and request object and verifies the code
// provided in the request. If the code is valid, it sets the secure factor to either true or false and updates the
// user's secret if necessary. It then returns a response and any errors that might have occurred.
func (a *Service) SetFactor(ctx context.Context, req *pbaccount.SetRequestFactor) (*pbaccount.ResponseFactor, error) {

	// The purpose of the above code is to declare three variables - response of type pbaccount.ResponseFactor, secret of
	// type string and secure of type bool. These variables are used to store values such as the response from a secure
	// account, a secret string and a boolean value to indicate whether a secure account is enabled.
	var (
		response pbaccount.ResponseFactor
		secret   string
		secure   bool
	)

	// This code is used to authenticate a user. The variable 'auth' stores the authentication information, and the 'err'
	// variable stores any potential errors that might occur during authentication. If an error occurs, the code returns an error response.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code attempts to query the user using the authentication (auth) provided. If there is an error during the
	// querying process, the return statement will pass back an error to the calling code and the Context.Error() method
	// will be called to return the error.
	user, err := a.QueryUser(auth)
	if err != nil {
		return &response, err
	}

	// This code is used to determine which secret value to use. If the user has factor secure enabled, then it will use the
	// user's factor secret value. If not, then it will use the secret value from the request and set secure to true.
	if user.GetFactorSecure() {
		secret = user.GetFactorSecret()
		secure = false
	} else {
		secret = req.GetSecret()
		secure = true
	}

	// This code is performing a TOTP (Time-based One-Time Password) validation. The purpose of this code is to check if the
	// code provided by the user (req.GetCode()) matches the secret code and is valid. If the code does not match the secret
	// code, the code returns an error message ("invalid secure code") with a status code (115654).
	if !totp.Validate(req.GetCode(), secret) {
		return &response, status.Error(115654, "invalid secure code")
	}

	// The purpose of this code is to check if user authentication is enabled, and if so, set the secret variable to an
	// empty string. This ensures that the user is secure and that no secret information is accessible.
	if user.GetFactorSecure() {
		secret = ""
	}

	// This if statement is checking to see if the Exec() method of the Db object from the Context object in the variable a
	// returns an error. If it does, then the statement returns a response and the error. This statement is likely part of a
	// function that is updating the accounts table in a database.
	if _, err := a.Context.Db.Exec("update accounts set factor_secure = $1, factor_secret = $2 where id = $3;", secure, secret, auth); err != nil {
		return &response, err
	}

	return &response, nil
}

// GetFactor - This function is a method of a Service struct. It is used to generate a two-factor authentication key for a user. It
// first checks the user's authentication, then checks if the user has already enabled two-factor authentication. If not,
// it generates a new key and returns the secret and URL associated with it as part of a ResponseSecure struct.
func (a *Service) GetFactor(ctx context.Context, _ *pbaccount.GetRequestFactor) (*pbaccount.ResponseFactor, error) {

	// The purpose of this code is to declare two variables: one of type pbaccount.ResponseFactor, and one of type error.
	// This is used to store the response of a function call, as well as any errors that may have occurred during the call.
	var (
		response pbaccount.ResponseFactor
		err      error
	)

	// This code checks for a valid authentication and if there is an error with the authentication it will return an error
	// response. It is used to ensure that only authorized users can access the requested resource.
	auth, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is used to query a user by authentication. The variable "user" is assigned the result of the query, and if
	// there is an error, the "err" variable will be assigned the error and the function will return a response and the error.
	user, err := a.QueryUser(auth)
	if err != nil {
		return &response, err
	}

	// This if statement is checking if the function GetFactorSecure returns a false boolean value. If it does, the code in
	// the if statement block will be executed. This could be used to determine if a user has enabled two-factor
	// authentication on their account or not.
	if !user.GetFactorSecure() {

		// This code is generating a Time-based One-Time Password (TOTP) for a user. The totp.GenerateOpts is a set of options
		// that are given to the Generate method. This includes the Issuer, which is the name of the service the user is
		// authenticating to, AccountName which is the user's email address, and SecretSize which is the length of the secret
		// key being generated. If an error is encountered, it is returned and the function is exited.
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "Envoys Exchange",
			AccountName: user.GetEmail(),
			SecretSize:  15,
		})
		if err != nil {
			return &response, err
		}

		// The purpose of this code is to set the response.Secret and response.URL variables to the values of the key.Secret()
		// and key.URL() functions, respectively. This code is likely used to return an API key to a user after they have
		// requested one. The key.Secret() and key.URL() functions are likely used to generate a random API key and a URL
		// associated with it, respectively.
		response.Secret = key.Secret()
		response.Url = key.URL()
	}

	return &response, nil
}
