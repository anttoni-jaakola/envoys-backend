package auth

import (
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/pbauth"
	"github.com/cryptogateway/backend-envoys/server/query"
	"github.com/golang-jwt/jwt/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/net/context"
	"time"
)

type Service struct {
	Context *assets.Context
}

// ReplayToken - auth token couple generate.
func (a *Service) ReplayToken(subject int64) (*pbauth.Response, error) {

	var (
		response pbauth.Response
		session  pbauth.Response_Session
	)

	signing := jwt.New(jwt.SigningMethodHS256)

	claims := signing.Claims.(jwt.MapClaims)
	claims["sub"] = subject
	claims["exp"] = time.Now().Add(15 * time.Minute).Unix()
	claims["iat"] = time.Now().Unix()

	access, err := signing.SignedString([]byte(a.Context.Secrets[0]))
	if err != nil {
		return &response, err
	}

	response.AccessToken, response.RefreshToken = access, uuid.NewV4().String()
	session.AccessToken, session.Subject = response.GetAccessToken(), subject

	marshal, err := msgpack.Marshal(&session)
	if err != nil {
		return &response, err
	}

	// Delete old refresh token.
	a.Context.RedisClient.Del(context.Background(), response.GetRefreshToken())

	if err = a.Context.RedisClient.Set(context.Background(), response.GetRefreshToken(), marshal, 24*time.Hour).Err(); err != nil {
		return &response, err
	}

	return &response, err
}

// setCode - auth set and send to email.
func (a *Service) setCode(email string) (code interface{}, err error) {

	var (
		migrate = query.Migrate{
			Context: a.Context,
		}
		q query.Query
	)

	code, err = help.KeyCode(6)
	if err != nil {
		return nil, err
	}

	if err := a.Context.Db.QueryRow("update accounts set secure = $2 where email = $1 returning id;", email, code).Scan(&q.Id); err != nil {
		return nil, err
	}

	go migrate.SamplePosts(q.Id, "secure", code)

	return code, err
}
