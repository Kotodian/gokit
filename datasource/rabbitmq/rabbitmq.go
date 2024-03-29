package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/publisher"
	"github.com/rabbitmq/amqp091-go"
)

var dialer *amqpextra.Dialer
var P *publisher.Publisher

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
	P, err = dialer.Publisher()
	if err != nil {
		panic(err)
	}
}

func Publish(ctx context.Context,
	key string,
	headers map[string]interface{}, body interface{}) error {
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return P.Publish(publisher.Message{
		Context: ctx,
		Key:     key,
		Publishing: amqp091.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        bytes,
		},
	})
}

func PublishJSON(ctx context.Context,
	key string, headers map[string]interface{}, body []byte) error {
	return P.Publish(publisher.Message{
		Context: ctx,
		Key:     key,
		Publishing: amqp091.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
		},
	})
}

func Consume(ctx context.Context, queue string, handler func(ctx context.Context, msg amqp091.Delivery) interface{}) error {
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
