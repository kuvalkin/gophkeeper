package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func Test_TokenInfo(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("no token info", func(t *testing.T) {
		info, ok := auth.GetTokenInfo(ctx)
		require.False(t, ok)
		require.Empty(t, info)

	})

	tokenInfo := user.TokenInfo{
		UserID: "user",
	}

	t.Run("set token info", func(t *testing.T) {
		ctxWithToken := auth.SetTokenInfo(ctx, tokenInfo)

		info, ok := auth.GetTokenInfo(ctxWithToken)
		require.True(t, ok)
		require.Equal(t, tokenInfo, info)
	})
}

func Test_AuthFunc(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("no token", func(t *testing.T) {
		authFunc := auth.NewAuthFunc(nil)
		_, err := authFunc(ctx)
		require.Error(t, err)
	})

	t.Run("valid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := NewMockService(ctrl)
		authFunc := auth.NewAuthFunc(service)

		ctxWithToken := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
			"authorization": "bearer token",
		}))

		service.EXPECT().ParseToken(ctxWithToken, "token").Return(&user.TokenInfo{
			UserID: "user",
		}, nil)

		newCtx, err := authFunc(ctxWithToken)
		require.NoError(t, err)
		info, ok := auth.GetTokenInfo(newCtx)
		require.True(t, ok)
		require.Equal(t, user.TokenInfo{
			UserID: "user",
		}, info)
	})

	t.Run("invalid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := NewMockService(ctrl)
		authFunc := auth.NewAuthFunc(service)

		ctxWithToken := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
			"authorization": "bearer token",
		}))

		service.EXPECT().ParseToken(ctxWithToken, "token").Return(nil, user.ErrInvalidToken)

		newCtx, err := authFunc(ctxWithToken)
		require.Error(t, err)
		require.Equal(t, codes.Unauthenticated, status.Code(err))
		require.Nil(t, newCtx)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := NewMockService(ctrl)
		authFunc := auth.NewAuthFunc(service)

		ctxWithToken := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
			"authorization": "bearer token",
		}))

		service.EXPECT().ParseToken(ctxWithToken, "token").Return(nil, user.ErrInternal)

		newCtx, err := authFunc(ctxWithToken)
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
		require.Nil(t, newCtx)
	})
}
