package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger = zap.NewNop().Sugar()

func Logger() *zap.SugaredLogger {
	return logger
}

func InitServerLogger() error {
	noSugarLogger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("logger build error: %w", err)
	}

	logger = noSugarLogger.Sugar()

	return nil
}

func InitClientLogger() error {
	conf := zap.NewDevelopmentConfig()
	conf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.DisableCaller = true
	conf.DisableStacktrace = true

	noSugarLogger, err := conf.Build()
	if err != nil {
		return fmt.Errorf("logger build error: %w", err)
	}

	logger = noSugarLogger.Sugar()

	return nil
}

func InitTestLogger(t *testing.T) {
	require.NoError(t, InitServerLogger())
}
