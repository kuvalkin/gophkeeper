package blob

import (
	"fmt"
	"io"
	"os"
	"path"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func NewFileBlobRepository(path string) *FileBlobRepository {
	return &FileBlobRepository{
		path: path,
		log:  log.Logger().Named("blobs"),
	}
}

type FileBlobRepository struct {
	path string
	log  *zap.SugaredLogger
}

const dirPerms = os.FileMode(0700)
const filePerms = os.FileMode(0600)

func (f *FileBlobRepository) Writer(key string) (io.WriteCloser, error) {
	fullPath := path.Join(f.path, key)

	err := os.MkdirAll(path.Dir(fullPath), dirPerms)
	if err != nil {
		return nil, fmt.Errorf("cant create directory: %w", err)
	}

	f.log.Debugw("opening for write", "path", fullPath)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerms)
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

func (f *FileBlobRepository) Delete(key string) error {
	fullPath := path.Join(f.path, key)
	f.log.Debugw("deleting", "path", fullPath)

	return os.Remove(fullPath)
}
