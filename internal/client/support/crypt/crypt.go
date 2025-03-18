package crypt

import (
	"io"
)

func NewAgeCrypter() (*AgeCrypter, error) {

}

type AgeCrypter struct {
}

func (a *AgeCrypter) Encrypt(dst io.Writer) (io.WriteCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (a *AgeCrypter) Decrypt(src io.Reader) (io.Reader, error) {
	//TODO implement me
	panic("implement me")
}
