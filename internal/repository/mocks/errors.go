package mocks

import "errors"

// Definition of errors for test mocks
var (
	ErrUserAlreadyExists = errors.New("user with this login already exists")
	ErrSecretNotFound    = errors.New("secret not found")
)
