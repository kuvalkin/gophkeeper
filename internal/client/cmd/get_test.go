package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/entries"
	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	clientUtils "github.com/kuvalkin/gophkeeper/internal/client/support/utils"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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

		container := NewMockContainer(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, "test", "test")
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

		container := NewMockContainer(ctrl)

		cmd, _ := newTestGetLoginCommand(container, "")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("entry not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, "test", "test")
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

		container := NewMockContainer(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authCtx := context.WithValue(ctx, "test", "test")
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

		container := NewMockContainer(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(nil, errors.New("error")).AnyTimes()

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant get auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(nil, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("error")).AnyTimes()

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant set token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)
		entryService := NewMockEntryService(ctrl)
		authService := NewMockAuthService(ctrl)

		container.EXPECT().GetEntryService(ctx).Return(entryService, nil).AnyTimes()
		container.EXPECT().GetAuthService(ctx).Return(authService, nil).AnyTimes()

		authService.EXPECT().AddAuthorizationHeader(ctx).Return(nil, errors.New("error"))

		cmd, _ := newTestGetLoginCommand(container, "name")
		err := cmd.Execute()
		require.Error(t, err)
	})
}