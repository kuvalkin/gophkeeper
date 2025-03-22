package sync

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func New(metaRepo MetadataRepository, blobRepo blob.Repository) Service {
	return &service{
		metaRepo: metaRepo,
		blobRepo: blobRepo,
		log:      log.Logger().Named("service.sync"),
	}
}

type service struct {
	log      *zap.SugaredLogger
	metaRepo MetadataRepository
	blobRepo blob.Repository
}

func (s *service) Set(ctx context.Context, userID string, md Metadata) (chan<- UploadChunk, <-chan UpdateEntryResult, error) {
	llog := s.log.WithLazy("userID", userID, "key", md.Key)

	dst, err := s.blobRepo.Writer(s.getBlobKey(userID, md.Key))
	if err != nil {
		llog.Errorw("cant get writer", "err", err)

		return nil, nil, ErrInternal
	}

	uploadChan := make(chan UploadChunk)
	resultChan := make(chan UpdateEntryResult, 1)

	go func() {
		defer close(resultChan)

		llog.Debug("waiting for chunks")

		for chunk := range uploadChan {
			if chunk.Err != nil {
				err = dst.Close()
				if err != nil {
					llog.Errorw("cant close writer", "err", err)

					resultChan <- UpdateEntryResult{Err: ErrInternal}
					return
				}

				err = s.blobRepo.Delete(s.getBlobKey(userID, md.Key))
				if err != nil {
					llog.Errorw("cant delete blob", "err", err)

					resultChan <- UpdateEntryResult{Err: ErrInternal}
					return
				}

				resultChan <- UpdateEntryResult{Err: chunk.Err}
				return
			}

			_, err = dst.Write(chunk.Content)
			if err != nil {
				llog.Errorw("write error", "err", err)

				err = dst.Close()
				if err != nil {
					llog.Errorw("cant close writer", "err", err)

					resultChan <- UpdateEntryResult{Err: ErrInternal}
					return
				}

				err = s.blobRepo.Delete(s.getBlobKey(userID, md.Key))
				if err != nil {
					llog.Errorw("cant delete blob", "err", err)

					resultChan <- UpdateEntryResult{Err: ErrInternal}
					return
				}

				resultChan <- UpdateEntryResult{Err: ErrInternal}
				return
			}

			llog.Debug("chunk written")
		}

		err = dst.Close()
		if err != nil {
			llog.Errorw("cant close writer", "err", err)

			resultChan <- UpdateEntryResult{Err: ErrInternal}
			return
		}

		err = s.metaRepo.Set(ctx, userID, md)
		if err != nil {
			llog.Errorw("cant set metadata", "err", err)

			resultChan <- UpdateEntryResult{Err: ErrInternal}
			return
		}

		resultChan <- UpdateEntryResult{}
	}()

	return uploadChan, resultChan, nil
}

func (s *service) Get(ctx context.Context, userID string, key string) (Metadata, io.ReadCloser, bool, error) {
	md, ok, err := s.metaRepo.Get(ctx, userID, key)
	if err != nil {
		return Metadata{}, nil, false, fmt.Errorf("cant get metadata: %w", err)
	}

	if !ok {
		return Metadata{}, nil, false, nil
	}

	rc, ok, err := s.blobRepo.Reader(s.getBlobKey(userID, key))
	if err != nil {
		return Metadata{}, nil, false, fmt.Errorf("cant get blob reader: %w", err)
	}
	if !ok {
		return Metadata{}, nil, false, errors.New("blob not found")
	}

	return md, rc, true, nil
}

func (s *service) Delete(ctx context.Context, userID string, key string) error {
	err := s.metaRepo.Delete(ctx, userID, key)
	if err != nil {
		return fmt.Errorf("cant delete metadata: %w", err)
	}

	err = s.blobRepo.Delete(s.getBlobKey(userID, key))
	if err != nil {
		return fmt.Errorf("cant delete blob: %w", err)
	}

	return nil
}

func (s *service) getBlobKey(userID string, key string) string {
	return fmt.Sprintf("%s/%s", userID, key)
}
