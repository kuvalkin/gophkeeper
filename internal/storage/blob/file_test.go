package blob_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
)

func TestFile_New(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		path, err := os.MkdirTemp("", "test-*")
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll(path))
		}()

		repo, err := blob.NewFileBlobRepository(path)
		require.NoError(t, err)
		require.NotNil(t, repo)
	})

	t.Run("invalid path", func(t *testing.T) {
		repo, err := blob.NewFileBlobRepository("")
		require.Error(t, err)
		require.Nil(t, repo)
	})
}

func TestFile_Writer(t *testing.T) {
	path, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(path))
	}()

	repo, err := blob.NewFileBlobRepository(path)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		t.Run("simple key", func(t *testing.T) {
			wc, err := repo.OpenBlobWriter("test-simple-key")
			require.NoError(t, err)
			require.NotNil(t, wc)
		})

		t.Run("with subdirs", func(t *testing.T) {
			wc, err := repo.OpenBlobWriter("test/test-with-subdirs")
			require.NoError(t, err)
			require.NotNil(t, wc)
		})

		t.Run("file already exists", func(t *testing.T) {
			wc, err := repo.OpenBlobWriter("test-already-exists")
			require.NoError(t, err)
			require.NotNil(t, wc)

			wc, err = repo.OpenBlobWriter("test-already-exists")
			require.NoError(t, err)
			require.NotNil(t, wc)
		})
	})

	t.Run("path traversal", func(t *testing.T) {
		wc, err := repo.OpenBlobWriter("../test")
		require.Error(t, err)
		require.Nil(t, wc)

		wc, err = repo.OpenBlobWriter("test/../../test")
		require.Error(t, err)
		require.Nil(t, wc)

		wc, err = repo.OpenBlobWriter("test/../../../../test")
		require.Error(t, err)
		require.Nil(t, wc)
	})

	t.Run("invalid path", func(t *testing.T) {
		wc, err := repo.OpenBlobWriter("")
		require.Error(t, err)
		require.Nil(t, wc)
	})
}

func TestFile_Reader(t *testing.T) {
	path, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(path))
	}()

	repo, err := blob.NewFileBlobRepository(path)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		_, err := repo.OpenBlobWriter("test")
		require.NoError(t, err)

		rc, exists, err := repo.OpenBlobReader("test")
		require.NoError(t, err)
		require.True(t, exists)
		require.NotNil(t, rc)
	})

	t.Run("not exists", func(t *testing.T) {
		rc, exists, err := repo.OpenBlobReader("not-exists")
		require.NoError(t, err)
		require.False(t, exists)
		require.Nil(t, rc)
	})

	t.Run("path traversal", func(t *testing.T) {
		rc, exists, err := repo.OpenBlobReader("../test")
		require.Error(t, err)
		require.False(t, exists)
		require.Nil(t, rc)

		rc, exists, err = repo.OpenBlobReader("test/../../test")
		require.Error(t, err)
		require.False(t, exists)
		require.Nil(t, rc)

		rc, exists, err = repo.OpenBlobReader("test/../../../../test")
		require.Error(t, err)
		require.False(t, exists)
		require.Nil(t, rc)
	})
}

func TestFile_Delete(t *testing.T) {
	path, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(path))
	}()

	repo, err := blob.NewFileBlobRepository(path)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		_, err := repo.OpenBlobWriter("test")
		require.NoError(t, err)

		rc, exists, err := repo.OpenBlobReader("test")
		require.NoError(t, err)
		require.True(t, exists)
		require.NotNil(t, rc)

		err = repo.DeleteBlob("test")
		require.NoError(t, err)

		_, exists, err = repo.OpenBlobReader("test")
		require.NoError(t, err)
		require.False(t, exists)
	})

	t.Run("path traversal", func(t *testing.T) {
		err := repo.DeleteBlob("../test")
		require.Error(t, err)

		err = repo.DeleteBlob("test/../../test")
		require.Error(t, err)

		err = repo.DeleteBlob("test/../../../../test")
		require.Error(t, err)
	})
}
