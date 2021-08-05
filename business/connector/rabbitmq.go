package connector

import (
	"context"
	"fmt"
	"time"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/rabbitmq"
	"github.com/streadway/amqp"
)

var connectorPushExchange rabbitmq.Exchange
var ConnectorDelayExchange rabbitmq.Exchange

func init() {
	connectorPushExchange = rabbitmq.Exchange{
		Name:    "connector_push",
		Type:    rabbitmq.ExchangeTypeFanout,
		Durable: false,
	}
	ConnectorDelayExchange = rabbitmq.Exchange{
		Name:    "connector_delay",
		Type:    rabbitmq.ExchangeTypeTopic,
		Durable: true,
		Delayed: true,
	}
}

//Push 根据ID推送到第三方平台
func Push(connectorID datasource.UUID, t time.Time, state int32) error {
	return connectorPushExchange.Publish(context.TODO(), "", amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(fmt.Sprintf("%d:%d:%d", connectorID, t.Unix(), state)),
		DeliveryMode: 2,
	})
}

//ProcessOffline 处理设备离线引起的订单问题
//func ProcessOffline(connectorID datasource.UUID, t time.Time) error {
//	return ConnectorDelayExchange.Publish(context.TODO(), "offline", amqp.Publishing{
//		MessageId: connectorID.String(),
//	})
//}
