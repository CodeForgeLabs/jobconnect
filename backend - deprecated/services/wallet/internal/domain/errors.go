package domain

import "errors"

var (
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrConflict          = errors.New("conflict")
)
