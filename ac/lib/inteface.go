package lib

import (
	"context"
	"net"

	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/protocol/interfaces"
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
	// Encrypt
	Encrypt() Encrypt
	// SetEncrypt
	SetEncrypt(Encrypt)
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
	SetBaseURL(string)
	BaseURL() string
}

type testClient struct {
	chargingStation interfaces.ChargeStation
	encrypt         Encrypt
	encryptKey      []byte
} // Send 直接发送消息

func NewTestClient() ClientInterface {
	return &testClient{
		chargingStation: interfaces.NewDefaultChargeStation("test", true, 0),
	}
}
func (*testClient) Send(msg []byte) error {
	return nil
}

// SubRegMQTT 订阅平台mqtt消息
func (*testClient) SubRegMQTT() {

}

// SubMQTT 订阅平台mqtt消息
func (*testClient) SubMQTT() {

}

// WritePump 往桩端写入消息
func (*testClient) WritePump() {

}

// ReadPump 从桩端读取消息
func (*testClient) ReadPump() {

}

// Reply 回复桩消息(将payload转换成协议消息发送)
func (*testClient) Reply(ctx context.Context, payload interface{}) {

}

// ReplyError 回复桩错误(将err转换成协议消息发送)
func (*testClient) ReplyError(ctx context.Context, err error, desc ...string) {

}

// Publish 发送消息到平台
func (*testClient) Publish(m mqtt.MqttMessage) {

}

// PublishReg 发送消息到平台
func (*testClient) PublishReg(m mqtt.MqttMessage) {

}

// Close 关闭连接
func (*testClient) Close(err error) error {
	return nil
}

// ChargeStation 桩实体
func (t *testClient) ChargeStation() interfaces.ChargeStation {
	return t.chargingStation
}

// SetChargeStation 设置桩实体
func (t *testClient) SetChargeStation(chargingStation interfaces.ChargeStation) {
	t.chargingStation = chargingStation
}

// Hub 所属的容器
func (t *testClient) Hub() *Hub {
	return nil
}

// KeepAlive 心跳时间
func (t *testClient) KeepAlive() int64 {
	return 60
}

// RemoteAddress 客户端地址
func (t *testClient) RemoteAddress() string {
	return "127.0.0.1"
}

// SetClientOfflineFunc 设置客户端离线通知
func (*testClient) SetClientOfflineFunc(_ func(err error)) {

}

// ClientOfflineFunc 客户端离线通知
func (*testClient) ClientOfflineFunc() func(err error) {
	return nil
}

// Lock 上锁
func (t *testClient) Lock() {

}

// Unlock 解锁
func (t *testClient) Unlock() {

}

// Coregw 需要回复消息的coregw
func (t *testClient) Coregw() string {
	return ""
}

// SetCoregw 设置coregw以便回复消息
func (t *testClient) SetCoregw(coregw string) {

}

// IsClose 该客户端是否关闭
func (t *testClient) IsClose() bool {
	return false
}

// Encrypt
func (t *testClient) Encrypt() Encrypt {
	return t.encrypt
}

// SetEncrypt
func (t *testClient) SetEncrypt(encrypt Encrypt) {
	t.encrypt = encrypt
}

// EncryptKey 加密的公钥
func (t *testClient) EncryptKey() []byte {
	return t.encryptKey
}

// SetEncryptKey 设置加密的公钥
func (t *testClient) SetEncryptKey(key string) {
	t.encryptKey = []byte(key)
}

// PingHandler 接收到ping时调用的函数
func (*testClient) PingHandler(msg string) error {
	return nil
}

// SetCertificateSN 设置证书sn
func (t *testClient) SetCertificateSN(sn string) {

}

// CertificateSN 证书sn
func (t *testClient) CertificateSN() string {
	return ""
}

func (t *testClient) MessageNumber() int16 {
	return 0
}

func (t *testClient) SetMessageNumber(_ int16) {

}

func (t *testClient) SetData(key interface{}, val interface{}) {

}

func (t *testClient) SetKeepalive(_ int64) {

}

func (t *testClient) GetData(key interface{}) interface{} {
	return nil
}

func (t *testClient) Conn() net.Conn {
	return nil
}

func (t *testClient) SetRemoteAddress(address string) {

}

func (t *testClient) SetOrderInterval(_ int) {

}

func (t *testClient) OrderInterval() int {
	return 30
}

func (t *testClient) SetBaseURL(_ string) {

}

func (t *testClient) BaseURL() string {
	return "jxcsmsuat.joysonquin.com"
}
