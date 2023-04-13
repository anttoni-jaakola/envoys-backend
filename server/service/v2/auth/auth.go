package auth

import (
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbauth"
	"github.com/golang-jwt/jwt/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/net/context"
	"time"
)

// Service - The type Service struct is a data structure which holds a pointer to an instance of the assets.Context. This is used
// to store information related to the service, and provide access to context-specific data.
type Service struct {
	Context *assets.Context
}

// ReplayToken - This function is used to create a new token and refresh token with a given subject ID. It first creates a new signing
// object with a HS256 signing method. Then it sets the claims such as "sub", "exp", and "iat" with the given subject,
// the expiration time 15 minutes from now, and the current time respectively. It then creates an access token and a
// refresh token, and creates a session object with the access token and the subject. It then marshals the session object
// and deletes any old refresh token with the same value. Finally, it sets the new refresh token with a 24-hour expiration time in the Redis client and returns the response.
func (a *Service) ReplayToken(subject int64) (*pbauth.Response, error) {

	// The two variables, response and session, are both declared as types of pbauth.Response and pbauth.Response_Session,
	// respectively. The purpose of this declaration is to create two variables that will be used to store data related to
	// the pbauth.Response and pbauth.Response_Session objects.
	var (
		response pbauth.Response
		session  pbauth.Response_Session
	)

	// The purpose of signing := jwt.New(jwt.SigningMethodHS256) is to create a new JWT object with the signing method HS256
	// and assign it to the variable signing. HS256 is a secure hashing algorithm used for encrypting data.
	signing := jwt.New(jwt.SigningMethodHS256)

	// This code is setting up the JWT claims when creating a JWT token. The "sub" claim is the subject of the token, "exp"
	// is the expiration time, and "iat" is the issued at time. This code is setting the expiration time to 15 minutes from
	// the current time and the issued at time to the current time.
	claims := signing.Claims.(jwt.MapClaims)
	claims["sub"] = subject
	claims["exp"] = time.Now().Add(15 * time.Minute).Unix()
	claims["iat"] = time.Now().Unix()

	// This code is attempting to sign a string using the secret stored in the Context.Secrets[0] array. The access variable
	// will store the signed string, and the if statement will return an error if the signing fails.
	access, err := signing.SignedString([]byte(a.Context.Secrets[0]))
	if err != nil {
		return &response, err
	}

	// This code assigns a new AccessToken and a new RefreshToken to the response object. The AccessToken is used to
	// authenticate a user, while the RefreshToken is used to generate a new AccessToken when it expires.
	response.AccessToken, response.RefreshToken = access, uuid.NewV4().String()

	// The purpose of this code is to assign the value of the Access Token and Subject returned from the
	// response.GetAccessToken() method to the session.AccessToken and session.Subject variables, respectively.
	session.AccessToken, session.Subject = response.GetAccessToken(), subject

	// The purpose of this code is to Marshal the 'session' variable into the MessagePack format. If there is an error
	// during the process, it will return an error as well as a response.
	marshal, err := msgpack.Marshal(&session)
	if err != nil {
		return &response, err
	}

	// The purpose of this statement is to delete the refresh token stored in Redis. This statement is using the
	// Context.RedisClient.Del() function to delete the refresh token based on the context and the response received.
	a.Context.RedisClient.Del(context.Background(), response.GetRefreshToken())

	// This code is checking for an error when setting a refresh token in a Redis database. If an error occurs, it is
	// returned along with the response.
	if err = a.Context.RedisClient.Set(context.Background(), response.GetRefreshToken(), marshal, 24*time.Hour).Err(); err != nil {
		return &response, err
	}

	return &response, err
}

// writeCode - This function sets a 6-character code for a given email address and sends an email containing that code for
// verification. The GO statement at the end allows the code to be sent asynchronously.
func (a *Service) writeCode(email string) (code interface{}, err error) {

	// The purpose of the code is to initialize two variables: migrate and q. The first variable, migrate, is set to an
	// instance of the Migrate type from the query package with the context set to the value of the a.Context variable. The
	// second variable, q, is set to an instance of the Query type from the query package.
	var (
		migrate = query.Migrate{
			Context: a.Context,
		}
		q query.Query
	)

	// The code provided is used to create a new 6-digit code that is set to true. This code can be used for a variety of
	// purposes, such as verifying the identity of a user or providing an access code for a secure system.
	code = help.NewCode(6, true)

	// This code is attempting to update the accounts table in a database with a given email_code for a given email address.
	// It is using a QueryRow and Scan to save the value of the id of the account returned from the query into the q.Id
	// variable. If there is an error in the query, the code will return nil and the error.
	if err := a.Context.Db.QueryRow("update accounts set email_code = $2 where email = $1 returning id;", email, code).Scan(&q.Id); err != nil {
		return nil, err
	}

	// The purpose of this code is to email the user with a secure code for migration. The code takes three
	// arguments, the first is the user's ID, the second is a string "secure" and the third is a code. The code then uses
	// these arguments to email the user with the secure code for migration.
	go migrate.SendMail(q.Id, "secure", code)

	return code, err
}
