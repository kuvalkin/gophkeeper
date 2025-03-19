package entry

import (
	"fmt"
	"io"
	"os"
	"path"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func NewFileBlobRepository(path string) (*FileBlobRepository, error) {
	return &FileBlobRepository{
		path: path,
		log:  log.Logger().Named("blobs"),
	}, nil
}

type FileBlobRepository struct {
	path string
	log  *zap.SugaredLogger
}

func (f *FileBlobRepository) Writer(key string) (io.WriteCloser, error) {
	fullPath := path.Join(f.path, key)
	f.log.Debugw("opening for write", "path", fullPath)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("cant open file: %w", err)
	}

	return file, nil
}

func (f *FileBlobRepository) Reader(key string) (io.ReadCloser, bool, error) {
	fullPath := path.Join(f.path, key)
	f.log.Debugw("opening for read", "path", fullPath)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("cant open file: %w", err)
	}

	return file, true, nil
}
