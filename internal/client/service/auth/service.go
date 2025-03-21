package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
)

func New(client pbAuth.AuthServiceClient, repo Repository) *Service {
	return &Service{
		client: client,
		repo:   repo,
	}
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

func (s *Service) SetToken(ctx context.Context) (context.Context, error) {
	token, ok, err := s.repo.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("token not found")
	}

	md := metadata.Pairs(
		"authorization",
		"bearer "+token,
	)

	return metadata.NewOutgoingContext(ctx, md), nil
}
