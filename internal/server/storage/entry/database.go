package entry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kuvalkin/gophkeeper/internal/server/service/entry"
)

func NewDatabaseMetadataRepository(db *sql.DB) *DatabaseMetadataRepository {
	return &DatabaseMetadataRepository{
		db: db,
	}
}

type DatabaseMetadataRepository struct {
	db *sql.DB
}

func (d *DatabaseMetadataRepository) Get(ctx context.Context, userID string, key string) (entry.Metadata, bool, error) {
	row := d.db.QueryRowContext(
		ctx,
		"SELECT key, name, notes FROM entries WHERE user_id = $1 AND key = $2",
		userID,
		key,
	)

	var md entry.Metadata
	err := row.Scan(&md.Key, &md.Name, &md.Notes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entry.Metadata{}, false, nil
		}

		return entry.Metadata{}, false, fmt.Errorf("query error: %w", err)
	}

	return md, true, nil
}

func (d *DatabaseMetadataRepository) Set(ctx context.Context, userID string, md entry.Metadata) error {
	_, err := d.db.ExecContext(
		ctx,
		"INSERT INTO entries (user_id, key, name, notes) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, key) DO UPDATE SET name = excluded.name, notes = excluded.notes", userID, md.Key, md.Name, md.Notes)
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}

func (d *DatabaseMetadataRepository) Delete(ctx context.Context, userID string, key string) error {
	_, err := d.db.ExecContext(
		ctx,
		"DELETE FROM entries WHERE user_id = $1 AND key = $2",
		userID,
		key,
	)

	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}
