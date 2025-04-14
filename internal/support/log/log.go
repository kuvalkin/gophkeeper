// Package log provides logging utilities for the application.
// It uses the zap logging library to configure and manage loggers.
package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger is the global logger instance used throughout the application.
var logger = zap.NewNop().Sugar()

// Logger returns the current global logger instance.
// This logger can be used for logging messages at various levels.
func Logger() *zap.SugaredLogger {
	return logger
}

// InitServerLogger initializes the logger for the server environment.
// It sets up a development logger with default configurations.
// Returns an error if the logger cannot be built.
func InitServerLogger() error {
	noSugarLogger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("logger build error: %w", err)
	}

	logger = noSugarLogger.Sugar()

	return nil
}

// InitClientLogger initializes the logger for the client environment.
// It customizes the logger configuration to include time formatting,
// colored log levels, and disables caller and stack trace information.
// Returns an error if the logger cannot be built.
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
