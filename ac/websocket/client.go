package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kotodian/gokit/datasource/redis"
	"go.uber.org/zap"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Kotodian/gokit/ac/lib"
	"github.com/Kotodian/gokit/workpool"
	"github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/Kotodian/protocol/interfaces"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	readWait  = 190 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	chargeStation           interfaces.ChargeStation
	hub                     *Hub            //中间件
	conn                    *websocket.Conn //socket连接
	send                    chan []byte     //发送消息的管道
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
}

func (c *Client) Send(msg []byte) (err error) {
	c.send <- msg
	c.log.Info("<-" + string(msg))
	return
}

func (c *Client) Close(err error) error {
	fmt.Println("关闭连接 1", c.chargeStation.CoreID(), c.chargeStation.SN())
	c.once.Do(func() {
		c.log.Error(err.Error())
		c.hub.Clients.Delete(c.chargeStation.SN())
		c.hub.RegClients.Delete(c.chargeStation.SN())
		fmt.Println("关闭连接 2", c.chargeStation.CoreID(), c.chargeStation.SN())
		//c.Hub.MqttClient.GetMQTT().Unsubscribe(c.subTopics...)
		_ = c.conn.Close()
		c.conn = nil
		close(c.send)
		close(c.close)
		close(c.mqttRegCh)
		close(c.mqttMsgCh)
		c.clientOfflineNotifyFunc(err)
	})
	return nil
}

// NewClient
// 连接客户端管理类
func NewClient(chargeStation interfaces.ChargeStation, hub *Hub, conn *websocket.Conn, keepalive int, remoteAddress string, log *zap.Logger) ClientInterface {
	return &Client{
		log:           log,
		chargeStation: chargeStation,
		hub:           hub,
		conn:          conn,
		remoteAddress: remoteAddress,
		send:          make(chan []byte, 5),
		sendPing:      make(chan struct{}, 1),
		mqttMsgCh:     make(chan MqttMessage, 5),
		mqttRegCh:     make(chan MqttMessage, 5),
		close:         make(chan struct{}),
		keepalive:     int64(keepalive),
	}
}

//SubRegMQTT 监听MQTT的注册报文回复信息
func (c *Client) SubRegMQTT() {
	c.hub.RegClients.Store(c.chargeStation.SN(), c)
	//if c.Evse.CoreID() == 0 {
	for {
		fmt.Println("------------> loop reg msg start", c.chargeStation.SN())
		select {
		case <-c.close:
			fmt.Println("------------> loop reg msg end", c.chargeStation.SN())
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
					err = fmt.Errorf("FromAPDU reg error, err:%s topic:%s", err.Error(), topic)
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
				} else if err = c.hub.SendMsgToDevice(c.chargeStation.SN(), bMsg); err != nil {
					return
				}
			}()
			//logrus.Infof("reg resp:%s, sn:%s", string(bMsg), c.Evse.SN())
		}
	}
}

//SubMQTT 监听MQTT非注册的一般信息
func (c *Client) SubMQTT() {
	c.hub.Clients.Store(c.chargeStation.SN(), c)
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
					} else if err = c.hub.SendMsgToDevice(c.chargeStation.SN(), bMsg); err != nil {
						return
					}
					//}()
					return
				}, Args: []interface{}{apdu},
			})
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	var err error
	defer func() {
		_ = c.Close(err)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err = c.conn.SetReadDeadline(time.Now().Add(readWait))
	if err != nil {
		return
	}
	c.conn.SetPingHandler(func(appData string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(readWait))
		c.log.Info("ping message received", zap.String("sn", c.chargeStation.SN()))
		err = c.conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(writeWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		redisConn := redis.GetRedis()
		defer redisConn.Close()
		_, err = redisConn.Do("expire", fmt.Sprintf("ac:%s:%s:%s:%s", "online", c.hub.Protocol, c.chargeStation.SN(), c.hub.Hostname), 190)
		if err != nil {
			c.log.Error(err.Error(), zap.String("sn", c.chargeStation.SN()))
		}
		return err

	})
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
		_, msg, err = c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseAbnormalClosure) {
				c.log.Sugar().Errorf("error: %v", err)
			}
			break
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

			if !trData.IsTelemetry {
				sendTopic = "coregw/" + c.hub.Hostname + "/command/" + c.chargeStation.SN()
				sendQos = 2
			} else {
				sendTopic = "coregw/" + c.hub.Hostname + "/telemetry/" + c.chargeStation.SN()
				sendQos = 2
			}

			c.hub.PubMqttMsg <- MqttMessage{
				Topic:    sendTopic,
				Qos:      sendQos,
				Retained: false,
				Payload:  toCoreMSG,
			}
		}(ctx, msg)

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

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	var err error
	for {
		select {
		case message, ok := <-c.send:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			var w io.WriteCloser
			w, err = c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.send)
			}

			if err = w.Close(); err != nil {
				return
			}
		//case <-ticker.C:
		//	if c.conn == nil {
		//		return
		//	}
		//	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		//	if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		//		return
		//	}
		case <-c.sendPing:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) PublishReg(m MqttMessage) {
	c.mqttRegCh <- m
}

func (c *Client) Publish(m MqttMessage) {
	c.mqttMsgCh <- m
}

func (c *Client) KeepAlive() int64 {
	return c.keepalive
}

func (c *Client) Logger() *zap.Logger {
	return c.log
}

func (c *Client) Hub() *Hub {
	return c.hub
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
