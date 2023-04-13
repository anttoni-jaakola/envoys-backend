package kyc

import (
	"context"
	"github.com/cryptogateway/backend-envoys/assets/common/kycaid"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbkyc"
	"github.com/cryptogateway/backend-envoys/server/service/v2/account"
	"strconv"
	"strings"
)

// SetCanceled The purpose of this code is to update a KYC (Know Your Customer) table in a database in order to set the secret, type,
// and process fields for a given user_id. It also ensures that only authorized users can access the requested resource
// by checking for a valid authentication. Finally, it returns a response containing a success message if the update was successful.
func (s *Service) SetCanceled(ctx context.Context, _ *pbkyc.SetRequestCanceled) (*pbkyc.ResponseCanceled, error) {

	// The purpose of this code is to declare a variable named "response" of type "pbkyc.ResponseCanceled". This variable
	// will be used to store a response from a KYC (Know Your Customer) request.
	var (
		response pbkyc.ResponseCanceled
	)

	// This code checks for a valid authentication and if there is an error with the authentication it will return an error
	// response. It is used to ensure that only authorized users can access the requested resource.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code is updating the KYC (Know Your Customer) table in a database. The code is updating the secret, type, and
	// process fields of the KYC table, where the user_id field matches the one given in the parameters. The purpose of
	// this code is to update the KYC table with new values for the given user_id.
	if _, err := s.Context.Db.Exec(`update kyc set secret = $2, level = $3, process = $4 where user_id = $1`, auth, "", "level_0", false); err != nil {
		return &response, err
	}
	response.Success = true

	return &response, nil
}

// GetStatus - This function is used to get the status of a KYC request from a database. It takes in a GetRequestStatus request and
// returns a ResponseStatus response. The code queries the database for a user_id that matches the given auth value and
// checks for any errors that occur. If an error occurs, the code returns with an error message. The code then scans the
// results of the query into two variables and returns the response and an error if the scan fails. Finally, the function
// returns the response and nil if successful.
func (s *Service) GetStatus(_ context.Context, req *pbkyc.GetRequestStatus) (*pbkyc.ResponseStatus, error) {

	// The variable 'response' is used to store a value of the type pbkyc.ResponseStatus. It is a variable declaration
	// statement used to create a variable of a certain type. The variable can then be used in the code to store values of that type.
	var (
		response pbkyc.ResponseStatus
	)

	// This code is querying a database for a user_id that matches a given auth value. The row variable will contain the
	// result of the query. The code will check for any errors that occur when querying the database. If an error occurs,
	// the code will return with an error message. The defer statement will ensure that the row is closed when the function
	// is finished executing.
	row, err := s.Context.Db.Query(`select process, secure, level from kyc where user_id = $1`, req.GetId())
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The purpose of the if statement is to determine whether the row.Next() function returns true or false. If the
	// function returns true, it means that there is a next row in the result set and the code inside the if statement will
	// be executed. If the function returns false, it means that there is no next row and the code inside the if statement
	// will not be executed.
	if row.Next() {

		// This code is part of an if statement that is used to scan the results of a database query into two variables,
		// response.Process and response.Secure and response.Type. If the scan is successful, it will return the response and continue. If the
		// scan fails, it will return the response and an error.
		if err = row.Scan(&response.Process, &response.Secure, &response.Level); err != nil {
			return &response, err
		}
	}

	return &response, nil
}

// GetPrivilege - The purpose of this function is to return a response containing a multiplication value based on the type of request
// (standard, premium, or corporate). The function takes in a request parameter of type pbkyc.GetRequestPrivilege, which
// is used to determine the type of request. The function then uses a switch statement to set the response.Multiplication
// value according to the matching key in the Kyc forms. Finally, the function returns a pointer to the response and a nil error.
func (s *Service) GetPrivilege(_ context.Context, _ *pbkyc.GetRequestPrivilege) (*pbkyc.ResponsePrivilege, error) {

	// This variable declaration creates a new variable called "response" of type "pbkyc.ResponsePrivilege". The purpose of
	// this variable is to store data of type "pbkyc.ResponsePrivilege".
	var (
		response pbkyc.ResponsePrivilege
	)

	// This code snippet is creating a map called multiplication that stores the multiplication factors for three different
	// types of KYC forms: standard, premium, and corporate. The factor for each type of form is retrieved from the
	// Context.Kyc.Forms object.
	multiplication := make(map[string]int32)
	multiplication["level_1"] = s.Context.Kyc.Forms.S.Multiplication
	multiplication["level_2"] = s.Context.Kyc.Forms.P.Multiplication
	multiplication["level_3"] = s.Context.Kyc.Forms.C.Multiplication

	response.Multiplication = multiplication

	return &response, nil
}

// SetProcess - This code is used to set a process for a KYC (Know Your Customer) request. It authenticates the user, creates an
// applicant, creates a form, and stores the form information in a response object. The code also checks for errors and
// returns an error response if one is encountered.
func (s *Service) SetProcess(ctx context.Context, req *pbkyc.SetRequestProcess) (*pbkyc.ResponseProcess, error) {

	// The purpose of this code is to declare a variable named "response" of type "pbkyc.ResponseKyc". This variable
	// will be used to store a response from a KYC (Know Your Customer) request.
	var (
		response pbkyc.ResponseProcess
	)

	// This code checks for a valid authentication and if there is an error with the authentication it will return an error
	// response. It is used to ensure that only authorized users can access the requested resource.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// The purpose of this code is to create a Service object that uses the context stored in the variable e. The Service
	// object is then assigned to the variable migrate.
	migrate := account.Service{
		Context: s.Context,
	}

	// This code is attempting to query a user from migrate using the provided authentication credentials (auth). If the
	// query fails, an error is returned.
	user, err := migrate.QueryUser(auth)
	if err != nil {
		return &response, err
	}

	// This code is creating applicants in a KYC (Know Your Customer) process. The purpose is to create an applicant in the
	// KYC process using the user's email and setting the type as "PERSON". If there is an error, it will return an error
	// response and an error message.
	applicants, err := s.Context.KycProvider.CreateApplicants(map[string]string{
		"type":                 strings.ToUpper(req.GetForm()),
		"email":                user.GetEmail(),
		"company_name":         "You company name",
		"registration_country": "UA",
		"phone":                "+199999999999",
		"business_activity_id": "02445235095c674b820a9d948e3b27824bc4",
	})
	if err != nil {
		return &response, err
	}

	// This code is querying a database for a user_id that matches a given auth value. The row variable will contain the
	// result of the query. The code will check for any errors that occur when querying the database. If an error occurs,
	// the code will return with an error message. The defer statement will ensure that the row is closed when the function
	// is finished executing.
	row, err := s.Context.Db.Query(`select user_id from kyc where user_id = $1`, auth)
	if err != nil {
		return &response, err
	}
	defer row.Close()

	// The purpose of the if statement is to determine whether the row.Next() function returns true or false. If the
	// function returns true, it means that there is a next row in the result set and the code inside the if statement will
	// be executed. If the function returns false, it means that there is no next row and the code inside the if statement
	// will not be executed.
	if row.Next() {

		// This code is updating the KYC (Know Your Customer) table in a database. The code is updating the secret, type, and
		// process fields of the KYC table, where the user_id field matches the one given in the parameters. The purpose of
		// this code is to update the KYC table with new values for the given user_id.
		if _, err := s.Context.Db.Exec(`update kyc set secret = $2, level = $3, process = $4 where user_id = $1`, auth, applicants.GetApplicantId(), req.GetLevel(), true); err != nil {
			return &response, err
		}

	} else {

		// This code is inserting data into a database table called 'kyc'. It is taking four values from the variables auth,
		// applicants.GetApplicantId(), req.GetType(), and true, and inserting them into the kyc table. The purpose of the 'if'
		// statement is to check for errors that may occur during the insertion of data. If an error is encountered, the
		// function will return an error response.
		if _, err = s.Context.Db.Exec("insert into kyc (user_id, secret, level, process) values ($1, $2, $3, $4)", auth, applicants.GetApplicantId(), req.GetLevel(), true); err != nil {
			return &response, err
		}
	}

	// This code is creating a form with specific information and assigning it to a variable. It is then checking if there
	// is an error while creating the form and if there is an error, it will return an error response. The purpose of this
	// code is to create a form to be used for KYC (Know Your Customer) verification with specific information included.
	form, err := s.Context.KycProvider.CreateForm(map[string]string{
		"applicant_id":          applicants.GetApplicantId(),
		"external_applicant_id": strconv.FormatInt(auth, 10),
		"redirect_url":          s.Context.Kyc.RedirectUrl,
	}, req.GetLevel())
	if err != nil {
		return &response, err
	}

	// The purpose of these lines of code is to store the form information in the response object. The formId, formUrl, and
	// the verificationId are all stored in the response object so that they can be accessed and used later.
	response.FormId = form.FormId
	response.FormUrl = form.FormUrl
	response.VerificationId = form.VerificationId

	return &response, nil
}

// SetCallback - This function is used to handle a KYC (Know Your Customer) verification process. It checks the type of request and if
// it matches a certain type, it updates the 'kyc_secure' column in the 'accounts' table to true or false depending on
// whether the verification is valid. If the request is of type 'pending', the code sets the response type to
// 'pending'. The code also adds any comments associated with the request's profile or document to the response. Finally,
// the code publishes a response to an exchange on the topic "account/kyc-verify".
func (s *Service) SetCallback(_ context.Context, req *pbkyc.SetRequestCallback) (*pbkyc.ResponseCallback, error) {

	// This variable called "response" is used to store the response of a pbaccount.ResponseKycVerification. This variable
	// is used to store the response of a KYC verification process, which is used to verify the identity of a person or
	// organization.
	var (
		response pbkyc.ResponseCallback
	)

	// The purpose of this statement is to check the type of request (req) and determine if it matches the type
	// "kycaid.TypeVerificationCompleted". If it does match, the statement will evaluate to true and the code following the
	// statement will be executed. If the request is of type 'pending', the code sets the response type to 'pending'.
	if req.GetType() == kycaid.TypeVerificationCompleted {

		// The purpose of this code is to check if the verification for a request is valid and if it has been verified. The code
		// checks if the verification ID matches the request and if the request has been verified. If both conditions are true,
		// then this code will execute.
		if req.GetVerified() && req.GetStatus() == kycaid.StatusCompleted {

			if err := s.Context.Db.QueryRow(`update kyc set secure = $2, process = $3 where secret = $1 returning user_id`, req.GetApplicantId(), true, false).Scan(&response.Id); err != nil {
				return &response, err
			}

			response.Status = "completed"
		}

		// This code is checking if the profile and document verifications have been completed. If they have not been verified,
		// then the code updates the 'secure' column in the 'accounts' table to 'false' and the 'process' column to
		// 'false'. It then returns an error message with any comments from the verifications.
		if !req.GetVerifications().Profile.GetVerified() || !req.GetVerifications().Document.GetVerified() {

			if err := s.Context.Db.QueryRow(`update kyc set secure = $2, process = $3 where secret = $1 returning user_id`, req.GetApplicantId(), false, false).Scan(&response.Id); err != nil {
				return &response, err
			}

			response.Status = "error"

			// This code checks if the length of the request's GetVerifications().Profile.GetComment() is greater than 0. If it
			// is, then the response.Messages array is appended with the contents of the request's
			// GetVerifications().Profile.GetComment(). The purpose of this code is to make sure that any comments associated with the request's profile are included in the response.
			if len(req.GetVerifications().Profile.GetComment()) > 0 {
				response.Messages = append(response.Messages, req.GetVerifications().Profile.GetComment())
			}

			// This code is checking if the length of the "GetComment" field in the "Document" field of the "GetVerifications"
			// request is greater than 0. If it is, it adds the value of the "GetComment" field to the "Messages" field of the
			// response. This is likely done to add a message to the response based on the content of the "GetComment" field.
			if len(req.GetVerifications().Document.GetComment()) > 0 {
				response.Messages = append(response.Messages, req.GetVerifications().Document.GetComment())
			}
		}

	} else {
		response.Status = "pending"
	}

	// This code is used to publish a response to an exchange on the topic "account/kyc-verify". If there is an error while
	// publishing the response, the function will return the response and an error. The a.Context.Debug function is used to
	// print out any errors encountered.
	if err := s.Context.Publish(&response, "exchange", "account/kyc-verify"); s.Context.Debug(err) {
		return &response, err
	}

	return &response, nil
}

// GetApplicant - This function is part of a service for a KYC (Know Your Customer) provider. It is used to retrieve applicants from the
// provider by their ID, and check if an error occurred during the process. If an error did occur, it will return an
// error response.
func (s *Service) GetApplicant(_ context.Context, req *pbkyc.GetRequestApplicant) (*pbkyc.ResponseApplicant, error) {

	// The purpose of the above code is to declare a new variable called 'response' of type 'pbkyc.ResponseApplicant'. This
	// variable is used to store a response from an Applicant when a request is made.
	var (
		response *pbkyc.ResponseApplicant
	)

	// This code is retrieving applicants from a KYC (Know Your Customer) Provider by their ID, then checking if an error
	// occurred during the process. If an error did occur, it will return an error response.
	response, err := s.Context.KycProvider.GetApplicantsById(req.GetId())
	if err != nil {
		return response, err
	}

	return response, nil //6145b9660cb3844d6818adf134cf7dc03655
}
