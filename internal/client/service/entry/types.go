// Package entry provides the core business logic for managing encrypted entries in the application.
// It includes interfaces and implementations for creating, retrieving, and deleting entries,
// as well as encrypting and decrypting their content and metadata.
package entry

import (
	"context"
	"errors"
	"io"
)

// ErrEntryExists is returned when an entry with the same key already exists.
var ErrEntryExists = errors.New("entry already exists")

// Service defines the interface for managing entries, including creating, retrieving, and deleting them.
type Service interface {
	// SetEntry creates or updates an entry with the given key, name, notes, and content.
	// If the entry already exists, the onOverwrite callback is invoked to determine whether to overwrite it.
	SetEntry(ctx context.Context, key string, name string, notes string, content io.ReadCloser, onOverwrite func() bool) error

	// GetEntry retrieves an entry by its key. It returns the notes, content, a boolean indicating existence, and an error if any.
	GetEntry(ctx context.Context, key string) (string, io.ReadCloser, bool, error)

	// DeleteEntry removes an entry by its key. Returns an error if the deletion fails.
	DeleteEntry(ctx context.Context, key string) error
}

// Crypt defines the interface for encryption and decryption operations.
type Crypt interface {
	// Encrypt creates a writer that encrypts data written to it and writes the encrypted data to the provided destination.
	Encrypt(dst io.Writer) (io.WriteCloser, error)

	// Decrypt creates a reader that decrypts data read from the provided source.
	Decrypt(src io.Reader) (io.Reader, error)
}
