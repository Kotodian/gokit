package order

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Kotodian/gokit/business"
	businessFeed "github.com/Kotodian/gokit/business/feed"
	"github.com/Kotodian/gokit/datasource"
	"github.com/streadway/amqp"
)

func AddEvent(orderID datasource.UUID, msg string, act ...string) {
	_act := "others"
	if len(act) > 0 {
		_act = act[0]
	}
	feedQueue := businessFeed.Queue{
		Payload: map[string]interface{}{
			"act": _act,
		},
		OrderID: datasource.NewUUID(orderID),
		Time:    time.Now().Local(),
		Msg:     msg,
		Kind:    businessFeed.KindCharge,
	}
	feedJson, _ := json.Marshal(feedQueue)
	_ = business.CommonTopicExchange.Publish(context.TODO(), "feed", amqp.Publishing{
		Body: feedJson,
	})

}

func AddEventWithPayload(orderID datasource.UUID, t time.Time, msg string, err error, payload map[string]interface{}) {
	//b, _ := json.Marshal(fmt.Sprintf("%d|%s", orderID, msg))
	//_ = business.CommonTopicExchange.Publish(context.TODO(), "order:event", amqp.Publishing{
	//	Body: b,
	//})
	feedQueue := businessFeed.Queue{
		Payload: payload,
		OrderID: datasource.NewUUID(orderID),
		Time:    t,
		Msg:     msg,
		Kind:    businessFeed.KindCharge,
		Error:   businessFeed.FormatError(err),
	}
	feedJson, _ := json.Marshal(feedQueue)
	_ = business.CommonTopicExchange.Publish(context.TODO(), "feed", amqp.Publishing{
		Body: feedJson,
	})

}
