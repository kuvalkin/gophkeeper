package keyring

import (
	"context"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
)

func NewRepository() *Repository {
	return &Repository{}
}

type Repository struct{}

func (d *Repository) GetToken(_ context.Context) (string, bool, error) {
	return keyring.Get("token")
}

func (d *Repository) SetToken(_ context.Context, token string) error {
	return keyring.Set("token", token)
}

func (d *Repository) DeleteToken(_ context.Context) error {
	return keyring.Delete("token")
}

func (d *Repository) GetSecret(_ context.Context) (string, bool, error) {
	return keyring.Get("secret")
}

func (d *Repository) SetSecret(_ context.Context, secret string) error {
	return keyring.Set("secret", secret)
}
