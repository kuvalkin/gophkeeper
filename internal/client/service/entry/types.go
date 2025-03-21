package entry

import (
	"io"
)

type Crypt interface {
	Encrypt(dst io.Writer) (io.WriteCloser, error)
	Decrypt(src io.Reader) (io.Reader, error)
}
