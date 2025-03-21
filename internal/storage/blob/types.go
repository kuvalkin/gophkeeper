package blob

import "io"

type Repository interface {
	Writer(key string) (ErrWriteCloser, error)
	Reader(key string) (io.ReadCloser, bool, error)
}

type ErrWriteCloser interface {
	io.WriteCloser
	CloseWithError(err error) error
}
