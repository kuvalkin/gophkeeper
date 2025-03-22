package auth

import (
	"context"
	"errors"
)

var ErrNoSecret = errors.New("secret not set")

type Service interface {
	Register(ctx context.Context, login string, password string) error
	SetToken(ctx context.Context) (context.Context, error)
}

type Repository interface {
	GetToken(ctx context.Context) (string, bool, error)
	SetToken(ctx context.Context, token string) error
}
