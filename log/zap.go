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

func Infof(format string, args ...interface{}) {
	logger.Sugar().Infof(format, args...)
}

func Debug(format string, fields ...zap.Field) {
	logger.Debug(format, fields...)
}

func Debugf(format string, args ...interface{}) {
	logger.Sugar().Debugf(format, args...)
}

func Warn(format string, fields ...zap.Field) {
	logger.Warn(format, fields...)
}

func Warnf(format string, args ...interface{}) {
	logger.Sugar().Warnf(format, args...)
}

func Error(format string, fields ...zap.Field) {
	logger.Error(format, fields...)
}

func Errorf(format string, args ...interface{}) {
	logger.Sugar().Errorf(format, args...)
}

func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Int32(key string, value int32) zap.Field {
	return zap.Int32(key, value)
}
