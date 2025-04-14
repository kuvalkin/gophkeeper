package entry_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entryService "github.com/kuvalkin/gophkeeper/internal/server/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/server/storage/entry"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestDatabaseMetadataRepository_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, cancel := utils.TestContext(t)
		defer cancel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectQuery("SELECT key, name, notes FROM entries WHERE user_id = \\$1 AND key = \\$2").
			WithArgs("user", "key").
			WillReturnRows(
				sqlmock.NewRows([]string{"key", "name", "notes"}).AddRow("key", "name", []byte("notes")),
			)

		repo := entry.NewDatabaseMetadataRepository(db)
		md, ok, err := repo.GetMetadata(ctx, "user", "key")
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, entryService.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}, md)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, cancel := utils.TestContext(t)
		defer cancel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectQuery("SELECT key, name, notes FROM entries WHERE user_id = \\$1 AND key = \\$2").
			WithArgs("user", "key").
			WillReturnError(sql.ErrNoRows)

		repo := entry.NewDatabaseMetadataRepository(db)
		md, ok, err := repo.GetMetadata(ctx, "user", "key")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, entryService.Metadata{}, md)
	})

	t.Run("query error", func(t *testing.T) {
		ctx, cancel := utils.TestContext(t)
		defer cancel()

		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectQuery("SELECT key, name, notes FROM entries WHERE user_id = \\$1 AND key = \\$2").
			WithArgs("user", "key").
			WillReturnError(errors.New("query error"))

		repo := entry.NewDatabaseMetadataRepository(db)
		md, ok, err := repo.GetMetadata(ctx, "user", "key")
		require.Error(t, err)
		assert.False(t, ok)
		assert.Equal(t, entryService.Metadata{}, md)
	})
}

func TestDatabaseMetadataRepository_Set(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("INSERT INTO entries \\(user_id, key, name, notes\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) ON CONFLICT \\(user_id, key\\) DO UPDATE SET name = excluded\\.name, notes = excluded\\.notes").
			WithArgs("user", "key", "name", []byte("notes")).
			WillReturnResult(sqlmock.NewResult(0, 1))

		repo := entry.NewDatabaseMetadataRepository(db)
		err = repo.SetMetadata(ctx, "user", entryService.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		})
		require.NoError(t, err)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("INSERT INTO entries \\(user_id, key, name, notes\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) ON CONFLICT \\(user_id, key\\) DO UPDATE SET name = excluded\\.name, notes = excluded\\.notes").
			WithArgs("user", "key", "name", []byte("notes")).
			WillReturnError(errors.New("query error"))

		repo := entry.NewDatabaseMetadataRepository(db)
		err = repo.SetMetadata(ctx, "user", entryService.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		})
		require.Error(t, err)
	})
}

func TestDatabaseMetadataRepository_Delete(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("DELETE FROM entries WHERE user_id = \\$1 AND key = \\$2").
			WithArgs("user", "key").
			WillReturnResult(sqlmock.NewResult(0, 1))

		repo := entry.NewDatabaseMetadataRepository(db)
		err = repo.DeleteMetadata(ctx, "user", "key")
		require.NoError(t, err)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("DELETE FROM entries WHERE user_id = \\$1 AND key = \\$2").
			WithArgs("user", "key").
			WillReturnError(errors.New("query error"))

		repo := entry.NewDatabaseMetadataRepository(db)
		err = repo.DeleteMetadata(ctx, "user", "key")
		require.Error(t, err)
	})

}
