package vertex

import (
	"fmt"
	"net/http"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/dvirsky/go-pylog/logging"
)

type internalError struct {
	Message string
	Code    int
}

const (
	// The request succeeded
	Ok = iota

	// General failure
	ErrGeneralFailure

	// Input validation failed
	ErrInvalidRequest

	// Missing parameter
	ErrMissingParam

	// Invalid parameter value
	ErrInvalidParam

	// The request was denied for auth reasons
	ErrUnauthorized

	// Insecure access denied
	ErrInsecureAccessDenied

	// We do not want to server this request, the client should not retry
	ErrResourceUnavailable

	// Please back off
	ErrBackOff

	// Some middleware took over the request, and the renderer should not render the response
	ErrHijacked
)

// ErrorString converts an error code to a user "friendly" string
func httpError(err error) (re int, rm string) {

	if err == nil {
		return http.StatusOK, http.StatusText(http.StatusOK)
	}

	incidentId := uuid.New()
	if err != Hijacked {
		logging.Error("[%s] Error processing request: %s", incidentId, err)
	}

	statusFunc := func(i int) (int, string) {
		return i, fmt.Sprintf("[%s] %s", incidentId, http.StatusText(i))
	}

	if e, ok := err.(*internalError); !ok {
		return statusFunc(http.StatusInternalServerError)
	} else {

		switch e.Code {
		case Ok:
			return http.StatusOK, "OK"
		case ErrHijacked:
			return http.StatusOK, "Request Hijacked By Handler"
		case ErrInvalidRequest:
			return statusFunc(http.StatusBadRequest)
		case ErrInvalidParam, ErrMissingParam:
			return http.StatusBadRequest, e.Message
		case ErrUnauthorized:
			return statusFunc(http.StatusUnauthorized)
		case ErrInsecureAccessDenied:
			return statusFunc(http.StatusForbidden)
		case ErrResourceUnavailable:
			return statusFunc(http.StatusServiceUnavailable)
		case ErrBackOff:
			return statusFunc(http.StatusServiceUnavailable)
		case ErrGeneralFailure:
			fallthrough
		default:
			return statusFunc(http.StatusInternalServerError)

		}
	}

}

// A special error that should be returned when hijacking a request, taking over response rendering from the renderer
var Hijacked = newErrorCode(ErrHijacked, "Request Hijacked, Do not rendere response")

// IsHijacked inspects an error and checks whether it represents a hijacked response
func IsHijacked(err error) bool {

	if err == Hijacked {
		return true
	}

	if e, ok := err.(*internalError); !ok {
		return false
	} else {
		return e.Code == ErrHijacked
	}

}

func newErrorCode(code int, msg string) error {

	return &internalError{
		Message: msg,
		Code:    code,
	}
}

func newErrorfCode(code int, format string, args ...interface{}) error {

	return &internalError{
		Message: fmt.Sprintf(format, args...),
		Code:    code,
	}
}

// Wrap a normal error object with an internal object
func NewError(err error) error {

	if _, ok := err.(*internalError); ok {
		fmt.Println("NOT WRAPPING ERROR %s", err)
		return err
	} else {
		return newErrorCode(ErrGeneralFailure, err.Error())
	}
}

//Format a new web error from message
func NewErrorf(format string, args ...interface{}) error {
	return &internalError{
		Message: fmt.Sprintf(format, args...),
		Code:    ErrGeneralFailure,
	}
}

// Error returns the error message of the underlying error object
func (e *internalError) Error() string {
	if e != nil {
		return fmt.Sprintf("%s", e.Message)
	}

	return ""
}

// MissingParamError Returns a formatted error stating that a parameter was missing.
//
// NOTE: The message will be returned to the client directly
func MissingParamError(msg string, args ...interface{}) error {
	return newErrorfCode(ErrMissingParam, msg, args...)
}

// InvalidRequest returns an error signifying something went bad reading the request data (not the validation process).
// This in general should not be used by APIs
func InvalidRequestError(msg string, args ...interface{}) error {
	return newErrorfCode(ErrInvalidRequest, msg, args...)
}

// InvalidParam returns an error signifying an invalid parameter value.
//
// NOTE: The error string will be returned directly to the client
func InvalidParamError(msg string, args ...interface{}) error {
	return newErrorfCode(ErrInvalidParam, msg, args...)
}

// Unauthorized returns an error signifying the request was not authorized, but the client may log-in and retry
func UnauthorizedError(msg string, args ...interface{}) error {
	return newErrorfCode(ErrUnauthorized, msg, args...)
}

// InsecureAccessDenied returns an error signifying the client has no access to the requested resource
func InsecureAccessDenied(msg string, args ...interface{}) error {
	return newErrorfCode(ErrInsecureAccessDenied, msg, args...)
}

// ResourceUnavailable returns an error meaning we do not want to server this request, the client should not retry
func ResourceUnavailableError(msg string, args ...interface{}) error {
	return newErrorfCode(ErrResourceUnavailable, msg, args...)
}

// BackOff returns a back-off error with a message formatted for the given amount of backoff time
func BackOffError(duration time.Duration) error {

	return newErrorfCode(ErrBackOff, fmt.Sprintf("Retry-Seconds: %.02f", duration.Seconds()))

}
