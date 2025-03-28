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

func NewDatabaseRepository(db *sql.DB) user.Repository {
	return &dbRepo{db: db}
}

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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}

	return false
}
