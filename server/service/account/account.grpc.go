package account

import (
	"context"
	"encoding/json"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/query"
)

// SetUser - set info user.
func (a *Service) SetUser(ctx context.Context, req *pbaccount.SetRequestUserManual) (*pbaccount.Response, error) {

	var (
		response *pbaccount.Response
		migrate  = query.Migrate{
			Context: a.Context,
		}
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return response, a.Context.Error(err)
	}

	if len(req.GetSample()) > 0 {
		if err := a.setSample(account, req.GetSample()); err != nil {
			return response, a.Context.Error(err)
		}
	}

	if len(req.GetOldPassword()) > 0 && len(req.GetNewPassword()) > 0 {
		if err := a.setPassword(account, req.GetOldPassword(), req.GetNewPassword()); err != nil {
			return response, a.Context.Error(err)
		}
	}

	return migrate.User(account)
}

// GetUser - get info user.
func (a *Service) GetUser(ctx context.Context, _ *pbaccount.GetRequestUser) (*pbaccount.Response, error) {

	var (
		response pbaccount.Response
		migrate  = query.Migrate{
			Context: a.Context,
		}
		err error
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	return migrate.User(account)
}

// GetActivities - get activities list.
func (a *Service) GetActivities(ctx context.Context, req *pbaccount.GetRequestActivities) (*pbaccount.Response, error) {

	var (
		response pbaccount.Response
		browser  []byte
	)

	account, err := a.Context.Auth(ctx)
	if err != nil {
		return &response, a.Context.Error(err)
	}

	if err := a.Context.Db.QueryRow("select count(*) from activities where user_id = $1", account).Scan(&response.Count); err != nil {
		return &response, a.Context.Error(err)
	}

	if response.Count > 0 {

		offset := req.GetLimit() * req.GetPage()
		if req.GetPage() > 0 {
			offset = req.GetLimit() * (req.GetPage() - 1)
		}

		rows, err := a.Context.Db.Query("select id, os, device, ip, browser, create_at from activities where user_id = $1 order by id desc limit $2 offset $3", account, req.GetLimit(), offset)
		if err != nil {
			return &response, a.Context.Error(err)
		}
		defer rows.Close()

		for rows.Next() {

			var item pbaccount.Activity
			if err := rows.Scan(&item.Id, &item.Os, &item.Device, &item.Ip, &browser, &item.CreateAt); err != nil {
				return &response, a.Context.Error(err)
			}

			if err := json.Unmarshal(browser, &item.Browser); err != nil {
				return &response, a.Context.Error(err)
			}

			response.Activities = append(response.Activities, &item)
		}

		if err = rows.Err(); err != nil {
			return &response, a.Context.Error(err)
		}

	}

	return &response, nil
}
