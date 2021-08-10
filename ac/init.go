package ac

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/gokit/datasource/redis"
	pCharger "github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/golang/protobuf/proto"
	rdlib "github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

var (
	ErrDataFormatKick = errors.New("data format error kick")
	ErrNotAuthKick    = errors.New("not auth error kick")
)

// Ctx ac上下文
type Ctx struct {
	Raw         interface{}
	evse        string
	EvseOffset  int
	Data        map[string]interface{}
	Log         *log.Entry
	Ignore      bool
	TopicSuffix string
	IsTelemetry bool
	Retained    bool
	APDU        *pCharger.APDU
	AfterFuncs  []func()
}

// IMqttBase Mqtt协议相关
type IMqttBase interface {
	//GetTopicPrefix mqtt topic的路径前缀，用于标示协议
	GetTopicPrefix() string

	//SetConnection 设置连接句柄
	SetMQTTConnection(client *mqtt.MQTTClient)

	//GetMQTTConnection 获取连接句柄
	GetMQTTConnection() *mqtt.MQTTClient

	//GetEvsePubRegisterTopic 获取设备注册发送的topic
	GetEvsePubRegisterTopic() string

	//GetEvseSubRegisterTopic 获取设备注册监听topic
	GetEvseSubRegisterTopic(string) string

	//GetEvseSubTopic 获取设备常规消息监听topic
	GetEvseSubTopic(ID string) string

	//GetEvsePubTopic 获取设备常规消息发送的topic
	GetEvsePubTopics() map[string]byte

	//Getevse 获取设备ID
	GetevseWithTopic(string) string
}

// ITranslateTCP mqtt->tcp基本转发
type ITranslateTCP interface {
	IMqttBase

	//ToAPDU 发送到core的protobuf的消息
	ToAPDU(ctx *Ctx) (to interface{}, ret interface{}, err error)

	//FromAPDU 从core的protobuf来的消息
	FromAPDU(ctx *Ctx) (to interface{}, ret interface{}, err error)
}

// ToApduFunc 设备发送
type ToApduFunc func(ctx *Ctx) (to proto.Message, err error)

// FromApduFunc 后台发送
type FromApduFunc func(ctx *Ctx) (to interface{}, err error)

// ITranslateBase mqtt基本通讯(mqtt->mqtt, tcp->mqtt)
type ITranslateBase interface {
	IMqttBase

	//ToAPDU 发送到core的protobuf的消息
	ToAPDU(ctx *Ctx) (ToApduFunc, error)

	//FromAPDU 从core的protobuf来的消息
	FromAPDU(ctx *Ctx) (FromApduFunc, error)
}

//ITranslate 协议翻译类
type ITranslate interface {
	ITranslateBase

	ErrorToAPDU(ctx *Ctx) (to proto.Message, err error)

	ErrorToDevice(ctx *Ctx) (to interface{}, err error)

	//开机启动请求
	BootReq(ctx *Ctx) (to proto.Message, err error)
	//开机启动回复
	BootConf(ctx *Ctx) (to interface{}, err error)

	RemoteStopConf(ctx *Ctx) (to proto.Message, err error)
	RemoteStopReq(ctx *Ctx) (to interface{}, err error)

	RemoteStartConf(ctx *Ctx) (to proto.Message, err error)
	RemoteStartReq(ctx *Ctx) (to interface{}, err error)

	StopTransactionReq(ctx *Ctx) (to proto.Message, err error)
	StopTransactionConf(ctx *Ctx) (to interface{}, err error)

	StartTransactionReq(ctx *Ctx) (to proto.Message, err error)
	StartTransactionConf(ctx *Ctx) (to interface{}, err error)

	RemoteControlConf(ctx *Ctx) (to proto.Message, err error)
	RemoteControlReq(ctx *Ctx) (to interface{}, err error)

	SetTariffConf(ctx *Ctx) (to proto.Message, err error)
	SetTariffReq(ctx *Ctx) (to interface{}, err error)

	TelemetryReq(ctx *Ctx) (to proto.Message, err error)
	ChargingInfoReq(ctx *Ctx) (to proto.Message, err error)

	UpdateFirmwareReq(ctx *Ctx) (to interface{}, err error)
	//UpdateFirmwareConf(ctx *Ctx) (to proto.Message, err error)

	TriggerReq(ctx *Ctx) (to interface{}, err error)
}

var sessions sync.Map

// Session ....
type Session struct {
	CH  chan interface{}
	Key string
}

// NewSession 创建
func NewSession(key string) *Session {
	s := &Session{
		CH:  make(chan interface{}, 1),
		Key: key,
	}
	sessions.Store(key, s)
	return s
}

// LoadSession 加载
func LoadSession(key string) *Session {
	if s, ok := sessions.Load(key); ok {
		return s.(*Session)
	}
	return nil
}

// Listen 监听
func (sess *Session) Listen(timeout time.Duration) (ret interface{}, err error) {
	select {
	case <-time.After(timeout):
		err = fmt.Errorf("timeout")
		return
	case ret = <-sess.CH:
		break
	}
	if ret == nil {
		err = fmt.Errorf("session chan had been closed")
		return
	}
	return
}

// Close 关闭通道
func (sess *Session) Close() {
	close(sess.CH)
	sessions.Delete(sess.Key)
}

var evses sync.Map

// LoadClientUUIDWithEvse 根据evse获取uuid
// func LoadClientUUIDWithEvse(evse string) (ID string, err error) {
// 	if data, ok := evses.Load(evse); ok {
// 		ID = data.(string)
// 		return
// 	}
// 	var evse datasource.UUID
// 	if evse, err = datasource.ParseUUID(evse); err != nil {
// 		return
// 	}
// 	reqEvse := &coregw.EvseReq{evse: evse.Uint64()}
// 	respEvse := &coregw.EvseResp{}
// 	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().UnixNano()), "ClientID", evse.String()))
// 	if err = grpc.Invoke(ctx, coregw.LogicServicesServer.Evse, reqEvse, respEvse); err != nil {
// 		return
// 	}

// 	ID = respEvse.Evse.Uuid
// 	evses.Store(evse, ID)
// 	return
// }

// LoadevseWithUUID 根据uuid获取evse
// func LoadevseWithUUID(topicPrefix, uuid string) (evse string, err error) {
// 	if data, ok := evses.Load(uuid); ok {
// 		evse = data.(string)
// 		return
// 	}
// 	reqEvse := &coregw.EvseWithTopicPrefixAndUUIDReq{TopicPrefix: topicPrefix, Uuid: uuid}
// 	respEvse := &coregw.EvseResp{}

// 	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("requestID", fmt.Sprintf("%d", time.Now().UnixNano()), "ClientID", fmt.Sprintf("core-ac")))
// 	if err = grpc.Invoke(ctx, coregw.ManageServiceServer.EvseWithTopicPrefixAndUUID, reqEvse, respEvse); err != nil {
// 		return
// 	}

// 	evse = fmt.Sprintf("%d", respEvse.Evse.ID)
// 	evses.Store(uuid, evse)
// 	return
// }

// Seq 请求序列号
// todo: 公共使用
type Seq uint64

// GetCMD 获取seq
func (seq Seq) GetCMD(uuid string) (uint32, error) {
	rd := redis.GetRedis()
	defer rd.Close()
	id, err := rdlib.Uint64(rd.Do("get", fmt.Sprintf("%d:%s:cl:seq:ac", seq, uuid)))
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}

// SetCMD 设置seq
func (seq Seq) SetCMD(uuid string, cmd uint32) error {
	rd := redis.GetRedis()
	defer rd.Close()
	_, err := rd.Do("set", fmt.Sprintf("%d:%s:cl:seq:ac", cmd, uuid), seq, "ex", 300)
	return err
}

// // TraficSize 进出流量监控
// func TraficSize(evse string, inOrOut string, size int) error {
// 	return redis.PublishStreamWithMaxlen("evses:metric", 10000, &common.MetricValue{
// 		Instance:   evse,
// 		EntryPoint: "trafic",
// 		Values: map[string]interface{}{
// 			"value": size,
// 		},
// 		Labels: map[string]string{
// 			"direction": inOrOut,
// 		},
// 		Timestamp: time.Now(),
// 	})
// }
