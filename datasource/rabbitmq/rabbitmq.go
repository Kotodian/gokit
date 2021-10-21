package rabbitmq

import (
	"context"
	"github.com/makasim/amqpextra"
	"github.com/makasim/amqpextra/consumer"
	"github.com/makasim/amqpextra/publisher"
	"github.com/streadway/amqp"
	"os"
)

var dialer *amqpextra.Dialer

func Init() {
	var err error
	url := "amqp://" + os.Getenv("RABBITMQ_USER") + ":" +
		os.Getenv("RABBITMQ_PASS") + "@" +
		os.Getenv("RABBITMQ_POOL") + "/"
	dialer, err = amqpextra.NewDialer(amqpextra.WithURL(url))
	if err != nil {
		panic(err)
	}
}

func Publish(ctx context.Context,
	exchange string, key string,
	headers map[string]interface{}, body []byte) error {
	p, err := dialer.Publisher()
	if err != nil {
		return err
	}
	defer p.Close()
	return p.Publish(publisher.Message{
		Context:   ctx,
		Exchange:  exchange,
		Key:       key,
		Immediate: true,
		Publishing: amqp.Publishing{
			Headers:     headers,
			ContentType: "application/json",
			Body:        body,
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
