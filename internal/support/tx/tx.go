package tx

import "context"

type Provider interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	// Rollback a transaction. If the transaction is already commited or rolled back, should return nil
	Rollback() error
}
