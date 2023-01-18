package query

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"gopkg.in/gomail.v2"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	RoleDefault = 0
	RoleSpot    = 1
	RoleStock   = 2
)

type Query struct {
	Id                                       int64
	Email, Subject, Text, Name, Type, Symbol string
	Sample, Rules                            []byte
	Buffer                                   bytes.Buffer
}

type Migrate struct {
	Context *assets.Context
}

// User - get user.
func (m *Migrate) User(id int64) (*pbaccount.Response, error) {

	var (
		response pbaccount.Response
		q        Query
	)

	if err := m.Context.Db.QueryRow("select id, name, email, status, sample, rules from accounts where id = $1", id).Scan(&response.Id, &response.Name, &response.Email, &response.Status, &q.Sample, &q.Rules); err != nil {
		return &response, err
	}

	if err := json.Unmarshal(q.Sample, &response.Sample); err != nil {
		return &response, err
	}

	if err := json.Unmarshal(q.Rules, &response.Rules); err != nil {
		return &response, err
	}

	return &response, nil
}

// Rules - user role admin.
func (m *Migrate) Rules(id int64, name string, tag int) bool {

	var (
		response Query
		roles    []string
		rules    pbaccount.Rules
	)

	if err := m.Context.Db.QueryRow("select rules from accounts where id = $1", id).Scan(&response.Rules); m.Context.Debug(err) {
		return false
	}

	err := json.Unmarshal(response.Rules, &rules)
	if err != nil {
		return false
	}

	switch tag {
	case RoleDefault:
		roles = rules.Default
	case RoleSpot:
		roles = rules.Spot
	case RoleStock:
		roles = rules.Stock
	}

	if help.IndexOf(roles, name) {
		return true
	}

	return false
}

// Rename - rename file.
func (m *Migrate) Rename(path, oldName, newName string) error {

	var (
		storage []string
	)

	storage = append(storage, []string{m.Context.StoragePath, "static", path}...)
	if err := os.Rename(filepath.Join(append(storage, []string{fmt.Sprintf("%v.png", oldName)}...)...), filepath.Join(append(storage, []string{fmt.Sprintf("%v.png", newName)}...)...)); err != nil {
		return err
	}

	return nil
}

// Remove - remove file.
func (m *Migrate) Remove(path, name string) error {

	var (
		storage []string
	)

	storage = append(storage, []string{m.Context.StoragePath, "static", path, fmt.Sprintf("%v.png", name)}...)
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(filepath.Join(storage...)); err != nil {
			return err
		}
	}

	return nil
}

// Image - upload image
func (m *Migrate) Image(img []byte, path, name string) error {

	var (
		response Query
	)

	response.Type = http.DetectContentType(img)

	if response.Type != "image/jpeg" && response.Type != "image/png" && response.Type != "image/gif" {
		return status.Error(12000, "image type is not correct")
	}

	if err := m.Remove(path, name); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join([]string{m.Context.StoragePath, "static", path, fmt.Sprintf("%v.png", name)}...))
	if err != nil {
		return err
	}

	serialize, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return err
	}

	defer file.Close()

	// Resize to width 300 using Lanczos resampling,
	// and preserve aspect ratio,
	// write new image to file.
	// Crop the original image to 300x300px size using the center anchor.
	if err := png.Encode(&response.Buffer, imaging.Fill(serialize, 300, 300, imaging.Center, imaging.Lanczos)); err != nil {
		return err
	}

	_, err = bufio.NewWriter(file).Write(response.Buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// SamplePosts - send to user mail sample.
func (m *Migrate) SamplePosts(userId int64, name string, params ...interface{}) {

	var (
		response Query
		buffer   bytes.Buffer
	)

	if err := m.Context.Db.QueryRow("select name, sample, email from accounts where id = $1", userId).Scan(&response.Name, &response.Sample, &response.Email); m.Context.Debug(err) {
		return
	}

	templates, err := template.ParseFiles(fmt.Sprintf("./static/sample/sample_%v.html", name))
	if m.Context.Debug(err) {
		return
	}

	switch name {
	case "order_filled":
		response.Subject = "Your order has been filled"

		switch params[4].(pbspot.Assigning) {
		case pbspot.Assigning_BUY:
			response.Symbol = strings.ToUpper(params[3].(string))
		case pbspot.Assigning_SELL:
			response.Symbol = strings.ToUpper(params[2].(string))
		}

		response.Text = fmt.Sprintf("Order ID: %d, Quantit: %v<b>%v</b>, Pair: <b>%v/%s</b>", params[0].(int64), params[1].(float64), response.Symbol, strings.ToUpper(params[2].(string)), strings.ToUpper(params[3].(string)))
		break
	case "withdrawal":
		response.Subject = "Withdrawal Successful"
		response.Text = fmt.Sprintf("You've successfully withdrawn %v <b>%s</b>.", params[0].(float64), strings.ToUpper(params[1].(string)))
		break
	case "login":
		response.Subject = "You just logged in Envoys"
		break
	case "news":
		response.Subject = "Latest news from Envoys"
		break
	case "secure":
		response.Subject = "Secure code Envoys"
		response.Text = fmt.Sprintf("Your secret code <b>%v</b>, do not give it to anyone", params[0].(string))
		break
	}

	err = templates.Execute(&buffer, &response)
	if m.Context.Debug(err) {
		return
	}

	if help.Comparable(response.Sample, name, "secure") {

		g := gomail.NewMessage()
		g.SetHeader("From", m.Context.Smtp.Sender)
		g.SetHeader("To", response.Email)
		g.SetHeader("Subject", response.Subject)
		g.SetBody("text/html", buffer.String())

		d := gomail.NewDialer(m.Context.Smtp.Host, m.Context.Smtp.Port, m.Context.Smtp.Sender, m.Context.Smtp.Password)
		if err := d.DialAndSend(g); m.Context.Debug(err) {
			return
		}
	}

	return
}
