// Package auth provides the core business logic for authentication in the client application.
// It handles user registration, login, logout, and token-based authorization.
package auth

import (
	"context"
	"errors"
)

// ErrLoginTaken is returned when a user attempts to register with a login that already exists.
var ErrLoginTaken = errors.New("user with this login already exists")

// ErrInvalidPair is returned when the provided login/password pair is invalid.
var ErrInvalidPair = errors.New("login/password pair is invalid")

// Service defines the interface for authentication-related operations.
type Service interface {
	// Register registers a new user with the given login and password.
	Register(ctx context.Context, login string, password string) error

	// AddAuthorizationHeader adds an authorization header to the context using the stored token.
	AddAuthorizationHeader(ctx context.Context) (context.Context, error)

	// Login authenticates a user with the given login and password.
	Login(ctx context.Context, login string, password string) error

	// IsLoggedIn checks if the user is currently logged in.
	IsLoggedIn(ctx context.Context) (bool, error)

	// Logout logs the user out by deleting the stored token.
	Logout(ctx context.Context) error
}

// Repository defines the interface for token storage operations.
type Repository interface {
	// GetToken retrieves the stored token from the repository.
	// Returns the token, a boolean indicating if the token exists, and an error if any.
	GetToken(ctx context.Context) (string, bool, error)

	// SetToken stores the given token in the repository.
	SetToken(ctx context.Context, token string) error

	// DeleteToken removes the stored token from the repository.
	DeleteToken(ctx context.Context) error
}
