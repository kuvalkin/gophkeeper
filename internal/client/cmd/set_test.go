package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	clientUtils "github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/client/tui/prompts"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

//go:generate mockgen -destination=./container_mock_test.go -package=cmd github.com/kuvalkin/gophkeeper/internal/client/service/container Container
//go:generate mockgen -destination=./auth_service_mock_test.go -package=cmd -mock_names Service=MockAuthService github.com/kuvalkin/gophkeeper/internal/client/service/auth Service
//go:generate mockgen -destination=./entry_service_mock_test.go -package=cmd -mock_names Service=MockEntryService github.com/kuvalkin/gophkeeper/internal/client/service/entry Service
//go:generate mockgen -destination=./prompter_mock_test.go -package=cmd github.com/kuvalkin/gophkeeper/internal/client/tui/prompts Prompter

func TestSetLogin(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestSetLoginCommand := func(container container.Container, name string, notes string) *cobra.Command {
		cmd := newSetCommand(container)
		// gkeep set login {name} --notes {notes}
		cmd.SetArgs([]string{"login", name, "--notes", notes})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("login", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(nil)

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := newTestSetLoginCommand(container, "", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt cancel", func(t *testing.T) {
		t.Run("login", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			cmd := newTestSetLoginCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})
		t.Run("password", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			cmd := newTestSetLoginCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})
	})

	t.Run("prompt error", func(t *testing.T) {
		t.Run("login", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			cmd := newTestSetLoginCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
		t.Run("password", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			cmd := newTestSetLoginCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
	})

	t.Run("cant get prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("error"))

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("error")).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant add auth header", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(nil, errors.New("error"))

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get entry service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(nil, errors.New("error"))

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)
		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant set entry", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		authService := NewMockAuthService(ctrl)
		entryService := NewMockEntryService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("login", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(errors.New("error"))

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("duplicate", func(t *testing.T) {
		t.Run("overwrite", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			authService := NewMockAuthService(ctrl)
			entryService := NewMockEntryService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

			authCtx := context.WithValue(ctx, "test", "test")
			authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

			entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("login", "name"), "name", "notes", gomock.Any(), gomock.Any()).Do(
				func(ctx context.Context, key string, name string, notes string, content io.ReadCloser, onOverwrite func() bool) {
					require.NotNil(t, onOverwrite)
					require.True(t, onOverwrite())
				},
			).Return(nil)
			prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(true)

			cmd := newTestSetLoginCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})

		t.Run("no overwrite", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			authService := NewMockAuthService(ctrl)
			entryService := NewMockEntryService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
			prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

			authCtx := context.WithValue(ctx, "test", "test")
			authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

			entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("login", "name"), "name", "notes", gomock.Any(), gomock.Any()).Do(
				func(ctx context.Context, key string, name string, notes string, content io.ReadCloser, onOverwrite func() bool) {
					require.NotNil(t, onOverwrite)
					require.False(t, onOverwrite())
				},
			).Return(errors.New("duplicate"))
			prompter.EXPECT().Confirm(ctx, gomock.Any()).Return(false)

			cmd := newTestSetLoginCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
	})

	t.Run("empty login", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("password", nil)

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("login", nil)
		prompter.EXPECT().AskPassword(ctx, gomock.Any(), gomock.Any()).Return("", nil)

		cmd := newTestSetLoginCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})
}

func TestSetFile(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestSetFileCommand := func(container container.Container, name string, path string, notes string) *cobra.Command {
		cmd := newSetCommand(container)
		// gkeep set file {name} {path} --notes {notes}
		cmd.SetArgs([]string{"file", name, path, "--notes", notes})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("file", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(nil)

		file, err := os.CreateTemp("", "set-file-test-*")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		_, err = io.WriteString(file, "text")
		require.NoError(t, err)
		require.NoError(t, file.Close())

		cmd := newTestSetFileCommand(container, "name", file.Name(), "notes")
		err = cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := newTestSetFileCommand(container, "", "path", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("no path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := newTestSetFileCommand(container, "name", "", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := newTestSetFileCommand(container, "name", "path", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("store error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("file", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(errors.New("error"))

		file, err := os.CreateTemp("", "set-file-test-*")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		_, err = io.WriteString(file, "text")
		require.NoError(t, err)
		require.NoError(t, file.Close())

		cmd := newTestSetFileCommand(container, "name", file.Name(), "notes")
		err = cmd.Execute()
		require.Error(t, err)
	})
}

func TestSetCard(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestSetCardCommand := func(container container.Container, name string, notes string) *cobra.Command {
		cmd := newSetCommand(container)
		// gkeep set card {name} --notes {notes}
		cmd.SetArgs([]string{"card", name, "--notes", notes})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(11, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(111, nil)

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("card", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(nil)

		cmd := newTestSetCardCommand(container, "name", "notes")
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := newTestSetCardCommand(container, "", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt cancel", func(t *testing.T) {
		t.Run("number", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})

		t.Run("holder name", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})
		t.Run("expiration year", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(0, prompts.ErrCanceled)

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})
		t.Run("expiration month", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(0, prompts.ErrCanceled)

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})
		t.Run("cvv", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(11, nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(0, prompts.ErrCanceled)

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.NoError(t, err)
		})
	})

	t.Run("cant get prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("error"))

		cmd := newTestSetCardCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt error", func(t *testing.T) {
		t.Run("number", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
		t.Run("holder name", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
		t.Run("expiration year", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(0, errors.New("error"))

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
		t.Run("expiration month", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(0, errors.New("error"))

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
		t.Run("cvv", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			container := NewMockContainer(ctrl)
			prompter := NewMockPrompter(ctrl)
			entryService := NewMockEntryService(ctrl)
			authService := NewMockAuthService(ctrl)

			container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
			container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
			container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
			prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(11, nil)
			prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(0, errors.New("error"))

			cmd := newTestSetCardCommand(container, "name", "notes")
			err := cmd.Execute()
			require.Error(t, err)
		})
	})

	t.Run("store error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(11, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(111, nil)

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("card", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(errors.New("error"))

		cmd := newTestSetCardCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("invalid number", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("1234567890123456", nil)
		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(2030, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(11, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(111, nil)

		cmd := newTestSetCardCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("invalid expiration year", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("4111111111111111", nil)
		prompter.EXPECT().AskString(ctx, gomock.Any(), gomock.Any()).Return("John Doe", nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(9999, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(11, nil)
		prompter.EXPECT().AskInt(ctx, gomock.Any(), gomock.Any()).Return(111, nil)

		cmd := newTestSetCardCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})
}

func TestSetText(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestSetTextCommand := func(container container.Container, name string, notes string) *cobra.Command {
		cmd := newSetCommand(container)
		// gkeep set text {name} --notes {notes}
		cmd.SetArgs([]string{"text", name, "--notes", notes})
		cmd.SetOut(io.Discard)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskText(ctx, gomock.Any(), gomock.Any()).Return("text", nil)

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("text", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(nil)

		cmd := newTestSetTextCommand(container, "name", "notes")
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := newTestSetTextCommand(container, "", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get prompter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(nil, errors.New("error"))

		cmd := newTestSetTextCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskText(ctx, gomock.Any(), gomock.Any()).Return("", errors.New("error"))

		cmd := newTestSetTextCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("prompt cancel", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskText(ctx, gomock.Any(), gomock.Any()).Return("", prompts.ErrCanceled)

		cmd := newTestSetTextCommand(container, "name", "notes")
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("store error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskText(ctx, gomock.Any(), gomock.Any()).Return("text", nil)

		authCtx := context.WithValue(ctx, "test", "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().SetEntry(authCtx, clientUtils.GetEntryKey("text", "name"), "name", "notes", gomock.Any(), gomock.Any()).Return(errors.New("error"))

		cmd := newTestSetTextCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("empty text", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		prompter := NewMockPrompter(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetPrompter(ctx).Return(prompter, nil).AnyTimes()
		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		prompter.EXPECT().AskText(ctx, gomock.Any(), gomock.Any()).Return("", nil)

		cmd := newTestSetTextCommand(container, "name", "notes")
		err := cmd.Execute()
		require.Error(t, err)
	})
}
