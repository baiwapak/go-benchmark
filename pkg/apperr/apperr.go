package apperr

import "net/http"

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
	Field      string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(httpStatus, code int, message string) *AppError {
	return &AppError{HTTPStatus: httpStatus, Code: code, Message: message}
}

func Validation(field, message string) *AppError {
	return &AppError{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       422,
		Message:    message,
		Field:      field,
	}
}

var (
	ErrNotFound   = New(http.StatusNotFound, 404, "not found")
	ErrForbidden  = New(http.StatusForbidden, 403, "forbidden")
	ErrBadRequest = New(http.StatusBadRequest, 400, "bad request")
	ErrConflict   = New(http.StatusConflict, 409, "conflict")
	ErrTooMany    = New(http.StatusTooManyRequests, 429, "too many requests")
)
