package crypt_test

import (
	"bytes"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/client/support/crypt"
	"github.com/stretchr/testify/require"
)

func Test_NewAgeCrypter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		crypter, err := crypt.NewAgeCrypter("secret")
		require.NoError(t, err)
		require.NotNil(t, crypter)
	})

	t.Run("invalid secret", func(t *testing.T) {
		crypter, err := crypt.NewAgeCrypter("")
		require.Error(t, err)
		require.Nil(t, crypter)
	})
}

func Test_AgeCrypter_Encrypt(t *testing.T) {
	crypter, err := crypt.NewAgeCrypter("secret")
	require.NoError(t, err)

	dst := bytes.NewBuffer(nil)

	wc, err := crypter.Encrypt(dst)
	require.NoError(t, err)
	require.NotNil(t, wc)
}

func Test_AgeCrypter_Decrypt(t *testing.T) {
	crypter, err := crypt.NewAgeCrypter("secret")
	require.NoError(t, err)

	t.Run("empty src", func(t *testing.T) {
		src := bytes.NewBuffer(nil)
		r, err := crypter.Decrypt(src)
		require.Error(t, err)
		require.Nil(t, r)
	})
}

