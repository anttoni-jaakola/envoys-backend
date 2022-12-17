package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/pbauth"
	"github.com/cryptogateway/backend-envoys/server/query"
	"github.com/tyler-smith/go-bip39"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"net"
	"net/mail"
	"strings"
)

// ActionSignup - auth action signup.
func (a *Service) ActionSignup(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	var (
		response pbauth.Response
	)

	// Metadata from incoming context.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, a.Context.Error(status.Error(10004, "permission denied"))
	}

	switch req.GetSignup() {
	case pbauth.Signup_ActionSignupAccount:

		if len(req.GetName()) < 5 {
			return &response, a.Context.Error(status.Error(19522, "the name must be at least 5 characters long"))
		}

		if len(req.GetPassword()) < 8 {
			return &response, a.Context.Error(status.Error(14563, "the password must be at least 8 characters long"))
		}

		if _, err := mail.ParseAddress(req.GetEmail()); err != nil {
			return &response, a.Context.Error(err)
		}

		row, err := a.Context.Db.Query("select id from accounts where email = $1", req.GetEmail())
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer row.Close()

		if !row.Next() {

			hashed := sha256.New()
			hashed.Write([]byte(fmt.Sprintf("%v-%v", req.GetPassword(), a.Context.Secrets[0])))

			entropy, err := bip39.NewEntropy(128)
			if err != nil {
				return &response, a.Context.Error(err)
			}

			if _, err := a.Context.Db.Exec("insert into accounts (name, email, password, entropy) values ($1, $2, $3, $4)", req.GetName(), req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)), entropy); err != nil {
				return &response, a.Context.Error(status.Error(15316, "a user with this address has already been registered before"))
			}

		} else {
			return &response, a.Context.Error(status.Error(64401, "a user with this email address is already registered"))
		}

		break
	case pbauth.Signup_ActionSignupCode:

		code, err := a.setCode(req.GetEmail())
		if err != nil {
			return &response, a.Context.Error(err)
		}

		if _, err = a.Context.Db.Exec("update accounts set secure = $3 where email = $1 and status = $2;", req.GetEmail(), false, code); err != nil {
			return &response, a.Context.Error(err)
		}

		break
	case pbauth.Signup_ActionSignupConfirm:

		if len(req.GetSecure()) != 6 {
			return &response, a.Context.Error(status.Error(14773, "the code must be 6 numbers"))
		}

		row, err := a.Context.Db.Query("select id from accounts where email = $1 and secure = $2 and status = $3", req.GetEmail(), req.GetSecure(), false)
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer row.Close()

		if row.Next() {

			if _, err := a.Context.Db.Exec("update accounts set status = $2 where email = $1;", req.GetEmail(), true); err != nil {
				return &response, a.Context.Error(err)
			}

		} else {
			return &response, status.Error(58042, "this code is invalid")
		}

		break
	default:
		return &response, a.Context.Error(status.Error(60001, "invalid input parameter"))
	}

	return &response, nil
}

// ActionSignin - auth action signup.
func (a *Service) ActionSignin(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	var (
		response pbauth.Response
	)

	// Metadata from incoming context.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, a.Context.Error(status.Error(10004, "permission denied"))
	}

	hashed := sha256.New()
	hashed.Write([]byte(fmt.Sprintf("%v-%v", req.GetPassword(), a.Context.Secrets[0])))

	switch req.GetSignin() {
	case pbauth.Signin_ActionSigninAccount:

		row, err := a.Context.Db.Query("select id from accounts where email = $1 and password = $2", req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)))
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer row.Close()

		if !row.Next() {
			return &response, a.Context.Error(status.Error(48512, "the email address or password was entered incorrectly"))
		}

		break
	case pbauth.Signin_ActionSigninCode:

		code, err := a.setCode(req.GetEmail())
		if err != nil {
			return &response, a.Context.Error(err)
		}

		if _, err = a.Context.Db.Exec("update accounts set secure = $3 where email = $1 and password = $2;", req.GetEmail(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)), code); err != nil {
			return &response, a.Context.Error(err)
		}

		break
	case pbauth.Signin_ActionSigninConfirm:

		if len(req.GetSecure()) != 6 {
			return &response, a.Context.Error(status.Error(14773, "the code must be 6 numbers"))
		}

		row, err := a.Context.Db.Query("select id from accounts where email = $1 and secure = $2 and password = $3", req.GetEmail(), req.GetSecure(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)))
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer row.Close()

		if row.Next() {

			var (
				params struct {
					ip string
					id int64
				}
				migrate = query.Migrate{
					Context: a.Context,
				}
			)

			if err := row.Scan(&params.id); err != nil {
				return &response, a.Context.Error(err)
			}

			token, err := a.ReplayToken(params.id)
			if err != nil {
				return &response, a.Context.Error(err)
			}

			if meta, ok := metadata.FromIncomingContext(ctx); ok {
				agent := help.MetaAgent(meta.Get("grpcgateway-user-agent")[0])

				browser, err := json.Marshal([]string{strings.ToLower(agent.Name), agent.Version})
				if err != nil {
					return &response, a.Context.Error(err)
				}

				if mp, ok := peer.FromContext(ctx); ok {
					if tcpAddr, ok := mp.Addr.(*net.TCPAddr); ok {
						params.ip = tcpAddr.IP.String()
					} else {
						params.ip = mp.Addr.String()
					}
				}

				if _, err = a.Context.Db.Exec("insert into actions (user_id, os, device, browser, ip) values ($1, $2, $3, $4, $5)", params.id, strings.ToLower(agent.OS), agent.Device, browser, params.ip); err != nil {
					return &response, a.Context.Error(err)
				}
			}

			if _, err = a.Context.Db.Exec("update accounts set secure = $2 where email = $1;", req.GetEmail(), ""); err != nil {
				return &response, a.Context.Error(err)
			}

			go migrate.SamplePosts(params.id, "login", nil)

			response.AccessToken, response.RefreshToken = token.AccessToken, token.RefreshToken

		} else {
			return &response, status.Error(58042, "this code is invalid")
		}

		break
	default:
		return &response, a.Context.Error(status.Error(60001, "invalid input parameter"))
	}

	return &response, nil
}

// ActionReset - auth action reset.
func (a *Service) ActionReset(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	var (
		response pbauth.Response
	)

	// Metadata from incoming context.
	meta, ok := metadata.FromIncomingContext(ctx)
	if ok && meta["authorization"] != nil {
		return &response, a.Context.Error(status.Error(10004, "permission denied"))
	}

	switch req.GetReset_() {
	case pbauth.Reset_ActionResetAccount:

		row, err := a.Context.Db.Query("select id from accounts where email = $1", req.GetEmail())
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer row.Close()

		if !row.Next() {
			return &response, a.Context.Error(status.Error(48512, "there is no user with this email"))
		}

		break
	case pbauth.Reset_ActionResetCode:

		code, err := a.setCode(req.GetEmail())
		if err != nil {
			return &response, a.Context.Error(err)
		}

		if _, err = a.Context.Db.Exec("update accounts set secure = $2 where email = $1;", req.GetEmail(), code); err != nil {
			return &response, a.Context.Error(err)
		}

		break
	case pbauth.Reset_ActionResetConfirm:

		if len(req.GetSecure()) != 6 {
			return &response, a.Context.Error(status.Error(14773, "the code must be 6 numbers"))
		}

		row, err := a.Context.Db.Query("select id from accounts where email = $1 and secure = $2", req.GetEmail(), req.GetSecure())
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer row.Close()

		if !row.Next() {
			return &response, status.Error(58042, "this code is invalid")
		}

		break
	case pbauth.Reset_ActionResetPassword:

		if len(req.GetSecure()) != 6 {
			return &response, a.Context.Error(status.Error(14773, "the code must be 6 numbers"))
		}

		if len(req.GetPassword()) < 8 {
			return &response, a.Context.Error(status.Error(14563, "the password must be at least 8 characters long"))
		}

		hashed := sha256.New()
		hashed.Write([]byte(fmt.Sprintf("%v-%v", req.GetPassword(), a.Context.Secrets[0])))

		if _, err := a.Context.Db.Exec("update accounts set password = $3, secure = $4 where email = $1 and secure = $2;", req.GetEmail(), req.GetSecure(), base64.URLEncoding.EncodeToString(hashed.Sum(nil)), ""); err != nil {
			return &response, a.Context.Error(err)
		}

		break
	default:
		return &response, a.Context.Error(status.Error(60001, "invalid input parameter"))
	}

	return &response, nil
}

// SetLogout - auth account exit.
func (a *Service) SetLogout(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	var (
		response pbauth.Response
	)

	// Metadata from incoming context.
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok && len(meta["authorization"]) != 1 && meta["authorization"] == nil {
		return &response, a.Context.Error(status.Error(10004, "permission denied"))
	}

	// Delete old refresh token.
	a.Context.RedisClient.Del(context.Background(), req.GetRefresh())

	if _, err := a.Context.Db.Exec("update accounts set secure = $2 where email = $1;", req.GetEmail(), ""); err != nil {
		return &response, a.Context.Error(err)
	}

	return &response, nil
}

// GetRefresh - auth token refresh.
func (a *Service) GetRefresh(ctx context.Context, req *pbauth.Request) (*pbauth.Response, error) {

	var (
		response  pbauth.Response
		serialize pbauth.Response_Session
	)

	// Metadata from incoming context.
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok && len(meta["authorization"]) != 1 && meta["authorization"] == nil {
		return &response, a.Context.Error(status.Error(10411, "missing metadata"))
	}

	session, err := a.Context.RedisClient.Get(context.Background(), req.GetRefresh()).Bytes()
	if err != nil {
		return &response, a.Context.Error(err)
	}

	err = msgpack.Unmarshal(session, &serialize)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	token := strings.Split(meta["authorization"][0], "Bearer ")[1]
	if serialize.AccessToken != token {
		return &response, a.Context.Error(status.Error(31754, "session not found"))
	}

	replayToken, err := a.ReplayToken(serialize.Subject)
	if err != nil {
		return nil, err
	}

	response.AccessToken, response.RefreshToken = replayToken.GetAccessToken(), replayToken.GetRefreshToken()

	return &response, nil
}
