// Package httpx contains HTTP response helpers used across the API layer. All
// error responses share a single JSON shape so clients can rely on one format.
package httpx

import (
	"errors"
	"net/http"
)

// Error is a structured API error carrying an HTTP status and a machine
// readable code.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

// NewError builds an Error with the given status, code and message.
func NewError(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// Constructors for the common error categories used by handlers.
func BadRequest(code, msg string) *Error {
	return NewError(http.StatusBadRequest, code, msg)
}

func Unauthorized(code, msg string) *Error {
	return NewError(http.StatusUnauthorized, code, msg)
}

func NotFound(code, msg string) *Error {
	return NewError(http.StatusNotFound, code, msg)
}

func Gone(code, msg string) *Error {
	return NewError(http.StatusGone, code, msg)
}

func UnsupportedMediaType(code, msg string) *Error {
	return NewError(http.StatusUnsupportedMediaType, code, msg)
}

func RequestEntityTooLarge(code, msg string) *Error {
	return NewError(http.StatusRequestEntityTooLarge, code, msg)
}

func Internal(msg string) *Error {
	return NewError(http.StatusInternalServerError, "internal", msg)
}

// AsError unwraps err into *Error, returning nil when it is not one.
func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return nil
}
