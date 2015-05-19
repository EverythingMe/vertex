package web2

import (
	"fmt"
)

type Error struct {
	Message string
	Code    int
}

const (
	// The request succeeded
	Ok = 1
	// General failure

	GeneralFailure = -1

	// Input validation failed
	InvalidRequest = -14

	// The request was denied for auth reasons
	Unauthorized = -9

	// Insecure access denied
	InsecureAccessDenied = -10

	// We do not want to server this request, the client should not retry
	ResourceUnavailable = -1337

	// Please back off
	BackOff = -100

	// Some middleware took over the request, and the renderer should not render the response
	Hijacked = 0
)

// A special error that should be returned when hijacking a request, taking over response rendering from the renderer
var ErrHijacked = NewErrorCode("Request Hijacked, Do not rendere response", Hijacked)

func IsHijacked(err error) bool {
	if e, ok := err.(*Error); !ok {
		return false
	} else {
		return e.Code == Hijacked
	}
	return false

}
func NewErrorCode(e string, code int) *Error {

	return &Error{
		Message: e,
		Code:    code,
	}
}

func NewError(e string) *Error {
	return NewErrorCode(e, GeneralFailure)
}

//Format a new web error from message
func NewErrorf(format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Code:    -1,
	}
}

func (e *Error) Error() string {
	if e != nil {
		return e.Message
	}

	return ""
}
