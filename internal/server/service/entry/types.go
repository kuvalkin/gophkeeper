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

type SetEntryResult struct {
	Err error
}

var ErrInternal = errors.New("internal error")
var ErrUploadChunk = errors.New("received error from upload chunk chan")
var ErrNoUpload = errors.New("upload closed with no data")
var ErrEntryExists = errors.New("entry already exists")

type Service interface {
	Set(ctx context.Context, userID string, md Metadata, overwrite bool) (chan<- UploadChunk, <-chan SetEntryResult, error)
	Get(ctx context.Context, userID string, key string) (Metadata, io.ReadCloser, bool, error)
	Delete(ctx context.Context, userID string, key string) error
}

type MetadataRepository interface {
	Set(ctx context.Context, userID string, md Metadata) error
	Get(ctx context.Context, userID string, key string) (Metadata, bool, error)
	Delete(ctx context.Context, userID string, key string) error
}
