package sync

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kuvalkin/gophkeeper/internal/support/transaction"
)

func NewDatabaseMetadataRepository() *DatabaseMetadataRepository {
	return &DatabaseMetadataRepository{}
}

type DatabaseMetadataRepository struct {
}

func (d *DatabaseMetadataRepository) GetVersion(ctx context.Context, tx transaction.Tx, userID string, key string) (version int64, exists bool, err error) {
	dbTx, ok := tx.(*transaction.DatabaseTx)
	if !ok {
		return 0, exists, errors.New("tx is not a database transaction")
	}

	err = dbTx.Tx.QueryRowContext(ctx, "SELECT version FROM entries WHERE user_id = $1 AND key = $2", userID, key).Scan(&version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("query error: %w", err)
	}

	return version, true, nil
}

func (d *DatabaseMetadataRepository) Set(ctx context.Context, tx transaction.Tx, userID string, key string, name string, notes []byte) (version int64, err error) {
	dbTx, ok := tx.(*transaction.DatabaseTx)
	if !ok {
		return 0, errors.New("tx is not a database transaction")
	}

	row := dbTx.Tx.QueryRowContext(ctx, "INSERT INTO entries (user_id, key, name, notes, version) VALUES ($1, $2, $3, $4, version + 1) ON CONFLICT (user_id, key) DO UPDATE SET name = excluded.name, notes = excluded.notes, version = 0 RETURNING version", userID, key, name, notes)

	err = row.Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	_, err = dbTx.Tx.ExecContext(ctx, "UPDATE users SET version = version + 1 WHERE id = $1", userID)
	if err != nil {
		return 0, fmt.Errorf("query error: %w", err)
	}

	return version, nil
}
