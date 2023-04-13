package admin_ads

import (
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets/common/query"
	admin_pbads "github.com/cryptogateway/backend-envoys/server/proto/v1/admin.pbads"
	"github.com/cryptogateway/backend-envoys/server/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc/status"
)

// SetAdvertising - This function is a service function that is used to set an advertisement rule for the application. It checks for
// authentication, and whether the user has the necessary rights to set the rule. It then checks the title and text
// lengths, and inserts or updates the record in the database. Finally, it uploads the image associated with the
// advertisement and sets the response status to true.
func (s *Service) SetAdvertising(ctx context.Context, req *admin_pbads.SetRequestAdvertising) (*admin_pbads.ResponseAdvertising, error) {

	// The purpose of the above code is to declare two variables in the same statement. The first variable, response, is of
	// type admin_pbads.ResponseAdvertising. The second variable, migrate, is of type query.Migrate and has a Context field set to s.Context.
	var (
		response admin_pbads.ResponseAdvertising
		migrate  = query.Migrate{
			Context: s.Context,
		}
	)

	// This code is checking for authentication. The purpose of this code is to ensure that the user is authenticated before
	// continuing with the request. The s.Context.Auth() function is used to check the authentication and if there is an
	// error it will be returned with the s.Context.Error() function. If the authentication check is successful, the request will continue.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This code checks if a user has the appropriate permissions to write and edit data. If the user does not have the
	// correct rules, then an error is returned.
	if !migrate.Rules(auth, "advertising", query.RoleDefault) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// The purpose of this if statement is to check the type of Advertising object within the req object. If the type of
	// Advertising object is equal to the types.Pattern_TEXT type, then the code within the block will be executed.
	if req.Advertising.GetPattern() == types.PatternText {

		// This code is checking if the title of an advertising request is less than 10 characters or the text of the request
		// is less than 20 characters, and if it is, it will return an error with a status code of 15431 saying "title must be
		// at least 10 characters, text must be at least 20 characters". This is likely to ensure that the advertising request
		// is valid and meets certain requirements.
		if len(req.Advertising.GetTitle()) < 10 && len(req.Advertising.GetTitle()) < 20 {
			return &response, status.Error(15431, "title must be at least 10 characters, text must be at least 20 characters")
		}
	}

	// This is a conditional statement that checks if the value of req.GetId() is greater than 0. This is likely being done
	// to ensure that the value of req.GetId() is valid and not a negative number.
	if req.GetId() > 0 {

		// This code is part of a function that updates an advertisement in a database. The purpose of this code is to execute
		// an SQL query that updates the advertisement in the database with the supplied data. The query takes the values of
		// the title, text, link, type, and id of the advertisement from the request and updates the corresponding fields in
		// the database. If the query fails, an error is returned.
		if _, err := s.Context.Db.Exec("update advertising set title = $1, text = $2, link = $3, pattern = $4 where id = $5;",
			req.Advertising.GetTitle(),
			req.Advertising.GetText(),
			req.Advertising.GetLink(),
			req.Advertising.GetPattern(),
			req.GetId(),
		); err != nil {
			return &response, err
		}

	} else {

		// This code is used to insert new advertising data into the database. Specifically, it is inserting data from the
		// request object, req, into the advertising table in the database. The data being inserted includes the title, text,
		// link, and type of the advertising as specified by the req object. After the data is inserted, the id associated with
		// the newly added advertising data is returned and stored in the request object, req. If there is an error encountered when inserting the data into the database, the code will return an error.
		if err := s.Context.Db.QueryRow("insert into advertising (title, text, link, pattern) values ($1, $2, $3, $4) returning id;",
			req.Advertising.GetTitle(),
			req.Advertising.GetText(),
			req.Advertising.GetLink(),
			req.Advertising.GetPattern(),
		).Scan(&req.Id); err != nil {
			return &response, err
		}

	}

	// This is a conditional statement that is checking to see if the length of the req.GetImage() is greater than 0. This
	// is used to check if there is anything in the req.GetImage() and if there is, then the code within the if statement will execute.
	if len(req.GetImage()) > 0 {

		// This code is checking for an error when migrating an image, and if an error is encountered, it returns an error
		// response. To migrate.Image() function is likely used to resize an image according to the given parameters (640 and
		// 340). The fmt.Sprintf() function is used to format the ID number as a string.
		if err := migrate.Image(req.GetImage(), "ads", fmt.Sprintf("%d", req.GetId()), 640, 340); err != nil {
			return &response, err
		}
	}
	response.Success = true

	return &response, nil
}

// DeleteAdvertising - This function is used to delete an advertising record from a database. It first checks for authorization and the
// necessary roles for writing and editing data. If the user has the necessary roles, it executes a delete statement to
// remove the record from the database and removes any related resources. Finally, it returns a success response.
func (s *Service) DeleteAdvertising(ctx context.Context, req *admin_pbads.DeleteRequestAdvertising) (*admin_pbads.ResponseAdvertising, error) {

	// The purpose of this code is to create a new variable 'response' of type admin_pbads.ResponseAdvertising and a new variable
	// 'migrate' of type query.Migrate. To migrate variable is given the s.Context value, which is the context associated
	// with the current request.
	var (
		response admin_pbads.ResponseAdvertising
		migrate  = query.Migrate{
			Context: s.Context,
		}
	)

	// This code is used to authenticate a user. The s.Context.Auth(ctx) function attempts to authenticate the user using
	// the ctx context. If it succeeds, it returns an auth object. If it fails, it returns an error. The if statement checks
	// if an error occurred. If so, it returns an empty response and the error.
	auth, err := s.Context.Auth(ctx)
	if err != nil {
		return &response, err
	}

	// This if statement is checking the user's authorization to write and edit data. If the user does not have the
	// necessary rules, the statement will return an error message indicating they do not have the appropriate permissions.
	if !migrate.Rules(auth, "advertising", query.RoleDefault) || migrate.Rules(auth, "deny-record", query.RoleDefault) {
		return &response, status.Error(12011, "you do not have rules for writing and editing data")
	}

	// This code is checking for an error when deleting an entry from the advertising table in a database. It is using the
	// Exec() method to attempt to delete the entry with the ID specified in the req parameter. If there is an error, it
	// will return &response and an error message.
	if _, err = s.Context.Db.Exec("delete from advertising where id = $1", req.GetId()); err != nil {
		return &response, err
	}

	// The purpose of this code is to remove files associated with an ID number that is being passed as a parameter to the function. The if statement is
	// used to check if an error occurs while attempting to remove the files, and if there is an error, it will be returned
	// as part of the response, along with a context error.
	if err := migrate.RemoveFiles("ads", fmt.Sprintf("%d", req.GetId())); err != nil {
		return &response, err
	}
	response.Success = true

	return &response, nil
}
