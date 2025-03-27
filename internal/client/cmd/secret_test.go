package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	prompts "github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	"github.com/kuvalkin/gophkeeper/internal/support/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestSecret(t *testing.T) {
	ctx, cancel := test.Context(t)
	defer cancel()

	newTestSecretCommand := func(container container.Container) *cobra.Command {
		secCmd := newSecretCommand(container)
		secCmd.SetContext(ctx)
		secCmd.SetOut(bytes.NewBuffer(nil))
		secCmd.SetErr(bytes.NewBuffer(nil))
		secCmd.SetArgs([]string{})
		return secCmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		service := NewMockSecretService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Get(ctx).Return("", false, nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("secret", nil)
		service.EXPECT().Set(ctx, "secret").Return(nil)

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("already set", func(t *testing.T) {
		t.Run("no overwrite", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			service := NewMockSecretService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

			service.EXPECT().Get(ctx).Return("secret", true, nil)
			prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(false)

			cmd := newTestSecretCommand(container)
			err := cmd.Execute()
			require.NoError(t, err)
		})

		t.Run("overwrite", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			service := NewMockSecretService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

			service.EXPECT().Get(ctx).Return("secret", true, nil)
			prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("new_secret", nil)
			service.EXPECT().Set(ctx, "new_secret").Return(nil)

			cmd := newTestSecretCommand(container)
			err := cmd.Execute()
			require.NoError(t, err)
		})
	})

	t.Run("prompt canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		service := NewMockSecretService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Get(ctx).Return("", false, nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("error getting service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		container.EXPECT().GetSecretService(ctx).Return(nil, errors.New("err"))

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error getting prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		service := NewMockSecretService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("error")).AnyTimes()
		container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error getting current secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		service := NewMockSecretService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Get(ctx).Return("", false, errors.New("err"))

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error setting secret", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		service := NewMockSecretService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Get(ctx).Return("", false, nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("secret", nil)
		service.EXPECT().Set(ctx, "secret").Return(errors.New("err"))

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt err", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		service := NewMockSecretService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetSecretService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Get(ctx).Return("", false, nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("err"))

		cmd := newTestSecretCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})
}