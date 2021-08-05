package rpc

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/rabbitmq"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RPCClient struct {
	ReplayRoutingKey string
}

func NewRPCClient(replyRoutingKey string) *RPCClient {
	return &RPCClient{ReplayRoutingKey: replyRoutingKey}
}

//NewCallContext
func NewCallContext(ctx context.Context, correlationId datasource.UUID, timeout ...time.Duration) context.Context {
	ctx = context.WithValue(ctx, "request_id", correlationId)
	if len(timeout) > 1 {
		ctx, _ = context.WithTimeout(ctx, timeout[0])
	}
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	return ctx
}

var callRespCH sync.Map

type callResp struct {
	ch      chan struct{}
	payload proto.Message
	err     error
}

func init() {
	exchange = rabbitmq.Exchange{
		Name: "rpc",
		Type: rabbitmq.ExchangeTypeFanout,
	}
}

func (r *RPCClient) CallRespDaemon(ctx context.Context) {
	routingKey := fmt.Sprintf("%s:resp:rpc", r.ReplayRoutingKey)

	exchange.Start(ctx, 1000, rabbitmq.Receiver{
		RouterKey: routingKey,
		QOSFn: func() (prefetchCount, prefetchSize int, global bool) {
			return 1000, 0, true
		},
		Exclusive: true,
		OnErrorFn: func(r rabbitmq.Receiver, err error) {
			logrus.Errorf("%s:%s error:%s", r.Name, r.RouterKey, err.Error())
		},
		OnReceiveFn: func(r rabbitmq.Receiver, msg amqp.Delivery) (ret bool) {
			ret = true
			ch, ok := callRespCH.Load(msg.CorrelationId)
			//fmt.Println("xxxxxxxxxxx", ok, msg.CorrelationId, string(msg.Body))
			if !ok {
				return
			}
			var err error
			ret = true
			defer func() {
				if e := recover(); e != nil {
					ch.(*callResp).err = err
					logrus.Error("%s:%s error:%s", r.Name, r.RouterKey, e)
				}
			}()
			defer func() {
				if err != nil {
					ch.(*callResp).err = err
				}
				ch.(*callResp).ch <- struct{}{}
			}()
			msgBodyArr := bytes.SplitN(msg.Body, []byte("|"), 2)
			if len(msgBodyArr) != 2 {
				err = fmt.Errorf("unknow response msg")
				return
			} else if errStr := string(msgBodyArr[0]); errStr == "1" {
				err = fmt.Errorf("%s", msgBodyArr[1])
				return
			} else if len(msgBodyArr[1]) == 0 {
				return
			}

			if err = proto.Unmarshal(msgBodyArr[1], ch.(*callResp).payload); err != nil {
				return
			}

			return
		},
	})
}

//Call 请求
func (r *RPCClient) Call(ctx context.Context, module string, method string, req proto.Message, resp proto.Message) (err error) {
	reqID := ctx.Value("request_id")
	if reqID == nil {
		err = fmt.Errorf("undefined request_id")
		return
	}
	routingKey := fmt.Sprintf("%s:req:rpc", module)

	var args []byte
	if args, err = proto.Marshal(req); err != nil {
		return
	}

	reqBody := []byte(fmt.Sprintf("%s|%s", method, args))
	ch := make(chan struct{}, 1)
	defer func() {
		close(ch)
	}()

	callResp := &callResp{
		ch:      ch,
		payload: resp,
	}
	callRespCH.Store(fmt.Sprintf("%d:%s", reqID.(datasource.UUID), method), callResp)
	if err = exchange.Publish(ctx, fmt.Sprintf("%s", routingKey), amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: fmt.Sprintf("%s:%s", reqID.(datasource.UUID).String(), method),
		Body:          reqBody,
		ReplyTo:       fmt.Sprintf("%s:resp:rpc", r.ReplayRoutingKey),
	}); err != nil {
		return
	}
	select {
	case <-ctx.Done():
		err = fmt.Errorf("timeout")
		return
	case <-ch:
		if callResp.err != nil {
			err = callResp.err
		} else {
			resp = callResp.payload
		}
		return
	}
}
