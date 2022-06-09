package rabbitmq

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/Kotodian/gokit/id"
	jsoniter "github.com/json-iterator/go"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type rabbitmqHook struct {
	hostName string
	queue    string
	service  string
	version  string
	ignore   func(map[string]interface{}) bool
}

func NewRabbitmqHook(queue string, service string, version string, ignore func(map[string]interface{}) bool) *rabbitmqHook {
	hostName, _ := os.Hostname()
	hook := &rabbitmqHook{
		queue:    queue,
		service:  service,
		hostName: hostName,
		version:  version,
	}
	if ignore != nil {
		hook.ignore = ignore
	}
	return hook
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
	if r.ignore != nil && r.ignore(dataMap) {
		return nil
	}
	dataMap["_id"] = id.Next()
	dataMap["date"] = now.Format("2006-01-02")
	dataMap["time"] = now.Format("15:04:05")
	delete(dataMap, "date_time")
	_ = Publish(ctx, r.queue, nil, dataMap)
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

type Options struct {
	LogFileDir    string //日志路径
	AppName       string // Filename是要写入日志的文件前缀
	AppVersion    string // 版本
	ErrorFileName string
	WarnFileName  string
	InfoFileName  string
	DebugFileName string
	MaxSize       int    // 一个文件多少Ｍ大于该数字开始切分文件
	MaxBackups    int    // MaxBackups是要保留的最大旧日志文件数
	MaxAge        int    // MaxAge是根据日期保留旧日志文件的最大天数
	Queue         string // rabbitmq queue
	IgnoreFunc    func(map[string]interface{}) bool
	zap.Config
}

func NewZapLogger(minLevel zapcore.Level, version, queue, service string, ignore ...func(map[string]interface{}) bool) *zap.Logger {
	writeSyncerList := make([]zapcore.WriteSyncer, 0)
	coreList := make([]zapcore.Core, 0)

	levelFunc := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= minLevel
	})
	if minLevel == zapcore.DebugLevel {
		writeSyncerList = append(writeSyncerList, zapcore.AddSync(os.Stdout))
	}
	var hook *rabbitmqHook
	if queue != "" {
		if len(ignore) > 0 {
			hook = NewRabbitmqHook(queue, service, version, ignore[0])
		} else {
			hook = NewRabbitmqHook(queue, service, version, nil)
		}
		writeSyncerList = append(writeSyncerList, zapcore.AddSync(hook))
	}
	coreList = append(coreList, zapcore.NewCore(zapcore.NewJSONEncoder(NewEncoderConfig()), zapcore.NewMultiWriteSyncer(writeSyncerList...), levelFunc))
	var logger *zap.Logger
	core := zapcore.NewTee(coreList...)
	logger = zap.New(core, zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return logger
}

var (
	sp                             = string(filepath.Separator)
	errWS, warnWS, infoWS, debugWS zapcore.WriteSyncer       // IO输出
	rabbitmqWS                     zapcore.WriteSyncer       // rabbitmq 输出
	debugConsoleWS                 = zapcore.Lock(os.Stdout) // 控制台标准输出
	errorConsoleWS                 = zapcore.Lock(os.Stderr)
)

type Logger struct {
	*zap.Logger
	Opts      *Options `json:"opts"`
	zapConfig zap.Config
	hostname  string
}

func InitLogger(cf ...*Options) *Logger {
	logger := &Logger{
		Opts: &Options{},
	}
	if len(cf) > 0 {
		logger.Opts = cf[0]
	}
	logger.loadCfg()
	logger.init()
	return logger
}

func (l *Logger) init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	l.hostname = hostname
	l.setSyncers()
	mylogger, err := l.zapConfig.Build(l.cores())
	if err != nil {
		panic(err)
	}
	l.Logger = mylogger
	defer l.Logger.Sync()
}

func (l *Logger) loadCfg() {
	if l.Opts.Development {
		l.zapConfig = zap.NewDevelopmentConfig()
		l.zapConfig.EncoderConfig = NewEncoderConfig()
	} else {
		l.zapConfig = zap.NewProductionConfig()
		l.zapConfig.EncoderConfig = NewEncoderConfig()
	}
	if l.Opts.OutputPaths == nil || len(l.Opts.OutputPaths) == 0 {
		l.zapConfig.OutputPaths = []string{"stdout"}
	}
	if l.Opts.ErrorOutputPaths == nil || len(l.Opts.ErrorOutputPaths) == 0 {
		l.zapConfig.OutputPaths = []string{"stderr"}
	}
	// 默认输出到程序运行目录的logs子目录
	if l.Opts.LogFileDir == "" {
		l.Opts.LogFileDir, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
		// l.Opts.LogFileDir += sp + "logs" + sp
	}
	if l.Opts.AppName == "" {
		l.Opts.AppName = "app"
	}
	if l.Opts.ErrorFileName == "" {
		l.Opts.ErrorFileName = "error.log"
	}
	if l.Opts.WarnFileName == "" {
		l.Opts.WarnFileName = "warn.log"
	}
	if l.Opts.InfoFileName == "" {
		l.Opts.InfoFileName = "info.log"
	}
	if l.Opts.DebugFileName == "" {
		l.Opts.DebugFileName = "debug.log"
	}
	if l.Opts.MaxSize == 0 {
		l.Opts.MaxSize = 100
	}
	if l.Opts.MaxBackups == 0 {
		l.Opts.MaxBackups = 30
	}
	if l.Opts.MaxAge == 0 {
		l.Opts.MaxAge = 30
	}
}

func (l *Logger) setSyncers() {
	f := func(fN string) zapcore.WriteSyncer {
		// 每小时一个文件
		logf, _ := rotatelogs.New(l.Opts.LogFileDir+sp+l.Opts.AppName+"-%Y-%m-%d.%H"+"."+fN,
			rotatelogs.WithLinkName(l.Opts.LogFileDir+sp+l.Opts.AppName+"-"+fN),
			// rotatelogs.WithMaxAge(30*24*time.Hour),
			rotatelogs.WithRotationTime(time.Minute),
		)
		return zapcore.AddSync(logf)
	}
	errWS = f(l.Opts.ErrorFileName)
	warnWS = f(l.Opts.WarnFileName)
	infoWS = f(l.Opts.InfoFileName)
	debugWS = f(l.Opts.DebugFileName)
	if l.Opts.Queue != "" {
		rf := func() zapcore.WriteSyncer {
			return zapcore.AddSync(NewRabbitmqHook(l.Opts.Queue, l.Opts.AppName, l.Opts.AppVersion, l.Opts.IgnoreFunc))
		}
		rabbitmqWS = rf()
	}
	return
}

func (l *Logger) cores() zap.Option {
	fileEncoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)
	consoleEncoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)

	errPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.WarnLevel && zapcore.WarnLevel-l.zapConfig.Level.Level() > -1
	})
	warnPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel && zapcore.WarnLevel-l.zapConfig.Level.Level() > -1
	})
	infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel && zapcore.InfoLevel-l.zapConfig.Level.Level() > -1
	})
	debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel && zapcore.DebugLevel-l.zapConfig.Level.Level() > -1
	})
	cores := []zapcore.Core{
		zapcore.NewCore(fileEncoder, errWS, errPriority),
		zapcore.NewCore(fileEncoder, warnWS, warnPriority),
		zapcore.NewCore(fileEncoder, infoWS, infoPriority),
		zapcore.NewCore(fileEncoder, debugWS, debugPriority),
	}
	if l.Opts.Development {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(consoleEncoder, errorConsoleWS, errPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, warnPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, infoPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, debugPriority),
		}...)
	}
	if l.Opts.Queue != "" {
		rabbitmqEncoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)
		l.defaultFields(rabbitmqEncoder)
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(rabbitmqEncoder, rabbitmqWS, errPriority),
			zapcore.NewCore(rabbitmqEncoder, rabbitmqWS, warnPriority),
			zapcore.NewCore(rabbitmqEncoder, rabbitmqWS, infoPriority),
			zapcore.NewCore(rabbitmqEncoder, rabbitmqWS, debugPriority),
		}...)
	}

	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func timeUnixNano(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.UnixNano() / 1e6)
}

func (l *Logger) defaultFields(ecd zapcore.Encoder) {
	ecd.AddString("host", l.hostname)
	ecd.AddString("edition", l.Opts.AppVersion)
}
