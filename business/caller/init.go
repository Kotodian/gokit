package caller

import "github.com/Kotodian/gokit/datasource/rabbitmq"

var Exchange rabbitmq.Exchange

func init() {
	Exchange = rabbitmq.Exchange{
		Name:    "caller_push_delay",
		Type:    rabbitmq.ExchangeTypeTopic,
		Delayed: true,
		Durable: true,
	}
}
