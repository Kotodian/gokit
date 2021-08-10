package boot

import (
	"context"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource/redis"

	log "github.com/sirupsen/logrus"
)

type initFunc func() error

type initService struct {
	name string
	fn   initFunc
}

var initFuncs []*initService
var WG *sync.WaitGroup
var bootDoneCH chan struct{}
var ctx context.Context
var cancel context.CancelFunc

func init() {
	initFuncs = make([]*initService, 0)
	WG = &sync.WaitGroup{}
	bootDoneCH = make(chan struct{}, 0)
	ctx, cancel = context.WithCancel(context.TODO())
	// initStack = stack.NewStack()
}

// RegisterInit 注册初始化服务
func RegisterInit(name string, fn initFunc) {
	log.Info("register init ", name)
	initFuncs = append(initFuncs, &initService{
		name: name,
		fn:   fn,
	})
}

func ReloadService(name string) {
	redis.PublishStreamWithMaxlen("reload:services:boot", 100, name)
}

// Init 加载所以boot
func Init() {
	defer func() {
		close(bootDoneCH)
	}()
	for _, service := range initFuncs {
		log.Infof("service:%s booting...", service.name)
		if err := service.fn(); err != nil {
			panic(err)
		}
		log.Infof("service:%s boot success", service.name)
	}
}

func Daemon(name string, fn func(ctx context.Context)) {
	log.Infof("daemon:%s listening...", name)

	WG.Add(1)
	//done := make(chan struct{}, 1)
	go func() {
		<-bootDoneCH
		defer WG.Done()
		for {
			select {
			case <-ctx.Done():
				//fmt.Println("dddddddaaaaa")
				return
			default:
			}
			// _log := log.WithFields(log.Fields{
			// 	"daemon": name,
			// })
			// ctx = context.WithValue(ctx, "log", _log)
			fn(ctx)
			time.Sleep(time.Second * 3)
		}
	}()
}

func Shutdown() {
	//close(shutdown)
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func Context() context.Context {
	return ctx
}
