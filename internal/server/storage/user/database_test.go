package user_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userService "github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/storage/user"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestDatabaseRepository_Add(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("INSERT INTO users \\(login, password_hash\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs("login", "hash").
			WillReturnResult(sqlmock.NewResult(1, 1))

		repo := user.NewDatabaseRepository(db)
		err = repo.AddUser(ctx, "login", "hash")
		require.NoError(t, err)
	})

	t.Run("login not unique", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("INSERT INTO users \\(login, password_hash\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs("login", "hash").
			WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

		repo := user.NewDatabaseRepository(db)
		err = repo.AddUser(ctx, "login", "hash")
		require.ErrorIs(t, err, userService.ErrLoginNotUnique)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectExec("INSERT INTO users \\(login, password_hash\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs("login", "hash").
			WillReturnError(errors.New("error"))

		repo := user.NewDatabaseRepository(db)
		err = repo.AddUser(ctx, "login", "hash")
		require.Error(t, err)
	})
}

func TestDatabaseRepository_Find(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectQuery("SELECT id, password_hash FROM users WHERE login = \\$1").
			WithArgs("login").
			WillReturnRows(sqlmock.NewRows([]string{"id", "password_hash"}).AddRow("id", "hash"))

		repo := user.NewDatabaseRepository(db)
		info, ok, err := repo.FindUser(ctx, "login")
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "id", info.ID)
		assert.Equal(t, "hash", info.PasswordHash)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectQuery("SELECT id, password_hash FROM users WHERE login = \\$1").
			WithArgs("login").
			WillReturnError(sql.ErrNoRows)

		repo := user.NewDatabaseRepository(db)
		info, ok, err := repo.FindUser(ctx, "login")
		require.NoError(t, err)
		require.False(t, ok)
		assert.Empty(t, info)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, mock.ExpectationsWereMet())
		}()

		mock.
			ExpectQuery("SELECT id, password_hash FROM users WHERE login = \\$1").
			WithArgs("login").
			WillReturnError(errors.New("error"))

		repo := user.NewDatabaseRepository(db)
		info, ok, err := repo.FindUser(ctx, "login")
		require.Error(t, err)
		require.False(t, ok)
		assert.Empty(t, info)
	})
}
