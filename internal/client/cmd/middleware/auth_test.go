package middleware_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/middleware"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestEnsureLoggedIn(t *testing.T) {
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

	t.Run("logged in", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(true, nil)

		mw := middleware.EnsureLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.NoError(t, err)
	})

	t.Run("not logged in", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(false, nil)

		mw := middleware.EnsureLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error getting auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("error getting auth service"))

		mw := middleware.EnsureLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error checking if logged in", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(false, errors.New("error checking if logged in"))

		mw := middleware.EnsureLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("wrapped function error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(true, nil)

		mw := middleware.EnsureLoggedIn(container)
		wrapped := mw(func(cmd *cobra.Command, args []string) error {
			return errors.New("wrapped function error")
		})
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})
}

func TestEnsureNotLoggedIn(t *testing.T) {
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

	t.Run("not logged in", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(false, nil)

		mw := middleware.EnsureNotLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.NoError(t, err)
	})

	t.Run("logged in", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(true, nil)

		mw := middleware.EnsureNotLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error getting auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("error getting auth service"))

		mw := middleware.EnsureNotLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("error checking if logged in", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(false, errors.New("error checking if logged in"))

		mw := middleware.EnsureNotLoggedIn(container)
		wrapped := mw(nil)
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})

	t.Run("wrapped function error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		auth := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(auth, nil).AnyTimes()
		auth.EXPECT().IsLoggedIn(ctx).Return(false, nil)

		mw := middleware.EnsureNotLoggedIn(container)
		wrapped := mw(func(cmd *cobra.Command, args []string) error {
			return errors.New("wrapped function error")
		})
		err := wrapped(newTestCommand(), []string{})
		require.Error(t, err)
	})
}
