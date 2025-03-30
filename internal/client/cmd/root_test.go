package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func TestNewRootCommand(t *testing.T) {
	t.Run("sanity check", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		cmd := NewRootCommand(container)
		require.NotNil(t, cmd)
	})

	t.Run("verbose", func(t *testing.T) {
		ctx, cancel := utils.TestContext(t)
		defer cancel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		container := NewMockContainer(ctrl)

		out := bytes.NewBuffer(nil)

		cmd := NewRootCommand(container)
		cmd.SetArgs([]string{"--verbose", "help"})
		cmd.SetIn(bytes.NewBuffer(nil))
		cmd.SetOut(out)
		cmd.SetErr(out)
		cmd.SetContext(ctx)

		err := cmd.Execute()
		require.NoError(t, err)

		require.Contains(t, out.String(), "Verbose output enabled")
	})
}

func TestNewConfig(t *testing.T) {
	t.Run("sanity check", func(t *testing.T) {
		conf, err := NewConfig()
		require.NoError(t, err)
		require.NotNil(t, conf)
	})
}

func TestConfigPath(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	out := bytes.NewBuffer(nil)

	cmd := newConfigPathCommand()
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)
	outString := out.String()
	pwd, err := os.Getwd()
	require.NoError(t, err)
	require.Contains(t, outString, pwd)
}
