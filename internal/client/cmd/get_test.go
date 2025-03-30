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

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/entries"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	clientUtils "github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestGetLogin(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestGetLoginCommand := func(container container.Container, name string) (*cobra.Command, *bytes.Buffer) {
		out := bytes.NewBuffer(nil)

		cmd := newGetCommand(container)
		// gkeep get login {name}
		cmd.SetArgs([]string{"login", name})
		cmd.SetOut(out)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd, out
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		pair := &entries.LoginPasswordPair{
			Login:    "mylogin",
			Password: "mypassword",
		}
		content, err := pair.Marshal()
		require.NoError(t, err)

		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("login", "name")).Return("mynotes", content, true, nil)

		cmd, out := newTestGetLoginCommand(container, "name")
		err = cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "mylogin")
		require.Contains(t, outString, "mypassword")
		require.Contains(t, outString, "mynotes")
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		cmd, _ := newTestGetLoginCommand(container, "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("entry not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("login", "name")).Return("", nil, false, nil)

		cmd, out := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "not found")
	})

	t.Run("service returns invalid data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("invalid data"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("login", "name")).Return("mynotes", content, true, nil)

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get entry service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(nil, errors.New("error")).AnyTimes()

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(nil, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("error")).AnyTimes()

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant set token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authService.EXPECT().AddAuthorizationHeader(ctx).Return(nil, errors.New("error"))

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})
}

func TestGetFile(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestGetFileCommand := func(container container.Container, name string, path string) (*cobra.Command, *bytes.Buffer) {
		out := bytes.NewBuffer(nil)

		cmd := newGetCommand(container)
		// gkeep get file {name} {savePath}
		cmd.SetArgs([]string{"file", name, path})
		cmd.SetOut(out)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd, out
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("file content"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("file", "name")).Return("mynotes", content, true, nil)

		file, err := os.CreateTemp("", "test-get-file-*")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		require.NoError(t, file.Close())

		cmd, out := newTestGetFileCommand(container, "name", file.Name())
		err = cmd.Execute()
		require.NoError(t, err)
		require.Contains(t, out.String(), "mynotes")
		require.FileExists(t, file.Name())
		written, err := os.ReadFile(file.Name())
		require.NoError(t, err)
		require.Equal(t, "file content", string(written))
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		cmd, _ := newTestGetFileCommand(container, "", "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("no path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		cmd, _ := newTestGetFileCommand(container, "name", "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("entry not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("file", "name")).Return("", nil, false, nil)

		file, err := os.CreateTemp("", "test-get-file-*")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		require.NoError(t, file.Close())

		cmd, out := newTestGetFileCommand(container, "name", file.Name())
		err = cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "not found")
	})

	t.Run("get error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("file content"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("file", "name")).Return("mynotes", content, true, errors.New("error"))

		file, err := os.CreateTemp("", "test-get-file-*")
		require.NoError(t, err)
		defer os.Remove(file.Name())
		require.NoError(t, file.Close())

		cmd, _ := newTestGetFileCommand(container, "name", file.Name())
		err = cmd.Execute()
		require.Error(t, err)
	})

	t.Run("invalid path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("file content"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("file", "name")).Return("mynotes", content, true, nil)

		// path is directory
		cmd, _ := newTestGetFileCommand(container, "name", os.TempDir())
		err := cmd.Execute()
		require.Error(t, err)
	})
}

func TestGetCard(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestGetCardCommand := func(container container.Container, name string) (*cobra.Command, *bytes.Buffer) {
		out := bytes.NewBuffer(nil)

		cmd := newGetCommand(container)
		// gkeep get card {name}
		cmd.SetArgs([]string{"card", name})
		cmd.SetOut(out)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd, out
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		card := &entries.BankCard{
			Number:     "4111111111111111",
			HolderName: "John Doe",
			ExpirationDate: entries.ExpirationDate{
				Year:  2030,
				Month: 12,
			},
			CVV: 123,
		}
		content, err := card.Marshal()
		require.NoError(t, err)

		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("card", "name")).Return("mynotes", content, true, nil)

		cmd, out := newTestGetCardCommand(container, "name")
		err = cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "4111111111111111")
		require.Contains(t, outString, "John Doe")
		require.Contains(t, outString, "2030")
		require.Contains(t, outString, "12")
		require.Contains(t, outString, "123")
		require.Contains(t, outString, "mynotes")
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		cmd, _ := newTestGetCardCommand(container, "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("entry not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("card", "name")).Return("", nil, false, nil)

		cmd, out := newTestGetCardCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "not found")
	})

	t.Run("get error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("file content"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("card", "name")).Return("mynotes", content, true, errors.New("error"))

		cmd, _ := newTestGetCardCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("service returns invalid data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("invalid data"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("card", "name")).Return("mynotes", content, true, nil)

		cmd, _ := newTestGetCardCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})
}

func TestGetText(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestGetTextCommand := func(container container.Container, name string) (*cobra.Command, *bytes.Buffer) {
		out := bytes.NewBuffer(nil)

		cmd := newGetCommand(container)
		// gkeep get text {name}
		cmd.SetArgs([]string{"text", name})
		cmd.SetOut(out)
		cmd.SetErr(io.Discard)
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetContext(ctx)

		return cmd, out
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("text content"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("text", "name")).Return("mynotes", content, true, nil)

		cmd, out := newTestGetTextCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "text content")
		require.Contains(t, outString, "mynotes")
	})

	t.Run("no name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		cmd, _ := newTestGetTextCommand(container, "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("entry not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("text", "name")).Return("", nil, false, nil)

		cmd, out := newTestGetTextCommand(container, "name")
		err := cmd.Execute()
		require.NoError(t, err)
		outString := out.String()
		require.Contains(t, outString, "not found")
	})

	t.Run("get error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		entryService := mocks.NewMockEntryService(ctrl)
		authService := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, testCtxKey("test"), "test")
		authService.EXPECT().AddAuthorizationHeader(ctx).Return(authCtx, nil)

		content := io.NopCloser(bytes.NewBufferString("file content"))
		entryService.EXPECT().GetEntry(authCtx, clientUtils.GetEntryKey("text", "name")).Return("mynotes", content, true, errors.New("error"))

		cmd, _ := newTestGetTextCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})
}
