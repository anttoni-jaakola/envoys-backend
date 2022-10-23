package service

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto"
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

type Query struct {
	Context *assets.Context

	InternalId     int64
	InternalName   string
	InternalType   string
	InternalBuffer bytes.Buffer
}

// Rules - user role admin.
func (q *Query) Rules(userId int64, name string) bool {

	var (
		response proto.Query
	)

	if err := q.Context.Db.QueryRow("select rules from accounts where id = $1", userId).Scan(&response.Rules); q.Context.Debug(err) {
		return false
	}

	if help.Comparable(response.GetRules(), name) {
		return true
	}

	return false
}

// Rename - rename file.
func (q *Query) Rename(path, oldName, newName string) error {

	var (
		storage []string
	)

	storage = append(storage, []string{q.Context.StoragePath, "static", path}...)
	if err := os.Rename(filepath.Join(append(storage, []string{fmt.Sprintf("%v.png", oldName)}...)...), filepath.Join(append(storage, []string{fmt.Sprintf("%v.png", newName)}...)...)); err != nil {
		return err
	}

	return nil
}

// Remove - remove file.
func (q *Query) Remove(path, name string) error {

	var (
		storage []string
	)

	storage = append(storage, []string{q.Context.StoragePath, "static", path, fmt.Sprintf("%v.png", name)}...)
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(filepath.Join(storage...)); err != nil {
			return err
		}
	}

	return nil
}

// Image - upload image
func (q *Query) Image(img []byte, path, name string) error {

	q.InternalType = http.DetectContentType(img)

	if q.InternalType != "image/jpeg" && q.InternalType != "image/png" && q.InternalType != "image/gif" {
		return status.Error(12000, "image type is not correct")
	}

	if err := q.Remove(path, name); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join([]string{q.Context.StoragePath, "static", path, fmt.Sprintf("%v.png", name)}...))
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
	if err := png.Encode(&q.InternalBuffer, imaging.Fill(serialize, 300, 300, imaging.Center, imaging.Lanczos)); err != nil {
		return err
	}

	_, err = bufio.NewWriter(file).Write(q.InternalBuffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// SamplePosts - send to user mail sample.
func (q *Query) SamplePosts(userId int64, name string, params ...interface{}) {

	var (
		response proto.Query
		buffer   bytes.Buffer
	)

	if err := q.Context.Db.QueryRow("select name, sample, email from accounts where id = $1", userId).Scan(&response.Name, &response.Sample, &response.Email); q.Context.Debug(err) {
		return
	}

	templates, err := template.ParseFiles(fmt.Sprintf("./static/sample/sample_%v.html", name))
	if q.Context.Debug(err) {
		return
	}

	switch name {
	case "order_filled":
		response.Subject = "Your order has been filled"

		switch params[4].(proto.Assigning) {
		case proto.Assigning_BUY:
			response.Symbol = strings.ToUpper(params[3].(string))
		case proto.Assigning_SELL:
			response.Symbol = strings.ToUpper(params[2].(string))
		}

		response.Text = fmt.Sprintf("Order ID: %d, Quantit: %v<b>%v</b>, Pair: <b>%v/%s</b>", params[0].(int64), params[1].(float64), response.GetSymbol(), strings.ToUpper(params[2].(string)), strings.ToUpper(params[3].(string)))
		break
	case "withdrawal":
		response.Subject = "Withdrawal Successful"
		response.Text = fmt.Sprintf("You've successfully withdrawn %v <b>%s</b>.", params[0].(float64), strings.ToUpper(params[1].(string)))
		break
	case "login":
		response.Subject = "You just logged in PayMex"
		break
	case "news":
		response.Subject = "Latest news from PayMex"
		break
	case "secure":
		response.Subject = "Secure code PayMex"
		response.Text = fmt.Sprintf("Your secret code <b>%v</b>, do not give it to anyone", params[0].(string))
		break
	}

	err = templates.Execute(&buffer, &response)
	if q.Context.Debug(err) {
		return
	}

	if help.Comparable(response.GetSample(), name, "secure") {

		m := gomail.NewMessage()
		m.SetHeader("From", q.Context.SmtpSender)
		m.SetHeader("To", response.GetEmail())
		m.SetHeader("Subject", response.GetSubject())
		m.SetBody("text/html", buffer.String())

		d := gomail.NewDialer(q.Context.SmtpHost, q.Context.SmtpPort, q.Context.SmtpSender, q.Context.SmtpPassword)
		if err := d.DialAndSend(m); q.Context.Debug(err) {
			return
		}
	}

	return
}
