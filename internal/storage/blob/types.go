package blob

import "io"

type Repository interface {
	OpenBlobWriter(key string) (io.WriteCloser, error)
	OpenBlobReader(key string) (io.ReadCloser, bool, error)
	DeleteBlob(key string) error
}
