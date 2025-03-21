package auth

import (
	"context"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
)

func NewKeyringRepository() *KeyringRepository {
	return &KeyringRepository{}
}

type KeyringRepository struct {
}

func (d *KeyringRepository) GetToken(_ context.Context) (string, bool, error) {
	return keyring.Get("token")
}

func (d *KeyringRepository) SetToken(_ context.Context, token string) error {
	return keyring.Set("token", token)
}
