package secret

import (
	"context"
	"errors"
	"fmt"
)

type Service interface {
	Set(ctx context.Context, secret string) error
	Get(ctx context.Context) (string, bool, error)
}

type Repository interface {
	Get(ctx context.Context) (string, bool, error)
	Set(ctx context.Context, secret string) error
}

func New(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo Repository
}

var ErrInternal = errors.New("internal error")
var ErrAlreadySet = errors.New("secret already set")

func (s *service) Set(ctx context.Context, secret string) error {
	_, exists, err := s.repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("error getting secret: %w", err)
	}

	if exists {
		return ErrAlreadySet
	}

	err = s.repo.Set(ctx, secret)
	if err != nil {
		return fmt.Errorf("error setting secret: %w", err)
	}

	return nil
}

func (s *service) Get(ctx context.Context) (string, bool, error) {
	secret, exists, err := s.repo.Get(ctx)
	if err != nil {
		return "", false, fmt.Errorf("error getting secret: %w", err)
	}

	return secret, exists, nil
}
