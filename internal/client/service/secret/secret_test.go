package secret_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/client/service/secret"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestService_Set(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("", false, nil)
		repo.EXPECT().SetSecret(ctx, "secret").Return(nil)

		err := service.SetSecret(ctx, "secret")
		require.NoError(t, err)
	})

	t.Run("already set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("secret", true, nil)

		err := service.SetSecret(ctx, "secret")
		require.Equal(t, secret.ErrAlreadySet, err)
	})

	t.Run("repo error on get", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("", false, errors.New("error"))

		err := service.SetSecret(ctx, "secret")
		require.Error(t, err)
	})

	t.Run("repo error on set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("", false, nil)
		repo.EXPECT().SetSecret(ctx, "secret").Return(errors.New("error"))

		err := service.SetSecret(ctx, "secret")
		require.Error(t, err)
	})
}

func TestService_Get(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("secret", true, nil)

		secret, exists, err := service.GetSecret(ctx)
		require.NoError(t, err)
		require.Equal(t, "secret", secret)
		require.True(t, exists)
	})

	t.Run("no secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("", false, nil)

		secret, exists, err := service.GetSecret(ctx)
		require.NoError(t, err)
		require.Empty(t, secret)
		require.False(t, exists)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockSecretRepository(ctrl)

		service := secret.New(repo)

		repo.EXPECT().GetSecret(ctx).Return("", false, errors.New("error"))

		secret, exists, err := service.GetSecret(ctx)
		require.Error(t, err)
		require.Empty(t, secret)
		require.False(t, exists)
	})
}
