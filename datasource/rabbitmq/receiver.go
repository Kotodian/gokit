package rabbitmq

import (
	"context"

	"github.com/streadway/amqp"
)

type Receiver struct {
	//ConsumeName string
	Name        string
	RouterKey   string
	Durable     bool
	AutoDelete  bool
	NoWait      bool
	Exclusive   bool
	Args        amqp.Table
	NoLocal     bool
	AutoAck     bool
	OnErrorFn   func(r Receiver, err error)
	OnReceiveFn func(r Receiver, msg amqp.Delivery) bool
	QOSFn       func() (prefetchCount, prefetchSize int, global bool) //default prefetchCount=1, prefetchSize=0, global=true
	QuitFunc    context.CancelFunc
	QuitCH      chan struct{}
}

func (r Receiver) QOS() (prefetchCount, prefetchSize int, global bool) {
	if r.QOSFn != nil {
		return r.QOSFn()
	}
	return 30, 0, true
}

func (r Receiver) OnError(err error) {
	if r.OnErrorFn != nil {
		r.OnErrorFn(r, err)
	}
	return
}

func (r Receiver) OnReceive(msg amqp.Delivery) bool {
	if r.OnReceiveFn != nil {
		return r.OnReceiveFn(r, msg)
	}
	return false
}
