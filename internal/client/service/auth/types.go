package auth

import "context"

type Repository interface {
	GetToken(ctx context.Context) ([]byte, bool, error)
	SetToken(ctx context.Context, token []byte) error
}
