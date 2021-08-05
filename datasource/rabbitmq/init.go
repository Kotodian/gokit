package rabbitmq

import (
	"context"
	"fmt"

	"github.com/Kotodian/gokit/workpool"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.uber.org/atomic"
)

type ExchangeType string

const (
	ExchangeTypeFanout          ExchangeType = "fanout"
	ExchangeTypeTopic           ExchangeType = "topic"
	ExchangeTypeDirect          ExchangeType = "direct"
	ExchangeTypeHeaders         ExchangeType = "headers"
	ExchangeTypeXDelayedMessage ExchangeType = "x-delayed-message"
)

func (e ExchangeType) String() string {
	return string(e)
}

// RabbitMQListener 用于管理和维护rabbitmq的对象
type Exchange struct {
	Name       string       // exchange的名称
	Type       ExchangeType // exchange的类型
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
	Delayed    bool
	//receivers    sync.Map
	//receivers    []Receiver
}

// New 创建一个新的操作RabbitMQ的对象
func NewExchange(exchangeType ExchangeType, exchangeName string) (rmq *Exchange, err error) {
	rmq = &Exchange{}
	//if rmq.channel, err = rmq.conn.Channel(); err != nil {
	//	return
	//}
	rmq.Name = exchangeName
	rmq.Type = exchangeType
	return
}

// Start 启动Rabbitmq的客户端
func (mq *Exchange) Start(ctx context.Context, workerNum int, receiver Receiver) {
	var err error
	logEntry := logrus.WithFields(logrus.Fields{
		"ex":    mq.Name,
		"act":   "listen",
		"queue": receiver.Name,
	})
	ctx = context.WithValue(ctx, "log", logEntry)
	defer func() {
		if err != nil {
			logEntry.Error(err.Error())
		}
	}()

	//isBreak := func(rawCtx context.Context) (isBreak bool) {
	if _, ch := mq.GetConnAndChannelWithContext(ctx); ch == nil {
		var closeFn func()
		if ctx, closeFn, err = mq.NewConnAndChannelIfNotExists(ctx); err != nil {
			logEntry.Errorf("new rabbitmq connection err:%s", err)
			return
		}
		defer func() {
			closeFn()
		}()
	}
	ctx, receiver.QuitFunc = context.WithCancel(ctx)
	receiver.QuitCH = make(chan struct{})
	defer func() {
		if receiver.QuitFunc != nil {
			receiver.QuitFunc()
		}
		close(receiver.QuitCH)
	}()

	//var conn *amqp.Connection
	_, ch := mq.GetConnAndChannelWithContext(ctx)
	//receiver.SetConnNotify(conn.NotifyClose(make(chan *amqp.Error)))
	//receiver.SetChannelNotify(channel.NotifyClose(make(chan *amqp.Error)))

	args := make(amqp.Table)
	if mq.Delayed {
		args["x-delayed-type"] = "topic"

		// 申明Exchange
		if err = ch.ExchangeDeclare(
			mq.Name,             // exchange
			"x-delayed-message", // type
			mq.Durable,          // durable
			mq.AutoDelete,       // autoDelete
			mq.Internal,         // internal
			mq.NoWait,           // noWait
			args,
		); err != nil {
			return
		}
	} else {
		// 申明Exchange
		if err = ch.ExchangeDeclare(
			mq.Name,          // exchange
			mq.Type.String(), // type
			mq.Durable,       // durable
			mq.AutoDelete,    // autoDelete
			mq.Internal,      // internal
			mq.NoWait,        // noWait
			args,
		); err != nil {
			return
		}
	}

	if err = mq.BindQueue(ctx, receiver); err != nil {
		return
	}

	prefetchCount, prefetchSize, global := receiver.QOS()
	// 获取消费通道
	_ = ch.Qos(prefetchCount, prefetchSize, global) // 确保rabbitmq会一个一个发消息

	var delivery <-chan amqp.Delivery
	delivery, err = ch.Consume(
		receiver.Name,      // queue
		"",                 // consumer tag
		receiver.AutoAck,   // auto-ack
		receiver.Exclusive, // exclusive
		receiver.NoLocal,   // no-local
		receiver.NoWait,    // no-wait
		//receiver.Args,      // args
		nil,
	)

	if err != nil {
		//fmt.Errorf("获取队列 %s 的消费通道失败: %s", receiver.Name, err.Error()))
		return
	}
	for {
		if mq.Run(ctx, receiver, workerNum, delivery) {
			break
		}
	}

}

func (mq *Exchange) Run(ctx context.Context, receiver Receiver, workerNum int, delivery <-chan amqp.Delivery) (isBreak bool) {
	//logEntry := ctx.Value("log").(*logrus.Entry)
	//rl := ratelimit.New(10, time.Second)
	wp := workpool.New(workerNum, 2*workerNum).Start()
	defer wp.Stop()
	i := atomic.NewInt32(0)

	for {
		select {
		case <-receiver.QuitCH:
			isBreak = true
			return
		case <-ctx.Done():
			isBreak = true
			return
		case msg := <-delivery:
			if !receiver.AutoAck {
				wp.PushTaskFunc(func(w *workpool.WorkPool, args ...interface{}) workpool.Flag {
					ok := receiver.OnReceiveFn(receiver, msg)
					if ok {
						_ = msg.Ack(false)
					} else {
						//当reject的消息过多的时候，就退出整个进程
						_ = msg.Reject(true)
					}
					if i.Add(1) == int32(1000) {
						receiver.QuitCH <- struct{}{}
					}
					return workpool.FLAG_OK
				})
			}
			//} else {
			//	加到死信队列
			//	if err = ch.Publish(mq.Name, receiver.Name, false, false, amqp.Publishing{
			//		ContentType:  "text/plain",
			//		Body:         msg.Body,
			//		MessageId:    msg.MessageId,
			//		DeliveryMode: 2,
			//	}); err != nil {
			//		logrus.Errorf("add to dlx error:%s", err.Error())
			//	}
			//	_ = msg.Reject(true)
			//}
		}
	}
}

func (mq *Exchange) BindQueue(ctx context.Context, receiver Receiver) (err error) {
	//var conn *amqp.Connection
	var ch *amqp.Channel
	_, ch = mq.GetConnAndChannelWithContext(ctx)
	//receiver.SetConnNotify(conn.NotifyClose(make(chan *amqp.Error)))
	//receiver.SetChannelNotify(channel.NotifyClose(make(chan *amqp.Error)))

	// 这里获取每个接收者需要监听的队列和路由
	queueName := receiver.Name
	routerKey := receiver.RouterKey

	// 申明Queue
	_, err = ch.QueueDeclare(
		queueName,           // name
		receiver.Durable,    // durable
		receiver.AutoDelete, // delete when usused
		receiver.Exclusive,  // exclusive(排他性队列)
		receiver.NoWait,     // no-wait
		receiver.Args,       // arguments
	)
	if nil != err {
		// 当队列初始化失败的时候，需要告诉这个接收者相应的错误
		return
	}

	// 将Queue绑定到Exchange上去
	err = ch.QueueBind(
		queueName,       // queue name
		routerKey,       // routing key
		mq.Name,         // exchange
		receiver.NoWait, // no-wait
		nil,
		//receiver.Args,
	)
	if nil != err {
		return
	}
	return nil
}

//
//func (mq *Exchange) reconnect(ctx context.Context, workerNum int, receiver IReceiver) {
//	for {
//		connNotify := receiver.ConnNotify()
//		channelNotify := receiver.ChannelNotify()
//		select {
//		case err := <-connNotify:
//			if err != nil {
//				logrus.Error("rabbitmq consumer - connection NotifyClose: ", err)
//			}
//		case err := <-channelNotify:
//			if err != nil {
//				logrus.Error("rabbitmq consumer - channel NotifyClose: ", err)
//			}
//		}
//
//		close(connNotify)
//		close(channelNotify)
//
//		 backstop
//		if !receiver.Conn().IsClosed() {
//			// close message delivery
//			if err := receiver.Channel().Cancel(receiver.ConsumeName(), true); err != nil {
//				logrus.Error("rabbitmq consumer - channel cancel failed: ", err)
//			}
//
//			if err := receiver.Conn().Close(); err != nil {
//				logrus.Error("rabbitmq consumer - channel cancel failed: ", err)
//			}
//		}
//
//		 //IMPORTANT: 必须清空 Notify，否则死连接不会释放
//		for err := range channelNotify {
//			logrus.Error(err)
//		}
//		for err := range connNotify {
//			logrus.Error(err)
//		}
//
//	quit:
//		for {
//			select {
//			default:
//				logrus.Error("rabbitmq consumer - reconnect")
//
//				if err := mq.run(workerNum, receiver); err != nil {
//					logrus.Error("rabbitmq consumer - failCheck: ", err)
//
//					// sleep 5s reconnect
//					time.Sleep(time.Second * 5)
//					continue
//				}
//
//				break quit
//			}
//		}
//	}
//}

func (mq *Exchange) NewConnAndChannelIfNotExists(ctx context.Context) (newCtx context.Context, closeFn func(), err error) {
	var closeFns []func()
	closeFn = func() {
		for i := len(closeFns); i > 0; i-- {
			closeFns[i-1]()
		}
	}
	var conn *amqp.Connection
	if _c := ctx.Value("conn"); _c == nil {
		if conn, err = Conn(); err != nil {
			return
		}
		newCtx = context.WithValue(ctx, "conn", conn)
		defer func() {
			closeFns = append(closeFns, func() {
				_ = Close(conn)
			})
		}()
	} else {
		conn = _c.(*amqp.Connection)
		newCtx = ctx
	}
	if _ch := ctx.Value("channel"); _ch == nil {
		var ch *amqp.Channel
		if ch, err = conn.Channel(); err != nil {
			return
		} else {
			newCtx = context.WithValue(newCtx, "channel", ch)
			defer func() {
				closeFns = append(closeFns, func() {
					_ = ch.Close()
				})
			}()
		}
	} else {
		newCtx = ctx
	}
	return
}

func (mq *Exchange) GetConnAndChannelWithContext(ctx context.Context) (*amqp.Connection, *amqp.Channel) {
	if conn := ctx.Value("conn"); conn == nil {
		return nil, nil
	} else if ch := ctx.Value("channel"); ch == nil {
		return conn.(*amqp.Connection), nil
	} else {
		return conn.(*amqp.Connection), ch.(*amqp.Channel)
	}
}

func (mq *Exchange) Publish(ctx context.Context, key string, msg amqp.Publishing, delayTime ...int64) (err error) {
	//var conn *amqp.Connection
	var channel *amqp.Channel
	var fn func()
	ctx, fn, err = mq.NewConnAndChannelIfNotExists(ctx)
	defer func() {
		if fn != nil {
			fn()
		}
	}()
	if err != nil {
		return
	}
	_, channel = mq.GetConnAndChannelWithContext(ctx)

	if len(delayTime) > 0 {
		if msg.Headers == nil {
			msg.Headers = make(amqp.Table)
		}
		if mq.Delayed {
			msg.Headers["x-delay"] = fmt.Sprintf("%d000", delayTime[0])
		} else {
			delete(msg.Headers, "x-delay")
			msg.Expiration = fmt.Sprintf("%d000", delayTime[0])
		}
	}
	if err = channel.Publish(
		mq.Name,
		key,
		// 如果为true, 会根据exchange类型和routkey规则，如果无法找到符合条件的队列那么会把发送的消息返回给发送者
		false,
		// 如果为true, 当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息发还给发送者
		false,
		msg,
	); err != nil {
		return
	}
	return
}
