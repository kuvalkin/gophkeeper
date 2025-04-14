package user_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/server/service/user"
	"github.com/kuvalkin/gophkeeper/internal/server/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

var defaultOptions = user.Options{
	TokenSecret:           []byte("secret"),
	PasswordSalt:          "salt",
	TokenExpirationPeriod: time.Hour,
}

func TestService_Register(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().AddUser(ctx, "login", "7a37b85c8918eac19a9089c0fa5a2ab4dce3f90528dcdeec108b23ddf3607b99").Return(nil)

		s := user.NewService(repo, defaultOptions)
		err := s.RegisterUser(ctx, "login", "password")
		require.NoError(t, err)
	})

	t.Run("login taken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().AddUser(ctx, gomock.Any(), gomock.Any()).Return(user.ErrLoginNotUnique)

		s := user.NewService(repo, defaultOptions)
		err := s.RegisterUser(ctx, "login", "password")
		require.ErrorIs(t, err, user.ErrLoginTaken)
	})

	t.Run("empty login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		s := user.NewService(repo, defaultOptions)
		err := s.RegisterUser(ctx, "", "password")
		require.ErrorIs(t, err, user.ErrInvalidLogin)
	})

	t.Run("repo returns unknown error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().AddUser(ctx, gomock.Any(), gomock.Any()).Return(errors.New("query failed"))

		s := user.NewService(repo, defaultOptions)
		err := s.RegisterUser(ctx, "login", "password")
		require.ErrorIs(t, err, user.ErrInternal)
	})
}

func TestService_Login(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().FindUser(ctx, "login").Return(user.UserInfo{
			ID:           uuid.New().String(),
			PasswordHash: "7a37b85c8918eac19a9089c0fa5a2ab4dce3f90528dcdeec108b23ddf3607b99",
		}, true, nil)

		s := user.NewService(repo, defaultOptions)
		token, err := s.LoginUser(ctx, "login", "password")
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().FindUser(ctx, "login").Return(user.UserInfo{}, false, nil)

		s := user.NewService(repo, defaultOptions)
		token, err := s.LoginUser(ctx, "login", "password")
		require.ErrorIs(t, err, user.ErrInvalidPair)
		require.Empty(t, token)
	})

	t.Run("repo return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().FindUser(ctx, "login").Return(user.UserInfo{}, false, errors.New("query failed"))

		s := user.NewService(repo, defaultOptions)
		token, err := s.LoginUser(ctx, "login", "password")
		require.ErrorIs(t, err, user.ErrInternal)
		require.Empty(t, token)
	})

	t.Run("password doesn't match", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockUserRepository(ctrl)

		repo.EXPECT().FindUser(ctx, "login").Return(user.UserInfo{
			ID:           uuid.New().String(),
			PasswordHash: "its password hash",
		}, true, nil)

		s := user.NewService(repo, defaultOptions)
		token, err := s.LoginUser(ctx, "login", "password")
		require.ErrorIs(t, err, user.ErrInvalidPair)
		require.Empty(t, token)
	})
}

func TestService_ParseToken(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		now := time.Now()

		userID := uuid.New().String()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultOptions.TokenExpirationPeriod)),
		})

		tokenString, err := token.SignedString(defaultOptions.TokenSecret)
		require.NoError(t, err)

		s := user.NewService(nil, defaultOptions)
		info, err := s.ParseAuthToken(ctx, tokenString)
		require.NoError(t, err)
		require.Equal(t, userID, info.UserID)
	})

	t.Run("invalid string", func(t *testing.T) {
		s := user.NewService(nil, defaultOptions)
		info, err := s.ParseAuthToken(ctx, "its definitely a valid token, trust me")
		require.ErrorIs(t, err, user.ErrInvalidToken)
		require.Nil(t, info)
	})

	t.Run("different signing method", func(t *testing.T) {
		now := time.Now()

		userID := uuid.New().String()

		token := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultOptions.TokenExpirationPeriod)),
		})

		tokenString, err := token.SignedString(defaultOptions.TokenSecret)
		require.NoError(t, err)

		s := user.NewService(nil, defaultOptions)
		info, err := s.ParseAuthToken(ctx, tokenString)
		require.ErrorIs(t, err, user.ErrInvalidToken)
		require.Nil(t, info)
	})

	t.Run("token expired", func(t *testing.T) {
		now := time.Now().Add(-24 * time.Hour)

		userID := uuid.New().String()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultOptions.TokenExpirationPeriod)),
		})

		tokenString, err := token.SignedString(defaultOptions.TokenSecret)
		require.NoError(t, err)

		s := user.NewService(nil, defaultOptions)
		info, err := s.ParseAuthToken(ctx, tokenString)
		require.ErrorIs(t, err, user.ErrInvalidToken)
		require.Nil(t, info)
	})
}
