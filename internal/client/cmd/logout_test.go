package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/client/service/container"
	"github.com/kuvalkin/gophkeeper/internal/client/support/mocks"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestLogout(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	newTestLogoutCommand := func(container container.Container) *cobra.Command {
		logout := newLogoutCommand(container)
		logout.SetArgs([]string{})
		logout.SetOut(bytes.NewBuffer(nil))
		logout.SetErr(bytes.NewBuffer(nil))
		logout.SetIn(bytes.NewBuffer(nil))
		logout.SetContext(ctx)
		return logout
	}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		service := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Logout(ctx).Return(nil)

		cmd := newTestLogoutCommand(container)
		err := cmd.Execute()
		require.NoError(t, err)
	})

	t.Run("cant get auth service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(nil, errors.New("test error")).AnyTimes()

		cmd := newTestLogoutCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})

	t.Run("cant logout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := mocks.NewMockContainer(ctrl)
		service := mocks.NewMockAuthService(ctrl)

		container.EXPECT().GetAuthService(ctx).Return(service, nil).AnyTimes()

		service.EXPECT().Logout(ctx).Return(errors.New("test error"))

		cmd := newTestLogoutCommand(container)
		err := cmd.Execute()
		require.Error(t, err)
	})
}
