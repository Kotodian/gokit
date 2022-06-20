package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
)

var dialer *amqpextra.Dialer

func Init() {
	var err error
	urls := make([]string, 0)
	for _, v := range strings.Split(os.Getenv("RABBITMQ_POOL"), ",") {
		urls = append(urls, "amqp://"+os.Getenv("RABBITMQ_USER")+":"+
			os.Getenv("RABBITMQ_PASS")+"@"+
			v+"/")
	}

	dialer, err = amqpextra.NewDialer(amqpextra.WithURL(urls...))
	if err != nil {
		panic(err)
	}
}

func Publish(ctx context.Context,
	key string,
	headers map[string]interface{}, body interface{}) error {
	p, err := dialer.Publisher()
	if err != nil {
		return err
	}
	defer p.Close()
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return p.Publish(publisher.Message{
		Context: ctx,
		Key:     key,
		Publishing: amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        bytes,
		},
	})
}

func PublishJSON(ctx context.Context,
	key string, headers map[string]interface{}, body []byte) error {
	p, err := dialer.Publisher()
	if err != nil {
		return err
	}
	defer p.Close()
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return p.Publish(publisher.Message{
		Context: ctx,
		Key:     key,
		Publishing: amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        bytes,
		},
	})
}

func Consume(ctx context.Context, queue string, handler func(ctx context.Context, msg amqp.Delivery) interface{}) error {
	h := consumer.HandlerFunc(handler)
	_, err := dialer.Consumer(
		consumer.WithContext(ctx),
		consumer.WithQueue(queue),
		consumer.WithHandler(h),
	)
	if err != nil {
		return err
	}
	return nil
}
