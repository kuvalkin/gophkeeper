// Package utils provides utility functions and helpers for common operations.
package utils

import (
	"io"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

// CloseAndLogError attempts to close the provided io.Closer and logs any error
// that occurs during the close operation. If no logger is provided, a default logger is used.
func CloseAndLogError(closer io.Closer, logger *zap.SugaredLogger) {
	if err := closer.Close(); err != nil {
		if logger == nil {
			logger = log.Logger()
		}

		logger.Errorw("error closing", "error", err)
	}
}
