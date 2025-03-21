package sync

import (
	"context"
	"errors"
	"io"

	"github.com/kuvalkin/gophkeeper/internal/support/transaction"
)

type Metadata struct {
	Key   string
	Name  string
	Notes []byte
}

type UpdateEntryChunk struct {
	Content []byte
	Err     error
}

type UpdateEntryResult struct {
	Err        error
	NewVersion int64
}

var ErrInternal = errors.New("internal error")
var ErrVersionMismatch = errors.New("newer version exists")

type Service interface {
	UpdateEntry(ctx context.Context, userID string, md Metadata, lastKnownVersion int64, force bool) (chan<- UpdateEntryChunk, <-chan UpdateEntryResult, error)
}

type MetadataRepository interface {
	GetVersion(ctx context.Context, tx transaction.Tx, userID string, key string) (version int64, exists bool, err error)
	Set(ctx context.Context, tx transaction.Tx, userID string, key string, name string, notes []byte) (version int64, err error)
}

type BlobRepository interface {
	CopyFrom(ctx context.Context, key string, r io.Reader) error
}
