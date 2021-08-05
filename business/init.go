package business

import "github.com/Kotodian/gokit/datasource/rabbitmq"

var CommonTopicExchange rabbitmq.Exchange
var CommonTopicDurableExchange rabbitmq.Exchange
var CommonTopicDelayedExchange rabbitmq.Exchange
var CommonTopicDurableDelayedExchange rabbitmq.Exchange
var XAExchange rabbitmq.Exchange

func init() {
	CommonTopicExchange = rabbitmq.Exchange{
		Name: "common",
		Type: rabbitmq.ExchangeTypeTopic,
	}
	CommonTopicDurableExchange = rabbitmq.Exchange{
		Name:    "common:durable",
		Type:    rabbitmq.ExchangeTypeTopic,
		Durable: true,
	}
	CommonTopicDelayedExchange = rabbitmq.Exchange{
		Name:    "common:delay",
		Type:    rabbitmq.ExchangeTypeTopic,
		Delayed: true,
	}
	CommonTopicDurableDelayedExchange = rabbitmq.Exchange{
		Name:    "common:delay:durable",
		Type:    rabbitmq.ExchangeTypeTopic,
		Delayed: true,
		Durable: true,
	}
	XAExchange = rabbitmq.Exchange{
		Name:    "xa:delay",
		Type:    rabbitmq.ExchangeTypeTopic,
		Delayed: true,
		Durable: true,
	}
}
