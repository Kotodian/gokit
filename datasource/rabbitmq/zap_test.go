package rabbitmq

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestZapLog(t *testing.T) {
	data := &Options{
		LogFileDir: "/Users/linqiankai/go/src/github.com/Kotodian/gokit/datasource/rabbitmq/logs",
		AppName:    "jx-ac-ocpp",
		MaxSize:    30,
		MaxBackups: 7,
		MaxAge:     7,
		Config:     zap.Config{},
	}
	data.Development = true
	logger := InitLogger(data)
	for i := 0; i < 2; i++ {
		time.Sleep(time.Second)
		logger.Debug(fmt.Sprint("debug log ", i), zap.Int("line", 999))
		logger.Info(fmt.Sprint("Info log ", i), zap.Any("line", "1231231231"))
		logger.Warn(fmt.Sprint("warn log ", i), zap.String("line", `{"a":"4","b":"5"}`))
		logger.Error(fmt.Sprint("err log ", i), zap.String("line", `{"a":"7","b":"8"}`))
	}
}
