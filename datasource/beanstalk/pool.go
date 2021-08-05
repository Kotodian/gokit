package beanstalk

import (
	"fmt"
	"os"
	"time"

	"github.com/beanstalkd/go-beanstalk"
	"github.com/silenceper/pool"
)

var p pool.Pool

// var ch chan struct{}

func init() {
	//factory 创建连接的方法
	factory := func() (interface{}, error) {
		return beanstalk.Dial("tcp", os.Getenv("BEANSTALKD_URL"))
	}

	//close 关闭链接的方法
	close := func(v interface{}) error {
		return v.(*beanstalk.Conn).Close()
	}

	//创建一个连接池： 初始化5，最大链接30
	poolConfig := &pool.Config{
		InitialCap: 5,
		MaxIdle:    100,
		MaxCap:     1000,
		Factory:    factory,
		Close:      close,
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 30 * time.Second,
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

	// 从连接池中取得一个链接
	// v, err := p.Get()

	// do something
	// onn=v.(net.Conn)

	// 将链接放回连接池中
	// p.Put(v)

	// 释放连接池中的所有链接
	// p.Release()

	// 查看当前链接中的数量
	// current := p.Len()
}

func Conn() (*beanstalk.Conn, error) {
	// <-ch
	fmt.Println("获取beanstalkd", p.Len())
	c, err := p.Get()
	if err != nil {
		return nil, err
	}
	return c.(*beanstalk.Conn), nil
}

func Close(c interface{}) error {
	//fmt.Println("放回beanstalkd", p.Len())
	return p.Put(c)
}
