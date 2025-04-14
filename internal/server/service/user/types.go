// Package user provides core application logic for handling user-related operations,
// including user registration, authentication, and token management.
package user

import (
	"context"
	"errors"
	"time"
)

// ErrInvalidLogin is returned when the provided login is invalid.
var ErrInvalidLogin = errors.New("invalid login")

// ErrLoginTaken is returned when a user with the given login already exists.
var ErrLoginTaken = errors.New("user with this login already exists")

// ErrInvalidPair is returned when the login/password pair is invalid.
var ErrInvalidPair = errors.New("login/password pair is invalid")

// ErrInvalidToken is returned when the provided token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrInternal is returned when an internal error occurs.
var ErrInternal = errors.New("internal error")

// TokenInfo contains information extracted from an authentication token.
type TokenInfo struct {
	// UserID is the unique identifier of the user associated with the token.
	UserID string
}

// Service defines the interface for user-related operations.
type Service interface {
	// RegisterUser registers a new user with the given login and password.
	RegisterUser(ctx context.Context, login string, password string) error
	// LoginUser authenticates a user and returns an authentication token on success.
	LoginUser(ctx context.Context, login string, password string) (string, error)
	// ParseAuthToken parses and validates an authentication token, returning the associated TokenInfo.
	ParseAuthToken(ctx context.Context, token string) (*TokenInfo, error)
}

// Options contains configuration options for the user service.
type Options struct {
	// TokenSecret is the secret key used for signing authentication tokens.
	TokenSecret []byte
	// PasswordSalt is the salt used for hashing passwords.
	PasswordSalt string
	// TokenExpirationPeriod defines the duration for which an authentication token is valid.
	TokenExpirationPeriod time.Duration
}

// ErrLoginNotUnique is returned when a user with the given login already exists.
var ErrLoginNotUnique = errors.New("user with this login already exists")

// UserInfo represents information about a user stored in the repository.
type UserInfo struct {
	// ID is the unique identifier of the user.
	ID string
	// PasswordHash is the hashed password of the user.
	PasswordHash string
}

// Repository defines the interface for user data storage operations.
type Repository interface {
	// AddUser adds a new user with the given login and password hash to the repository.
	AddUser(ctx context.Context, login string, passwordHash string) error
	// FindUser retrieves a user by login. Returns the user info, a boolean indicating if the user was found, and an error if any.
	FindUser(ctx context.Context, login string) (UserInfo, bool, error)
}
