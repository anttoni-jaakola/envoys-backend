package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto"
	"google.golang.org/grpc/status"
	"strings"
)

// GetAccountsRule - get all users.
func (a *AccountService) GetAccountsRule(ctx context.Context, req *proto.GetAccountRequestUsersRule) (*proto.ResponseAccount, error) {

	var (
		response proto.ResponseAccount
		migrate  = Query{
			Context: a.Context,
		}
		query []string
		rules []byte
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts") {
		return &response, a.Context.Error(status.Error(12011, "you do not have rules for writing and editing data"))
	}

	if len(req.GetSearch()) > 0 {
		query = append(query, fmt.Sprintf("where name like %[1]s or email like %[1]s", "'%"+req.GetSearch()+"%'"))
	}

	_ = a.Context.Db.QueryRow(fmt.Sprintf("select count(*) as count from accounts %s", strings.Join(query, " "))).Scan(&response.Count)

	if response.GetCount() > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := a.Context.Db.Query(fmt.Sprintf("select id, name, email, status, create_at, rules from accounts %s order by id desc limit %d offset %d", strings.Join(query, " "), req.GetLimit(), offset))
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var (
				item   proto.User
				counts proto.User_Counts
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

			_ = a.Context.Db.QueryRow("select count(*) as count from transactions where user_id = $1", item.Id).Scan(&counts.Transaction)
			_ = a.Context.Db.QueryRow("select count(*) as count from orders where user_id = $1", item.Id).Scan(&counts.Order)
			_ = a.Context.Db.QueryRow("select count(*) as count from assets where user_id = $1", item.Id).Scan(&counts.Asset)

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
func (a *AccountService) GetAccountRule(ctx context.Context, req *proto.GetAccountRequestUserRule) (*proto.ResponseAccount, error) {

	var (
		response proto.ResponseAccount
		migrate  = Query{
			Context: a.Context,
		}
		rules []byte
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts") || migrate.Rules(account, "deny-record") {
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
func (a *AccountService) SetAccountRule(ctx context.Context, req *proto.SetAccountRequestUserManual) (*proto.ResponseAccount, error) {

	var (
		response proto.ResponseAccount
		migrate  = Query{
			Context: a.Context,
		}
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if !migrate.Rules(account, "accounts") || migrate.Rules(account, "deny-record") {
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
