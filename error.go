package vertex

import (
	"fmt"
	"net/http"
)

type internalError struct {
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

// ErrorString converts an error code to a user "friendly" string
func errorString(errorCode int) string {
	switch errorCode {
	case Ok:
		return "OK"
	case Hijacked:
		return "Request Hijacked By Handler"
	case GeneralFailure:
		return "Request Failed"
	case InvalidRequest:
		return "Invalid/missing parameters for request"
	case Unauthorized:
		return "Unauthorized Request"
	case InsecureAccessDenied:
		return "Insecure Access Denied"
	case ResourceUnavailable:
		return "Resource Temporary Unavailable"
	case BackOff:
		return "Please Back-off and Retry in a While"

	}

	return fmt.Sprintf("Unknown error code: %d", errorCode)
}

// A special error that should be returned when hijacking a request, taking over response rendering from the renderer
var ErrHijacked = NewErrorCode("Request Hijacked, Do not rendere response", Hijacked)

// IsHijacked inspects an error and checks whether it represents a hijacked response
func IsHijacked(err error) bool {

	if err == ErrHijacked {
		return true
	}

	if e, ok := err.(*internalError); !ok {
		return false
	} else {
		return e.Code == Hijacked
	}

}

// convert an internal error (or any other error) to an http code
func httpCode(errorCode int) int {
	switch errorCode {

	case Ok, Hijacked:
		return http.StatusOK
	case GeneralFailure:
		return http.StatusInternalServerError
	case InvalidRequest:
		return http.StatusBadRequest
	case Unauthorized:
		return http.StatusUnauthorized
	case InsecureAccessDenied:
		return http.StatusForbidden
	case ResourceUnavailable:
		return http.StatusServiceUnavailable
	case BackOff:
		return http.StatusServiceUnavailable

	}

	return http.StatusInternalServerError
}

func NewErrorCode(e string, code int) error {

	return &internalError{
		Message: e,
		Code:    code,
	}
}

func NewError(e string) error {
	return NewErrorCode(e, GeneralFailure)
}

//Format a new web error from message
func NewErrorf(format string, args ...interface{}) error {
	return &internalError{
		Message: fmt.Sprintf(format, args...),
		Code:    -1,
	}
}

func (e *internalError) Error() string {
	if e != nil {
		return e.Message
	}

	return ""
}
