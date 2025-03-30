package middleware_test

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd/middleware"
)

func TestCombine(t *testing.T) {
	t.Run("no hooks", func(t *testing.T) {
		combined := middleware.Combine()
		err := combined(&cobra.Command{}, []string{})
		assert.NoError(t, err)
	})

	t.Run("single hook success", func(t *testing.T) {
		hook := func(cmd *cobra.Command, args []string) error {
			return nil
		}
		combined := middleware.Combine(hook)
		err := combined(&cobra.Command{}, []string{})
		assert.NoError(t, err)
	})

	t.Run("single hook failure", func(t *testing.T) {
		hook := func(cmd *cobra.Command, args []string) error {
			return errors.New("hook error")
		}
		combined := middleware.Combine(hook)
		err := combined(&cobra.Command{}, []string{})
		assert.EqualError(t, err, "hook error")
	})

	t.Run("multiple hooks success", func(t *testing.T) {
		hook1 := func(cmd *cobra.Command, args []string) error {
			return nil
		}
		hook2 := func(cmd *cobra.Command, args []string) error {
			return nil
		}
		combined := middleware.Combine(hook1, hook2)
		err := combined(&cobra.Command{}, []string{})
		assert.NoError(t, err)
	})

	t.Run("multiple hooks with failure", func(t *testing.T) {
		hook1 := func(cmd *cobra.Command, args []string) error {
			return nil
		}
		hook2 := func(cmd *cobra.Command, args []string) error {
			return errors.New("hook2 error")
		}
		hook3 := func(cmd *cobra.Command, args []string) error {
			return nil
		}
		combined := middleware.Combine(hook1, hook2, hook3)
		err := combined(&cobra.Command{}, []string{})
		assert.EqualError(t, err, "hook2 error")
	})

	t.Run("nil hook", func(t *testing.T) {
		hook := func(cmd *cobra.Command, args []string) error {
			return nil
		}
		combined := middleware.Combine(hook, nil)
		err := combined(&cobra.Command{}, []string{})
		assert.NoError(t, err)
	})
}
