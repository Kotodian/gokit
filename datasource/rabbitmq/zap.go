package rabbitmq

import (
	"context"
	"github.com/Kotodian/gokit/id"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type rabbitmqHook struct {
	hostName string
	queue    string
	service  string
	version  string
}

func NewRabbitmqHook(queue string, service string, version string) *rabbitmqHook {
	hostName, _ := os.Hostname()
	return &rabbitmqHook{
		queue:    queue,
		service:  service,
		hostName: hostName,
		version:  version,
	}
}

func (r *rabbitmqHook) Write(p []byte) (n int, err error) {
	if err = r.push(p); err != nil {
		return 0, err
	}
	return 0, nil
}

func (r *rabbitmqHook) push(data []byte) (err error) {
	ctx := context.Background()
	now := time.Now()
	var object interface{}
	if err = jsoniter.Unmarshal(data, &object); err != nil {
		return err
	}

	dataMap := object.(map[string]interface{})
	dataMap["_id"] = id.Next()
	dataMap["host"] = r.hostName
	dataMap["date"] = now.Format("2006-01-02")
	dataMap["time"] = now.Format("15:04:05")
	dataMap["version"] = r.version
	delete(dataMap, "date_time")
	_ = Publish(ctx, r.queue, nil, data)
	return nil
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
	hook := NewRabbitmqHook(queue, service, version)
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
