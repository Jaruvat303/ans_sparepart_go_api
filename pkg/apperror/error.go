package apperror

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrUserForbidden     = errors.New("forbidden")
	ErrConflict          = errors.New("resource already exists")
	ErrInternalServer    = errors.New("internal server error")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidSKU        = errors.New("sku is invalid or contains restricted characters")
)
