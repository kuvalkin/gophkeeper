package entry_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/server/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/support/test"
)

func TestService_Set(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("no overwrite", func(t *testing.T) {
		t.Run("entry exists", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)

			metaRepo.EXPECT().Get(ctx, "user", "key").Return(entry.Metadata{}, true, nil)

			s := entry.New(metaRepo, blobRepo)
			upload, result, err := s.Set(ctx, "user", entry.Metadata{Key: "key"}, false)
			require.ErrorIs(t, err, entry.ErrEntryExists)
			require.Nil(t, upload)
			require.Nil(t, result)
		})

		t.Run("repo returns err", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)

			metaRepo.EXPECT().Get(ctx, "user", "key").Return(entry.Metadata{}, false, errors.New("query error"))

			s := entry.New(metaRepo, blobRepo)
			upload, result, err := s.Set(ctx, "user", entry.Metadata{Key: "key"}, false)
			require.ErrorIs(t, err, entry.ErrInternal)
			require.Nil(t, upload)
			require.Nil(t, result)
		})

		t.Run("entry does not exist (success)", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			metaRepo.EXPECT().Get(ctx, "user", "key").Return(entry.Metadata{}, false, nil)
			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Write([]byte("chunk")).Return(0, nil)
			writer.EXPECT().Close().Return(nil)
			metaRepo.EXPECT().Set(ctx, "user", md).Return(nil)

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, false)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
			close(uploadChan)

			result := <-resultChan
			require.NoError(t, result.Err)
		})
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)
		writer := NewMockWriteCloser(ctrl)

		md := entry.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}

		blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
		writer.EXPECT().Write([]byte("chunk")).Return(0, nil)
		writer.EXPECT().Close().Return(nil)
		metaRepo.EXPECT().Set(ctx, "user", md).Return(nil)

		s := entry.New(metaRepo, blobRepo)
		uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
		require.NoError(t, err)
		require.NotNil(t, uploadChan)
		require.NotNil(t, resultChan)

		uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
		close(uploadChan)

		result := <-resultChan
		require.NoError(t, result.Err)
	})

	t.Run("metadata set err", func(t *testing.T) {
		t.Run("delete ok", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Write([]byte("chunk")).Return(0, nil)
			writer.EXPECT().Close().Return(nil)
			metaRepo.EXPECT().Set(ctx, "user", md).Return(errors.New("query failed"))
			blobRepo.EXPECT().Delete("user/key").Return(nil)

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
			close(uploadChan)

			result := <-resultChan
			require.Error(t, result.Err, entry.ErrInternal)
		})

		t.Run("delete err", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Write([]byte("chunk")).Return(0, nil)
			writer.EXPECT().Close().Return(nil)
			metaRepo.EXPECT().Set(ctx, "user", md).Return(errors.New("query failed"))
			blobRepo.EXPECT().Delete("user/key").Return(errors.New("close fail"))

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
			close(uploadChan)

			result := <-resultChan
			require.Error(t, result.Err, entry.ErrInternal)
		})
	})

	t.Run("writer close err", func(t *testing.T) {
		t.Run("delete ok", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Write([]byte("chunk")).Return(0, nil)
			writer.EXPECT().Close().Return(errors.New("close failed"))
			blobRepo.EXPECT().Delete("user/key").Return(nil)

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
			close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, entry.ErrInternal)
		})

		t.Run("delete err", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Write([]byte("chunk")).Return(0, nil)
			writer.EXPECT().Close().Return(errors.New("close failed"))
			blobRepo.EXPECT().Delete("user/key").Return(errors.New("delete failed"))

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
			close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, entry.ErrInternal)
		})
	})

	t.Run("can open writer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)

		md := entry.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}

		blobRepo.EXPECT().Writer("user/key").Return(nil, errors.New("cant open writer"))

		s := entry.New(metaRepo, blobRepo)
		uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
		require.ErrorIs(t, err, entry.ErrInternal)
		require.Nil(t, uploadChan)
		require.Nil(t, resultChan)
	})

	t.Run("upload failed", func(t *testing.T) {
		t.Run("closed without writes", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Close().Return(nil)
			blobRepo.EXPECT().Delete("user/key").Return(nil)

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, entry.ErrNoUpload)
		})

		t.Run("closed without writes, close and delete err", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Close().Return(errors.New("close failed"))
			blobRepo.EXPECT().Delete("user/key").Return(errors.New("delete failed"))

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, entry.ErrInternal)
		})

		t.Run("ctx done", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Close().Return(nil)
			blobRepo.EXPECT().Delete("user/key").Return(nil)

			localCtx, cancel := context.WithCancel(ctx)
			// already closed
			cancel()

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(localCtx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			defer close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, localCtx.Err())
		})

		t.Run("err in upload chan", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Close().Return(nil)
			blobRepo.EXPECT().Delete("user/key").Return(nil)

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Err: errors.New("some error")}
			close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, entry.ErrUploadChunk)
		})

		t.Run("write err", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			metaRepo := NewMockMetadataRepository(ctrl)
			blobRepo := NewMockRepository(ctrl)
			writer := NewMockWriteCloser(ctrl)

			md := entry.Metadata{
				Key:   "key",
				Name:  "name",
				Notes: []byte("notes"),
			}

			blobRepo.EXPECT().Writer("user/key").Return(writer, nil)
			writer.EXPECT().Write([]byte("chunk")).Return(0, errors.New("write failed"))
			writer.EXPECT().Close().Return(nil)
			blobRepo.EXPECT().Delete("user/key").Return(nil)

			s := entry.New(metaRepo, blobRepo)
			uploadChan, resultChan, err := s.Set(ctx, "user", md, true)
			require.NoError(t, err)
			require.NotNil(t, uploadChan)
			require.NotNil(t, resultChan)

			uploadChan <- entry.UploadChunk{Content: []byte("chunk")}
			close(uploadChan)

			result := <-resultChan
			require.ErrorIs(t, result.Err, entry.ErrInternal)
		})
	})
}

func TestService_Get(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)
		reader := io.NopCloser(bytes.NewBuffer(nil))

		md := entry.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}

		metaRepo.EXPECT().Get(ctx, "user", "key").Return(md, true, nil)
		blobRepo.EXPECT().Reader("user/key").Return(reader, true, nil)

		s := entry.New(metaRepo, blobRepo)
		meta, r, ok, err := s.Get(ctx, "user", "key")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, md, meta)
		require.NotNil(t, r)
	})

	t.Run("metadata not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)

		metaRepo.EXPECT().Get(ctx, "user", "key").Return(entry.Metadata{}, false, nil)

		s := entry.New(metaRepo, blobRepo)
		meta, r, ok, err := s.Get(ctx, "user", "key")
		require.NoError(t, err)
		require.False(t, ok)
		require.Equal(t, entry.Metadata{}, meta)
		require.Nil(t, r)
	})

	t.Run("metadata repo err", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)

		md := entry.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}

		metaRepo.EXPECT().Get(ctx, "user", "key").Return(md, false, errors.New("query failed"))

		s := entry.New(metaRepo, blobRepo)
		meta, r, ok, err := s.Get(ctx, "user", "key")
		require.ErrorIs(t, err, entry.ErrInternal)
		require.False(t, ok)
		require.Equal(t, entry.Metadata{}, meta)
		require.Nil(t, r)
	})

	t.Run("blob repo err", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)

		md := entry.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}

		metaRepo.EXPECT().Get(ctx, "user", "key").Return(md, true, nil)
		blobRepo.EXPECT().Reader("user/key").Return(nil, false, errors.New("query failed"))

		s := entry.New(metaRepo, blobRepo)
		meta, r, ok, err := s.Get(ctx, "user", "key")
		require.ErrorIs(t, err, entry.ErrInternal)
		require.False(t, ok)
		require.Equal(t, entry.Metadata{}, meta)
		require.Nil(t, r)
	})

	t.Run("blob not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metaRepo := NewMockMetadataRepository(ctrl)
		blobRepo := NewMockRepository(ctrl)

		md := entry.Metadata{
			Key:   "key",
			Name:  "name",
			Notes: []byte("notes"),
		}

		metaRepo.EXPECT().Get(ctx, "user", "key").Return(md, true, nil)
		blobRepo.EXPECT().Reader("user/key").Return(nil, false, nil)

		s := entry.New(metaRepo, blobRepo)
		meta, r, ok, err := s.Get(ctx, "user", "key")
		require.ErrorIs(t, err, entry.ErrInternal)
		require.False(t, ok)
		require.Equal(t, entry.Metadata{}, meta)
		require.Nil(t, r)
	})
}
