package sync

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kuvalkin/gophkeeper/internal/server/service/sync"
)

func NewDatabaseMetadataRepository(db *sql.DB) *DatabaseMetadataRepository {
	return &DatabaseMetadataRepository{
		db: db,
	}
}

type DatabaseMetadataRepository struct {
	db *sql.DB
}

func (d *DatabaseMetadataRepository) Set(ctx context.Context, userID string, md sync.Metadata) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO entries (user_id, key, name, notes) VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, key) DO UPDATE SET name = excluded.name, notes = excluded.notes", userID, md.Key, md.Name, md.Notes)
	if err != nil {
		return fmt.Errorf("query error: %w", err)
	}

	return nil
}
