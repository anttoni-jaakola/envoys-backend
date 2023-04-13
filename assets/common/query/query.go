package query

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/types"
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

// The purpose of this code is to define two constants, RoleDefault and RoleSpot, with numerical values of 0 and 1,
// respectively. These constants could be used to assign roles to different users in a program or system. For example,
// RoleDefault might refer to a normal user, while RoleSpot, RoleMarket could refer to an administrator or moderator.
const (
	RoleDefault = 0
	RoleSpot    = 1
	RoleMarket  = 2
)

// Query -This type Query struct is used to store data related to a query which might be sent to a database. It contains fields
// of various types such as integers, strings, and slices of bytes. The bytes.Buffer field is used to store data in a
// buffer, which is a temporary storage area used to store data in memory before it is written to a file or database.
type Query struct {
	Id                                       int64
	Email, Subject, Text, Name, Type, Symbol string
	Sample, Rules                            []byte
	Buffer                                   bytes.Buffer
}

// Migrate - The type Migrate struct is used to store a pointer to a Context object from the assets package. The Context object
// contains information and methods that can be used to help with the migration of assets. This type is used to ensure
// that the necessary information and methods are always accessible when the migration process needs to be executed.
type Migrate struct {
	Context *assets.Context
}

// Rules - This function is used to check if a user has a certain role in a specific context. It takes an id, name and tag as
// parameters and returns a boolean. It first queries the database to retrieve the user's rules from the accounts table.
// It then unmasrshals the rules into a types.Rules structure. The tag parameter is used to determine which roles to
// check - either the Default or Spot roles. Finally, it uses the help.IndexOf() function to check if the role is in the
// array of roles. If so, it returns true, otherwise it returns false.
func (m *Migrate) Rules(id int64, name string, tag int) bool {

	// The purpose of the code above is to declare three variables: response, roles, and rules. The response variable is of
	// type Query, the roles variable is of type string slice, and the rules variable is of type pbaccount.Rules. These
	// variables can then be used in the code that follows.
	var (
		response Query
		roles    []string
		rules    types.Rules
	)

	// This code is used to query a database for a row in the "accounts" table that matches a given 'id'. If a matching row
	// is found, the 'rules' field is stored in the 'response.Rules' variable. If an error is encountered, the
	// 'm.Context.Debug(err)' line is used to debug the issue and the function returns 'false'.
	if err := m.Context.Db.QueryRow("select rules from accounts where id = $1", id).Scan(&response.Rules); m.Context.Debug(err) {
		return false
	}

	// This code is attempting to decode a JSON response into a data structure. The json.Unmarshal() function takes two
	// arguments: a JSON string and a pointer to a data structure. It then attempts to decode the JSON into the specified
	// data structure. The "if err != nil" condition tests for any errors that occurred during decoding. If an error does
	// occur, the function returns false.
	err := json.Unmarshal(response.Rules, &rules)
	if err != nil {
		return false
	}

	// This switch tag is used to assign different roles to a given set of rules. The switch statement checks the value of
	// the tag and depending on its value, it assigns the roles in the rules to either the Default or the Spot roles.
	switch tag {
	case RoleDefault:
		roles = rules.Default
	case RoleSpot:
		roles = rules.Spot
	case RoleMarket:
		roles = rules.Market
	}

	// The purpose of this code is to check if a given name is present in a list of roles. The help.IndexOf() function is
	// used to search for the name in the list of roles and if it is found, the function returns true.
	if help.IndexOf(roles, name) {
		return true
	}

	return false
}

// Rename - This function is used to rename a file in a given directory. It takes three parameters,
// the path of the file, the old name of the file, and the new name of the file.
// It appends the path to the context storage path and then uses the os.Rename() and filepath.Join() functions to rename the file. It returns an error if the file cannot be renamed.
func (m *Migrate) Rename(path, oldName, newName string) error {

	// The purpose of this statement is to declare a variable called 'storage' which will be an empty slice of strings. This
	// slice can then be used to store strings for later use.
	var (
		storage []string
	)

	// This statement is appending the elements of the array ['m.Context.StoragePath', 'static', path] to the array
	// 'storage'. This is used to add the elements of one array to another array.
	storage = append(storage, []string{m.Context.StoragePath, "static", path}...)

	// This code is renaming a file, using the os.Rename function. The purpose of this code is to change the name of a file
	// stored at the specified filepath. The oldName and newName variables are used to specify what the file is being
	// renamed from and to. The filepath is constructed using the storage variable, and the fmt.Sprintf command, which
	// creates a string with formatting based on the given parameters. If there is an error, the function will return an error.
	if err := os.Rename(filepath.Join(append(storage, []string{fmt.Sprintf("%v.png", oldName)}...)...), filepath.Join(append(storage, []string{fmt.Sprintf("%v.png", newName)}...)...)); err != nil {
		return err
	}

	return nil
}

// RemoveFiles - The purpose of this function is to remove a file located at the given path and name. It first checks if the file
// exists and then removes it if it does.
func (m *Migrate) RemoveFiles(path, name string) error {

	// The purpose of the above code is to declare a variable named "storage" as an empty slice of strings. The slice is
	// empty and will need to be filled before it can be used.
	var (
		storage []string
	)

	// This is adding elements to an existing slice of strings called "storage". It is adding the string elements of
	// m.Context.StoragePath, "static", path, and a formatted string of the variable "name" with a ".png" extension. This is
	// likely part of a larger function which uses the storage slice to create a file path.
	storage = append(storage, []string{m.Context.StoragePath, "static", path, fmt.Sprintf("%v.png", name)}...)

	// This code is checking if a file exists in the given filepath. If it does exist, it is then removed. This process is
	// used to delete a file that is located in the specified filepath.
	if _, err := os.Stat(filepath.Join(storage...)); !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(filepath.Join(storage...)); err != nil {
			return err
		}
	}

	return nil
}

// Image - This function is used to convert an image to a different file format, such as from JPEG to PNG. It also resizes the
// image to the specified width and height. Additionally, it checks that the provided image is of an acceptable type
// (JPEG, PNG, or GIF) and removes any existing files with the same name and path. Finally, it writes the converted image
// to a file in the specified path.
func (m *Migrate) Image(img []byte, path, name string, width, height int) error {

	// The purpose of this code is to declare a variable called "response" of type "Query". This variable can be used to
	// store information related to a query that is being processed.
	var (
		response Query
	)

	// The purpose of this line of code is to set the response type to the content type of the image file. This is necessary
	// so that the web server can properly handle the request and respond with the correct type of data. By setting the
	// response type to the content type of the image file, the web server can properly interpret the request and respond accordingly.
	response.Type = http.DetectContentType(img)

	// This code is checking to make sure that the response type is an image in one of the three accepted formats (JPEG,
	// PNG, or GIF). If the response type is not one of the accepted formats, an error is returned.
	if response.Type != "image/jpeg" && response.Type != "image/png" && response.Type != "image/gif" {
		return status.Error(12000, "image type is not correct")
	}

	// The purpose of this code is to remove a file (or directory) from a given path with the given name. If there is an
	// error when trying to remove the file, the code will return an error.
	if err := m.RemoveFiles(path, name); err != nil {
		return err
	}

	// This code is used to create a file with the path and name provided. The filepath.Join command creates a single path
	// out of all the strings provided in the array, which is then used to create a file with the os.Create command and
	// stored in the "file" variable. If there is an error, it will be returned.
	file, err := os.Create(filepath.Join([]string{m.Context.StoragePath, "static", path, fmt.Sprintf("%v.png", name)}...))
	if err != nil {
		return err
	}

	// This code is attempting to decode an image stored in a byte slice (img) using the image.Decode() function. This
	// function will produce a serialized image and an error value. The 'if' statement is checking to see if an error
	// occurred during the decoding process. If an error did occur, it returns the error to the calling function.
	serialize, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		return err
	}

	// To defer keyword is used to delay the execution of a function until the surrounding function returns. In this case,
	// the defer file.Close() statement delays the execution of the Close() function until the surrounding function returns.
	// This ensures that the file is always closed, even if the surrounding function returns early due to an error.
	defer file.Close()

	// This code is checking for any errors that arise when encoding a PNG image and filling it with specific criteria
	// (width, height, center and lanczos). If an error is returned, the code returns that error.
	if err := png.Encode(&response.Buffer, imaging.Fill(serialize, width, height, imaging.Center, imaging.Lanczos)); err != nil {
		return err
	}

	// This code is used to write the response.Buffer.Bytes() to the file using bufio. It checks if there is an error and
	// returns it if so.
	_, err = bufio.NewWriter(file).Write(response.Buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// SendMail - This function is part of a Migrate struct and is used to email a user with a given user ID and name.
// The params parameter is a variadic argument which can contain a variable number of additional parameters to be used in the email.
func (m *Migrate) SendMail(userId int64, name string, params ...interface{}) {

	// The purpose of the above code is to declare two variables: response, of type Query, and buffer, of type bytes.Buffer.
	// These two variables can then be used in the code to store and manipulate data.
	var (
		response Query
		buffer   bytes.Buffer
	)

	// This code is used to query a database for information related to a user with a given ID (userId). The code uses the
	// Scan() method to assign the values returned from the query to response.Name, response.Sample, and response.Email. If
	// there is an error, it is logged and the function returns.
	if err := m.Context.Db.QueryRow("select name, sample, email from accounts where id = $1", userId).Scan(&response.Name, &response.Sample, &response.Email); m.Context.Debug(err) {
		return
	}

	// This code is used to parse an HTML template file with a dynamic name (sample_%v.html). The fmt.Sprintf function is
	// used to construct the filename with the name parameter. The template.ParseFiles function is used to parse the file,
	// and it returns a slice of templates. The if statement checks for errors and stops the execution of the code if any are encountered.
	templates, err := template.ParseFiles(fmt.Sprintf("./static/sample/sample_%v.html", name))
	if m.Context.Debug(err) {
		return
	}

	// The switch statement is used to execute certain code depending on the value of the given expression. The expression
	// used in the switch statement is known as the switch name. The code within the switch statement body is executed
	// depending on the value of the expression.
	switch name {
	case "order_filled":
		response.Subject = "Your order has been filled"

		// This code is part of a switch statement which checks the value of the 4th parameter in the 'params' array and then
		// executes different code depending on the value. In this case, the code checks whether the 4th parameter is of type
		// 'types.Assigning' and, if it is, then checks whether the value is 'types.AssigningBuy' or
		// 'types.AssigningSell'. Depending on the result, the code either sets the 'Symbol' property of the 'response'
		// object to the uppercase version of the 3rd parameter in the 'params' array, or the uppercase version of the 2nd
		// parameter in the 'params' array. This code is used to ensure that the 'Symbol' property of the 'response' object is
		// set correctly depending on the value of the 4th parameter in the 'params' array.
		switch params[4].(string) {
		case types.AssigningBuy:
			response.Symbol = strings.ToUpper(params[3].(string))
		case types.AssigningSell:
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
	case "new_password":
		response.Subject = "Reset password"
		response.Text = fmt.Sprintf("Your new password <b>%v</b>", params[0].(string))
		break
	}

	// The code is likely part of a program that generates an HTML response to a client. The first line executes a template
	// (likely an HTML file) and stores the resulting HTML in a buffer. The second line checks if an error has occurred
	// while executing the template. If an error has occurred, the debug method is used to log the error and the program
	// returns without sending a response to the client.
	err = templates.Execute(&buffer, &response)
	if m.Context.Debug(err) {
		return
	}

	// This if statement is checking if the response.Sample, name, "secure", and "new_password" parameters are comparable.
	// If they are comparable, the statement will evaluate to true and the code inside the block will be executed. If not,
	// the statement will evaluate to false and the code inside the block will not be executed.
	if help.Comparable(response.Sample, name, "secure", "new_password") {

		// The purpose of the line of code "g := gomail.NewMessage()" is to create a new instance of a gomail message, which is
		// used to send emails. The "g" is a variable that holds the reference to the newly created message.
		g := gomail.NewMessage()

		// The purpose of this line of code is to set the "From" header in an email message. The m.Context.Smtp.Sender value is
		// used to specify the sender of the email message.
		g.SetHeader("From", m.Context.Smtp.Sender)

		// The purpose of the following is to set the "To" header of an email to the response.Email address. This will ensure
		// that the email is delivered to the intended recipient.
		g.SetHeader("To", response.Email)

		// The purpose of this code is to set the Subject header of an email message to the response subject of an email.
		g.SetHeader("Subject", response.Subject)

		// The purpose of g.SetBody("text/html", buffer.String()) is to set the response body of an HTTP request. The
		// "text/html" argument is the content type and the buffer.String() argument is the data being sent in the response
		// body. This allows the server to respond with an HTML page to the client.
		g.SetBody("text/html", buffer.String())

		// This code snippet is creating a new dialer and using it to dial and send an email using the gomail library. The
		// dialer is initialized with the SMTP host, port, sender, and password from the m.Context object. If an error occurs
		// while dialing and sending the email, the m.Context.Debug() method is used to log the error, then the function returns.
		d := gomail.NewDialer(m.Context.Smtp.Host, m.Context.Smtp.Port, m.Context.Smtp.Sender, m.Context.Smtp.Password)
		if err := d.DialAndSend(g); m.Context.Debug(err) {
			return
		}
	}

	return
}
