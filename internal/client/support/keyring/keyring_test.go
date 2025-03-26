package keyring_test

import (
	"errors"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/client/support/keyring"
	"github.com/stretchr/testify/require"
	goKeyring "github.com/zalando/go-keyring"
)

func TestKeyring_Set(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		goKeyring.MockInit()

		err := keyring.Set("key", "value")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		goKeyring.MockInitWithError(errors.New("error"))

		err := keyring.Set("key", "value")
		require.Error(t, err)
	})
}

func TestKeyring_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		goKeyring.MockInit()
		err := keyring.Set("key", "value")
		require.NoError(t, err)

		value, ok, err := keyring.Get("key")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "value", value)
	})

	t.Run("not found", func(t *testing.T) {
		goKeyring.MockInitWithError(goKeyring.ErrNotFound)

		value, ok, err := keyring.Get("key")
		require.NoError(t, err)
		require.False(t, ok)
		require.Empty(t, value)
	})

	t.Run("error", func(t *testing.T) {
		goKeyring.MockInitWithError(errors.New("error"))

		value, ok, err := keyring.Get("key")
		require.Error(t, err)
		require.False(t, ok)
		require.Empty(t, value)
	})
}

func TestKeyring_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		goKeyring.MockInit()
		err := keyring.Set("key", "value")
		require.NoError(t, err)

		err = keyring.Delete("key")
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		goKeyring.MockInitWithError(errors.New("error"))

		err := keyring.Delete("key")
		require.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		goKeyring.MockInitWithError(goKeyring.ErrNotFound)

		err := keyring.Delete("key")
		require.NoError(t, err)
	})
}