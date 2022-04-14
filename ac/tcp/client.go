package tcp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kotodian/gokit/ac/lib"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/gokit/datasource/redis"
	"github.com/Kotodian/gokit/workpool"
	"github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/Kotodian/protocol/golang/keys"
	"github.com/Kotodian/protocol/interfaces"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
)

const (
	writeWait = 10 * time.Second
	readWait  = 125 * time.Second
)

type Client struct {
	// 桩实体
	chargeStation interfaces.ChargeStation
	// 存储桩的容器
	hub *lib.Hub
	// 发送消息的管道
	send chan []byte
	// 退出的通知 以后可以做一些其他的监听订阅
	close chan struct{}
	// 加锁,一次处理一个请求
	lock sync.RWMutex
	// 关闭时通知
	clientOfflineNotifyFunc func(err error)
	// 注册信息管道
	mqttRegCh chan mqtt.MqttMessage
	// 返回或者下发的消息
	mqttMsgCh chan mqtt.MqttMessage
	// 客户端地址
	remoteAddress string
	// 日志组件
	log *zap.Logger
	// 维持连接的超时时间
	keepalive int64
	// 记录coregw发送的请求host,以便回复给请求的coregw
	coregw string
	// 加密key
	encryptKey []byte
	// 平台端id
	id string
	// 连接
	conn    *net.TCPConn
	isClose bool
	once    sync.Once
	// 证书sn
	certificateSN string
	// 发送消息的序号
	messageNumber int16

	data sync.Map
}

func NewClient(hub *lib.Hub, conn *net.TCPConn, keepalive int64, remoteAddress string, log *zap.Logger) lib.ClientInterface {
	client := &Client{
		log:           log,
		hub:           hub,
		conn:          conn,
		remoteAddress: remoteAddress,
		send:          make(chan []byte, 5),
		mqttMsgCh:     make(chan mqtt.MqttMessage, 5),
		mqttRegCh:     make(chan mqtt.MqttMessage, 5),
		close:         make(chan struct{}),
		keepalive:     keepalive,
		isClose:       false,
		messageNumber: 0,
	}
	return client
}

func (c *Client) Send(msg []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	if c.hub.Encrypt != nil && len(c.encryptKey) > 0 {
		msg, err = c.hub.Encrypt.Encode(msg, c.encryptKey)
		if err != nil {
			return err
		}
	}
	c.send <- msg
	return
}

func (c *Client) Close(err error) error {
	c.once.Do(func() {
		if err == nil {
			err = errors.New("平台关闭")
		}
		c.log.Error(err.Error())
		c.hub.Clients.Delete(c.chargeStation.CoreID())
		c.hub.RegClients.Delete(c.chargeStation.CoreID())
		_ = c.conn.Close()
		c.log.Sugar().Info(c.chargeStation.SN(), "关闭连接")
		c.conn = nil
		c.data = sync.Map{}
		close(c.send)
		close(c.close)
		close(c.mqttRegCh)
		close(c.mqttMsgCh)

		c.clientOfflineNotifyFunc(err)
		c.isClose = true
	})

	return nil
}

func (c *Client) SubRegMQTT() {
	c.hub.RegClients.Store(c.chargeStation.CoreID(), c)
	for {
		select {
		case <-c.close:
			c.log.Sugar().Info("register msg end", c.chargeStation.SN())
			return
		case m := <-c.mqttRegCh:
			func() {
				var apdu charger.APDU
				var err error
				topic := m.Topic
				defer func() {
					if err != nil {
						c.log.Error(err.Error(), zap.String("sn", c.chargeStation.SN()))
					}
				}()
				if err = proto.Unmarshal(m.Payload, &apdu); err != nil {
					return
				}
				trData := &lib.TRData{
					APDU: &apdu,
				}

				// 将客户端信息以及消息放入到上下文中
				ctx := context.WithValue(context.TODO(), "client", c)
				ctx = context.WithValue(ctx, "trData", trData)

				var msg interface{}

				// 处理并翻译成桩端需要的结果
				if msg, err = c.hub.TR.FromAPDU(ctx, &apdu); err != nil {
					err = fmt.Errorf("FromAPDU register error, err: %s topic: %s", err.Error(), topic)
					return
				} else if msg == nil {
					return
				}

				if trData.Ignore {
					return
				}
				if apdu.IsRequest() {
					c.sendCommand(ctx, msg)
				} else {
					c.Reply(ctx, msg)
				}
			}()
		}
	}
}

func (c *Client) Reply(ctx context.Context, payload interface{}) {
	resp, err := c.hub.ResponseFn(ctx, payload)
	if err != nil {
		return
	}
	_ = c.Send(resp)
}

func (c *Client) sendCommand(ctx context.Context, payload interface{}) {
	command, err := c.hub.CommandFn(ctx, payload)
	if err != nil {
		return
	}
	_ = c.Send(command)
}

func (c *Client) ReplyError(ctx context.Context, err error, desc ...string) {
	b := c.hub.ResponseErrFn(ctx, err, desc...)
	if b != nil {
		_ = c.Send(b)
	}
}

func (c *Client) SubMQTT() {
	c.hub.Clients.Store(c.chargeStation.CoreID(), c)
	wp := workpool.New(1, 5).Start()
	for {
		select {
		case <-c.close:
			return
		case m := <-c.mqttMsgCh:
			if len(m.Payload) == 0 {
				break
			}
			var apdu charger.APDU
			if err := proto.Unmarshal(m.Payload, &apdu); err != nil {
				break
			}
			wp.PushTask(workpool.Task{
				F: func(w *workpool.WorkPool, args ...interface{}) (flag workpool.Flag) {
					flag = workpool.FLAG_OK
					topic := m.Topic

					apdu := args[0].(charger.APDU)
					trData := &lib.TRData{
						APDU:  &apdu,
						Topic: topic,
					}
					ctx := context.WithValue(context.TODO(), "client", c)
					ctx = context.WithValue(ctx, "trData", trData)
					var msg interface{}

					var err error
					defer func() {
						//如果没有错误就转发到设备上，否则写日志，回复到平台的错误日志有FromAPDU实现了
						if err != nil {
							c.log.Error(err.Error(), zap.String("sn", c.chargeStation.SN()))
							if trData.Ignore == false && (int32(apdu.MessageId)>>7 == 0 || apdu.MessageId == charger.MessageID_ID_MessageError) {
								if apdu.MessageId != charger.MessageID_ID_MessageError {
									apdu.MessageId = charger.MessageID_ID_MessageError
									apdu.Payload, _ = proto.Marshal(&charger.MessageError{
										Error:       charger.ErrorCode_EC_GenericError,
										Description: err.Error(),
									})
								}
								apduEncoded, _ := proto.Marshal(&apdu)
								pubMqttMsg := mqtt.MqttMessage{
									Topic:    strings.Replace(trData.Topic, c.hub.Hostname, "coregw", 1),
									Qos:      2,
									Retained: false,
									Payload:  apduEncoded,
								}
								c.hub.PubMqttMsg <- pubMqttMsg
							}
						}
					}()
					if msg, err = c.hub.TR.FromAPDU(ctx, &apdu); err != nil {
						return
					} else if msg == nil {
						return
					}
					if trData.Ignore {
						return
					}
					if apdu.IsRequest() {
						c.sendCommand(ctx, msg)
					} else {
						c.Reply(ctx, msg)
					}
					return
				}, Args: []interface{}{apdu},
			})
		}
	}
}

func (c *Client) ReadPump() {
	var err error
	defer func() {
		// todo 关闭连接
		if err != nil {
			_ = c.Close(err)
		}
	}()
	err = c.conn.SetReadDeadline(time.Now().Add(readWait))
	if err != nil {
		return
	}
	for {
		if c.conn == nil {
			return
		}
		err = c.conn.SetReadDeadline(time.Now().Add(readWait))
		if err != nil {
			break
		}
		msg := make([]byte, 256)
		reader := bufio.NewReader(c.conn)
		_, err = reader.Read(msg)
		if err != nil {
			break
		}
		if c.hub.Encrypt != nil && len(c.encryptKey) > 0 {
			msg, err = c.hub.Encrypt.Decode(msg, c.encryptKey)
			if err != nil {
				break
			}
		}
		ctx := context.WithValue(context.TODO(), "client", c)

		go func(ctx context.Context, msg []byte) {
			trData := &lib.TRData{}
			ctx = context.WithValue(ctx, "trData", trData)
			var err error

			var payload proto.Message
			if payload, err = c.hub.TR.ToAPDU(ctx, msg); err != nil {
				return
			}

			if payload == nil {
				return
			}

			if trData.Ignore {
				return
			}

			if trData.APDU.Payload, err = proto.Marshal(payload); err != nil {
				err = fmt.Errorf("encode cmd req payload error, err:%s", err.Error())
				return
			}
			var toCoreMSG []byte
			if toCoreMSG, err = proto.Marshal(trData.APDU); err != nil {
				err = fmt.Errorf("encode cmd req apdu error, err:%s", err.Error())
				return
			}

			var sendTopic string
			var sendQos byte
			if trData.IsTelemetry {
				sendTopic = "coregw/" + c.hub.Hostname + "/telemetry/" + datasource.UUID(c.chargeStation.CoreID()).String()
			} else if !trData.Sync {
				sendTopic = "coregw/" + c.hub.Hostname + "/command/" + datasource.UUID(c.chargeStation.CoreID()).String()
			} else {
				sendTopic = c.coregw + "/sync/" + datasource.UUID(c.chargeStation.CoreID()).String()
			}
			sendQos = 2
			if c.chargeStation != nil {
				c.hub.PubMqttMsg <- mqtt.MqttMessage{
					Topic:    sendTopic,
					Qos:      sendQos,
					Retained: false,
					Payload:  toCoreMSG,
				}
			}
		}(ctx, msg)
	}
}

func (c *Client) WritePump() {
	var err error
	defer func() {
		if err != nil {
			_ = c.Close(err)
		}
	}()
	for {
		select {
		case <-c.close:
			return
		case message, ok := <-c.send:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				err = errors.New("send on closed channel")
				return
			}
			w := bufio.NewWriter(c.conn)
			_, err = w.Write(message)
			if err != nil {
				return
			}
			err = w.Flush()
			if err != nil {
				return
			}
		}
	}
}
func (c *Client) Hub() *lib.Hub {
	return c.hub
}
func (c *Client) PublishReg(m mqtt.MqttMessage) {
	c.mqttRegCh <- m
}

func (c *Client) Publish(m mqtt.MqttMessage) {
	c.mqttMsgCh <- m
}

func (c *Client) KeepAlive() int64 {
	return c.keepalive
}

func (c *Client) Logger() *zap.Logger {
	return c.log
}

func (c *Client) RemoteAddress() string {
	return c.remoteAddress
}

func (c *Client) ChargeStation() interfaces.ChargeStation {
	return c.chargeStation
}

func (c *Client) SetClientOfflineFunc(clientOfflineFunc func(err error)) {
	c.clientOfflineNotifyFunc = clientOfflineFunc
}

func (c *Client) ClientOfflineFunc() func(err error) {
	return c.clientOfflineNotifyFunc
}

func (c *Client) Lock() {
	c.lock.Lock()
}

func (c *Client) Unlock() {
	c.lock.Unlock()
}

func (c *Client) Coregw() string {
	return c.coregw
}

func (c *Client) SetCoregw(coregw string) {
	c.coregw = coregw
}

func (c *Client) SetEncryptKey(encryptKey string) {
	c.encryptKey = []byte(encryptKey)
}

func (c *Client) IsClose() bool {
	return c.isClose
}

func (c *Client) EncryptKey() []byte {
	return c.encryptKey
}

func (c *Client) PingHandler(msg string) error {
	redisConn := redis.GetRedis()
	defer redisConn.Close()
	_, err := redisConn.Do("expire", keys.Equipment(strconv.FormatUint(c.chargeStation.CoreID(), 10)), 190)
	if err != nil {
		c.log.Error(err.Error(), zap.String("sn", c.chargeStation.SN()))
	}
	return nil
}

func (c *Client) CertificateSN() string {
	return c.certificateSN
}

func (c *Client) SetCertificateSN(sn string) {
	c.certificateSN = sn
}

func (c *Client) SetMessageNumber(i int16) {
	c.messageNumber = i
}

func (c *Client) GetMessageNumber() int16 {
	return c.messageNumber
}

func (c *Client) SetData(key, value interface{}) {
	if value == nil {
		c.data.Delete(key)
	} else {
		c.data.Store(key, value)
	}
}

func (c *Client) GetData(key interface{}) interface{} {
	if value, ok := c.data.Load(key); ok {
		return value
	}
	return nil
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) MessageNumber() int16 {
	return c.messageNumber
}

func (c *Client) SetChargeStation(cs interfaces.ChargeStation) {
	c.chargeStation = cs
}

func (c *Client) SetKeepalive(keepalive int64) {
	c.keepalive = keepalive
}
