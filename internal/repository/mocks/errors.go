package mocks

import "errors"

// Определение ошибок для тестовых моков
var (
	ErrUserAlreadyExists = errors.New("пользователь с таким логином уже существует")
	ErrSecretNotFound    = errors.New("секрет не найден")
)
