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

type UpdateEntryResult struct {
	Err        error
	NewVersion int64
}

var ErrInternal = errors.New("internal error")
var ErrVersionMismatch = errors.New("newer version exists")

// todo move
type Service interface {
	// put bytes to writer and close it when you done, wait for result in chan
	UpdateEntry(ctx context.Context, userID string, md Metadata, lastKnownVersion int64, force bool) (*io.PipeWriter, error, <-chan UpdateEntryResult)
}

type MetadataRepository interface {
	GetAndLock(ctx context.Context, tx transaction.Tx, userID string, key string) (version int64, err error)
	SetAndUnlock(ctx context.Context, tx transaction.Tx, userID string, key string, name string, notes []byte) (version int64, err error)
}

type BlobRepository interface {
	CopyFrom(ctx context.Context, key string, r io.Reader) error
}
