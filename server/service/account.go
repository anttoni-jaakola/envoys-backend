package service

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"google.golang.org/grpc/status"
	"hash"
	"strings"
)

type AccountService struct {
	Context *assets.Context
}

// setPassword - update new password.
func (a *AccountService) setPassword(id int64, oldPassword, newPassword string) error {

	var (
		hashed = make([]hash.Hash, 2)
	)

	hashed[0] = sha256.New()
	hashed[0].Write([]byte(fmt.Sprintf("%v-%v", oldPassword, a.Context.Secrets[0])))

	hashed[1] = sha256.New()
	hashed[1].Write([]byte(fmt.Sprintf("%v-%v", newPassword, a.Context.Secrets[0])))

	if len(newPassword) < 8 {
		return status.Error(18863, "the password must be at least 8 characters long")
	}

	if string(hashed[0].Sum(nil)) == string(hashed[1].Sum(nil)) {
		return status.Error(72554, "the new password must not be identical to the old one")
	}

	row, err := a.Context.Db.Query("select id from accounts where id = $1 and password = $2", id, base64.URLEncoding.EncodeToString(hashed[0].Sum(nil)))
	if err != nil {
		return a.Context.Error(err)
	}
	defer row.Close()

	if row.Next() {
		_, err := a.Context.Db.Exec(`update accounts set password = $2 where id = $1`, id, base64.URLEncoding.EncodeToString(hashed[1].Sum(nil)))
		if err != nil {
			return err
		}

		return nil
	}

	return status.Error(44754, "the old password was entered incorrectly")
}

// setSample - add to array column notify.
func (a *AccountService) setSample(id int64, index string) error {

	var (
		response proto.ResponseAccount
		column   = []string{"order_filled", "withdrawal", "login", "news"}
		query    []string
	)

	if !help.IndexOf(column, index) {
		return status.Error(10504, "incorrect sample index")
	}

	if err := a.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count from accounts where id = %d and sample @> '"%s"'::jsonb`, id, index)).Scan(&response.Count); err != nil && err != sql.ErrNoRows {
		return err
	}

	if response.Count > 0 {
		query = append(query, fmt.Sprintf(`sample = sample - '%s'`, index))
	} else {
		query = append(query, fmt.Sprintf(`sample = sample || '"%s"'`, index))
	}

	_, err := a.Context.Db.Exec(fmt.Sprintf(`update accounts set %[2]s where id = %[1]d`, id, strings.Join(query, "")))
	if err != nil {
		return err
	}

	return nil
}

// getUser - get user information.
func (a *AccountService) getUser(id int64) (*proto.ResponseAccount, error) {

	var (
		response      proto.ResponseAccount
		sample, rules []byte
	)

	if err := a.Context.Db.QueryRow("select id, name, email, status, sample, rules from accounts where id = $1", id).Scan(&response.Id, &response.Name, &response.Email, &response.Status, &sample, &rules); err != nil {
		return &response, a.Context.Error(err)
	}

	if err := json.Unmarshal(sample, &response.Sample); err != nil {
		return &response, a.Context.Error(err)
	}

	if err := json.Unmarshal(rules, &response.Rules); err != nil {
		return &response, a.Context.Error(err)
	}

	return &response, nil
}
