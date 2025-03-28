package entry

import (
	"context"
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

func (s *service) SetEntry(ctx context.Context, userID string, md Metadata, overwrite bool) (chan<- UploadChunk, <-chan SetEntryResult, error) {
	llog := s.log.WithLazy("userID", userID, "key", md.Key, "method", "Set")

	if !overwrite {
		_, ok, err := s.metaRepo.GetMetadata(ctx, userID, md.Key)
		if err != nil {
			llog.Errorw("cant get metadata", "err", err)

			return nil, nil, ErrInternal
		}

		if ok {
			llog.Debug("entry already exists")

			return nil, nil, ErrEntryExists
		}
	}

	blobKey := s.getBlobKey(userID, md.Key)

	dst, err := s.blobRepo.Writer(blobKey)
	if err != nil {
		llog.Errorw("cant get writer", "err", err)

		return nil, nil, ErrInternal
	}

	uploadChan := make(chan UploadChunk)
	// don't wait for caller to read the result
	resultChan := make(chan SetEntryResult, 1)

	go func() {
		defer close(resultChan)

		err = s.processUpload(ctx, uploadChan, dst, llog.Named("upload"))
		if err != nil {
			cderr := s.closeAndDelete(dst, blobKey, llog)
			if cderr != nil {
				resultChan <- SetEntryResult{Err: ErrInternal}
				return
			}

			resultChan <- SetEntryResult{Err: err}
			return
		}

		err = dst.Close()
		if err != nil {
			llog.Errorw("cant close writer", "err", err)

			err = s.blobRepo.Delete(blobKey)
			if err != nil {
				llog.Errorw("cant delete blob", "err", err)
			}

			resultChan <- SetEntryResult{Err: ErrInternal}
			return
		}

		err = s.metaRepo.SetMetadata(ctx, userID, md)
		if err != nil {
			llog.Errorw("cant set metadata", "err", err)

			err = s.blobRepo.Delete(blobKey)
			if err != nil {
				llog.Errorw("cant delete blob", "err", err)
			}

			resultChan <- SetEntryResult{Err: ErrInternal}
			return
		}

		resultChan <- SetEntryResult{}
	}()

	return uploadChan, resultChan, nil
}

func (s *service) processUpload(ctx context.Context, uploadChan <-chan UploadChunk, dst io.WriteCloser, llog *zap.SugaredLogger) error {
	llog.Debug("waiting for chunks")

	writtenAnything := false

	for {
		select {
		case <-ctx.Done():
			llog.Debug("context done")
			return ctx.Err()

		case chunk, ok := <-uploadChan:
			if !ok {
				llog.Debug("no more chunks")

				if !writtenAnything {
					llog.Error("upload closed with no data")

					return ErrNoUpload
				}

				return nil
			}

			if chunk.Err != nil {
				llog.Errorw("received error from upload chunk chan", "err", chunk.Err)

				return ErrUploadChunk
			}

			llog.Debug("received chunk")

			_, err := dst.Write(chunk.Content)
			if err != nil {
				llog.Errorw("write chunk error", "err", err)

				return ErrInternal
			}

			llog.Debug("chunk written")
			writtenAnything = true
		}
	}
}

func (s *service) closeAndDelete(c io.Closer, blobKey string, llog *zap.SugaredLogger) error {
	llog.Debug("deleting blob")

	err := c.Close()
	if err != nil {
		llog.Errorw("cant close writer", "err", err)
	}

	err = s.blobRepo.Delete(blobKey)
	if err != nil {
		llog.Errorw("cant delete blob", "err", err)

		return err
	}

	return nil
}

func (s *service) GetEntry(ctx context.Context, userID string, key string) (Metadata, io.ReadCloser, bool, error) {
	llog := s.log.WithLazy("userID", userID, "key", key, "method", "Get")

	md, ok, err := s.metaRepo.GetMetadata(ctx, userID, key)
	if err != nil {
		llog.Errorw("cant get metadata", "err", err)

		return Metadata{}, nil, false, ErrInternal
	}

	if !ok {
		return Metadata{}, nil, false, nil
	}

	rc, ok, err := s.blobRepo.Reader(s.getBlobKey(userID, key))
	if err != nil {
		llog.Errorw("cant get blob reader", "err", err)

		return Metadata{}, nil, false, ErrInternal
	}
	if !ok {
		llog.Errorw("blob not found")

		return Metadata{}, nil, false, ErrInternal
	}

	return md, rc, true, nil
}

func (s *service) DeleteEntry(ctx context.Context, userID string, key string) error {
	llog := s.log.WithLazy("userID", userID, "key", key, "method", "Delete")

	err := s.metaRepo.DeleteMetadata(ctx, userID, key)
	if err != nil {
		llog.Errorw("cant delete metadata", "err", err)

		return ErrInternal
	}

	err = s.blobRepo.Delete(s.getBlobKey(userID, key))
	if err != nil {
		llog.Errorw("cant delete blob", "err", err)

		return ErrInternal
	}

	return nil
}

func (s *service) getBlobKey(userID string, key string) string {
	return fmt.Sprintf("%s/%s", userID, key)
}
