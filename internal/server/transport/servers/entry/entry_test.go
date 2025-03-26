package entry_test

import (
	"errors"
	"io"
	"testing"

	pb "github.com/kuvalkin/gophkeeper/internal/proto/entry/v1"
	entryService "github.com/kuvalkin/gophkeeper/internal/server/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/servers/entry"
	"github.com/kuvalkin/gophkeeper/internal/support/test"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_GetEntry(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	ctxWithToken := auth.SetTokenInfo(ctx, user.TokenInfo{
		UserID: "user",
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)
		content := NewMockReadCloser(ctrl)

		stream.EXPECT().Context().Return(ctxWithToken).AnyTimes()

		service.EXPECT().Get(ctxWithToken, "user", "key").Return(
			entryService.Metadata{
				Key: "key",
				Name: "name",
				Notes: []byte("encrypted notes"),
			},
			content,
			true,
			nil,
		)

		// send metadata
		stream.EXPECT().Send(&pb.Entry{
			Key: "key",
			Name: "name",
			Notes: []byte("encrypted notes"),
		}).Return(nil)

		// stream content
		content.EXPECT().Read(gomock.Any()).SetArg(0, []byte("encrypted content")).Return(17, nil)
		stream.EXPECT().Send(&pb.Entry{
			Content: []byte("encrypted content"),
		}).Return(nil)
		content.EXPECT().Read(gomock.Any()).SetArg(0, []byte{}).Return(0, io.EOF)
		content.EXPECT().Close().Return(nil)

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.NoError(t, err)
	})

	t.Run("no token info", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)

		stream.EXPECT().Context().Return(ctx).AnyTimes()

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.Error(t, err)
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("cant get entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)

		stream.EXPECT().Context().Return(ctxWithToken).AnyTimes()

		service.EXPECT().Get(ctxWithToken, "user", "key").Return(
			entryService.Metadata{},
			nil,
			false,
			errors.New("cant get entry"),
		)

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)

		stream.EXPECT().Context().Return(ctxWithToken).AnyTimes()

		service.EXPECT().Get(ctxWithToken, "user", "key").Return(
			entryService.Metadata{},
			nil,
			false,
			nil,
		)

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.Error(t, err)
		require.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("cant send metadata", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)
		content := NewMockReadCloser(ctrl)

		stream.EXPECT().Context().Return(ctxWithToken).AnyTimes()

		service.EXPECT().Get(ctxWithToken, "user", "key").Return(
			entryService.Metadata{
				Key: "key",
				Name: "name",
				Notes: []byte("encrypted notes"),
			},
			content,
			true,
			nil,
		)

		// send metadata
		stream.EXPECT().Send(&pb.Entry{
			Key: "key",
			Name: "name",
			Notes: []byte("encrypted notes"),
		}).Return(errors.New("cant send metadata"))

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("cant read entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)
		content := NewMockReadCloser(ctrl)

		stream.EXPECT().Context().Return(ctxWithToken).AnyTimes()

		service.EXPECT().Get(ctxWithToken, "user", "key").Return(
			entryService.Metadata{
				Key: "key",
				Name: "name",
				Notes: []byte("encrypted notes"),
			},
			content,
			true,
			nil,
		)

		// send metadata
		stream.EXPECT().Send(&pb.Entry{
			Key: "key",
			Name: "name",
			Notes: []byte("encrypted notes"),
		}).Return(nil)

		// stream content
		content.EXPECT().Read(gomock.Any()).SetArg(0, []byte{}).Return(0, errors.New("cant read entry"))
		content.EXPECT().Close().Return(nil)

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("cant send entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		stream := NewMockServerStreamingServer[pb.Entry](ctrl)
		service := NewMockService(ctrl)
		content := NewMockReadCloser(ctrl)

		stream.EXPECT().Context().Return(ctxWithToken).AnyTimes()

		service.EXPECT().Get(ctxWithToken, "user", "key").Return(
			entryService.Metadata{
				Key: "key",
				Name: "name",
				Notes: []byte("encrypted notes"),
			},
			content,
			true,
			nil,
		)

		// send metadata
		stream.EXPECT().Send(&pb.Entry{
			Key: "key",
			Name: "name",
			Notes: []byte("encrypted notes"),
		}).Return(nil)

		// stream content
		content.EXPECT().Read(gomock.Any()).SetArg(0, []byte("encrypted content")).Return(17, nil)
		stream.EXPECT().Send(&pb.Entry{
			Content: []byte("encrypted content"),
		}).Return(errors.New("cant send entry"))
		content.EXPECT().Close().Return(nil)

		s := entry.New(service, 1024)
		err := s.GetEntry(&pb.GetEntryRequest{
			Key: "key",
		}, stream)
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})
}