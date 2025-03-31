// Package blob defines the interface for blob storage repositories.
// It provides abstractions for storing, retrieving, and deleting binary large objects (blobs).
package blob

import "io"

// Repository defines the interface for a blob storage repository.
type Repository interface {
	// OpenBlobWriter opens a writer for the blob identified by the given key.
	// Returns an io.WriteCloser for writing to the blob or an error if the operation fails.
	OpenBlobWriter(key string) (io.WriteCloser, error)

	// OpenBlobReader opens a reader for the blob identified by the given key.
	// Returns an io.ReadCloser for reading the blob, a boolean indicating if the blob exists,
	// and an error if the operation fails.
	OpenBlobReader(key string) (io.ReadCloser, bool, error)

	// DeleteBlob deletes the blob identified by the given key.
	// Returns an error if the operation fails.
	DeleteBlob(key string) error
}
