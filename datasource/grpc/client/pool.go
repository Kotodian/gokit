package client

import (
	"time"

	"github.com/silenceper/pool"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var p pool.Pool

// var ch chan struct{}

func init() {
	//factory 创建连接的方法
	factory := func() (interface{}, error) {
		//fmt.Println("链接 grpc-proxy")
		//ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
		//defer ()
		return grpcLib.Dial("grpc-proxy:8071", grpcLib.WithInsecure(), grpcLib.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}))
		// return grpcLib.Dial("106.14.61.23:8071", grpcLib.WithInsecure())
	}

	//close 关闭链接的方法
	_close := func(v interface{}) error {
		// fmt.Println("关闭beanstalkd")
		//fmt.Println("关闭 grpc-proxy")
		return v.(*grpcLib.ClientConn).Close()
	}
	poolConfig := &pool.Config{
		InitialCap: 100,
		MaxIdle:    300,
		MaxCap:     2000,
		Factory:    factory,
		Close:      _close,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 300 * time.Second,
	}
	// ch = make(chan struct{}, 1)

	// go func() {
	var err error
	p, err = pool.NewChannelPool(poolConfig)
	if err != nil {
		panic(err)
	}
	// close(ch)
	// }()

	// //从连接池中取得一个链接
	// v, err := p.Get()

	// //do something
	// //conn=v.(net.Conn)

	// //将链接放回连接池中
	// p.Put(v)

	// //释放连接池中的所有链接
	// p.Release()

	// //查看当前链接中的数量
	// current := p.Len()
}

func Conn() (*grpcLib.ClientConn, error) {
	// <-ch
	// fmt.Println("取出 grpc-proxy")
	c, err := p.Get()
	if err != nil {
		return nil, err
	}
	return c.(*grpcLib.ClientConn), nil
}

func Close(c interface{}) error {
	// fmt.Println("放回 grpc-proxy")
	return p.Put(c)
}
