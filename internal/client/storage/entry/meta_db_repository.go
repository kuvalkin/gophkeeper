package entry

import (
	"context"
	"database/sql"

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

func (d *DatabaseMetadataRepository) Set(ctx context.Context, tx tx.Tx, key string, name string, version int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DatabaseMetadataRepository) GetVersion(ctx context.Context, tx tx.Tx, key string) (int64, bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DatabaseMetadataRepository) MarkAsDownloaded(ctx context.Context, tx tx.Tx, key string) error {
	//TODO implement me
	panic("implement me")
}

func (d *DatabaseMetadataRepository) MarkAsDeleted(ctx context.Context, tx tx.Tx, key string) error {
	//TODO implement me
	panic("implement me")
}
