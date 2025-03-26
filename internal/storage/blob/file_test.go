package blob_test

import (
	"os"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
	"github.com/stretchr/testify/require"
)

func TestFile_Writer(t *testing.T) {
	path, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	repo := blob.NewFileBlobRepository(path)

	t.Run("success", func(t *testing.T) {
		t.Run("simple key", func(t *testing.T) {
			wc, err := repo.Writer("test-simple-key")
			require.NoError(t, err)
			require.NotNil(t, wc)
		})

		t.Run("with subdirs", func(t *testing.T) {
			wc, err := repo.Writer("test/test-with-subdirs")
			require.NoError(t, err)
			require.NotNil(t, wc)
		})

		t.Run("file already exists", func(t *testing.T) {
			wc, err := repo.Writer("test-already-exists")
			require.NoError(t, err)
			require.NotNil(t, wc)

			wc, err = repo.Writer("test-already-exists")
			require.NoError(t, err)
			require.NotNil(t, wc)
		})
	})
}

func TestFile_Reader(t *testing.T) {
	path, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	repo := blob.NewFileBlobRepository(path)

	t.Run("success", func(t *testing.T) {
		_, err := repo.Writer("test")
		require.NoError(t, err)

		rc, exists, err := repo.Reader("test")
		require.NoError(t, err)
		require.True(t, exists)
		require.NotNil(t, rc)
	})

	t.Run("not exists", func(t *testing.T) {
		rc, exists, err := repo.Reader("not-exists")
		require.NoError(t, err)
		require.False(t, exists)
		require.Nil(t, rc)
	})
}

func TestFile_Delete(t *testing.T) {
	path, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	repo := blob.NewFileBlobRepository(path)

	t.Run("success", func(t *testing.T) {
		_, err := repo.Writer("test")
		require.NoError(t, err)

		rc, exists, err := repo.Reader("test")
		require.NoError(t, err)
		require.True(t, exists)
		require.NotNil(t, rc)

		err = repo.Delete("test")
		require.NoError(t, err)

		_, exists, err = repo.Reader("test")
		require.NoError(t, err)
		require.False(t, exists)
	})
}