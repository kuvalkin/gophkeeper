package entry

import (
	"context"
	"io"
)

type Service interface {
	Set(ctx context.Context, key string, name string, entry Entry) error
	Get(ctx context.Context, key string, entry Entry) (bool, error)
	Delete(ctx context.Context, key string) error
}

type Entry interface {
	Bytes() (io.ReadCloser, error)
	FromBytes(reader io.Reader) error
	Notes() string
	SetNotes(notes string) error
}

type Crypt interface {
	Encrypt(dst io.Writer) (io.WriteCloser, error)
	Decrypt(src io.Reader) (io.Reader, error)
}
