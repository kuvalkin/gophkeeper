// Package keyring provides a wrapper around the go-keyring library to manage
// secure storage of key-value pairs in the system's keyring.
package keyring

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const service = "com.kuvalkin.gophkeeper"

// Set stores a key-value pair in the system's keyring.
// If the key already exists, its value will be updated.
func Set(key string, value string) error {
	err := keyring.Set(service, key, value)
	if err != nil {
		return fmt.Errorf("error setting value in keyring: %w", err)
	}

	return nil
}

// Get retrieves the value associated with a key from the system's keyring.
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

// Delete removes a key-value pair from the system's keyring. Returns error if the operation fails, or nil if successful.
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
