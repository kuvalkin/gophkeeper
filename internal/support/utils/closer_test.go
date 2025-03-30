package utils_test

import (
	"errors"
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/support/utils"
)

type mockCloser struct {
	shouldFail bool
}

func (m *mockCloser) Close() error {
	if m.shouldFail {
		return errors.New("mock close error")
	}
	return nil
}

func TestCloseAndLogError(t *testing.T) {
	t.Run("no panic with nil logger", func(t *testing.T) {
		t.Run("close fail", func(t *testing.T) {
			closer := &mockCloser{shouldFail: true}

			utils.CloseAndLogError(closer, nil)
		})

		t.Run("close ok", func(t *testing.T) {
			closer := &mockCloser{shouldFail: false}

			utils.CloseAndLogError(closer, nil)
		})
	})
}
