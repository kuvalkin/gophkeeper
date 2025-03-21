package sync

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
	"github.com/kuvalkin/gophkeeper/internal/support/transaction"
)

func New(metaRepo MetadataRepository, blobRepo blob.Repository, txProvider transaction.Provider) *ServiceImpl {
	return &ServiceImpl{
		metaRepo:   metaRepo,
		blobRepo:   blobRepo,
		txProvider: txProvider,
		log:        log.Logger().Named("sync-service"),
	}
}

type ServiceImpl struct {
	txProvider transaction.Provider
	log        *zap.SugaredLogger
	metaRepo   MetadataRepository
	blobRepo   blob.Repository
}

func (s *ServiceImpl) UpdateEntry(ctx context.Context, userID string, md Metadata, lastKnownVersion int64, force bool) (chan<- UpdateEntryChunk, <-chan UpdateEntryResult, error) {
	llog := s.log.WithLazy("userID", userID, "key", md.Key, "lastKnownVersion", lastKnownVersion)

	tx, err := s.txProvider.BeginTx(ctx)
	if err != nil {
		llog.Errorw("cant begin tx", "err", err)

		return nil, nil, ErrInternal
	}

	version, exists, err := s.metaRepo.GetVersion(ctx, tx, userID, md.Key)
	if err != nil {
		llog.Errorw("cant get version and lock", "err", err)

		err = tx.Rollback()
		if err != nil {
			llog.Errorw("cant rollback tx", "err", err)
		}

		return nil, nil, ErrInternal
	}
	if !exists {
		version = 0
	}

	if version != lastKnownVersion && !force {
		llog.Debugw("version mismatch", "version", version, "lastKnownVersion", lastKnownVersion)

		return nil, nil, ErrVersionMismatch
	}

	dst, err := s.blobRepo.Writer(fmt.Sprintf("%s_%s", userID, md.Key))
	if err != nil {
		llog.Errorw("cant get writer", "err", err)

		err = tx.Rollback()
		if err != nil {
			llog.Errorw("cant rollback tx", "err", err)
		}

		return nil, nil, ErrInternal
	}

	uploadDoneChan := make(chan UpdateEntryChunk)
	resultChan := make(chan UpdateEntryResult, 1)
	go func() {
		defer close(resultChan)

		defer func() {
			err = tx.Rollback()
			if err != nil {
				llog.Errorw("cant rollback tx", "err", err)
			}
		}()

		for chunk := range uploadDoneChan {
			if chunk.Err != nil {
				err = dst.CloseWithError(chunk.Err)
				if err != nil {
					llog.Errorw("cant close with error", "err", err)

					resultChan <- UpdateEntryResult{Err: ErrInternal}
					return
				}

				resultChan <- UpdateEntryResult{Err: chunk.Err}
				return
			}

			_, err = dst.Write(chunk.Content)
			if err != nil {
				llog.Errorw("write error", "err", err)

				resultChan <- UpdateEntryResult{Err: ErrInternal}
				return
			}
		}

		newVersion, err := s.metaRepo.Set(ctx, tx, userID, md.Key, md.Name, md.Notes)
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

	return uploadDoneChan, resultChan, nil
}
