package secret

import (
	"context"
	"errors"
	"fmt"
)

type Service interface {
	SetSecret(ctx context.Context, secret string) error
	GetSecret(ctx context.Context) (string, bool, error)
}

type Repository interface {
	GetSecret(ctx context.Context) (string, bool, error)
	SetSecret(ctx context.Context, secret string) error
}

func New(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo Repository
}

var ErrAlreadySet = errors.New("secret already set")

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

func (s *service) GetSecret(ctx context.Context) (string, bool, error) {
	secret, exists, err := s.repo.GetSecret(ctx)
	if err != nil {
		return "", false, fmt.Errorf("error getting secret: %w", err)
	}

	return secret, exists, nil
}
