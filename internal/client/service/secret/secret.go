// Package secret provides core application business logic for managing secrets.
// It includes functionality for storing and retrieving secrets, and defines interfaces
// for the service and repository layers to enable flexible and testable implementations.
package secret

import (
	"context"
	"errors"
	"fmt"
)

// Service defines the interface for managing secrets.
type Service interface {
	// SetSecret stores a new secret. Returns an error if the secret is already set.
	SetSecret(ctx context.Context, secret string) error
	// GetSecret retrieves the stored secret. Returns the secret, a boolean indicating if it exists, and an error if any.
	GetSecret(ctx context.Context) (string, bool, error)
}

// Repository defines the interface for the underlying storage of secrets.
type Repository interface {
	// GetSecret retrieves the stored secret from the repository.
	// Returns the secret, a boolean indicating if it exists, and an error if any.
	GetSecret(ctx context.Context) (string, bool, error)
	// SetSecret stores a new secret in the repository.
	// Returns an error if the operation fails.
	SetSecret(ctx context.Context, secret string) error
}

// New creates a new Service instance with the provided Repository.
func New(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo Repository
}

// ErrAlreadySet is returned when attempting to set a secret that already exists.
var ErrAlreadySet = errors.New("secret already set")

// SetSecret stores a new secret. Returns ErrAlreadySet if the secret already exists.
func (s *service) SetSecret(ctx context.Context, secret string) error {
	_, exists, err := s.repo.GetSecret(ctx)
	if err != nil {
		return fmt.Errorf("error getting secret: %w", err)
	}

	if exists {
		return ErrAlreadySet
	}

	err = s.repo.SetSecret(ctx, secret)
	if err != nil {
		return fmt.Errorf("error setting secret: %w", err)
	}

	return nil
}

// GetSecret retrieves the stored secret. Returns the secret, a boolean indicating if it exists, and an error if any.
func (s *service) GetSecret(ctx context.Context) (string, bool, error) {
	secret, exists, err := s.repo.GetSecret(ctx)
	if err != nil {
		return "", false, fmt.Errorf("error getting secret: %w", err)
	}

	return secret, exists, nil
}
