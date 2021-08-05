package evse

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Kotodian/gokit/business"
	businessFeed "github.com/Kotodian/gokit/business/feed"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/rabbitmq"
	"github.com/streadway/amqp"
)

var EvseExchange rabbitmq.Exchange
var EvseDelayExchange rabbitmq.Exchange
var EvseFanoutExchange rabbitmq.Exchange

func init() {
	EvseExchange = rabbitmq.Exchange{
		Name:    "evse",
		Type:    rabbitmq.ExchangeTypeTopic,
		Durable: false,
	}
	EvseFanoutExchange = rabbitmq.Exchange{
		Name:    "evse_fanout",
		Type:    rabbitmq.ExchangeTypeFanout,
		Durable: false,
	}
	EvseDelayExchange = rabbitmq.Exchange{
		Name:    "evse:delay",
		Type:    rabbitmq.ExchangeTypeTopic,
		Durable: false,
		Delayed: true,
	}
}

type State struct {
	EvseID datasource.UUID `json:"evse_id,string"`
	Time   time.Time       `json:"time"`
	State  int32           `json:"state"`
}

//SyncState 同步队列
func SyncState(s State) error {
	if s.State == 0 {
		return EvseDelayExchange.Publish(context.TODO(), "offline_state", amqp.Publishing{
			MessageId: s.EvseID.String(),
		}, 120)
	} else {
		return EvseDelayExchange.Publish(context.TODO(), "online_state", amqp.Publishing{
			MessageId: s.EvseID.String(),
		})
	}

	//b, _ := json.Marshal(s)
	//return EvseExchange.Publish(context.TODO(), "state", amqp.Publishing{
	//	Body: b,
	//ContentType:  "text/plain",
	//DeliveryMode: 2,
	//})
}

//UpdateLicense 更新证书
func UpdateLicense(evseID datasource.UUID) error {
	b, _ := json.Marshal(evseID)
	return EvseExchange.Publish(context.TODO(), "evse_license", amqp.Publishing{
		Body: b,
		//ContentType:  "text/plain",
		//DeliveryMode: 2,
	})
}

//AddEvseFeed 增加设备Feed
func AddEvseFeed(evseID datasource.UUID, t time.Time, msg string) {
	feedQueue := businessFeed.Queue{
		EvseID: datasource.NewUUID(evseID),
		Time:   t,
		Kind:   businessFeed.KindEvse,
		Msg:    msg,
	}
	feedJson, _ := json.Marshal(feedQueue)
	_ = business.CommonTopicExchange.Publish(context.TODO(), "feed", amqp.Publishing{
		Body: feedJson,
	})
}

//AddConnectorFeed 增加设备Feed
func AddConnectorFeed(evseID datasource.UUID, connectorID datasource.UUID, t time.Time, msg string) {
	feedQueue := businessFeed.Queue{
		EvseID:      datasource.NewUUID(evseID),
		Time:        t,
		Kind:        businessFeed.KindEvse,
		ConnectorID: datasource.NewUUID(connectorID),
		Msg:         msg,
	}
	feedJson, _ := json.Marshal(feedQueue)
	_ = business.CommonTopicExchange.Publish(context.TODO(), "feed", amqp.Publishing{
		Body: feedJson,
	})
}
