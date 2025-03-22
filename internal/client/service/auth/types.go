package auth

import (
	"context"
)

type Service interface {
	Register(ctx context.Context, login string, password string) error
	SetToken(ctx context.Context) (context.Context, error)
	Login(ctx context.Context, login string, password string) error
	IsLoggedIn(ctx context.Context) (bool, error)
	Logout(ctx context.Context) error
}

type Repository interface {
	Get(ctx context.Context) (string, bool, error)
	Set(ctx context.Context, token string) error
	Delete(ctx context.Context) error
}
