package blob

import "io"

type Repository interface {
	Writer(key string) (io.WriteCloser, error)
	Reader(key string) (io.ReadCloser, bool, error)
	Delete(key string) error
}
