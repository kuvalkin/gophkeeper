package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	prompts "github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestLogin(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestLoginCommand := func(container container.Container) *cobra.Command {
		cmd := newLoginCommand(container)

		cmd.SetContext(ctx)
		cmd.SetOut(bytes.NewBuffer(nil))
		cmd.SetErr(bytes.NewBuffer(nil))
		cmd.SetArgs([]string{})

		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		service := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(service, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		service.EXPECT().Login(ctx, "login", "password").Return(nil)

		cmd := newTestLoginCommand(container)

		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("prompt canceled", func(t *testing.T) {
		t.Run("login", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			cmd := newTestLoginCommand(container)

			err := cmd.Execute()
			require.NoError(t, err)
		})

		t.Run("password", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			cmd := newTestLoginCommand(container)

			err := cmd.Execute()
			require.NoError(t, err)
		})
	})

	t.Run("cant get prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("error")).AnyTimes()

		cmd := newTestLoginCommand(container)

		err := cmd.Execute()
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

		cmd := newTestLoginCommand(container)

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		service := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(service, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		service.EXPECT().Login(ctx, "login", "password").Return(errors.New("error"))

		cmd := newTestLoginCommand(container)

		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt err", func(t *testing.T) {
		t.Run("login", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			cmd := newTestLoginCommand(container)

			err := cmd.Execute()
			require.Error(t, err)
		})

		t.Run("password", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := mocks.NewMockContainer(ctrl)
			prompter := mocks.NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			cmd := newTestLoginCommand(container)

			err := cmd.Execute()
			require.Error(t, err)
		})
	})
}
