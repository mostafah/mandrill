// Package mandrill gives a simple interface for sending email through
// Mandrill's API, documented at https://mandrillapp.com/api/docs/.
//
// This is not a full implementation of the API and only provides some
// essential calls.
package mandrill

import (
	"fmt"
	"github.com/jmcvetta/restclient"
)

// API key for Mandrill user. You should set this to your API key before calling
// any of the functions. You can get a API key for your account in your
// Mandrill account settings.
var Key string

// type Error holds error return messages from API calls.
type Error struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// newError returns a new Error instance.
func newError() *Error {
	return &Error{}
}

// Error porduces error message for err.
func (err *Error) Error() string {
	return fmt.Sprintf("mandrill: %s: %s", err.Name, err.Message)
}

// do is an easy function for performing requests against Mandrill's API.
func do(url string, data interface{}, result interface{}) error {
	err := newError()

	rr := &restclient.RequestResponse{
		Url:    "https://mandrillapp.com/api/1.0",
		Method: "POST",
		Data:   data,
		Result: result,
		Error:  err}

	status, _ := restclient.Do(rr)
	if status == 200 {
		return nil
	}
	return err
}

// Ping validates your API key. Call this to make sure your API key is correct.
// It should return nil as error if everything is OK.
func Ping() error {
	var data struct {
		Key string `json:"key"`
	}
	data.Key = Key
	return do("/users/ping", &data, nil)
}
