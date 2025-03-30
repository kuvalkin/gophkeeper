package utils_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

func Test_CopyContext(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, cancel := utils.TestContext(t)
		defer cancel()

		src := bytes.NewBuffer([]byte("data"))
		dst := bytes.NewBuffer(nil)

		n, err := utils.CopyContext(ctx, dst, src)
		require.NoError(t, err)
		require.Equal(t, int64(4), n)
		require.Equal(t, "data", dst.String())
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := utils.TestContext(t)
		cancel()

		src := bytes.NewBuffer([]byte("data"))
		dst := bytes.NewBuffer(nil)

		n, err := utils.CopyContext(ctx, dst, src)
		require.Error(t, err)
		require.ErrorIs(t, err, ctx.Err())
		require.Zero(t, n)
		require.Empty(t, dst.String())
	})
}
