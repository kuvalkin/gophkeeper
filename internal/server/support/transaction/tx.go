package transaction

import (
	"context"
	"database/sql"
	"errors"
)

type Provider interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	// Rollback a transaction. If the transaction is already commited or rolled back, should return nil
	Rollback() error
}

func NewDatabaseTransactionProvider(db *sql.DB) *DatabaseTransactionProvider {
	return &DatabaseTransactionProvider{
		db: db,
	}
}

type DatabaseTransactionProvider struct {
	db *sql.DB
}

func (d *DatabaseTransactionProvider) BeginTx(ctx context.Context) (Tx, error) {
	dbTx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &DatabaseTx{
		Tx: dbTx,
	}, nil
}

type DatabaseTx struct {
	Tx *sql.Tx
}

func (d *DatabaseTx) Commit() error {
	return d.Tx.Commit()
}

func (d *DatabaseTx) Rollback() error {
	err := d.Tx.Rollback()
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
}
