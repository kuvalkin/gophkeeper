// Package entry provides the core business logic for managing entries on the server.
// It handles operations such as creating, retrieving, and deleting entries,
// while coordinating between metadata storage and blob storage systems.
package entry

import (
	"context"
	"errors"
	"io"
)

// Metadata represents the metadata associated with an entry.
type Metadata struct {
	Key   string // Key is the unique identifier for the entry.
	Name  string // Name is the human-readable name of the entry.
	Notes []byte // Notes contains additional information about the entry.
}

// UploadChunk represents a chunk of data being uploaded.
type UploadChunk struct {
	Content []byte // Content is the data of the chunk.
	Err     error  // Err indicates any error encountered during the upload.
}

// SetEntryResult represents the result of a SetEntry operation.
type SetEntryResult struct {
	Err error // Err indicates any error encountered during the operation.
}

// ErrInternal is returned when an internal error occurs.
var ErrInternal = errors.New("internal error")

// ErrUploadChunk is returned when an error is received from the upload chunk channel.
var ErrUploadChunk = errors.New("received error from upload chunk chan")

// ErrNoUpload is returned when an upload is closed without any data.
var ErrNoUpload = errors.New("upload closed with no data")

// ErrEntryExists is returned when an entry with the same key already exists.
var ErrEntryExists = errors.New("entry already exists")

// Service defines the interface for managing entries.
type Service interface {
	// SetEntry starts the process of uploading an entry.
	// It returns channels for uploading chunks and receiving the result.
	SetEntry(ctx context.Context, userID string, md Metadata, overwrite bool) (chan<- UploadChunk, <-chan SetEntryResult, error)

	// GetEntry retrieves an entry's metadata and content.
	// It returns the metadata, a reader for the content, a boolean indicating existence, and an error if any.
	GetEntry(ctx context.Context, userID string, key string) (Metadata, io.ReadCloser, bool, error)

	// DeleteEntry deletes an entry by its key.
	DeleteEntry(ctx context.Context, userID string, key string) error
}

// MetadataRepository defines the interface for managing metadata storage.
type MetadataRepository interface {
	// SetMetadata stores metadata for an entry.
	SetMetadata(ctx context.Context, userID string, md Metadata) error

	// GetMetadata retrieves metadata for an entry by its key.
	// It returns the metadata, a boolean indicating existence, and an error if any.
	GetMetadata(ctx context.Context, userID string, key string) (Metadata, bool, error)

	// DeleteMetadata deletes metadata for an entry by its key.
	DeleteMetadata(ctx context.Context, userID string, key string) error
}
