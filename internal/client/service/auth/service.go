package auth

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/kuvalkin/gophkeeper/internal/client/service"
	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
)

func New(client pbAuth.AuthServiceClient, repo Repository, crypt service.Crypt) (*Service, error) {
	return &Service{
		client: client,
		repo:   repo,
		crypt:  crypt,
	}, nil
}

type Service struct {
	client pbAuth.AuthServiceClient
	repo   Repository
	crypt  service.Crypt
}

func (s *Service) Register(ctx context.Context, login string, password string) error {
	response, err := s.client.Register(ctx, &pbAuth.RegisterRequest{Login: login, Password: password})

	if err != nil {
		//todo handle errors (conflict)
		return fmt.Errorf("error registering user: %w", err)
	}

	encryptedToken, err := s.encryptToken(response.Token)
	if err != nil {
		return fmt.Errorf("error encrypting token: %w", err)
	}

	err = s.repo.SetToken(ctx, encryptedToken)
	if err != nil {
		return fmt.Errorf("error saving token: %w", err)
	}

	return nil
}

func (s *Service) encryptToken(token string) (string, error) {
	var buf bytes.Buffer
	encryptWriter, err := s.crypt.Encrypt(&buf)
	if err != nil {
		return "", fmt.Errorf("could not create encrypt writer: %w", err)
	}

	_, err = io.WriteString(encryptWriter, token)
	if err != nil {
		_ = encryptWriter.Close()

		return "", fmt.Errorf("could not write token to encrypt writer: %w", err)
	}

	err = encryptWriter.Close()
	if err != nil {
		return "", fmt.Errorf("could not close encrypt writer: %w", err)
	}

	return buf.String(), nil
}
