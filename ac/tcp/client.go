package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/Kotodian/gokit/ac/lib"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/gokit/workpool"
	"github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/Kotodian/protocol/interfaces"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"time"
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
	conn *net.TCPConn

	isClose bool

	once sync.Once
}

func NewClient(chargeStation interfaces.ChargeStation, hub *lib.Hub, conn *net.TCPConn, keepalive int64, remoteAddress string, log *zap.Logger) *Client {
	client := &Client{
		log:           log,
		chargeStation: chargeStation,
		hub:           hub,
		conn:          conn,
		remoteAddress: remoteAddress,
		send:          make(chan []byte, 5),
		mqttMsgCh:     make(chan mqtt.MqttMessage, 5),
		mqttRegCh:     make(chan mqtt.MqttMessage, 5),
		close:         make(chan struct{}),
		keepalive:     keepalive,
		isClose:       false,
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
				c.Reply(ctx, msg)
			}()
		}
	}
}

func (c *Client) Reply(ctx context.Context, payload interface{}) {
	resp, err := c.hub.ResponseFn(ctx, payload)
	if err != nil {
		return
	}
	//resp := protocol.NewCallResult(ctx, payload)
	//b, _ := json.Marshal(resp)
	_ = c.Send(resp)
	//if _client := ctx.Value("client"); _client != nil {
	//	client := _client.(*Client)
	//	client.send <- b
	//}
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

					//平台下发的命令都要回复，不存在不回复的情况
					//else if apdu.NoNeedReply {
					//	return
					//}
					c.Reply(ctx, msg)
					//}()
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
		var msg []byte
		_, err = c.conn.Read(msg)
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
			defer func() {
				//如果发生了错误，都回复给设备，否则发送到平台
				if err != nil {
					c.ReplyError(ctx, err)
				}
			}()

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

			c.hub.PubMqttMsg <- mqtt.MqttMessage{
				Topic:    sendTopic,
				Qos:      sendQos,
				Retained: false,
				Payload:  toCoreMSG,
			}
		}(ctx, msg)
	}
}

func (c *Client) WritePump() {
	var err error
	if err != nil {
		// todo 关闭连接
	}
	for {
		select {
		case message, ok := <-c.send:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				err = errors.New("send on closed channel")
				return
			}
			_, err = c.conn.Write(message)
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
