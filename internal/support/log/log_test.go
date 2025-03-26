package log_test

import (
	"testing"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	logger := log.Logger()
	require.NotNil(t, logger)
}

func TestInitServerLogger(t *testing.T) {
	err := log.InitServerLogger()
	require.NoError(t, err)

	logger := log.Logger()
	require.NotNil(t, logger)
}

func TestInitClientLogger(t *testing.T) {
	err := log.InitClientLogger()
	require.NoError(t, err)

	logger := log.Logger()
	require.NotNil(t, logger)
}