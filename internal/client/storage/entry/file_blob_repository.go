package entry

import (
	"fmt"
	"io"
	"os"
	"path"
)

func NewFileBlobRepository(path string) (*FileBlobRepository, error) {
	return &FileBlobRepository{
		path: path,
	}, nil
}

type FileBlobRepository struct {
	path string
}

func (f *FileBlobRepository) Writer(key string) (io.WriteCloser, error) {
	file, err := os.OpenFile(path.Join(f.path, key), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("cant open file: %w", err)
	}

	return file, nil
}

func (f *FileBlobRepository) Reader(key string) (io.ReadCloser, bool, error) {
	file, err := os.Open(path.Join(f.path, key))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("cant open file: %w", err)
	}

	return file, true, nil
}
