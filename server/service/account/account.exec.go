package account

import (
	"encoding/json"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
)

// Modify - struct is a type of struct used to store two slices of bytes. The purpose of this struct is to store data
// related to modifications, such as sample data and rules for the modifications. This data can then be used for various
// purposes, such as for making changes to a system or for providing input to a program.
type Modify struct {
	Sample, Rules []byte
}

// QueryUser - This function is used to query a user from a database given an ID. It scans the database for the requested user, and
// then uses JSON unmarshalling to convert the data from the database into the appropriate fields in the response object.
// It returns the response object, containing the requested user's information, or an error if one occurs.
func (a *Service) QueryUser(id int64) (*pbaccount.User, error) {

	// The purpose of the above code is to declare two variables. The first variable, response, is a User type from the
	// pbaccount package. The second variable, q, is a Modify type. These two variables can then be used in the code
	// following this declaration.
	var (
		response pbaccount.User
		q        Modify
	)

	// This code is used to query the database for a specific row using the "id" variable. It then assigns the retrieved row
	// values to the response struct, which holds the values to be returned to the user. If an error occurs during the
	// query, it is returned to the user instead.
	if err := a.Context.Db.QueryRow("select id, name, email, status, sample, rules, factor_secure, factor_secret from accounts where id = $1", id).Scan(&response.Id, &response.Name, &response.Email, &response.Status, &q.Sample, &q.Rules, &response.FactorSecure, &response.FactorSecret); err != nil {
		return &response, err
	}

	// This code is using the json.Unmarshal function to convert a JSON object into a variable of type response.Sample. If
	// there is an error during the conversion, it returns the response variable and the error.
	if err := json.Unmarshal(q.Sample, &response.Sample); err != nil {
		return &response, err
	}

	// This code is trying to convert a JSON object stored in the variable "q.Rules" into a response.Rules object. If there
	// is an error while trying to do this, the code returns the response object and the error.
	if err := json.Unmarshal(q.Rules, &response.Rules); err != nil {
		return &response, err
	}

	return &response, nil
}
