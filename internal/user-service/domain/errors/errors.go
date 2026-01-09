package errors

import "errors"

var (
	ErrLoginRequired        = errors.New("login and password are required")
	ErrLoginAlreadyExists   = errors.New("login already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid login or password")
	ErrTokenRequired        = errors.New("token is required")
	ErrInvalidToken         = errors.New("invalid token")
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}


