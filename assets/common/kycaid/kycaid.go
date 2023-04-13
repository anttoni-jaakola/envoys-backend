package kycaid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbkyc"
	"github.com/cryptogateway/backend-envoys/server/types"
	"github.com/pkg/errors"
	"net/http"
)

const (
	MethodApplicants          = "applicants"
	MethodForms               = "forms"
	TypeVerificationCompleted = "VERIFICATION_COMPLETED"
	StatusCompleted           = "completed"
)

// Kyc - The Kyc struct is used to store information related to Know Your Customer (KYC) processes. It contains an API key, as
// well as three structs (S, P, and C) that store information related to the documents required to complete a KYC
// process. These documents may include a passport, proof of address, and proof of identity. Each struct contains a key
// and multiplication value, which is likely used to specify the number of documents that need to be submitted for the
// KYC process. The Kyc struct also contains a RedirectUrl field, which is likely used to redirect the user to a specific
// page after the KYC process is completed.
type Kyc struct {
	ApiKey string
	Forms  struct {
		S struct {
			Key            string
			Multiplication int
		}
		P struct {
			Key            string
			Multiplication int
		}
		C struct {
			Key            string
			Multiplication int
		}
	}
	RedirectUrl string
}

// Api - is a struct is used to define a type for creating API objects that are used to make API requests. The Key string is
// used to store an authentication key, and the Client *http.Client is used to store an http client object that is used
// to make requests.
type Api struct {
	Kyc    Kyc
	Client *http.Client
}

// NewApi - The purpose of this code is to create a new Api object from the given params and http.Client objects. It does this by
// first marshaling the params object into a serialized format and then unmarshaling it into a kyc object. It then checks
// if the client object is nil, and if it is, it creates a new http.Client object and assigns it to the client variable.
// Finally, it returns a new Api object with the Kyc and Client objects set.
func NewApi(params interface{}, client *http.Client) (*Api, error) {

	var (
		kyc Kyc
	)

	// This code is used to convert a given data structure (params) into a serialized format using the json.Marshal()
	// function. If an error occurs during the conversion, the function returns nil and the error.
	serialize, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// The purpose of this code snippet is to unmarshal a serialized JSON object into a kyc object, and to return an error
	// if it is not successful.
	if err = json.Unmarshal(serialize, &kyc); err != nil {
		return nil, err
	}

	// This code is checking if the "client" variable is nil or not. If it is, it creates a new http.Client object and
	// assigns it to the "client" variable. This is done so that the code has a valid http.Client object to use when making requests.
	if client == nil {
		client = &http.Client{}
	}

	return &Api{
		Kyc:    kyc,
		Client: client,
	}, nil
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
	req.Header.Set("Authorization", "Token "+p.Kyc.ApiKey)

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
func (p *Api) CreateApplicants(param map[string]string) (*pbkyc.ResponseApplicant, error) {

	// The purpose of this statement is to declare a variable called "response" of type ApplicantsResponse. This variable
	// can be used to store values or references to objects of type ApplicantsResponse.
	var (
		response pbkyc.ResponseApplicant
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
func (p *Api) CreateForm(param map[string]string, _type string) (*pbkyc.FormResponse, error) {

	// This is a variable declaration, which is used to create a variable named "response" of type "FormResponse". The
	// purpose of this is to allocate memory and store a value, which can be accessed and used in the program.
	var (
		response pbkyc.FormResponse
	)

	// The purpose of the switch form is to assign the correct form ID to the response based on the type of form. The switch
	// statement is used to evaluate an expression and depending on the value of the expression, it will perform a different
	// code block. In this case, the expression is the type of form and depending on the type of form, the response.FormId
	// will be set to the appropriate key.
	switch _type {
	case types.KYCLevel1:
		response.FormId = p.Kyc.Forms.S.Key
	case types.KYCLevel2:
		response.FormId = p.Kyc.Forms.P.Key
	case types.KYCLevel3:
		response.FormId = p.Kyc.Forms.C.Key
	}

	//This code is making an HTTP request to a given URL with a set of parameters. The request function is making the
	//request and fmt.Sprintf is formatting the URL with the given parameters. If there is an error, the code returns an error.
	request, err := p.request(fmt.Sprintf("%v/%v/urls", MethodForms, response.GetFormId()), param, http.MethodPost)
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
func (p *Api) GetApplicantsById(id string) (*pbkyc.ResponseApplicant, error) {

	// The variable 'response' is declared to be of type 'ApplicantsResponse'. This variable is used to store data related
	// to an applicant's response. It may contain information like a response to a survey, answers to questions on a job application, etc.
	var (
		response pbkyc.ResponseApplicant
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
