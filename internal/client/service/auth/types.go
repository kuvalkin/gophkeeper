package auth

import "context"

type Repository interface {
	GetToken(ctx context.Context) (string, bool, error)
	SetToken(ctx context.Context, token string) error
}
