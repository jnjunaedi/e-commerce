package apperror

import "errors"

var (
	ErrBadRequest         = errors.New("bad request data")
	ErrNotFound           = errors.New("data not found")
	ErrInternalServer     = errors.New("internal server error")
	ErrConflict           = errors.New("data conflict")
	ErrInvalidCredentials = errors.New("invalid login credentials")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrForbidden          = errors.New("access forbidden")
)

type AppError struct {
	Category error
	Message  string
}

func (a *AppError) Error() string {
	return a.Message
}

func (a *AppError) Unwrap() error {
	return a.Category
}

func New(category error, message string) error {
	return &AppError{Category: category, Message: message}
}
