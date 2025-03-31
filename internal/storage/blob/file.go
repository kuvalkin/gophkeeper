package blob

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

// NewFileBlobRepository creates a new instance of FileBlobRepository.
// The `path` parameter specifies the root directory where blobs will be stored.
func NewFileBlobRepository(path string) (*FileBlobRepository, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("path must be absolute")
	}

	return &FileBlobRepository{
		path: path,
		log:  log.Logger().Named("blobs"),
	}, nil
}

// FileBlobRepository is a file-based implementation of the blob.Repository interface.
// It manages blob storage operations such as writing, reading, and deleting files.
type FileBlobRepository struct {
	path string
	log  *zap.SugaredLogger
}

const dirPerms = os.FileMode(0700)
const filePerms = os.FileMode(0600)

// OpenBlobWriter opens a writer for the blob identified by the given key.
// If the blob does not exist, it will be created. If it exists, it will be truncated.
// Returns an io.WriteCloser for writing to the blob or an error if the operation fails.
func (f *FileBlobRepository) OpenBlobWriter(key string) (io.WriteCloser, error) {
	fullPath, err := f.getFullPath(key)
	if err != nil {
		return nil, fmt.Errorf("cant get full path: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(fullPath), dirPerms)
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

// OpenBlobReader opens a reader for the blob identified by the given key.
// Returns an io.ReadCloser for reading the blob, a boolean indicating if the blob exists,
// and an error if the operation fails.
func (f *FileBlobRepository) OpenBlobReader(key string) (io.ReadCloser, bool, error) {
	fullPath, err := f.getFullPath(key)
	if err != nil {
		return nil, false, fmt.Errorf("cant get full path: %w", err)
	}

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

// DeleteBlob deletes the blob identified by the given key.
// Returns an error if the operation fails.
func (f *FileBlobRepository) DeleteBlob(key string) error {
	fullPath, err := f.getFullPath(key)
	if err != nil {
		return fmt.Errorf("cant get full path: %w", err)
	}

	f.log.Debugw("deleting", "path", fullPath)

	return os.Remove(fullPath)
}

func (f *FileBlobRepository) getFullPath(key string) (string, error) {
	full := filepath.Join(f.path, key)

	rel, err := filepath.Rel(f.path, full)
	if err != nil {
		return "", fmt.Errorf("cant get relative path: %w", err)
	}

	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	return full, nil
}
