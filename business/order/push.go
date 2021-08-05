package order

import (
	"context"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/rabbitmq"
	"github.com/streadway/amqp"
)

var OrderPushExchange rabbitmq.Exchange
var OrderPartnerPushExchange rabbitmq.Exchange
var OrderDelayExchange rabbitmq.Exchange

func init() {
	OrderPushExchange = rabbitmq.Exchange{
		Name:    "order_push",
		Type:    rabbitmq.ExchangeTypeTopic,
		Durable: true,
	}
	OrderPartnerPushExchange = rabbitmq.Exchange{
		Name:    "order_push_partner",
		Type:    rabbitmq.ExchangeTypeFanout,
		Durable: true,
	}
	OrderDelayExchange = rabbitmq.Exchange{
		Name:    "order_delay",
		Type:    rabbitmq.ExchangeTypeTopic,
		Durable: true,
		Delayed: true,
	}
}

//Push 根据ID推送到第三方平台
func Push(orderID datasource.UUID, isLowPriority bool) error {
	//order_push_low_priority
	key := "id:priority:high"
	if isLowPriority {
		key = "id:priority:low"
	}
	return OrderPushExchange.Publish(context.TODO(), key, amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(orderID.String()),
		DeliveryMode: 2,
	})
}
