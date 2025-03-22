package entry

import (
	"context"
	"errors"
	"io"
)

type Metadata struct {
	Key   string
	Name  string
	Notes []byte
}

type UploadChunk struct {
	Content []byte
	Err     error
}

type UpdateEntryResult struct {
	Err error
}

var ErrInternal = errors.New("internal error")

type Service interface {
	Set(ctx context.Context, userID string, md Metadata) (chan<- UploadChunk, <-chan UpdateEntryResult, error)
	Get(ctx context.Context, userID string, key string) (Metadata, io.ReadCloser, bool, error)
	Delete(ctx context.Context, userID string, key string) error
}

type MetadataRepository interface {
	Set(ctx context.Context, userID string, md Metadata) error
	Get(ctx context.Context, userID string, key string) (Metadata, bool, error)
	Delete(ctx context.Context, userID string, key string) error
}
