package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
)

type dbRepo struct {
	db      *sql.DB
	timeout time.Duration
}

func NewDatabaseRepository(db *sql.DB, timeout time.Duration) user.Repository {
	return &dbRepo{db: db, timeout: timeout}
}

func (d *dbRepo) Add(ctx context.Context, login string, passwordHash string) error {
	localCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	_, err := d.db.ExecContext(localCtx, "INSERT INTO users (login, password_hash) VALUES ($1, $2)", login, passwordHash)
	if err != nil {
		if isUniqueViolation(err) {
			return user.ErrLoginNotUnique
		}

		return fmt.Errorf("query error: %w", err)
	}

	return nil
}

func (d *dbRepo) Find(ctx context.Context, login string) (string, string, bool, error) {
	localCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	row := d.db.QueryRowContext(localCtx, "SELECT id, password_hash FROM users WHERE login = $1", login)

	var userID string
	var hash string
	err := row.Scan(&userID, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", false, nil
		}

		return "", "", false, fmt.Errorf("query error: %w", err)
	}

	return userID, hash, true, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}
