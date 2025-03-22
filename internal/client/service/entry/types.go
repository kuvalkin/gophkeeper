package entry

import (
	"context"
	"io"
)

type Service interface {
	Set(ctx context.Context, key string, name string, notes string, content io.ReadCloser) error
	Get(ctx context.Context, key string) (string, io.ReadCloser, bool, error)
	Delete(ctx context.Context, key string) error
}

type Crypt interface {
	Encrypt(dst io.Writer) (io.WriteCloser, error)
	Decrypt(src io.Reader) (io.Reader, error)
}
