package auth_test

import (
	"testing"

	authMW "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userService "github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/servers/auth"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
	pb "github.com/kuvalkin/gophkeeper/pkg/proto/auth/v1"
)

func TestAuth_AuthFuncOverride(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	server := auth.New(nil)
	override, ok := server.(authMW.ServiceAuthFuncOverride)
	require.True(t, ok)
	_, err := override.AuthFuncOverride(ctx, "")
	require.NoError(t, err)
}

func TestAuth_Register(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().RegisterUser(ctx, "login", "password").Return(nil)
		service.EXPECT().LoginUser(ctx, "login", "password").Return("token", nil)

		server := auth.New(service)
		resp, err := server.Register(ctx, &pb.RegisterRequest{Login: "login", Password: "password"})
		require.NoError(t, err)
		require.Equal(t, "token", resp.Token)
	})

	t.Run("login taken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().RegisterUser(ctx, "login", "password").Return(userService.ErrLoginTaken)

		server := auth.New(service)
		_, err := server.Register(ctx, &pb.RegisterRequest{Login: "login", Password: "password"})
		require.Error(t, err)
		require.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().RegisterUser(ctx, "login", "password").Return(userService.ErrInternal)

		server := auth.New(service)
		_, err := server.Register(ctx, &pb.RegisterRequest{Login: "login", Password: "password"})
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("invalid login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().RegisterUser(ctx, "login", "password").Return(userService.ErrInvalidLogin)

		server := auth.New(service)
		_, err := server.Register(ctx, &pb.RegisterRequest{Login: "login", Password: "password"})
		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("login error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().RegisterUser(ctx, "login", "password").Return(nil)
		service.EXPECT().LoginUser(ctx, "login", "password").Return("", userService.ErrInternal)

		server := auth.New(service)
		_, err := server.Register(ctx, &pb.RegisterRequest{Login: "login", Password: "password"})
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestAuth_Login(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().LoginUser(ctx, "login", "password").Return("token", nil)

		server := auth.New(service)
		resp, err := server.Login(ctx, &pb.LoginRequest{Login: "login", Password: "password"})
		require.NoError(t, err)
		require.Equal(t, "token", resp.Token)
	})

	t.Run("invalid pair", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().LoginUser(ctx, "login", "password").Return("", userService.ErrInvalidPair)

		server := auth.New(service)
		_, err := server.Login(ctx, &pb.LoginRequest{Login: "login", Password: "password"})
		require.Error(t, err)
		require.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		service := mocks.NewMockUserService(ctrl)
		service.EXPECT().LoginUser(ctx, "login", "password").Return("", userService.ErrInternal)

		server := auth.New(service)
		_, err := server.Login(ctx, &pb.LoginRequest{Login: "login", Password: "password"})
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
	})
}
