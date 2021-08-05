package rabbitmq

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func Test_ReceiverDeadletter(t *testing.T) {
	dlxEx := &Exchange{
		Name:    "test_delayed",
		Type:    ExchangeTypeTopic,
		Durable: true,
		Delayed: true,
	}

	go func() {
		dlxEx.Start(context.TODO(), 100, Receiver{
			Name:      "test_dlx_delay",
			RouterKey: "delay:dlx:test",
			OnReceiveFn: func(r Receiver, msg amqp.Delivery) bool {
				fmt.Println("xxxxxxxxxxx", fmt.Sprintf("%+v", msg.MessageId))
				return true
			},
		})
	}()

	//go func() {
	//	dlxEx.Start(context.TODO(), 1, Receiver{
	//		Name:      "test_dlx",
	//		RouterKey: "dlx:test",
	//		OnReceiveFn: func(r Receiver, msg amqp.Delivery) bool {
	//			fmt.Println("ddddddd  ss", fmt.Sprintf("%+v", msg))
	//			return true
	//		},
	//	})
	//}()

	fmt.Println("1")
	dlxEx.Publish(context.TODO(), "delay:dlx:test", amqp.Publishing{
		MessageId: "10",
	})

	fmt.Println("2")
	dlxEx.Publish(context.TODO(), "delay:dlx:test", amqp.Publishing{
		MessageId: "3",
	})

	time.Sleep(100 * time.Second)
}

func Test_ReceiverRouterKey(t *testing.T) {
	dlxEx := &Exchange{
		Name:    "dlx",
		Type:    ExchangeTypeFanout,
		Durable: true,
	}

	go func() {
		dlxEx.Start(context.TODO(), 1, Receiver{
			Name: "dlx rc",
			OnReceiveFn: func(r Receiver, b amqp.Delivery) bool {
				fmt.Println(r.Name, string(b.Body))
				return true
			},
		})
	}()

	ex := &Exchange{
		Name:    "test1",
		Type:    ExchangeTypeTopic,
		Durable: true,
	}
	rc := Receiver{Name: "rc", RouterKey: "rc.*", Args: amqp.Table{"x-dead-letter-exchange": "dlx"}}
	rc2 := Receiver{
		Name:      "rc2",
		RouterKey: "rc2.*",
		OnReceiveFn: func(r Receiver, b amqp.Delivery) bool {
			fmt.Println(r.Name, "got message:", string(b.Body))
			time.Sleep(5 * time.Second)
			fmt.Println(r.Name, "done message:", string(b.Body))
			return true
		},
		QOSFn: func() (prefetchCount, prefetchSize int, global bool) {
			return 2, 0, true
		},
	}
	ctx, closeFn, _ := ex.NewConnAndChannelIfNotExists(context.TODO())
	defer func() {
		closeFn()
	}()
	_ = ex.BindQueue(ctx, rc)
	//go func() {
	//	ex.Start(context.TODO(), 1, rc)
	//}()
	go func() {
		ex.Start(context.TODO(), 2, rc2)
	}()
	time.Sleep(time.Second)

	if err := ex.Publish(context.TODO(), "rc2.1", amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte("hello world"),
		DeliveryMode: 2,
		MessageId:    "1",
		//Expiration:   "5000",
	}); err != nil {
		t.Fatal(err.Error())
	}

	if err := ex.Publish(context.TODO(), "rc2.1", amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte("edwardhey"),
		DeliveryMode: 2,
		MessageId:    "2",
		//Expiration:   "5000",
	}); err != nil {
		t.Fatal(err.Error())
	}
	//ex.PublishSimple(rc, "test")
	//fmt.Println(rc.QueueName())
	time.Sleep(100 * time.Second)
}
