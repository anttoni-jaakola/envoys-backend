package account

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/query"
	"google.golang.org/grpc/status"
	"strings"
)

// GetAccountsRule - get all users.
func (a *Service) GetAccountsRule(ctx context.Context, req *pbaccount.GetRequestUsersRule) (*pbaccount.Response, error) {

	var (
		response pbaccount.Response
		migrate  = query.Migrate{
			Context: a.Context,
		}
		maps  []string
		rules []byte
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts", query.RoleDefault) {
		return &response, a.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.GetSearch()) > 0 {
		maps = append(maps, fmt.Sprintf("where name like %[1]s or email like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	_ = a.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from accounts %s", strings.Join(maps, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := a.Context.Db.Query(fmt.Sprintf("select id, name, email, status, create_at, rules from accounts %s order by id desc limit %d offset %d", strings.Join(maps, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item   pbaccount.User
				counts pbaccount.User_Counts
			)

			if err = rows.Scan(
				&item.Id,
				&item.Name,
				&item.Email,
				&item.Status,
				&item.CreateAt,
				&rules,
			); err != nil {
				return &response, a.Context.Error(err)
			}

			if err := json.Unmarshal(rules, &item.Rules); err != nil {
				return &response, a.Context.Error(err)
			}

			_ = a.Context.Db.QueryRow("select count(*) as count from spot_transactions where user_id = $1", item.Id).Scan(&counts.Transaction)
			_ = a.Context.Db.QueryRow("select count(*) as count from spot_orders where user_id = $1", item.Id).Scan(&counts.Order)
			_ = a.Context.Db.QueryRow("select count(*) as count from spot_assets where user_id = $1", item.Id).Scan(&counts.Asset)

			item.Counts = &counts
			response.Fields = append(response.Fields, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, a.Context.Error(err)
		}
	}

	return &response, nil
}

// GetAccountRule - get user information.
func (a *Service) GetAccountRule(ctx context.Context, req *pbaccount.GetRequestUserRule) (*pbaccount.Response, error) {

	var (
		response pbaccount.Response
		migrate  = query.Migrate{
			Context: a.Context,
		}
		rules []byte
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts", query.RoleDefault) {
		return &response, a.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if err := a.Context.Db.QueryRow("select id, name, email, status, rules from accounts where id = $1", req.GetId()).Scan(&response.Id, &response.Name, &response.Email, &response.Status, &rules); err != nil {
		return &response, a.Context.Error(err)
	}

	if err := json.Unmarshal(rules, &response.Rules); err != nil {
		return &response, a.Context.Error(err)
	}

	return &response, nil
}

// SetAccountRule - set user information.
func (a *Service) SetAccountRule(ctx context.Context, req *pbaccount.SetRequestUserManual) (*pbaccount.Response, error) {

	var (
		response pbaccount.Response
		migrate  = query.Migrate{
			Context: a.Context,
		}
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts", query.RoleDefault) || migrate.Rules(account, "deny-record", query.RoleDefault) {
		return &response, a.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	serialize, err := json.Marshal(req.User.GetRules())
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if _, err := a.Context.Db.Exec("update accounts set name = $1, status = $2, rules = $3 where id = $4;",
		req.User.GetName(),
		req.User.GetStatus(),
		serialize,
		req.GetId(),
	); err != nil {
		return &response, a.Context.Error(err)
	}

	return &response, nil
}
