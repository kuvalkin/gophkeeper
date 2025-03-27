package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/client/service/auth"
	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	"github.com/kuvalkin/gophkeeper/internal/support/test"
)

func TestService_Register(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Register(ctx, &pbAuth.RegisterRequest{Login: "login", Password: "password"}).Return(&pbAuth.RegisterResponse{Token: "token"}, nil)
		repo.EXPECT().Set(ctx, "token").Return(nil)

		err := service.Register(ctx, "login", "password")
		require.NoError(t, err)
	})

	t.Run("client error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Register(ctx, &pbAuth.RegisterRequest{Login: "login", Password: "password"}).Return(nil, errors.New("error"))

		err := service.Register(ctx, "login", "password")
		require.Error(t, err)
	})

	t.Run("login taken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Register(ctx, &pbAuth.RegisterRequest{Login: "login", Password: "password"}).Return(nil, status.Error(codes.AlreadyExists, "error"))

		err := service.Register(ctx, "login", "password")
		require.ErrorIs(t, err, auth.ErrLoginTaken)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Register(ctx, &pbAuth.RegisterRequest{Login: "login", Password: "password"}).Return(&pbAuth.RegisterResponse{Token: "token"}, nil)
		repo.EXPECT().Set(ctx, "token").Return(errors.New("error"))

		err := service.Register(ctx, "login", "password")
		require.Error(t, err)
	})
}

func TestService_Login(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Login(ctx, &pbAuth.LoginRequest{Login: "login", Password: "password"}).Return(&pbAuth.LoginResponse{Token: "token"}, nil)
		repo.EXPECT().Set(ctx, "token").Return(nil)

		err := service.Login(ctx, "login", "password")
		require.NoError(t, err)
	})

	t.Run("client error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Login(ctx, &pbAuth.LoginRequest{Login: "login", Password: "password"}).Return(nil, errors.New("error"))

		err := service.Login(ctx, "login", "password")
		require.Error(t, err)
	})

	t.Run("invalid pair", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Login(ctx, &pbAuth.LoginRequest{Login: "login", Password: "password"}).Return(nil, status.Error(codes.Unauthenticated, "error"))

		err := service.Login(ctx, "login", "password")
		require.ErrorIs(t, err, auth.ErrInvalidPair)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		client.EXPECT().Login(ctx, &pbAuth.LoginRequest{Login: "login", Password: "password"}).Return(&pbAuth.LoginResponse{Token: "token"}, nil)
		repo.EXPECT().Set(ctx, "token").Return(errors.New("error"))

		err := service.Login(ctx, "login", "password")
		require.Error(t, err)
	})
}

func TestService_IsLoggedIn(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Get(ctx).Return("token", true, nil)

		ok, err := service.IsLoggedIn(ctx)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Get(ctx).Return("", false, errors.New("error"))

		ok, err := service.IsLoggedIn(ctx)
		require.Error(t, err)
		require.False(t, ok)
	})
}

func TestService_Logout(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Delete(ctx).Return(nil)

		err := service.Logout(ctx)
		require.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Delete(ctx).Return(errors.New("error"))

		err := service.Logout(ctx)
		require.Error(t, err)
	})
}

func TestService_AddAuthorizationHeader(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Get(ctx).Return("token", true, nil)

		ctxWithToken, err := service.AddAuthorizationHeader(ctx)
		require.NoError(t, err)
		require.NotNil(t, ctxWithToken)
	})

	t.Run("no token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Get(ctx).Return("", false, nil)

		ctxWithToken, err := service.AddAuthorizationHeader(ctx)
		require.Error(t, err)
		require.Nil(t, ctxWithToken)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockAuthServiceClient(ctrl)
		repo := NewMockRepository(ctrl)

		service := auth.New(client, repo)

		repo.EXPECT().Get(ctx).Return("", false, errors.New("error"))

		ctxWithToken, err := service.AddAuthorizationHeader(ctx)
		require.Error(t, err)
		require.Nil(t, ctxWithToken)
	})
}
