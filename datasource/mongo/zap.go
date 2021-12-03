package mongo

import (
	"context"
	"github.com/Kotodian/gokit/id"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type MongoLogHook struct {
	collection string
	version    string
}

func NewMongoLogHook(collection, version string) (*MongoLogHook, error) {
	hook := &MongoLogHook{
		collection: collection,
		version:    version,
	}
	db, err := connect()
	if err != nil {
		return nil, err
	}
	defer db.Client().Disconnect(context.Background())
	_ = db.CreateCollection(context.Background(), collection)
	return hook, nil
}

func (m *MongoLogHook) Write(data []byte) (n int, err error) {
	if err = m.insertLogToMongo(data); err != nil {
		return 0, err
	}
	return
}

func (m *MongoLogHook) insertLogToMongo(data []byte) (err error) {
	ctx := context.Background()
	db, err := connect()
	if err != nil {
		return err
	}
	defer db.Client().Disconnect(context.Background())
	now := time.Now()

	collection := db.Collection(m.collection + "-" + now.Format("0601"))
	var object interface{}
	if err = jsoniter.Unmarshal(data, &object); err != nil {
		return
	}

	//转为map类型
	dataMap := object.(map[string]interface{})
	host, _ := os.Hostname()
	dataMap["_id"] = id.Next()
	dataMap["host"] = host
	dataMap["date"] = now.Format("2006-01-02")
	dataMap["time"] = now.Format("15:04:05")
	dataMap["version"] = m.version
	delete(dataMap, "date_time")
	_, err = collection.InsertOne(ctx, dataMap)
	if err != nil {
		return
	}
	return
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

func NewZapLogger(minLevel zapcore.Level, version string, collection ...string) *zap.Logger {
	writeSyncerList := make([]zapcore.WriteSyncer, 0)
	coreList := make([]zapcore.Core, 0)
	if len(collection) > 0 {
		levelFunc := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level >= minLevel
		})
		hook, err := NewMongoLogHook(collection[0], version)
		if err != nil {
			return nil
		}
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
	logger, _ := zap.NewDevelopment()
	return logger
}
