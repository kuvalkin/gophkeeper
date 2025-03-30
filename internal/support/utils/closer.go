package utils

import (
	"io"

	"go.uber.org/zap"

	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func CloseAndLogError(closer io.Closer, logger *zap.SugaredLogger) {
	if err := closer.Close(); err != nil {
		if logger == nil {
			logger = log.Logger()
		}

		logger.Errorw("error closing", "error", err)
	}
}
