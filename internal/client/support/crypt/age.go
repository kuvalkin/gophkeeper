// Package crypt provides encryption and decryption utilities using the age encryption tool.
package crypt

import (
	"fmt"
	"io"

	"filippo.io/age"
)

// NewAgeCrypter creates a new AgeCrypter instance using the provided secret.
// It initializes both the recipient and identity for encryption and decryption.
// Returns an AgeCrypter or an error if initialization fails.
func NewAgeCrypter(secret string) (*AgeCrypter, error) {
	recipient, err := age.NewScryptRecipient(secret)
	if err != nil {
		return nil, fmt.Errorf("cant create age recipient: %w", err)
	}

	identity, err := age.NewScryptIdentity(secret)
	if err != nil {
		return nil, fmt.Errorf("cant create age identity: %w", err)
	}

	return &AgeCrypter{
		recipient: recipient,
		identity:  identity,
	}, nil
}

// AgeCrypter is a wrapper around the age encryption library that provides
// methods for encrypting and decrypting data streams.
type AgeCrypter struct {
	recipient age.Recipient
	identity  age.Identity
}

// Encrypt creates an io.WriteCloser that encrypts data written to it
// using the recipient initialized in the AgeCrypter.
// Returns an io.WriteCloser or an error if encryption setup fails.
func (a *AgeCrypter) Encrypt(dst io.Writer) (io.WriteCloser, error) {
	return age.Encrypt(dst, a.recipient)
}

// Decrypt creates an io.Reader that decrypts data read from it
// using the identity initialized in the AgeCrypter.
// Returns an io.Reader or an error if decryption setup fails.
func (a *AgeCrypter) Decrypt(src io.Reader) (io.Reader, error) {
	return age.Decrypt(src, a.identity)
}
