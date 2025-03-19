package entry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kuvalkin/gophkeeper/internal/support/tx"
)

func NewDatabaseMetadataRepository(db *sql.DB) (*DatabaseMetadataRepository, error) {
	return &DatabaseMetadataRepository{
		db: db,
	}, nil
}

type DatabaseMetadataRepository struct {
	db *sql.DB
}

func (d *DatabaseMetadataRepository) Set(ctx context.Context, tx tx.Tx, key string, name string, notes []byte, version int64) error {
	_, err := d.db.ExecContext(
		ctx,
		"INSERT INTO entries (key, name, notes, version) VALUES (?, ?, ?, ?) ON CONFLICT (key) DO UPDATE SET name = excluded.name, notes = excluded.notes, version = excluded.version",
		key,
		name,
		notes,
		version,
	)

	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}

func (d *DatabaseMetadataRepository) GetVersion(ctx context.Context, tx tx.Tx, key string) (int64, bool, error) {
	var version int64
	err := d.db.QueryRowContext(ctx, "SELECT version FROM entries WHERE key = ?", key).Scan(&version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("query error: %w", err)
	}

	return version, true, nil
}

func (d *DatabaseMetadataRepository) MarkAsDownloaded(ctx context.Context, tx tx.Tx, key string) error {
	_, err := d.db.ExecContext(ctx, "UPDATE entries SET is_downloaded = true WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}

func (d *DatabaseMetadataRepository) MarkAsDeleted(ctx context.Context, tx tx.Tx, key string) error {
	_, err := d.db.ExecContext(ctx, "UPDATE entries SET is_deleted = true WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}
