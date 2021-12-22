package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Kotodian/gokit/ac/lib"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/workpool"
	"github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/Kotodian/protocol/interfaces"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPClient struct {
	chargeStation           interfaces.ChargeStation
	hub                     *Hub        //中间件
	conn                    net.Conn    //socket连接
	send                    chan []byte //发送消息的管道
	sendPing                chan struct{}
	close                   chan struct{}    //退出的通知
	once                    sync.Once        //主要处理关闭通道
	lock                    sync.RWMutex     //加锁，一次只能同步一个报文，减少并发
	clientOfflineNotifyFunc func(err error)  // 网络断开同步到core的函数
	mqttRegCh               chan MqttMessage //注册信息
	mqttMsgCh               chan MqttMessage //返回或下发的信息
	remoteAddress           string
	log                     *zap.Logger
	keepalive               int64
	coregw                  string
	isClose                 bool
	encryptKey              []byte
	id                      string
}

func (c *TCPClient) Send(msg []byte) (err error) {
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

func (c *TCPClient) SubRegMQTT() {
	c.hub.RegClients.Store(c.chargeStation.CoreID(), c)
	//if c.Evse.CoreID() == 0 {
	for {
		c.log.Sugar().Info("------------> register msg start", c.chargeStation.SN())
		select {
		case <-c.close:
			c.log.Sugar().Info("------------> register msg end", c.chargeStation.SN())
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
				ctx := context.WithValue(context.TODO(), "client", c)
				ctx = context.WithValue(ctx, "trData", trData)

				var msg interface{}
				//var f lib.FromAPDUFunc
				if msg, err = c.hub.TR.FromAPDU(ctx, &apdu); err != nil {
					err = fmt.Errorf("FromAPDU register error, err:%s topic:%s", err.Error(), topic)
					return
				} else if msg == nil {
					return
				}

				//b, err := f(ctx)
				//if err != nil {
				//	log.Errorf("translate register conf error, err:%v topic:%s", err.Error(), topic)
				//	return
				if trData.Ignore {
					return
				}
				var bMsg []byte
				if bMsg, err = json.Marshal(msg); err != nil {
					return
				} else if err = c.hub.SendMsgToDevice(c.chargeStation.CoreID(), bMsg); err != nil {
					return
				}
			}()
			//logrus.Infof("reg resp:%s, sn:%s", string(bMsg), c.Evse.SN())
		}
	}
}

func (c *TCPClient) SubMQTT() {
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
								pubMqttMsg := MqttMessage{
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
					var bMsg []byte
					if bMsg, err = json.Marshal(msg); err != nil {
						return
					} else if err = c.hub.SendMsgToDevice(c.chargeStation.CoreID(), bMsg); err != nil {
						return
					}
					//}()
					return
				}, Args: []interface{}{apdu},
			})
		}
	}
}

func (c *TCPClient) ReadPump() {
	var err error
	defer func() {
		_ = c.Close(err)
	}()

	err = c.conn.SetReadDeadline(time.Now().Add(readWait))
	if err != nil {
		return
	}
	for {
		ctx := context.WithValue(context.TODO(), "client", c)
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseAbnormalClosure) {
				c.log.Sugar().Errorf("error: %v", err)
			}
			break
		}

		if c.hub.Encrypt != nil && len(c.encryptKey) > 0 {
			msg, err = c.hub.Encrypt.Decode(msg, c.encryptKey)
			if err != nil {
				break
			}
		}

		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))

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
				sendTopic = c.Coregw() + "/sync/" + datasource.UUID(c.chargeStation.CoreID()).String()
			}
			sendQos = 2

			c.hub.PubMqttMsg <- MqttMessage{
				Topic:    sendTopic,
				Qos:      sendQos,
				Retained: false,
				Payload:  toCoreMSG,
			}
		}(ctx, msg)
	}
}

func (c *TCPClient) WritePump() {
	var err error
	if err != nil {
		defer c.Close(err)
	}
	for {
		select {
		case message, ok := <-c.send:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				return
			}
			_, _ = c.conn.Write(message)
		case <-c.sendPing:
			if c.conn == nil {
				return
			}
		}
	}
}

func (c *TCPClient) Reply(ctx context.Context, payload interface{}) {
	resp, err := c.hub.ResponseFn(ctx, payload)
	if err != nil {
		return
	}
	//resp := protocol.NewCallResult(ctx, payload)
	//b, _ := json.Marshal(resp)
	_ = c.Send(resp)
}

func (c *TCPClient) ReplyError(ctx context.Context, err error, desc ...string) {
	b := c.hub.ResponseErrFn(ctx, err, desc...)
	if b != nil {
		_ = c.Send(b)
	}
}

func (c *TCPClient) Publish(m MqttMessage) {
	c.mqttMsgCh <- m
}

func (c *TCPClient) PublishReg(m MqttMessage) {
	c.mqttRegCh <- m
}

func (c *TCPClient) Close(err error) error {
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

func (c *TCPClient) ChargeStation() interfaces.ChargeStation {
	return c.chargeStation
}

func (c *TCPClient) Hub() *Hub {
	return c.hub
}

func (c *TCPClient) KeepAlive() int64 {
	return c.keepalive
}

func (c *TCPClient) Logger() *zap.Logger {
	return c.log
}

func (c *TCPClient) RemoteAddress() string {
	return c.remoteAddress
}

func (c *TCPClient) SetClientOfflineFunc(f func(err error)) {
	c.clientOfflineNotifyFunc = f
}

func (c *TCPClient) ClientOfflineFunc() func(err error) {
	return c.clientOfflineNotifyFunc
}

func (c *TCPClient) Lock() {
	c.lock.Lock()
}

func (c *TCPClient) Unlock() {
	c.lock.Unlock()
}

func (c *TCPClient) Coregw() string {
	return c.coregw
}

func (c *TCPClient) SetCoregw(coregw string) {
	c.coregw = coregw
}

func (c *TCPClient) IsClose() bool {
	return c.isClose
}

func (c *TCPClient) EncryptKey() []byte {
	return c.encryptKey
}

func (c *TCPClient) SetEncryptKey(s string) {
	c.encryptKey = []byte(s)
}
