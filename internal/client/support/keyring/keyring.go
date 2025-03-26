package keyring

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const service = "com.kuvalkin.gophkeeper"

func Set(key string, value string) error {
	err := keyring.Set(service, key, value)
	if err != nil {
		return fmt.Errorf("error setting value in keyring: %w", err)
	}

	return nil
}

func Get(key string) (string, bool, error) {
	value, err := keyring.Get(service, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", false, nil
		}

		return "", false, fmt.Errorf("error getting value from keyring: %w", err)
	}

	return value, true, nil
}

func Delete(key string) error {
	err := keyring.Delete(service, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}

		return fmt.Errorf("error deleting value from keyring: %w", err)
	}

	return nil
}
