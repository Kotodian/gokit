package rpc

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Kotodian/gokit/datasource/rabbitmq"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type TReq struct {
	//ReceiveRoutingKey string
	//ReplayRoutingKey  string
	Func   func(message proto.Message) (resp proto.Message, err error)
	Method string
	Args   proto.Message
}

//type TReq struct {
//	Method  string
//	Payload proto.Message
//}

//type TResp struct {
//	Payload proto.Message
//	Err     error
//	Code    int
//}

var methods map[string]TReq

func Register(fn TReq) {
	if _, ok := methods[fn.Method]; ok {
		panic(fmt.Sprintf("undefind method:%s", fn.Method))
	}
	methods[fn.Method] = fn
}

func init() {
	//business.CommonTopicExchange.Start(context.TODO())
	methods = make(map[string]TReq)
}

var exchange rabbitmq.Exchange

func init() {
	exchange = rabbitmq.Exchange{
		Name: "rpc",
		Type: rabbitmq.ExchangeTypeTopic,
	}
}

//Server 服务端
func Server(ctx context.Context, module string) {
	exchange.Start(ctx, 1000, rabbitmq.Receiver{
		//Name:      routerKey,
		//RouterKey: fmt.Sprintf("req:%s:1", routerKey),
		RouterKey: fmt.Sprintf("%s:req:rpc", module),
		OnErrorFn: func(r rabbitmq.Receiver, err error) {
			logrus.Errorf("%s:%s error:%s", r.Name, r.RouterKey, err.Error())
		},
		QOSFn: func() (prefetchCount, prefetchSize int, global bool) {
			return 1000, 0, true
		},
		Exclusive: true,
		OnReceiveFn: func(r rabbitmq.Receiver, msg amqp.Delivery) (ret bool) {
			//fmt.Println("req msg:", msg.CorrelationId, msg.ReplyTo, string(msg.Body))
			ret = true
			var err error
			var respMsg proto.Message
			defer func() {
				var resp string
				if err != nil {
					logrus.Errorf("%s error:%v", r.RouterKey, err)
					resp = fmt.Sprintf("1|%s", err.Error())
				} else if respMsg == nil {
					resp = "0|"
				} else {
					_respMsg, _ := proto.Marshal(respMsg)
					resp = fmt.Sprintf("0|%s", string(_respMsg))
				}

				//respMsg, _ := json.Marshal(resp)
				//routingKey := fmt.Sprintf("resp:%s:%s", msg.ReplyTo, msg.CorrelationId)
				//fmt.Println("response msg:", resp, msg.CorrelationId, msg.ReplyTo)

				_ = exchange.Publish(ctx, msg.ReplyTo, amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: msg.CorrelationId,
					Body:          []byte(resp),
				})
			}()

			_body := bytes.SplitN(msg.Body, []byte("|"), 2)
			if len(_body) != 2 {
				err = fmt.Errorf("unknown msg")
				return
			}

			req := TReq{
				Method: string(_body[0]),
			}

			if method, ok := methods[req.Method]; !ok {
				err = fmt.Errorf("undefined method:%s", req.Method)
				return
			} else {
				req.Args = proto.Clone(method.Args)
				if err = proto.Unmarshal(_body[1], req.Args); err != nil {
					return
				} else if respMsg, err = method.Func(req.Args); err != nil {
					return
				}
			}
			return
		},
	})
}
