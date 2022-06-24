package rabbitmq

import (
	"fmt"
	"testing"
	"time"

	"github.com/Kotodian/gokit/datasource/elasticsearch"
	"go.uber.org/zap"
)

func TestZapLog(t *testing.T) {
	elasticsearch.Init()
	data := &Options{
		AppName:    "jx-coregw",
		// MaxSize:    30,
		// MaxBackups: 7,
		// MaxAge:     7,
		Config:     zap.Config{},
		// Index:      "jx-coregw",
	}
	data.Development = true
	logger := InitLogger(data)
	for i := 0; i < 2; i++ {
		time.Sleep(time.Second)
		logger.Debug(fmt.Sprint("debug log ", i), zap.Int("line", 999))
		logger.Info(fmt.Sprint("Info log ", i), zap.Any("line", "1231231231"))
		logger.Info(fmt.Sprint("Info log ", i), zap.Any("line", "1231231231"))
		logger.Info(fmt.Sprint("Info log ", i), zap.Any("line", "1231231231"))
		logger.Info(fmt.Sprint("Info log ", i), zap.Any("line", "1231231231"))
		logger.Error("error log")
	}
}
