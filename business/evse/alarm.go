package evse

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Kotodian/gokit/business"
	"github.com/Kotodian/gokit/datasource"
	pCharger "github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/streadway/amqp"
)

type Alarm struct {
	EvseID  datasource.UUID     `json:"evse_id"` //设备ID
	Payload pCharger.WarningReq `json:"payload"` //告警内容
}

//PutAlarm 增加告警（队列做持久化）
func PutAlarm(evseID datasource.UUID, payload pCharger.WarningReq) error {
	alarm := Alarm{
		EvseID:  evseID,
		Payload: payload,
	}
	b, _ := json.Marshal(alarm)
	return business.CommonTopicDurableExchange.Publish(context.TODO(), "evse_alarm", amqp.Publishing{
		Body:      b,
		Timestamp: time.Now().Local(),
	})
}
