package web2

import (
	"fmt"
)

type Error struct {
	Message string
	Code    int
}

const (
	ErrOkay                = 1
	ErrFail                = -1
	ErrInvalidInput        = -14
	ErrUnauthorized        = -9
	ErrResourceUnavailable = -1337
)

func NewErrorCode(e string, code int) *Error {

	return &Error{
		Message: e,
		Code:    code,
	}
}

func NewError(e string) *Error {
	return NewErrorCode(e, ErrFail)
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
