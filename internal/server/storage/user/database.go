// Package user provides the implementation of the user repository for database storage.
package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
)

type dbRepo struct {
	db *sql.DB
}

// NewDatabaseRepository creates a new instance of Repository.
// It takes a database connection pool as input and returns a user.Repository implementation.
func NewDatabaseRepository(db *sql.DB) user.Repository {
	return &dbRepo{db: db}
}

// AddUser adds a new user to the database with the given login and password hash.
// Returns an error if the operation fails, including a specific error if the login is not unique.
func (d *dbRepo) AddUser(ctx context.Context, login string, passwordHash string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", login, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return user.ErrLoginNotUnique
		}

		return fmt.Errorf("query error: %w", err)
	}

	return nil
}

// FindUser retrieves a user's information from the database by their login.
// Returns the user's information, a boolean indicating if the user was found, and an error if the operation fails.
func (d *dbRepo) FindUser(ctx context.Context, login string) (user.UserInfo, bool, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, password_hash FROM users WHERE login = $1", login)

	info := user.UserInfo{}
	err := row.Scan(&info.ID, &info.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.UserInfo{}, false, nil
		}

		return user.UserInfo{}, false, fmt.Errorf("query error: %w", err)
	}

	return info, true, nil
}

// isUniqueViolation checks if the given error is a PostgreSQL unique constraint violation error.
// Returns true if the error is a unique violation, otherwise false.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}
