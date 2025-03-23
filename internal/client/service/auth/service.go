package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
)

func New(client pbAuth.AuthServiceClient, repo Repository) Service {
	return &service{
		client: client,
		repo:   repo,
	}
}

type service struct {
	client pbAuth.AuthServiceClient
	repo   Repository
}

func (s *service) Register(ctx context.Context, login string, password string) error {
	response, err := s.client.Register(ctx, &pbAuth.RegisterRequest{Login: login, Password: password})

	if err != nil {
		if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.AlreadyExists {
			return ErrLoginTaken
		}

		return fmt.Errorf("error registering user: %w", err)
	}

	err = s.repo.Set(ctx, response.Token)
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	return nil
}

func (s *service) Login(ctx context.Context, login string, password string) error {
	response, err := s.client.Login(ctx, &pbAuth.LoginRequest{Login: login, Password: password})

	if err != nil {
		if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.Unauthenticated {
			return ErrInvalidPair
		}

		return fmt.Errorf("error logging in: %w", err)
	}

	err = s.repo.Set(ctx, response.Token)
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	return nil
}

func (s *service) IsLoggedIn(ctx context.Context) (bool, error) {
	_, ok, err := s.repo.Get(ctx)
	if err != nil {
		return false, fmt.Errorf("error getting token: %w", err)
	}

	return ok, nil
}

func (s *service) Logout(ctx context.Context) error {
	err := s.repo.Delete(ctx)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}

	return nil
}

func (s *service) SetToken(ctx context.Context) (context.Context, error) {
	token, ok, err := s.repo.Get(ctx)
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
