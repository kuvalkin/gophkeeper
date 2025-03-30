package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	auth "github.com/kuvalkin/gophkeeper/internal/client/service/auth"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	prompts "github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestRegister(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestRegisterCommand := func(container container.Container) *cobra.Command {
		regCmd := newRegisterCommand(container)

		regCmd.SetContext(ctx)
		regCmd.SetOut(bytes.NewBuffer(nil))
		regCmd.SetErr(bytes.NewBuffer(nil))
		regCmd.SetArgs([]string{})

		return regCmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		authService.EXPECT().Register(ctx, "login", "password").Return(nil)

		regCmd := newTestRegisterCommand(container)

		err := regCmd.Execute()
		require.NoError(t, err)
	})

	t.Run("prompt canceled", func(t *testing.T) {
		t.Run("login", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)
			authService := mocks.NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			regCmd := newTestRegisterCommand(container)

			err := regCmd.Execute()
			require.NoError(t, err)
		})

		t.Run("password", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)
			authService := mocks.NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			regCmd := newTestRegisterCommand(container)

			err := regCmd.Execute()
			require.NoError(t, err)
		})
	})

	t.Run("login taken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		authService.EXPECT().Register(ctx, "login", "password").Return(auth.ErrLoginTaken)

		regCmd := newTestRegisterCommand(container)

		err := regCmd.Execute()
		require.Error(t, err)
		require.ErrorContains(t, err, "user with this login already exists, pick another one")
	})

	t.Run("cant get prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("error")).AnyTimes()

		regCmd := newTestRegisterCommand(container)

		err := regCmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("error")).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		regCmd := newTestRegisterCommand(container)

		err := regCmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant register", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		authService.EXPECT().Register(ctx, "login", "password").Return(errors.New("error"))

		regCmd := newTestRegisterCommand(container)

		err := regCmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt err", func(t *testing.T) {
		t.Run("login", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)
			authService := mocks.NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			regCmd := newTestRegisterCommand(container)

			err := regCmd.Execute()
			require.Error(t, err)
		})

		t.Run("password", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)
			authService := mocks.NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			regCmd := newTestRegisterCommand(container)

			err := regCmd.Execute()
			require.Error(t, err)
		})
	})
}
