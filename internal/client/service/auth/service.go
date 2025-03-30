package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pbAuth "github.com/kuvalkin/gophkeeper/pkg/proto/auth/v1"
)

// New creates a new instance of the authentication service.
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

// Register registers a new user with the given login and password.
// It saves the authentication token in the repository upon successful registration.
func (s *service) Register(ctx context.Context, login string, password string) error {
	response, err := s.client.Register(ctx, &pbAuth.RegisterRequest{Login: login, Password: password})

	if err != nil {
		if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.AlreadyExists {
			return ErrLoginTaken
		}

		return fmt.Errorf("error registering user: %w", err)
	}

	err = s.repo.SetToken(ctx, response.Token)
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	return nil
}

// Login authenticates a user with the given login and password.
// It saves the authentication token in the repository upon successful login.
func (s *service) Login(ctx context.Context, login string, password string) error {
	response, err := s.client.Login(ctx, &pbAuth.LoginRequest{Login: login, Password: password})

	if err != nil {
		if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.Unauthenticated {
			return ErrInvalidPair
		}

		return fmt.Errorf("error logging in: %w", err)
	}

	err = s.repo.SetToken(ctx, response.Token)
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	return nil
}

// IsLoggedIn checks if the user is currently logged in by verifying the existence of a stored token.
func (s *service) IsLoggedIn(ctx context.Context) (bool, error) {
	_, ok, err := s.repo.GetToken(ctx)
	if err != nil {
		return false, fmt.Errorf("error getting token: %w", err)
	}

	return ok, nil
}

// Logout logs the user out by deleting the stored token from the repository.
func (s *service) Logout(ctx context.Context) error {
	err := s.repo.DeleteToken(ctx)
	if err != nil {
		return fmt.Errorf("error deleting token: %w", err)
	}

	return nil
}

// AddAuthorizationHeader adds an authorization header to the context using the stored token.
// Returns an updated context with the authorization header or an error if the token is not found.
func (s *service) AddAuthorizationHeader(ctx context.Context) (context.Context, error) {
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
