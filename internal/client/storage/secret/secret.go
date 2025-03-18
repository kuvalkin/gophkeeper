package secret

import (
	"context"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
)

func NewKeyringRepository() (*KeyringSecretStorage, error) {
	return &KeyringSecretStorage{}, nil
}

type KeyringSecretStorage struct {
}

func (k *KeyringSecretStorage) GetSecret(_ context.Context) (string, bool, error) {
	return keyring.Get("secret")
}

func (k *KeyringSecretStorage) SetSecret(_ context.Context, secret string) error {
	return keyring.Set("secret", secret)
}
