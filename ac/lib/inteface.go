package lib

import (
	"context"
	"github.com/Kotodian/gokit/datasource/mqtt"
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
	Publish(m mqtt.MqttMessage)
	PublishReg(m mqtt.MqttMessage)
	Close(err error) error
	ChargeStation() interfaces.ChargeStation
	Hub() *Hub
	KeepAlive() int64
	Logger() *zap.Logger
	RemoteAddress() string
	SetClientOfflineFunc(func(err error))
	ClientOfflineFunc() func(err error)
	Lock()
	Unlock()
	Coregw() string
	SetCoregw(coregw string)
	IsClose() bool
	EncryptKey() []byte
	SetEncryptKey(string)
}
