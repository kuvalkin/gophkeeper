package auth

import (
	"context"
	"fmt"

	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
)

func New(client pbAuth.AuthServiceClient, repo Repository) (*Service, error) {
	return &Service{
		client: client,
		repo:   repo,
	}, nil
}

type Service struct {
	client pbAuth.AuthServiceClient
	repo   Repository
}

func (s *Service) Register(ctx context.Context, login string, password string) error {
	response, err := s.client.Register(ctx, &pbAuth.RegisterRequest{Login: login, Password: password})

	if err != nil {
		//todo handle errors (conflict)
		return fmt.Errorf("error registering user: %w", err)
	}

	err = s.repo.SetToken(ctx, response.Token)
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	return nil
}
