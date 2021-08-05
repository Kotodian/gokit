package station

import (
	"github.com/Kotodian/gokit/datasource/rabbitmq"
)

var StationPushExchange rabbitmq.Exchange

//var ConnectorDelayExchange rabbitmq.Exchange

func init() {
	StationPushExchange = rabbitmq.Exchange{
		Name:    "station_push",
		Type:    rabbitmq.ExchangeTypeFanout,
		Durable: false,
	}
	//ConnectorDelayExchange = rabbitmq.Exchange{
	//	Name:    "connector_delay",
	//	Type:    rabbitmq.ExchangeTypeTopic,
	//	Durable: false,
	//	Delayed: true,
	//}
}

//
////Push 根据ID推送到第三方平台
//func Push(connectorID datasource.UUID, t time.Time, state int32) error {
//	return stationPushExchange.Publish(context.TODO(), "", amqp.Publishing{
//		ContentType:  "text/plain",
//		Body:         []byte(fmt.Sprintf("%d:%d:%d", connectorID, t.Unix(), state)),
//		DeliveryMode: 2,
//	})
//}

//ProcessOffline 处理设备离线引起的订单问题
//func ProcessOffline(connectorID datasource.UUID, t time.Time) error {
//	return ConnectorDelayExchange.Publish(context.TODO(), "offline", amqp.Publishing{
//		MessageId: connectorID.String(),
//	})
//}
