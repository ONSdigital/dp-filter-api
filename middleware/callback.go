package middleware

import (
	"errors"
)

// statusCode attempts to extract a status code from an error,
// or returns 0 if not found
func statusCode(err error) int {
	var cerr coder

	if errors.As(err, &cerr) {
		return cerr.Code()
	}

	return 0
}
