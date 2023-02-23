package kycaid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"github.com/pkg/errors"
	"net/http"
)

const (
	MethodApplicants          = "applicants"
	MethodForms               = "forms"
	TypeVerificationCompleted = "VERIFICATION_COMPLETED"
	StatusCompleted           = "completed"
)

// Api - is a struct is used to define a type for creating API objects that are used to make API requests. The Key string is
// used to store an authentication key, and the Client *http.Client is used to store an http client object that is used
// to make requests.
type Api struct {
	Id     string
	Key    string
	Client *http.Client
}

// NewApi - this function creates and returns a new Api struct with the given key and http client. If a client is not provided, it
// will create a new http.Client automatically.
func NewApi(id, key string, client *http.Client) *Api {

	// This code is checking if the "client" variable is nil or not. If it is, it creates a new http.Client object and
	// assigns it to the "client" variable. This is done so that the code has a valid http.Client object to use when making requests.
	if client == nil {
		client = &http.Client{}
	}

	// The purpose of this code is to create and initialize a new instance of an Api struct, which is a type of struct used
	// to store data related to an API. The code assigns values to the fields of the struct and then returns a pointer to
	// the new instance of the Api struct.
	return &Api{
		Id:     id,
		Key:    key,
		Client: client,
	}
}

// request - the purpose of this code is to create an HTTP request with a given path, body, and method. It sets the authorization
// header, content-type header, and makes the request to a remote server. It then returns the response, or an error if one occurs.
func (p *Api) request(path string, body map[string]string, method string) (response *http.Response, err error) {

	var (
		serialize []byte
	)

	// This code is checking the length of the variable "body" and, if it is greater than 0, it is attempting to serialize
	// the variable "body" into JSON format. If there is an error encountered while attempting to serialize, the code will
	// return an error response.
	if len(body) > 0 {
		serialize, err = json.Marshal(body)
		if err != nil {
			return response, err
		}
	}

	// This code is creating a new HTTP request using the "POST" method and a URL with a specified method. It is also
	// providing the request body with a JSON data. If an error occurs, it will return nil and the error.
	req, err := http.NewRequest(method, fmt.Sprintf("https://api.kycaid.com/%v", path), bytes.NewBuffer(serialize))
	if err != nil {
		return response, err
	}

	// This code is setting the Authorization header to a value of "Token" followed by the value of the p.Key variable. This
	// is likely being used in an HTTP request to authenticate the user who is making the request.
	req.Header.Set("Authorization", "Token "+p.Key)

	// The purpose of req.Header.Set("Content-Type", "application/json") is to set the Content-Type header of an HTTP
	// request to application/json. This informs the server of the type of data that is being sent in the request body, so
	// that the server can correctly process and respond to the request.
	req.Header.Set("Content-Type", "application/json")

	// This code is used to make an HTTP request to a remote server and retrieve the response. It uses the Client.Do()
	// function from the http package to make the request, and the resp and err variables to store the response and any
	// errors that may occur. If an error occurs, the code returns nil and the error to the calling function.
	response, err = p.Client.Do(req)
	if err != nil {
		return response, err
	}

	return response, nil
}

// CreateApplicants - this function is used to create an applicant using a mapping of parameters. It sends a request to the MethodApplicants
// endpoint using the specified parameters and then decodes the response as an ApplicantsResponse type. Finally, it
// returns the applicant ID as a string or an error.
func (p *Api) CreateApplicants(param map[string]string) (*pbaccount.ResponseKycApplicant, error) {

	// The purpose of this statement is to declare a variable called "response" of type ApplicantsResponse. This variable
	// can be used to store values or references to objects of type ApplicantsResponse.
	var (
		response pbaccount.ResponseKycApplicant
	)

	// This code is used to make an API request using a specified method (MethodApplicants) with a given set of parameters
	// (param). If there is an error, it returns an empty string and the error. It then closes the body of the response once
	// the request is complete.
	request, err := p.request(MethodApplicants, param, http.MethodPost)
	if err != nil {
		return &response, err
	}
	defer request.Body.Close()

	// This code is used to decode a JSON object from a http request body and store it in the response variable. If there
	// is an error during decoding, the function will return the response variable and the error.
	if err = json.NewDecoder(request.Body).Decode(&response); err != nil {
		return &response, err
	}

	// The code snippet is checking if the length of the 'response.ApplicantID' is 0, and if it is, it is returning an error
	// message that "Not create new applicants". This code is used to check if the ApplicantID is empty or not and to return
	// an error if it is.
	if len(response.ApplicantId) == 0 {
		return &response, errors.New("not created new applicant")
	}

	return &response, nil
}

// CreateForm - the purpose of this code is to make an HTTP request to a given URL with a set of parameters, decode the request body
// in the JSON format, and check if the response is valid. If it is valid, it will return the response, otherwise, it
// will return an error message.
func (p *Api) CreateForm(param map[string]string) (*pbaccount.FormResponse, error) {

	// This is a variable declaration, which is used to create a variable named "response" of type "FormResponse". The
	// purpose of this is to allocate memory and store a value, which can be accessed and used in the program.
	var (
		response pbaccount.FormResponse
	)

	//This code is making an HTTP request to a given URL with a set of parameters. The request function is making the
	//request and fmt.Sprintf is formatting the URL with the given parameters. If there is an error, the code returns an error.
	request, err := p.request(fmt.Sprintf("%v/%v/urls", MethodForms, p.Id), param, http.MethodPost)
	if err != nil {
		return &response, err
	}
	defer request.Body.Close()

	// This code is used to decode a JSON object from a http request body and store it in the response variable. If there
	// is an error during decoding, the function will return the response variable and the error.
	if err = json.NewDecoder(request.Body).Decode(&response); err != nil {
		return &response, err
	}

	// The code snippet is checking if the length of the 'response.FormUrl' is 0, and if it is, it is returning an error
	// message that "Not create new applicants". This code is used to check if the ApplicantID is empty or not and to return
	// an error if it is.
	if len(response.FormUrl) == 0 {
		return &response, errors.New("not created new form")
	}

	return &response, nil
}

// GetApplicantsById - the purpose of this code is to make an API request to get data about an applicant and store it in an
// ApplicantsResponse variable. It then checks if the ApplicantID is empty and returns an error if it is. Finally, it
// returns an ApplicantsResponse variable and a nil error if the request is successful.
func (p *Api) GetApplicantsById(id string) (*pbaccount.ResponseKycApplicant, error) {

	// The variable 'response' is declared to be of type 'ApplicantsResponse'. This variable is used to store data related
	// to an applicant's response. It may contain information like a response to a survey, answers to questions on a job application, etc.
	var (
		response pbaccount.ResponseKycApplicant
	)

	// This code is used to make an API request using a specified method (MethodApplicants) with a given set of parameters
	// (param). If there is an error, it returns an empty string and the error. It then closes the body of the response once
	// the request is complete.
	request, err := p.request(fmt.Sprintf("%v/%v", MethodApplicants, id), nil, http.MethodGet)
	if err != nil {
		return &response, err
	}
	defer request.Body.Close()

	// This code is used to decode a JSON object from a http request body and store it in the response variable. If there
	// is an error during decoding, the function will return the response variable and the error.
	if err = json.NewDecoder(request.Body).Decode(&response); err != nil {
		return &response, err
	}

	// The code snippet is checking if the length of the 'response.ApplicantID' is 0, and if it is, it is returning an error
	// message that "Not create new applicants". This code is used to check if the ApplicantID is empty or not and to return
	// an error if it is.
	if len(response.ApplicantId) == 0 {
		return &response, errors.New("not get applicant")
	}

	return &response, nil
}
