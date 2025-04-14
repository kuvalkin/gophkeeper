package entry_test

import (
	"errors"
	io "io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/client/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
	pb "github.com/kuvalkin/gophkeeper/pkg/proto/entry/v1"
)

const chunkSize = 1024

func TestService_Set(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		encryptWriter := mocks.NewMockWriteCloser(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)

		// encrypt blob
		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
		rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
		rawContent.EXPECT().Close().Return(nil)
		encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
		encryptWriter.EXPECT().Close().Return(nil)
		blobWriter.EXPECT().Close().Return(nil)

		// encrypt notes
		notesEncrypter := mocks.NewMockWriteCloser(ctrl)
		crypt.EXPECT().Encrypt(gomock.Any()).Return(notesEncrypter, nil)
		notesEncrypter.EXPECT().Write([]byte("notes")).Return(5, nil)
		notesEncrypter.EXPECT().Close().Return(nil)

		encryptedContent := mocks.NewMockReadCloser(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)
		stream := mocks.NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

		// send metadata
		blobRepo.EXPECT().OpenBlobReader("key").Return(encryptedContent, true, nil)
		client.EXPECT().SetEntry(ctx).Return(stream, nil)
		stream.EXPECT().Send(&pb.SetEntryRequest{
			Entry: &pb.Entry{
				Key:     "key",
				Name:    "name",
				Notes:   nil, // encrypter will return nil bytes
				Content: nil,
			},
		}).Return(nil)
		stream.EXPECT().Recv().Return(&pb.SetEntryResponse{}, nil)

		// upload
		encryptedContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("encrypted")).Return(9, nil)
		encryptedContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte{}).Return(0, io.EOF)
		encryptedContent.EXPECT().Close().Return(nil)
		stream.EXPECT().Send(&pb.SetEntryRequest{
			Entry: &pb.Entry{
				Content: []byte("encrypted"),
			},
		}).Return(nil)
		stream.EXPECT().CloseSend().Return(nil).MinTimes(1)
		stream.EXPECT().Recv().Return(nil, io.EOF)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "notes", rawContent, nil)
		require.NoError(t, err)
	})

	t.Run("error creating blob writer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)

		blobRepo.EXPECT().OpenBlobWriter("key").Return(nil, errors.New("error"))
		rawContent.EXPECT().Close().Return(nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "notes", rawContent, nil)
		require.Error(t, err)
	})

	t.Run("error creating encrypt writer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)

		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(nil, errors.New("error"))
		rawContent.EXPECT().Close().Return(nil)
		blobWriter.EXPECT().Close().Return(nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "notes", rawContent, nil)
		require.Error(t, err)
	})

	t.Run("error reading raw content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		encryptWriter := mocks.NewMockWriteCloser(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)

		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
		rawContent.EXPECT().Read(gomock.Any()).Return(0, errors.New("error"))
		rawContent.EXPECT().Close().Return(nil)
		blobWriter.EXPECT().Close().Return(nil)
		encryptWriter.EXPECT().Close().Return(nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "notes", rawContent, nil)
		require.Error(t, err)
	})

	t.Run("error closing encrypt writer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		encryptWriter := mocks.NewMockWriteCloser(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)

		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
		rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
		rawContent.EXPECT().Close().Return(nil)
		encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
		encryptWriter.EXPECT().Close().Return(errors.New("error"))
		blobWriter.EXPECT().Close().Return(nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "notes", rawContent, nil)
		require.Error(t, err)
	})

	t.Run("already exists", func(t *testing.T) {
		t.Run("no callback", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			crypt := mocks.NewMockCrypt(ctrl)
			blobRepo := mocks.NewMockBlobRepository(ctrl)
			blobWriter := mocks.NewMockWriteCloser(ctrl)
			encryptWriter := mocks.NewMockWriteCloser(ctrl)
			rawContent := mocks.NewMockReadCloser(ctrl)

			// encrypt blob
			blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
			crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
			rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
			rawContent.EXPECT().Close().Return(nil)
			encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
			encryptWriter.EXPECT().Close().Return(nil)
			blobWriter.EXPECT().Close().Return(nil)

			encryptedContent := mocks.NewMockReadCloser(ctrl)
			client := mocks.NewMockEntryServiceClient(ctrl)
			stream := mocks.NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

			// send metadata
			blobRepo.EXPECT().OpenBlobReader("key").Return(encryptedContent, true, nil)
			client.EXPECT().SetEntry(ctx).Return(stream, nil)
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Entry: &pb.Entry{
					Key:     "key",
					Name:    "name",
					Notes:   nil, // encrypter will return nil bytes
					Content: nil,
				},
			}).Return(nil)
			stream.EXPECT().Recv().Return(&pb.SetEntryResponse{
				AlreadyExists: true,
			}, nil)
			stream.EXPECT().CloseSend().Return(nil)
			encryptedContent.EXPECT().Close().Return(nil)

			service := entry.New(crypt, client, blobRepo, chunkSize)
			err := service.SetEntry(ctx, "key", "name", "", rawContent, nil)
			require.ErrorIs(t, err, entry.ErrEntryExists)
		})

		t.Run("callback returns false", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			crypt := mocks.NewMockCrypt(ctrl)
			blobRepo := mocks.NewMockBlobRepository(ctrl)
			blobWriter := mocks.NewMockWriteCloser(ctrl)
			encryptWriter := mocks.NewMockWriteCloser(ctrl)
			rawContent := mocks.NewMockReadCloser(ctrl)

			// encrypt blob
			blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
			crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
			rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
			rawContent.EXPECT().Close().Return(nil)
			encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
			encryptWriter.EXPECT().Close().Return(nil)
			blobWriter.EXPECT().Close().Return(nil)

			encryptedContent := mocks.NewMockReadCloser(ctrl)
			client := mocks.NewMockEntryServiceClient(ctrl)
			stream := mocks.NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

			// send metadata
			blobRepo.EXPECT().OpenBlobReader("key").Return(encryptedContent, true, nil)
			client.EXPECT().SetEntry(ctx).Return(stream, nil)
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Entry: &pb.Entry{
					Key:     "key",
					Name:    "name",
					Notes:   nil, // encrypter will return nil bytes
					Content: nil,
				},
			}).Return(nil)
			stream.EXPECT().Recv().Return(&pb.SetEntryResponse{
				AlreadyExists: true,
			}, nil)
			stream.EXPECT().CloseSend().Return(nil)
			encryptedContent.EXPECT().Close().Return(nil)

			service := entry.New(crypt, client, blobRepo, chunkSize)
			err := service.SetEntry(ctx, "key", "name", "", rawContent, func() bool {
				return false
			})
			require.ErrorIs(t, err, entry.ErrEntryExists)
		})

		t.Run("callback returns true", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			crypt := mocks.NewMockCrypt(ctrl)
			blobRepo := mocks.NewMockBlobRepository(ctrl)
			blobWriter := mocks.NewMockWriteCloser(ctrl)
			encryptWriter := mocks.NewMockWriteCloser(ctrl)
			rawContent := mocks.NewMockReadCloser(ctrl)

			// encrypt blob
			blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
			crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
			rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
			rawContent.EXPECT().Close().Return(nil)
			encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
			encryptWriter.EXPECT().Close().Return(nil)
			blobWriter.EXPECT().Close().Return(nil)

			// encrypt notes
			notesEncrypter := mocks.NewMockWriteCloser(ctrl)
			crypt.EXPECT().Encrypt(gomock.Any()).Return(notesEncrypter, nil)
			notesEncrypter.EXPECT().Write([]byte("notes")).Return(5, nil)
			notesEncrypter.EXPECT().Close().Return(nil)

			encryptedContent := mocks.NewMockReadCloser(ctrl)
			client := mocks.NewMockEntryServiceClient(ctrl)
			stream := mocks.NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

			// send metadata
			blobRepo.EXPECT().OpenBlobReader("key").Return(encryptedContent, true, nil)
			client.EXPECT().SetEntry(ctx).Return(stream, nil)
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Entry: &pb.Entry{
					Key:     "key",
					Name:    "name",
					Notes:   nil, // encrypter will return nil bytes
					Content: nil,
				},
			}).Return(nil)
			stream.EXPECT().Recv().Return(&pb.SetEntryResponse{
				AlreadyExists: true,
			}, nil)

			// confirm overwrite
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Overwrite: true,
			}).Return(nil)

			// upload
			encryptedContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("encrypted")).Return(9, nil)
			encryptedContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte{}).Return(0, io.EOF)
			encryptedContent.EXPECT().Close().Return(nil)
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Entry: &pb.Entry{
					Content: []byte("encrypted"),
				},
			}).Return(nil)
			stream.EXPECT().CloseSend().Return(nil).MinTimes(1)
			stream.EXPECT().Recv().Return(nil, io.EOF)

			service := entry.New(crypt, client, blobRepo, chunkSize)
			err := service.SetEntry(ctx, "key", "name", "notes", rawContent, func() bool {
				return true
			})
			require.NoError(t, err)
		})

		t.Run("callback returns true, error sending confirmation", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			crypt := mocks.NewMockCrypt(ctrl)
			blobRepo := mocks.NewMockBlobRepository(ctrl)
			blobWriter := mocks.NewMockWriteCloser(ctrl)
			encryptWriter := mocks.NewMockWriteCloser(ctrl)
			rawContent := mocks.NewMockReadCloser(ctrl)

			// encrypt blob
			blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
			crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
			rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
			rawContent.EXPECT().Close().Return(nil)
			encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
			encryptWriter.EXPECT().Close().Return(nil)
			blobWriter.EXPECT().Close().Return(nil)

			// encrypt notes
			notesEncrypter := mocks.NewMockWriteCloser(ctrl)
			crypt.EXPECT().Encrypt(gomock.Any()).Return(notesEncrypter, nil)
			notesEncrypter.EXPECT().Write([]byte("notes")).Return(5, nil)
			notesEncrypter.EXPECT().Close().Return(nil)

			encryptedContent := mocks.NewMockReadCloser(ctrl)
			client := mocks.NewMockEntryServiceClient(ctrl)
			stream := mocks.NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

			// send metadata
			blobRepo.EXPECT().OpenBlobReader("key").Return(encryptedContent, true, nil)
			client.EXPECT().SetEntry(ctx).Return(stream, nil)
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Entry: &pb.Entry{
					Key:     "key",
					Name:    "name",
					Notes:   nil, // encrypter will return nil bytes
					Content: nil,
				},
			}).Return(nil)
			stream.EXPECT().Recv().Return(&pb.SetEntryResponse{
				AlreadyExists: true,
			}, nil)

			// confirm overwrite
			stream.EXPECT().Send(&pb.SetEntryRequest{
				Overwrite: true,
			}).Return(errors.New("error"))

			encryptedContent.EXPECT().Close().Return(nil)
			stream.EXPECT().CloseSend().Return(nil).MinTimes(1)

			service := entry.New(crypt, client, blobRepo, chunkSize)
			err := service.SetEntry(ctx, "key", "name", "notes", rawContent, func() bool {
				return true
			})
			require.Error(t, err)
		})
	})

	t.Run("upload error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		encryptWriter := mocks.NewMockWriteCloser(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)

		// encrypt blob
		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
		rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
		rawContent.EXPECT().Close().Return(nil)
		encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
		encryptWriter.EXPECT().Close().Return(nil)
		blobWriter.EXPECT().Close().Return(nil)

		encryptedContent := mocks.NewMockReadCloser(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)
		stream := mocks.NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

		// send metadata
		blobRepo.EXPECT().OpenBlobReader("key").Return(encryptedContent, true, nil)
		client.EXPECT().SetEntry(ctx).Return(stream, nil)
		stream.EXPECT().Send(&pb.SetEntryRequest{
			Entry: &pb.Entry{
				Key:     "key",
				Name:    "name",
				Notes:   nil,
				Content: nil,
			},
		}).Return(nil)
		stream.EXPECT().Recv().Return(&pb.SetEntryResponse{}, nil)

		// upload
		encryptedContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("encrypted")).Return(9, nil)
		encryptedContent.EXPECT().Close().Return(nil)
		stream.EXPECT().Send(&pb.SetEntryRequest{
			Entry: &pb.Entry{
				Content: []byte("encrypted"),
			},
		}).Return(errors.New("error"))

		stream.EXPECT().CloseSend().Return(nil).MinTimes(1)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "", rawContent, nil)
		require.Error(t, err)
	})

	t.Run("error encrypting notes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		encryptWriter := mocks.NewMockWriteCloser(ctrl)
		rawContent := mocks.NewMockReadCloser(ctrl)

		// encrypt blob
		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
		rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
		rawContent.EXPECT().Close().Return(nil)
		encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
		encryptWriter.EXPECT().Close().Return(nil)
		blobWriter.EXPECT().Close().Return(nil)

		// encrypt notes
		crypt.EXPECT().Encrypt(gomock.Any()).Return(nil, errors.New("error"))

		service := entry.New(crypt, mocks.NewMockEntryServiceClient(ctrl), blobRepo, chunkSize)
		err := service.SetEntry(ctx, "key", "name", "notes", rawContent, nil)
		require.Error(t, err)
	})
}

func TestService_Get(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)
		stream := mocks.NewMockServerStreamingClient[pb.Entry](ctrl)

		// send request
		client.EXPECT().GetEntry(ctx, &pb.GetEntryRequest{
			Key: "key",
		}).Return(stream, nil)

		// receive metadata
		stream.EXPECT().Recv().Return(&pb.Entry{
			Key:     "key",
			Name:    "name",
			Notes:   []byte("encrypted notes"),
			Content: nil,
		}, nil)

		// decrypt notes
		notesReader := mocks.NewMockReadCloser(ctrl)
		crypt.EXPECT().Decrypt(gomock.Any()).Return(notesReader, nil)
		notesReader.EXPECT().Read(gomock.Any()).SetArg(0, []byte("notes")).Return(5, nil)
		notesReader.EXPECT().Read(gomock.Any()).SetArg(0, []byte{}).Return(0, io.EOF)

		// receive content
		blobWriter := mocks.NewMockWriteCloser(ctrl)
		blobRepo.EXPECT().OpenBlobWriter("key").Return(blobWriter, nil)
		stream.EXPECT().Recv().Return(&pb.Entry{
			Content: []byte("encrypted content"),
		}, nil)
		stream.EXPECT().Recv().Return(nil, io.EOF)
		blobWriter.EXPECT().Write([]byte("encrypted content")).Return(16, nil)
		blobWriter.EXPECT().Close().Return(nil).MinTimes(1)

		// wrap in decrypt
		blobReader := mocks.NewMockReadCloser(ctrl)
		blobRepo.EXPECT().OpenBlobReader("key").Return(blobReader, true, nil)
		decryptReader := mocks.NewMockReadCloser(ctrl)
		crypt.EXPECT().Decrypt(blobReader).Return(decryptReader, nil)

		// close stream
		stream.EXPECT().CloseSend().Return(nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		notes, content, found, err := service.GetEntry(ctx, "key")
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, "notes", notes)
		require.NotNil(t, content)

		// read is forwarded to decrypt
		decryptReader.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, nil)
		decryptReader.EXPECT().Read(gomock.Any()).SetArg(0, []byte{}).Return(0, io.EOF)
		data, err := io.ReadAll(content)
		require.NoError(t, err)
		require.Equal(t, "content", string(data))

		// close is forwarded to blob
		blobReader.EXPECT().Close().Return(nil)
		require.NoError(t, content.Close())
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)
		stream := mocks.NewMockServerStreamingClient[pb.Entry](ctrl)

		// send request
		client.EXPECT().GetEntry(ctx, &pb.GetEntryRequest{
			Key: "key",
		}).Return(stream, nil)

		stream.EXPECT().Recv().Return(nil, status.Error(codes.NotFound, "not found"))

		// close stream
		stream.EXPECT().CloseSend().Return(nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		notes, content, found, err := service.GetEntry(ctx, "key")
		require.NoError(t, err)
		require.False(t, found)
		require.Empty(t, notes)
		require.Nil(t, content)
	})
}

func TestService_Delete(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)

		client.EXPECT().DeleteEntry(ctx, &pb.DeleteEntryRequest{
			Key: "key",
		}).Return(nil, nil)

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.DeleteEntry(ctx, "key")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := mocks.NewMockCrypt(ctrl)
		blobRepo := mocks.NewMockBlobRepository(ctrl)
		client := mocks.NewMockEntryServiceClient(ctrl)

		client.EXPECT().DeleteEntry(ctx, &pb.DeleteEntryRequest{
			Key: "key",
		}).Return(nil, errors.New("error"))

		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.DeleteEntry(ctx, "key")
		require.Error(t, err)
	})
}
