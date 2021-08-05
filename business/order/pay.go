package order

import (
	"github.com/Kotodian/gokit/datasource/rabbitmq"
)

var (
	PayExchange = &rabbitmq.Exchange{Name: "pay", Type: rabbitmq.ExchangeTypeTopic, Durable: true}
)
