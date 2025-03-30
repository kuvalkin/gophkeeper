package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	clientUtils "github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

type testCtxKey string

func TestDeleteLogin(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestDeleteLoginCommand := func(container container.Container, name string) *cobra.Command {
		cmd := newDeleteCommand(container)
		// gkeep delete login <name>
		cmd.SetArgs([]string{"login", name})
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetContext(ctx)
		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().DeleteEntry(authCtx, clientUtils.GetEntryKey("login", "name")).Return(nil)

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("confirm returns false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(false)

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("cant get prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("err")).AnyTimes()

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("empty name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		cmd := newTestDeleteLoginCommand(container, "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error getting entry service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(nil, errors.New("err")).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error getting auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("err")).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error adding authorization header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		authService.EXPECT().AddAuthorizationHeader(ctx).Return(nil, errors.New("err"))

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("error deleting entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().DeleteEntry(authCtx, clientUtils.GetEntryKey("login", "name")).Return(errors.New("err"))

		cmd := newTestDeleteLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})
}

func TestDeleteFile(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestDeleteFileCommand := func(container container.Container, name string) *cobra.Command {
		cmd := newDeleteCommand(container)
		// gkeep delete file <name>
		cmd.SetArgs([]string{"file", name})
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetContext(ctx)
		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().DeleteEntry(authCtx, clientUtils.GetEntryKey("file", "name")).Return(nil)

		cmd := newTestDeleteFileCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
	})
}

func TestDeleteCard(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestDeleteCardCommand := func(container container.Container, name string) *cobra.Command {
		cmd := newDeleteCommand(container)
		// gkeep delete card <name>
		cmd.SetArgs([]string{"card", name})
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetContext(ctx)
		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().DeleteEntry(authCtx, clientUtils.GetEntryKey("card", "name")).Return(nil)

		cmd := newTestDeleteCardCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
	})
}

func TestDeleteText(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestDeleteTextCommand := func(container container.Container, name string) *cobra.Command {
		cmd := newDeleteCommand(container)
		// gkeep delete test <name>
		cmd.SetArgs([]string{"text", name})
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetContext(ctx)
		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)
		prompter := mocks.NewMockPrompter(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

		prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().DeleteEntry(authCtx, clientUtils.GetEntryKey("text", "name")).Return(nil)

		cmd := newTestDeleteTextCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
	})
}
