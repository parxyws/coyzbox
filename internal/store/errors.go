package store

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrUnauthorized  = errors.New("unauthorized action")
	ErrInvalidInput  = errors.New("invalid input data")
)
