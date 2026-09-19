package services

import "errors"

// Common domain errors used across services
var (
	ErrDBNotConnected      = errors.New("database connection is not initialized")
	ErrValidation          = errors.New("validation failed")
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenUsed           = errors.New("token has already been used")
)
