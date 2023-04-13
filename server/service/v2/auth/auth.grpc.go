package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbauth"
	"github.com/pquerna/otp/totp"
	"github.com/tyler-smith/go-bip39"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"net"
	"net/mail"
	"strings"
)

// ActionSignup - This function is a signup action for a service. It receives a context, a request and returns a response and an error.
// It checks if the user is authorized and if not, denies the request. It also checks if the name, password, and email
// are of valid lengths before allowing the user account to be created. If the account already exists, it will return an
// error. If the account does not exist, it will create the account and return a response. Finally, it checks if the
// confirmation code is valid before allowing the user to confirm their account.
func (a *Service) ActionSignup(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	// The purpose of the code is to declare a variable named response of type pbauth.Response. This is a variable
	// declaration statement, and it allocates memory to store the data of type pbauth.Response.
	var (
		response pbauth.Response
	)

	// This code is checking if the incoming context contains a metadata key called "authorization". If it does, it is
	// returning an error with code 10004 and message "permission denied". This is likely being used to check if the user
	// has the proper authorization to access the resource being requested.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, status.Error(10004, "permission denied")
	}

	// The switch statement is used to check the value of req.GetSignup() and then execute code based on the result. The
	// switch statement is often used to execute different pieces of code depending on the value of the variable being tested.
	switch req.GetSignup() {
	case pbauth.Signup_ActionSignupAccount:

		// This code is checking to make sure that the length of the name sent in the request (req.GetName()) is at least 5
		// characters long. If the name is not at least 5 characters long, then it will return an error with status code 19522
		// and a message saying "the name must be at least 5 characters long".
		if len(req.GetName()) < 5 {
			return &response, status.Error(19522, "the name must be at least 5 characters long")
		}

		// This if statement checks is the password provided is at least 8 characters long. If it is not, an error is returned
		// indicating that the password must be at least 8 characters long. This is an important security measure to ensure
		// that passwords are sufficiently complex.
		if len(req.GetPassword()) < 8 {
			return &response, status.Error(14563, "the password must be at least 8 characters long")
		}

		// This code is checking if the email address provided in the request is valid. If it is not valid, an error is returned.
		if _, err := mail.ParseAddress(req.GetEmail()); err != nil {
			return &response, err
		}

		// This code is querying a database for an account with a specified email address. The row variable is used to store
		// the result of the query, and err is used to store any potential errors that might have been encountered while
		// querying the database. Finally, the defer statement is used to ensure that the row is closed after the query is finished.
		row, err := a.Context.Db.Query("select id from accounts where email = $1", req.GetEmail())
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The if statement is used to check if the row.Next() method returns true or false. If it is false, the code within
		// the if statement will execute.
		if !row.Next() {

			// The purpose of the code above is to hash the password and secret combination. The hashed variable is a new instance
			// of the sha256 hashing algorithm. The Write method takes a byte slice as an argument, which is created by formatting
			// the password and secret using the fmt.Sprintf function.
			hashed := sha256.New()
			hashed.Write([]byte(fmt.Sprintf("%v-%v", req.GetPassword(), a.Context.Secrets[0])))

			// This code is used to generate a new entropy value for a BIP39 wallet. Entropy is a random string of data used to
			// generate a seed phrase for the wallet. The code checks for any errors and will return an error if one occurs.
			entropy, err := bip39.NewEntropy(128)
			if err != nil {
				return &response, err
			}

			// This code is used to check if a user has already been registered with the same email address. If the user has
			// already been registered, an error is returned. Otherwise, the user is registered by inserting the name, email,
			// password and entropy into the accounts table.
			if _, err := a.Context.Db.Exec("insert into accounts (name, email, password, entropy) values ($1, $2, $3, $4)", req.GetName(), req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)), entropy); err != nil {
				return &response, status.Error(15316, "a user with this address has already been registered before")
			}

		} else {
			return &response, status.Error(64401, "a user with this email address is already registered")
		}

		break
	case pbauth.Signup_ActionSignupCode:

		// This code is used to set a code for the given email address. If there is an error while setting the code, the error
		// is returned and the process is stopped.
		code, err := a.writeCode(req.GetEmail())
		if err != nil {
			return &response, err
		}

		// This code is checking for errors when executing an update query on the "accounts" table in a database. The query is
		// updating the "email_code" field with a given code for a specified email address, where the status is false. If an
		// error occurs, the code returns an error response.
		if _, err = a.Context.Db.Exec("update accounts set email_code = $3 where email = $1 and status = $2;", req.GetEmail(), false, code); err != nil {
			return &response, err
		}

		break
	case pbauth.Signup_ActionSignupConfirm:

		// This code is checking the length of a value called req.GetEmailCode(), and if the length is not equal to 6, it
		// returns an error message stating that the code must be 6 numbers. This is likely a part of a larger program that
		// checks the validity of a code with 6 numbers.
		if len(req.GetEmailCode()) != 6 {
			return &response, status.Error(14773, "the code must be 6 numbers")
		}

		// This code is trying to get the id of an account from a database where the email, email code, and status all match
		// the given values. It is using a parameterized query to prevent SQL injection and make sure that the query is valid.
		// If an error occurs, it will return a response and an error. Finally, the row is closed when the function exits.
		row, err := a.Context.Db.Query("select id from accounts where email = $1 and email_code = $2 and status = $3", req.GetEmail(), req.GetEmailCode(), false)
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The if statement with row.Next() checks if there is another row in the result set of a query. If there is, the code
		// within the if statement block will be executed.
		if row.Next() {

			//This code is part of a function that is updating the status of an account in a database. Specifically, it is
			//updating the status of an account associated with the email provided in the request.
			// The `if` statement is checking for an error during the execution of the SQL query. If an error occurs, it is returned in the response, and the function is exited using the `a.Context.Error` method.
			if _, err := a.Context.Db.Exec("update accounts set status = $2 where email = $1;", req.GetEmail(), true); err != nil {
				return &response, err
			}

		} else {
			return &response, status.Error(58042, "this code is invalid")
		}

		break
	default:
		return &response, status.Error(60001, "invalid input parameter")
	}

	return &response, nil
}

// ActionSignin - This function is a service that is used to sign in a user. It checks the incoming context metadata to see if the user
// is authorized and then checks the sign in type (Account, Code, or Confirm). Depending on the sign in type, it will
// perform different actions such as checking the password, setting a code, and confirming the code. It also checks for a
// valid two-factor authentication code if the user has secure two-factor authentication enabled. If the sign in is
// successful, it will send a login email to the user and return a access and refresh token.
func (a *Service) ActionSignin(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	// The purpose of this line of code is to declare a variable named 'response' of type 'pbauth.Response'. This allows the
	// program to use this variable as an object or struct of type 'pbauth.Response'.
	var (
		response pbauth.Response
	)

	// This code is checking a context (ctx) to see if it contains a metadata element with the key "authorization". If the
	// metadata element exists, then the code returns an "Error" with status code 10004 and the message "permission denied".
	// This is likely part of a larger authorization process where the code is determining if the user has the necessary
	// permissions to access a resource.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, status.Error(10004, "permission denied")
	}

	// The purpose of the code above is to create a hash of the password and secret combination using the SHA256 hash
	// algorithm. The hashed variable holds the hash which can then be used for authentication or other security purposes.
	hashed := sha256.New()
	hashed.Write([]byte(fmt.Sprintf("%v-%v", req.GetPassword(), a.Context.Secrets[0])))

	// This code snippet is part of a function that is checking for a sign in. The switch statement is used to check the
	// value of the "req.GetSignin()" request. Depending on the value, the code block following the switch statement will be
	// executed. This allows the function to act differently based on the value of the "req.GetSignin()" request.
	switch req.GetSignin() {
	case pbauth.Signin_ActionSigninAccount:

		// This code is used to query a database for a given email and password. The row variable is used to hold the results
		// of the query. To err variable is used to check for any errors that occur when executing the query. If an error
		// occurs, the function returns an error response. To defer row.Close() statement is used to close the database
		// connection after the query has been executed.
		row, err := a.Context.Db.Query("select id from accounts where email = $1 and password = $2", req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)))
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// This if statement checks to see if the row in the database has the next value. If it does not, then it will return a
		// response with an error code of 48512 and a corresponding error message that states "the email address or password
		// was entered incorrectly". This is useful in cases where the user has entered incorrect credentials and the
		// application needs to inform them of that fact.
		if !row.Next() {
			return &response, status.Error(48512, "the email address or password was entered incorrectly")
		}

		break
	case pbauth.Signin_ActionSigninCode:

		// This code is setting a code for a given request. The "setCode" function is called with the email from the request,
		// and if an error is returned, the error is handled and the response is returned.
		code, err := a.writeCode(req.GetEmail())
		if err != nil {
			return &response, err
		}

		// This code is part of a function that is updating an account in a database. The code is specifically updating the
		// email_code field in the accounts table of the database, where the email and password match the supplied parameters.
		// The code is also using $3 to set the email_code field to the value of code. This could be used to verify a user's
		// email address or to reset a user's password.
		if _, err = a.Context.Db.Exec("update accounts set email_code = $3 where email = $1 and password = $2;", req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)), code); err != nil {
			return &response, err
		}

		break
	case pbauth.Signin_ActionSigninConfirm:

		// This code checks the length of the "req.GetEmailCode()" to make sure it is 6 numbers long. If it is not, it returns
		// an error code 14773 and the message "the email code must be 6 numbers". This is likely used to ensure that the email
		// code entered is the correct length.
		if len(req.GetEmailCode()) != 6 {
			return &response, status.Error(14773, "the email code must be 6 numbers")
		}

		// This code is querying an account table in a database. It is attempting to find an account with the given email,
		// email code, and a hashed password. The row variable holds the result of the query. If an error occurs it will return
		// an error message. The defer statement will close the row when the function returns.
		row, err := a.Context.Db.Query("select id, factor_secret, factor_secure from accounts where email = $1 and email_code = $2 and password = $3", req.GetEmail(), req.GetEmailCode(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)))
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The "if row.Next()" statement is used to check if the row has a next row. It is used to iterate through all the rows
		// of a database query result. If the row has a next row, it will execute the code within the "if" statement block.
		if row.Next() {

			// The purpose of the above code is to declare a structure called "params" that contains variables of different data
			// types including strings, an integer, and a boolean. It also declares a variable called "migrate" which assigns a
			// Migrate object to it and sets its context to the context of the application. This code is likely used to set up the
			// parameters and objects needed to make some sort of database query.
			var (
				params struct {
					ip, secret string
					id         int64
					secure     bool
				}
				migrate = query.Migrate{
					Context: a.Context,
				}
			)

			// This if statement is used to scan the row data and assign it to the params.id, params.secret, and params.secure
			// variables. If an error occurs, it will return the response and the Context.Error to handle the error.
			if err := row.Scan(&params.id, &params.secret, &params.secure); err != nil {
				return &response, err
			}

			// The purpose of this code snippet is to check if a two-factor authentication (2FA) code is valid. If the code is
			// invalid, an error response is returned. The if statement checks that the params.secure is true, then it uses the
			// totp.Validate() function to check if the provided code matches the secret. If the code does not match, an error is returned.
			if params.secure {
				if !totp.Validate(req.GetFactorCode(), params.secret) {
					return &response, status.Error(115654, "invalid 2fa secure code")
				}
			}

			// This code is used to obtain a token and check for any errors that occurred while attempting to obtain it. If an
			// error is found, the response is returned with the error.
			token, err := a.ReplayToken(params.id)
			if err != nil {
				return &response, err
			}

			// This code is checking if there is any incoming context in the metadata. If there is, the code is setting the
			// variable meta to the value of the incoming context and setting the ok variable to true.
			if meta, ok := metadata.FromIncomingContext(ctx); ok {

				// The purpose of this statement is to set the user agent for a grpc gateway. The statement assigns the user agent
				// string from the meta.Get() function to the agent variable. This allows the grpc gateway to identify the user agent
				// when processing requests.
				agent := help.MetaAgent(meta.Get("grpcgateway-user-agent")[0])

				// This code is attempting to marshal a slice of strings into JSON, taking the agent name and converting it to
				// lowercase and the agent version. If an error is encountered, the error is returned with the response.
				browser, err := json.Marshal([]string{strings.ToLower(agent.Name), agent.Version})
				if err != nil {
					return &response, err
				}

				// This code is attempting to obtain the IP address of a peer from a given context. The first "if" statement checks
				// if the peer is available in the context, and if so, assigns it to the variable "mp". The second "if" statement
				// checks if the address of the peer is a TCP address, and if so, assigns the IP address to the variable "params.ip".
				// If the address is not a TCP address, the code assigns the full address to "params.ip".
				if mp, ok := peer.FromContext(ctx); ok {
					if tcpAddr, ok := mp.Addr.(*net.TCPAddr); ok {
						params.ip = tcpAddr.IP.String()
					} else {
						params.ip = mp.Addr.String()
					}
				}

				// This code is inserting data into the 'actions' table of a database. It is taking the user ID, operating system,
				// device, browser, and IP address from the given parameters and inserting them into the database. The
				// strings.ToLower() function ensures that the OS is stored in lowercase. The code also checks for any errors that
				// may occur during the insertion and returns an error response if necessary.
				if _, err = a.Context.Db.Exec("insert into actions (user_id, os, device, browser, ip) values ($1, $2, $3, $4, $5)", params.id, strings.ToLower(agent.OS), agent.Device, browser, params.ip); err != nil {
					return &response, err
				}
			}

			// This code is part of a function that is updating a record in a database. The purpose of this code is to update the
			// "accounts" table in the database and set the "email_code" column to an empty string ("") for the record that has
			// the matching "email" column value as the value provided in the "req" parameter. If any errors occur while executing
			// the database update query, the function will return an error response.
			if _, err = a.Context.Db.Exec("update accounts set email_code = $2 where email = $1;", req.GetEmail(), ""); err != nil {
				return &response, err
			}

			// This code is used to send a login email to a user. The parameters passed in are the user's ID and the type of email
			// being sent. The final parameter is set to nil, indicating that there is no additional data to be included in the email.
			go migrate.SendMail(params.id, "login", nil)

			// This code assigns the AccessToken and RefreshToken from the token object to the response object. This allows the
			// AccessToken and RefreshToken to be stored in the response object for future use.
			response.AccessToken, response.RefreshToken = token.AccessToken, token.RefreshToken

		} else {
			return &response, status.Error(58042, "this code is invalid")
		}

		break
	default:
		return &response, status.Error(60001, "invalid input parameter")
	}

	return &response, nil
}

// ActionReset - This function is a method of a service that is used to reset a user's password. It checks the incoming context for
// authorization, sets a code for the user, confirms the code, and then sets a new password for the user. It also sends
// an email to the user with the new password.
func (a *Service) ActionReset(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	// The purpose of this code is to initialize a response variable of type pbauth.Response, a migrated variable of type
	// query.Migrate, and a q variable of type query.Query. These variables will be used in the program for different purposes.
	var (
		response pbauth.Response
		migrate  = query.Migrate{
			Context: a.Context,
		}
		q query.Query
	)

	// This code is checking if a key called "authorization" is present in the incoming context (ctx) and if it is, it is
	// returning an error message with a status code of 10004 ("permission denied"). This code is likely used to check if
	// the user has permission to access the requested resource. If the key is not present, the code allows the user to continue.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, status.Error(10004, "permission denied")
	}

	// The switch statement in this example is used to check the value of the req.GetReset_() expression, which is a method
	// call that returns a value. Depending on the value returned, different actions may be taken.
	switch req.GetReset_() {
	case pbauth.Reset_ActionResetAccount:

		// This code is querying a database to find the id associated with a given email address. The row variable is set with
		// the result of the query and err is set with any error that occurred while executing the query. If an error occurred,
		// the context.Error() method is called which will return the response and the error. The row variable is then closed
		// with the defer keyword which ensures that it is closed at the end of the function even if an error occurred.
		row, err := a.Context.Db.Query("select id from accounts where email = $1", req.GetEmail())
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// The code snippet is used to check if a row with a given email exists in a database. If there is no row, then an
		// error is returned indicating that "there is no user with this email".
		if !row.Next() {
			return &response, status.Error(48512, "there is no user with this email")
		}

		break
	case pbauth.Reset_ActionResetCode:

		// This code is setting a code based on the email provided in the request (req.GetEmail()) and checking for an error.
		// If an error is found, it is returned with the response and the context error is triggered.
		code, err := a.writeCode(req.GetEmail())
		if err != nil {
			return &response, err
		}

		// This code is used to update a row in the accounts table of a database. Specifically, it sets the email_code field to
		// the value of the variable 'code' for the account with the email address given in the variable 'req'.
		// If the operation is unsuccessful, the code returns an error to the caller.
		if _, err = a.Context.Db.Exec("update accounts set email_code = $2 where email = $1;", req.GetEmail(), code); err != nil {
			return &response, err
		}

		break
	case pbauth.Reset_ActionResetConfirm:

		// This code checks to make sure that the email code provided is 6 numbers long. If it is not 6 numbers long, it will
		// return an error message indicating that the code must be 6 numbers.
		if len(req.GetEmailCode()) != 6 {
			return &response, status.Error(14773, "the email code must be 6 numbers")
		}

		// This code is used to query the database for entries that have a specific email and email code. The row variable
		// stores the result of the query, which is then checked for errors. If there is an error, the error is returned.
		// Otherwise, the deferred row.Close() function is called to close the row once it's no longer needed.
		row, err := a.Context.Db.Query("select id from accounts where email = $1 and email_code = $2", req.GetEmail(), req.GetEmailCode())
		if err != nil {
			return &response, err
		}
		defer row.Close()

		// This code checks if the row has a next value. If it does not, it will return an error status code 58042 and message
		// "this code is invalid".
		if !row.Next() {
			return &response, status.Error(58042, "this code is invalid")
		}

		break
	case pbauth.Reset_ActionResetPassword:

		// This code checks the length of the email code to make sure it is 6 characters long. If the length is not 6
		// characters, it will return an error.
		if len(req.GetEmailCode()) != 6 {
			return &response, status.Error(14773, "the code must be 6 numbers")
		}

		// The purpose of the code is to generate a new 15 character password with false indicating that the password should
		// not include special characters.
		password := help.NewCode(15, false)

		// The code above is used to create a hash of the user's password combined with a unique secret string.  This is a
		// security measure to ensure that a user's password is not stored in plaintext, but is instead stored as a hashed
		// string which is much more secure.
		hashed := sha256.New()
		hashed.Write([]byte(fmt.Sprintf("%v-%v", password, a.Context.Secrets[0])))

		// This code is updating the password and email code of an account in a database using the values provided in the
		// request (req). The query is using the email address, email code, and hashed password to update the record. If the
		// query fails, an error is returned.
		if err := a.Context.Db.QueryRow("update accounts set password = $3, email_code = $4 where email = $1 and email_code = $2 returning id;", req.GetEmail(), req.GetEmailCode(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)), "").Scan(&q.Id); err != nil {
			return &response, err
		}

		// The purpose of this code is to email a user's Id with a new password. The variable "password" is the new
		// password that is being sent. The function migrate.SendMail allows the code to send an email with the new password to
		// the user's I'd.
		go migrate.SendMail(q.Id, "new_password", password)

		break
	default:
		return &response, status.Error(60001, "invalid input parameter")
	}

	return &response, nil
}

// SetLogout - This function is part of the Service struct and is used to log out a user from the system. It sets the email code in
// the accounts table to an empty string and deletes the refresh token from the Redis Client. It also checks for
// permission before executing the logout.
func (a *Service) SetLogout(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	// This line of code declares a variable called response of type pbauth.Response. The purpose of this line is to create
	// a variable that will store a value of type pbauth.Response.
	var (
		response pbauth.Response
	)

	// This code is used to check for authorization in an incoming context. It uses the FromIncomingContext() method from
	// the metadata package to retrieve the metadata from the context. If the metadata is not found or the authorization
	// field is not equal to 1 or is nil, an error is returned with the status code 10004 and the message "permission denied".
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok && len(meta["authorization"]) != 1 && meta["authorization"] == nil {
		return &response, status.Error(10004, "permission denied")
	}

	// The purpose of this code is to delete a key-value pair from a Redis database using the background context. The key is
	// given by the req.GetRefresh() method.
	a.Context.RedisClient.Del(context.Background(), req.GetRefresh())

	// This code is used to update the email_code field in the accounts table in a database. The req.GetEmail() is used to
	// get the email address from a request. The "" is used as the new value for the email_code field. If there is an error
	// during the execution of the update query, it will return an error and the response to the caller.
	if _, err := a.Context.Db.Exec("update accounts set email_code = $2 where email = $1;", req.GetEmail(), ""); err != nil {
		return &response, err
	}

	return &response, nil
}

// GetRefresh - This function is part of a Service struct and is used to get a refreshed Access Token and Refresh Token. It takes in a
// context object and a Request object as parameters, and returns a Response object and an error if there is one.
// The function first checks that the incoming context contains the necessary metadata. It then attempts to retrieve the session from Redis using the Refresh Token in the Request object.
// It then unmarshals the session and compares the Access Token in the session to the Access Token in the context's metadata.
// If they match, it calls the ReplayToken function to get a new Access Token and Refresh Token, which it sets in the Response object and returns.
func (a *Service) GetRefresh(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	// The purpose of this code is to declare two variables, response and serialize, of type pbauth.Response and
	// pbauth.Response_Session, respectively. The pbauth package is most likely related to authentication or authorization,
	// so these variables may be used to store a response from an authentication or authorization operation.
	var (
		response  pbauth.Response
		serialize pbauth.Response_Session
	)

	// This code checks if the incoming context (ctx) has any metadata associated with it and, if not, returns an error to
	// the caller (10411). If the metadata exists, it checks if the "authorization" key is present and has a length of 1. If
	// either of these conditions are not met, an error is returned.
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok && len(meta["authorization"]) != 1 && meta["authorization"] == nil {
		return &response, status.Error(10411, "missing metadata")
	}

	// This code is retrieving a session from a RedisClient using the Get() function. It is passing in the
	// context.Background() as the first argument and the req.GetRefresh() as the second argument. If there is an error in
	// retrieving the session, the code returns the response and an error.
	session, err := a.Context.RedisClient.Get(context.Background(), req.GetRefresh()).Bytes()
	if err != nil {
		return &response, err
	}

	// This code is attempting to unmarshal a session object from a msgpack-encoded byte array, and assign it to the
	// variable serialize. If an error occurs while attempting to unmarshal the byte array, it will return an error response.
	err = msgpack.Unmarshal(session, &serialize)
	if err != nil {
		return &response, err
	}

	// This code is checking if the authorization token provided in the meta field matches the serialized access token. If
	// the two tokens do not match, an error is returned.
	token := strings.Split(meta["authorization"][0], "Bearer ")[1]
	if serialize.AccessToken != token {
		return &response, status.Error(31754, "session not found")
	}

	// The purpose of this code is to generate a replay token associated with the given subject. If an error is encountered
	// while generating the token, the function returns an error.
	replayToken, err := a.ReplayToken(serialize.Subject)
	if err != nil {
		return nil, err
	}

	// This code is used to assign the values of the AccessToken and RefreshToken returned by the
	// replayToken.GetAccessToken() and replayToken.GetRefreshToken() functions to the response object. This is necessary in
	// order to store the tokens in the response object so they can be used by the application later.
	response.AccessToken, response.RefreshToken = replayToken.GetAccessToken(), replayToken.GetRefreshToken()

	return &response, nil
}

// GetSecure - This function is used to get a secure factor from an incoming context. It takes a context and a request as input, then
// checks the metadata from the incoming context. If the authorization key is not included, it returns an error.
// Otherwise, it hashes the password with the first secret found in the context and queries the database to get the
// factor secure. Finally, it returns the response or an error if one occurs.
func (a *Service) GetSecure(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	// The variable 'response' is of type 'pbauth.Response', which is used to store the response data returned from an
	// authentication request. The purpose of this variable is to store the authentication response data so that it can be
	// accessed and used.
	var (
		response pbauth.Response
	)

	// This code is used to check the incoming context for a specific key, "authorization". If the key is present, the code
	// returns an error with an accompanying status code of 10004. This is likely part of a larger authorization or
	// authentication process, where a user must provide the correct credentials in order to access a particular resource.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, status.Error(10004, "permission denied")
	}

	// The purpose of this code is to hash the password and context secret using the sha256 algorithm. The hashed value is
	// then stored in the "hashed" variable. This is often used to securely store and authenticate passwords, as sha256 is a
	// secure hashing algorithm.
	hashed := sha256.New()
	hashed.Write([]byte(fmt.Sprintf("%v-%v", req.GetPassword(), a.Context.Secrets[0])))

	// This code is querying a database for a specific row and scanning the result into a response object. The query is
	// looking for a row with an email and password that matches the given parameters. The purpose of the code is to fetch
	// the factor_secure value from the database for the given account.
	if err := a.Context.Db.QueryRow("select factor_secure from accounts where email = $1 and password = $2", req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil))).Scan(&response.FactorSecure); err != nil {
		return &response, err
	}

	return &response, nil
}
