package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kuvalkin/gophkeeper/internal/client/support/utils"
)

func Test_GetEntryKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		key := utils.GetEntryKey("prefix", "name")
		require.Equal(t, "88d31d4b6ee626c2316d2f12260a26bd2fd5504fd4a66dfa2263c22131acd1ac", key)
	})
}
