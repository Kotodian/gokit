package rabbitmq

import (
	"os"
	"time"

	"github.com/silenceper/pool"
	"github.com/streadway/amqp"
)

// MQURL 格式 amqp://账号：密码@rabbitmq服务器地址：端口号/vhost
// const MQURL = "amqp://guest:guest@rabbitmq-1:5672/"
var MQURL = "amqp://" + os.Getenv("RABBITMQ_USER") + ":" + os.Getenv("RABBITMQ_PASS") + "@" + os.Getenv("RABBITMQ_POOL") + "/"
var cp pool.Pool

func Init() {
	// go func() {
	var err error
	cp, err = pool.NewChannelPool(&pool.Config{
		InitialCap: 5,
		MaxIdle:    30,
		MaxCap:     2000,
		Factory: func() (interface{}, error) {
			return amqp.Dial(MQURL)
		},
		Close: func(v interface{}) error {
			return v.(*amqp.Connection).Close()
		},
		//链接最大空闲时间，超过该时间的链接 将会关闭，可避免空闲时链接EOF，自动失效的问题
		IdleTimeout: 30 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	// close(ch)
	// }()

	//从连接池中取得一个链接
	// v, err := p.Get()

	//do something
	//conn=v.(net.Conn)

	//将链接放回连接池中
	// p.Put(v)

	//释放连接池中的所有链接
	// p.Release()

	//查看当前链接中的数量
	// current := p.Len()
}

func Conn() (*amqp.Connection, error) {
	c, err := cp.Get()
	if err != nil {
		return nil, err
	}
	return c.(*amqp.Connection), nil
}

func Close(c interface{}) error {
	return cp.Put(c)
}

func NewChannelWithConn(c *amqp.Connection) (*amqp.Channel, error) {
	return c.Channel()
}

func NewChannel() (c *amqp.Connection, ch *amqp.Channel, err error) {
	if c, err = Conn(); err != nil {
		return
	} else if ch, err = NewChannelWithConn(c); err != nil {
		return
	}
	return
}
