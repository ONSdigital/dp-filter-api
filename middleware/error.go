package middleware

import (
	"github.com/ONSdigital/log.go/v2/log"
)

// errorResponse is the generic ONS error response body for HTTP errors
type erResponse struct {
	Errors []string `json:"errors"`
}

// er is the package's internal error struct
type er struct {
	err    error
	msg    string
	status int
	data   log.Data
}

func (e er) Error() string {
	if e.err == nil {
		return "nil"
	}
	return e.err.Error()
}

// Unwrap implements the standard library Go unwrapper interface
func (e er) Unwrap() error {
	return e.err
}

// LogData satisfies the dataLogger interface which is used to recover
// log data from an error
func (e er) LogData() map[string]interface{} {
	return e.data
}

// Code satisfies the coder interface which is used to recover a
// HTTP status code from an error
func (e er) Code() int {
	return e.status
}
