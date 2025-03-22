package crypt

import (
	"fmt"
	"io"

	"filippo.io/age"
)

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

type AgeCrypter struct {
	recipient age.Recipient
	identity  age.Identity
}

func (a *AgeCrypter) Encrypt(dst io.Writer) (io.WriteCloser, error) {
	return age.Encrypt(dst, a.recipient)
}

func (a *AgeCrypter) Decrypt(src io.Reader) (io.Reader, error) {
	return age.Decrypt(src, a.identity)
}
