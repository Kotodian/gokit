package mongo

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLog(t *testing.T) {
	os.Setenv("MONGO_AUTH_DB", "admin")
	os.Setenv("MONGO_DB", "jx-csms")
	os.Setenv("MONGO_USER", "root")
	os.Setenv("MONGO_PASSWD", "abc123,.")
	os.Setenv("MONGO_HOST", "10.43.0.13")
	os.Setenv("MONGO_PORT", "27017")
	InitEnv()

	//按实际需求灵活定义日志级别
	debugLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.DebugLevel
	})
	development := zap.Development()
	writeSyncerList := make([]zapcore.WriteSyncer, 0)
	coreList := make([]zapcore.Core, 0)
	hook, err := NewMongoLogHook("test", "v1.0")
	if err != nil {
		t.Error(err)
		return
	}
	writeSyncerList = append(writeSyncerList, zapcore.AddSync(os.Stdout))
	writeSyncerList = append(writeSyncerList, zapcore.AddSync(hook))
	coreList = append(coreList, zapcore.NewCore(zapcore.NewJSONEncoder(NewEncoderConfig()), zapcore.NewMultiWriteSyncer(writeSyncerList...), debugLevel))
	core := zapcore.NewTee(coreList...)

	logger := zap.New(core, development)
	logger.Info("ping message received", zap.String("sn", "T1641735210"))
}
