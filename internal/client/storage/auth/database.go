package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func NewDatabaseRepository(db *sql.DB) (*DatabaseRepository, error) {
	return &DatabaseRepository{db: db}, nil
}

type DatabaseRepository struct {
	db *sql.DB
}

func (d *DatabaseRepository) GetToken(ctx context.Context) (string, bool, error) {
	var token []byte
	err := d.db.QueryRowContext(ctx, "SELECT token FROM accounts WHERE name = 'default'").Scan(&token)

	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("query error: %w", err)
	}

	return string(token), true, nil
}

func (d *DatabaseRepository) SetToken(ctx context.Context, token string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO accounts (name, token) VALUES ('default', ?) ON CONFLICT (name) DO UPDATE SET token = excluded.token", []byte(token))
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}
