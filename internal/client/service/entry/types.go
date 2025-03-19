package entry

import (
	"context"
	"io"

	"github.com/kuvalkin/gophkeeper/internal/support/tx"
)

type Entry interface {
	Bytes() (io.ReadCloser, error)
	FromBytes(reader io.Reader) error
	Notes() string
	SetNotes(notes string) error
}

type MetadataRepository interface {
	Set(ctx context.Context, tx tx.Tx, key string, name string, version int64) error
	GetVersion(ctx context.Context, tx tx.Tx, key string) (int64, bool, error)
	MarkAsDownloaded(ctx context.Context, tx tx.Tx, key string) error
	MarkAsDeleted(ctx context.Context, tx tx.Tx, key string) error
}

type BlobRepository interface {
	Writer(key string) (io.WriteCloser, error)
	Reader(key string) (io.ReadCloser, bool, error)
}
