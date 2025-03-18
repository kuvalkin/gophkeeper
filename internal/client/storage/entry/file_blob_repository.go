package entry

import (
	"io"
)

func NewFileBlobRepository() (*FileBlobRepository, error) {

}

type FileBlobRepository struct {
}

func (f *FileBlobRepository) Writer(key string) (io.WriteCloser, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FileBlobRepository) Reader(key string) (io.ReadCloser, bool, error) {
	//TODO implement me
	panic("implement me")
}
