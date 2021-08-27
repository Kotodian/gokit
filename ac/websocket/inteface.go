package websocket

import "context"

type ClientInterface interface {
	Send(msg []byte) error
	SubRegMQTT()
	SubMQTT()
	ReadPump()
	Reply(ctx context.Context, payload interface{})
	ReplyError(ctx context.Context, err error, desc ...string)
	Publish(m MqttMessage)
	PublishReg(m MqttMessage)
}
