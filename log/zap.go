package log

import (
	"os"

	"go.uber.org/zap"
)

var logger *zap.Logger
var err error

func failOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	if os.Getenv("LOG_LEVEL") == "debug" || os.Getenv("LOG_LEVEL") == "" {
		logger, err = zap.NewDevelopment()
		failOnError(err)
	} else if os.Getenv("LOG_LEVEL") == "info" {
		logger, err = zap.NewProduction()
		failOnError(err)
	}
}

func Info(format string, fields ...zap.Field) {
	logger.Info(format, fields...)
}

func Debug(format string, fields ...zap.Field) {
	logger.Debug(format, fields...)
}

func Warn(format string, fields ...zap.Field) {
	logger.Warn(format, fields...)
}

func Error(format string, fields ...zap.Field) {
	logger.Error(format, fields...)
}
