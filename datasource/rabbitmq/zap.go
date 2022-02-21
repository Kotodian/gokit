package rabbitmq

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type rabbitmqHook struct {
	queue   string
	service string
}

func NewRabbitmqHook(queue string, service string) *rabbitmqHook {
	return &rabbitmqHook{
		queue:   queue,
		service: service,
	}
}

func (r *rabbitmqHook) Write(p []byte) (n int, err error) {
	return 0, nil
}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "date_time",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func NewZapLogger(minLevel zapcore.Level, version, queue, service string) *zap.Logger {
	writeSyncerList := make([]zapcore.WriteSyncer, 0)
	coreList := make([]zapcore.Core, 0)

	levelFunc := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= minLevel
	})
	hook := NewRabbitmqHook(queue, service)
	writeSyncerList = append(writeSyncerList, zapcore.AddSync(os.Stdout))
	writeSyncerList = append(writeSyncerList, zapcore.AddSync(hook))
	coreList = append(coreList, zapcore.NewCore(zapcore.NewJSONEncoder(NewEncoderConfig()), zapcore.NewMultiWriteSyncer(writeSyncerList...), levelFunc))
	core := zapcore.NewTee(coreList...)
	var logger *zap.Logger
	if minLevel == zapcore.DebugLevel {
		logger = zap.New(core, zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		logger = zap.New(core)
	}
	return logger
}
