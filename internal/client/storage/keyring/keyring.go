package keyring

import (
	"context"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
)

func NewRepository(key string) *Repository {
	return &Repository{
		key: key,
	}
}

type Repository struct {
	key string
}

func (d *Repository) Get(_ context.Context) (string, bool, error) {
	return keyring.Get(d.key)
}

func (d *Repository) Set(_ context.Context, token string) error {
	return keyring.Set(d.key, token)
}

func (d *Repository) Delete(_ context.Context) error {
	return keyring.Delete(d.key)
}
