// Package keyring provides a repository for securely storing and retrieving
// sensitive data such as tokens and secrets using the keyring support package.
package keyring

import (
	"context"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
)

// NewRepository creates and returns a new instance of Repository.
func NewRepository() *Repository {
	return &Repository{}
}

// Repository is a struct that provides methods to interact with the keyring
// for storing and retrieving sensitive data.
type Repository struct{}

// GetToken retrieves the stored token from the keyring.
// Returns the token, a boolean indicating if it exists, and an error if any.
func (d *Repository) GetToken(_ context.Context) (string, bool, error) {
	return keyring.Get("token")
}

// SetToken stores the given token in the keyring.
// Returns an error if the operation fails.
func (d *Repository) SetToken(_ context.Context, token string) error {
	return keyring.Set("token", token)
}

// DeleteToken removes the stored token from the keyring.
// Returns an error if the operation fails.
func (d *Repository) DeleteToken(_ context.Context) error {
	return keyring.Delete("token")
}

// GetSecret retrieves the stored secret from the keyring.
// Returns the secret, a boolean indicating if it exists, and an error if any.
func (d *Repository) GetSecret(_ context.Context) (string, bool, error) {
	return keyring.Get("secret")
}

// SetSecret stores the given secret in the keyring.
// Returns an error if the operation fails.
func (d *Repository) SetSecret(_ context.Context, secret string) error {
	return keyring.Set("secret", secret)
}
