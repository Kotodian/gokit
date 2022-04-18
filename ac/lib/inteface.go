package lib

import (
	"context"
	"net"

	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/protocol/interfaces"
	"go.uber.org/zap"
)

type ClientInterface interface {
	// Send 直接发送消息
	Send(msg []byte) error
	// SubRegMQTT 订阅平台mqtt消息
	SubRegMQTT()
	// SubMQTT 订阅平台mqtt消息
	SubMQTT()
	// WritePump 往桩端写入消息
	WritePump()
	// ReadPump 从桩端读取消息
	ReadPump()
	// Reply 回复桩消息(将payload转换成协议消息发送)
	Reply(ctx context.Context, payload interface{})
	// ReplyError 回复桩错误(将err转换成协议消息发送)
	ReplyError(ctx context.Context, err error, desc ...string)
	// Publish 发送消息到平台
	Publish(m mqtt.MqttMessage)
	// PublishReg 发送消息到平台
	PublishReg(m mqtt.MqttMessage)
	// Close 关闭连接
	Close(err error) error
	// ChargeStation 桩实体
	ChargeStation() interfaces.ChargeStation
	// SetChargeStation 设置桩实体
	SetChargeStation(interfaces.ChargeStation)
	// Hub 所属的容器
	Hub() *Hub
	// KeepAlive 心跳时间
	KeepAlive() int64
	// Logger 日志服务
	Logger() *zap.Logger
	// RemoteAddress 客户端地址
	RemoteAddress() string
	// SetClientOfflineFunc 设置客户端离线通知
	SetClientOfflineFunc(func(err error))
	// ClientOfflineFunc 客户端离线通知
	ClientOfflineFunc() func(err error)
	// Lock 上锁
	Lock()
	// Unlock 解锁
	Unlock()
	// Coregw 需要回复消息的coregw
	Coregw() string
	// SetCoregw 设置coregw以便回复消息
	SetCoregw(coregw string)
	// IsClose 该客户端是否关闭
	IsClose() bool
	// EncryptKey 加密的公钥
	EncryptKey() []byte
	// SetEncryptKey 设置加密的公钥
	SetEncryptKey(string)
	// PingHandler 接收到ping时调用的函数
	PingHandler(msg string) error
	// SetCertificateSN 设置证书sn
	SetCertificateSN(sn string)
	// CertificateSN 证书sn
	CertificateSN() string
	MessageNumber() int16
	SetMessageNumber(int16)
	SetData(key, val interface{})
	SetKeepalive(int64)
	GetData(key interface{}) interface{}
	Conn() net.Conn
	SetRemoteAddress(address string)
	SetOrderInterval(int)
	OrderInterval() int
}
