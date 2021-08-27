package websocket

import (
	"context"
	"github.com/Kotodian/protocol/interfaces"
	"go.uber.org/zap"
)

type ClientInterface interface {
	Send(msg []byte) error
	SubRegMQTT()
	SubMQTT()
	WritePump()
	ReadPump()
	Reply(ctx context.Context, payload interface{})
	ReplyError(ctx context.Context, err error, desc ...string)
	Publish(m MqttMessage)
	PublishReg(m MqttMessage)
	Close(err error) error
	ChargeStation() interfaces.ChargeStation
	Hub() *Hub
	KeepAlive() int64
	Logger() *zap.Logger
	RemoteAddress() string
}
