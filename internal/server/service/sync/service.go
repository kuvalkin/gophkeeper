package sync

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/transaction"
)

func New() (*MyService, error) {
	return &MyService{}, nil
}

type MyService struct {
	txProvider transaction.Provider
	log        *zap.SugaredLogger
	metaRepo   MetadataRepository
	blobRepo   BlobRepository
}

func (m *MyService) UpdateEntry(ctx context.Context, userID string, md Metadata, lastKnownVersion int64, force bool) (*io.PipeWriter, error, <-chan UpdateEntryResult) {
	llog := m.log.WithLazy("userID", userID, "key", md.Key, "lastKnownVersion", lastKnownVersion)

	tx, err := m.txProvider.BeginTx(ctx)
	if err != nil {
		llog.Errorw("cant begin tx", "err", err)

		return nil, ErrInternal, nil
	}

	version, err := m.metaRepo.GetAndLock(ctx, tx, userID, md.Key)
	if err != nil {
		llog.Errorw("cant get version and lock", "err", err)

		err = tx.Rollback()
		if err != nil {
			llog.Errorw("cant rollback tx", "err", err)
		}

		return nil, ErrInternal, nil
	}

	if version != lastKnownVersion && !force {
		llog.Debugw("version mismatch", "version", version, "lastKnownVersion", lastKnownVersion)

		return nil, ErrVersionMismatch, nil
	}

	pr, pw := io.Pipe()

	resultChan := make(chan UpdateEntryResult, 1)
	go func() {
		defer pr.Close()
		defer func() {
			err = tx.Rollback()
			if err != nil {
				llog.Errorw("cant rollback tx", "err", err)
			}
		}()

		err = m.blobRepo.CopyFrom(ctx, fmt.Sprintf("%s/%s", userID, md.Key), pr)
		if err != nil {
			llog.Errorw("cant copy from", "err", err)

			resultChan <- UpdateEntryResult{Err: ErrInternal}
			return
		}

		newVersion, err := m.metaRepo.SetAndUnlock(ctx, tx, userID, md.Key, md.Name, md.Notes)
		if err != nil {
			llog.Errorw("cant set and unlock", "err", err)

			resultChan <- UpdateEntryResult{Err: ErrInternal}
			return
		}

		err = tx.Commit()
		if err != nil {
			llog.Errorw("cant commit tx", "err", err)

			resultChan <- UpdateEntryResult{Err: ErrInternal}
			return
		}

		resultChan <- UpdateEntryResult{NewVersion: newVersion}
	}()

	return pw, nil, resultChan
}
