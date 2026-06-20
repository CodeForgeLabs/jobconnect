package domain

import "errors"

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrSessionExpired     = errors.New("session expired")
	ErrKYCNotVerified     = errors.New("user KYC not verified")
	ErrGatewayUnavailable = errors.New("payment gateway unavailable")
)
