// Package entry provides functionality for managing metadata storage in a database.
// It includes a repository implementation for interacting with metadata entries.
package entry

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kuvalkin/gophkeeper/internal/server/service/entry"
)

// NewDatabaseMetadataRepository creates a new instance of DatabaseMetadataRepository.
// It requires a database connection as input.
func NewDatabaseMetadataRepository(db *sql.DB) *DatabaseMetadataRepository {
	return &DatabaseMetadataRepository{
		db: db,
	}
}

// DatabaseMetadataRepository is a repository for managing metadata entries in a database.
type DatabaseMetadataRepository struct {
	db *sql.DB
}

// GetMetadata retrieves metadata for a given user and key from the database.
// It returns the metadata, a boolean indicating if the metadata was found, and an error if any occurred.
func (d *DatabaseMetadataRepository) GetMetadata(ctx context.Context, userID string, key string) (entry.Metadata, bool, error) {
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

// SetMetadata inserts or updates metadata for a given user and key in the database.
// If the key already exists, the metadata is updated.
func (d *DatabaseMetadataRepository) SetMetadata(ctx context.Context, userID string, md entry.Metadata) error {
	_, err := d.db.ExecContext(
		ctx,
		"INSERT INTO entries (user_id, key, name, notes) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, key) DO UPDATE SET name = excluded.name, notes = excluded.notes", userID, md.Key, md.Name, md.Notes)
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}

// DeleteMetadata removes metadata for a given user and key from the database.
// It returns an error if the operation fails.
func (d *DatabaseMetadataRepository) DeleteMetadata(ctx context.Context, userID string, key string) error {
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
