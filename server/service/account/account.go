package account

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/assets/common/help"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"google.golang.org/grpc/status"
	"hash"
	"strings"
)

// Service - The purpose of this code is to declare a Service struct which contains a Context pointer. The Context pointer is of
// type assets.Context, which likely contains parameters or other values that are relevant to the Service.
type Service struct {
	Context *assets.Context
}

// setPassword - This function sets a new password for a user given their ID, old password, and new password. It first checks if the
// new password is at least 8 characters long and is not the same as the old one. It then uses a database query to check
// if the given ID and old password match. If it does, it updates the database with the new password. If it doesn't, it
// returns an error.
func (a *Service) setPassword(id int64, oldPassword, newPassword string) error {

	// The purpose of this code is to create a slice of two hash.Hash elements. The make function is used to create a new
	// slice with the given length and capacity. The hash.Hash elements will contain the information needed to create a hash.
	var (
		hashed = make([]hash.Hash, 2)
	)

	// The purpose of this code is to use the SHA256 cryptographic hash algorithm to create a new hash of the old password
	// combined with the first secret stored in the application's context. The result of the hash is stored in the hashed
	// array at index 0.
	hashed[0] = sha256.New()
	hashed[0].Write([]byte(fmt.Sprintf("%v-%v", oldPassword, a.Context.Secrets[0])))

	// The purpose of this code is to use the SHA256 cryptographic hash algorithm to create a new hash of the old password
	// combined with the first secret stored in the application's context. The result of the hash is stored in the hashed
	// array at index 1.
	hashed[1] = sha256.New()
	hashed[1].Write([]byte(fmt.Sprintf("%v-%v", newPassword, a.Context.Secrets[0])))

	// This code is checking to see if the length of the new password is at least 8 characters long. If the new password is
	// not 8 characters long, it will return an error message that the password must be at least 8 characters long.
	if len(newPassword) < 8 {
		return status.Error(18863, "the password must be at least 8 characters long")
	}

	// This code is used to compare two hashed passwords. If comparing two hashed passwords returns true, then the code will
	// return an error with the status code 72554 and a message indicating that the new password must not be identical to
	// the old one. The purpose of this code is to ensure that users are not able to set the same password as their previous one.
	if string(hashed[0].Sum(nil)) == string(hashed[1].Sum(nil)) {
		return status.Error(72554, "the new password must not be identical to the old one")
	}

	// This code is attempting to query a database to find the ID of an account that matches the given id and hashed
	// password. The hashed password is encoded to a URL encoded string before being compared. If an error occurs while
	// querying the database, the a.Context.Error(err) function is called to handle the error. The row.Close() function is
	// used to defer closing the row until the end of the function, ensuring that all queries are properly closed.
	row, err := a.Context.Db.Query("select id from accounts where id = $1 and password = $2", id, base64.URLEncoding.EncodeToString(hashed[0].Sum(nil)))
	if err != nil {
		return a.Context.Error(err)
	}
	defer row.Close()

	// The purpose of the if row.Next() statement is to check whether the current row is valid or not and move to the next
	// row if it is. This allows the program to iterate through the database and access the required data.
	if row.Next() {

		// This code is updating the password of an account with a given id in a database. The password is encoded with
		// base64.URLEncoding.EncodeToString and the hashed[1].Sum(nil) is passed as the new password. The if statement checks
		// for any errors that may occur during the update, and if an error is encountered, it is returned.
		if _, err := a.Context.Db.Exec(`update accounts set password = $2 where id = $1`, id, base64.URLEncoding.EncodeToString(hashed[1].Sum(nil))); err != nil {
			return err
		}

		return nil
	}

	return status.Error(44754, "the old password was entered incorrectly")
}

// setSample - This code is part of a Service class in the pbaccount package. The purpose of this function is to set the sample field
// of a specific account identified by the id int64 parameter. It will check if the index string parameter is in the
// column array and if it is, it will either remove or add the index to the sample field of the account. It will then
// return an error if any of the operations fail.
func (a *Service) setSample(id int64, index string) error {

	// The purpose of this code is to create three variables, response, column, and query, for use in a program. The first
	// variable, response, is a pbaccount.ResponseUser type. The second variable, column, is an array of strings containing
	// the values "order_filled", "withdrawal", "login", and "news". The third variable, query, is an empty string array.
	var (
		response pbaccount.ResponseUser
		column   = []string{"order_filled", "withdrawal", "login", "news"}
		query    []string
	)

	// The purpose of this code is to check if an index exists in a column. If the index is not found, the code returns an
	// error with the code 10504 and the message "incorrect sample index".
	if !help.IndexOf(column, index) {
		return status.Error(10504, "incorrect sample index")
	}

	// This statement is used to query a database for a specific record that matches the given parameters. It is checking to
	// see if a record exists in the accounts table with the given id and index values. If a record is found, the
	// response.Count will be set to the count of the found records, otherwise, an error is returned.
	if err := a.Context.Db.QueryRow(fmt.Sprintf(`select count(*) as count from accounts where id = %d and sample @> '"%s"'::jsonb`, id, index)).Scan(&response.Count); err != nil && err != sql.ErrNoRows {
		return err
	}

	// The code above is assigning a value to the query variable depending on the value of the response.Count variable. If
	// response.Count is greater than 0, query will be assigned the value of the formatted string using the fmt.Sprintf
	// function - `sample = sample - '%s'`, with the index variable as the argument. If response.Count is 0 or less, query
	// will be assigned the value of the formatted string using the fmt.Sprintf function - `sample = sample || '"%s"'`, with
	// the index variable as the argument.
	if response.Count > 0 {
		query = append(query, fmt.Sprintf(`sample = sample - '%s'`, index))
	} else {
		query = append(query, fmt.Sprintf(`sample = sample || '"%s"'`, index))
	}

	// This code is used to execute an update statement in a database. The first line of code is using the Db.Exec()
	// function to execute a SQL query. The query updates the accounts table with the specified data. The ID of the account
	// to be updated is specified using the id argument. The query argument is a string that contains the fields and values
	// that the account should be updated with. The strings.Join() function is used to join the query string with the other
	// components of the query. If the query is successful, no error is returned. Otherwise, an error is returned.
	_, err := a.Context.Db.Exec(fmt.Sprintf(`update accounts set %[2]s where id = %[1]d`, id, strings.Join(query, "")))
	if err != nil {
		return err
	}

	return nil
}
