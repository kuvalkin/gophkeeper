package middleware_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/middleware"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEnsureSecretSet(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestCommand := func() *cobra.Command {
		cmd := &cobra.Command{}
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetArgs([]string{})
		cmd.SetContext(ctx)
		return cmd
	}

	t.Run("secret set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		secret := mocks.NewMockSecretService(ctrl)

		container.EXPECT().GetSecretService(ctx).Return(secret, nil).AnyTimes()
		secret.EXPECT().GetSecret(ctx).Return("", true, nil)

		mw := middleware.EnsureSecretSet(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.NoError(t, err)
	})

	t.Run("secret not set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		secret := mocks.NewMockSecretService(ctrl)

		container.EXPECT().GetSecretService(ctx).Return(secret, nil).AnyTimes()
		secret.EXPECT().GetSecret(ctx).Return("", false, nil)

		mw := middleware.EnsureSecretSet(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error getting secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		secret := mocks.NewMockSecretService(ctrl)

		container.EXPECT().GetSecretService(ctx).Return(secret, nil).AnyTimes()
		secret.EXPECT().GetSecret(ctx).Return("", false, errors.New("error getting secret"))

		mw := middleware.EnsureSecretSet(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error getting secret service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetSecretService(ctx).Return(nil, errors.New("error getting secret service"))

		mw := middleware.EnsureSecretSet(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error wrapping function", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		secret := mocks.NewMockSecretService(ctrl)

		container.EXPECT().GetSecretService(ctx).Return(secret, nil).AnyTimes()
		secret.EXPECT().GetSecret(ctx).Return("", true, nil)

		mw := middleware.EnsureSecretSet(container)
		wrapped := mw(func(cmd *cobra.Command, args []string) error {
			return errors.New("error wrapping function")
		})
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})
}