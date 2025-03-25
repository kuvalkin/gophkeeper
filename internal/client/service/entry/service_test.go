package entry_test

import (
	io "io"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/client/service/entry"
	pb "github.com/kuvalkin/gophkeeper/internal/proto/entry/v1"
	"github.com/kuvalkin/gophkeeper/internal/support/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const chunkSize = 1024

func TestService_Set(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		crypt := NewMockCrypt(ctrl)
		blobRepo := NewMockRepository(ctrl)
		blobWriter := NewMockWriteCloser(ctrl)
		encryptWriter := NewMockWriteCloser(ctrl)
		rawContent := NewMockReadCloser(ctrl)

		// encrypt blob
		blobRepo.EXPECT().Writer("key").Return(blobWriter, nil)
		crypt.EXPECT().Encrypt(blobWriter).Return(encryptWriter, nil)
		rawContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("content")).Return(7, io.EOF)
		rawContent.EXPECT().Close().Return(nil)
		encryptWriter.EXPECT().Write([]byte("content")).Return(7, nil)
		encryptWriter.EXPECT().Close().Return(nil)
		blobWriter.EXPECT().Close().Return(nil)

		// encrypt notes
		notesEncrypter := NewMockWriteCloser(ctrl)
		crypt.EXPECT().Encrypt(gomock.Any()).Return(notesEncrypter, nil)
		notesEncrypter.EXPECT().Write([]byte("notes")).Return(5, nil)
		notesEncrypter.EXPECT().Close().Return(nil)

		encryptedContent := NewMockReadCloser(ctrl)
		client := NewMockEntryServiceClient(ctrl)
		stream := NewMockBidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse](ctrl)

		// send metadata
		blobRepo.EXPECT().Reader("key").Return(encryptedContent, true, nil)
		client.EXPECT().SetEntry(ctx).Return(stream, nil)
		stream.EXPECT().Send(&pb.SetEntryRequest{
			Entry: &pb.Entry{
				Key: "key",
				Name: "name",
				Notes: nil, // encrypter will return nil bytes
				Content: nil,
			},
		}).Return(nil)
		stream.EXPECT().Recv().Return(&pb.SetEntryResponse{}, nil)

		// upload
		encryptedContent.EXPECT().Read(gomock.Any()).SetArg(0, []byte("encrypted")).Return(9, io.EOF)
		encryptedContent.EXPECT().Close().Return(nil)
		stream.EXPECT().Send(&pb.SetEntryRequest{
			Entry: &pb.Entry{
				Content: []byte("encrypted"),
			},
		}).Return(nil)
		stream.EXPECT().CloseSend().Return(nil).MinTimes(1)
		stream.EXPECT().Recv().Return(nil, io.EOF)


		service := entry.New(crypt, client, blobRepo, chunkSize)
		err := service.Set(ctx, "key", "name", "notes", rawContent, nil)
		require.NoError(t, err)
	})
}