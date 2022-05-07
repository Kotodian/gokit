package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/gokit/datasource/redis"
	"github.com/Kotodian/protocol/golang/keys"
	"github.com/valyala/bytebufferpool"
	"go.uber.org/zap"

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
	readWait  = 80 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
	readBufferSize = 2048
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	chargeStation           interfaces.ChargeStation
	hub                     *lib.Hub        //中间件
	conn                    *websocket.Conn //socket连接
	send                    chan []byte     //发送消息的管道
	sendPing                chan struct{}
	close                   chan struct{}         //退出的通知
	once                    sync.Once             //主要处理关闭通道
	lock                    sync.RWMutex          //加锁，一次只能同步一个报文，减少并发
	clientOfflineNotifyFunc func(err error)       // 网络断开同步到core的函数
	mqttRegCh               chan mqtt.MqttMessage //注册信息
	mqttMsgCh               chan mqtt.MqttMessage //返回或下发的信息
	remoteAddress           string
	log                     *zap.Logger
	keepalive               int64
	coregw                  string
	isClose                 bool
	encryptKey              []byte
	id                      string
	certificateSN           string
	orderInterval           int
	baseURL                 string // 上传日志、下载固件基本地址
}

func (c *Client) MessageNumber() int16 {
	return 0
}

func (c *Client) SetMessageNumber(i int16) {
	return
}

func (c *Client) SetData(key, val interface{}) {
	return
}

func (c *Client) GetData(key interface{}) interface{} {
	return nil
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
		c.log.Error(err.Error(), zap.String("sn", c.chargeStation.SN()))
		c.hub.Clients.Delete(c.chargeStation.CoreID())
		c.hub.RegClients.Delete(c.chargeStation.CoreID())
		_ = c.conn.Close()
		c.log.Info("关闭连接", zap.String("sn", c.chargeStation.SN()))
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

// NewClient
// 连接客户端管理类
func NewClient(chargeStation interfaces.ChargeStation, hub *lib.Hub, conn *websocket.Conn, keepalive int, remoteAddress string, log *zap.Logger) lib.ClientInterface {
	return &Client{
		log:           log,
		chargeStation: chargeStation,
		hub:           hub,
		conn:          conn,
		remoteAddress: remoteAddress,
		send:          make(chan []byte, 5),
		sendPing:      make(chan struct{}, 1),
		mqttMsgCh:     make(chan mqtt.MqttMessage, 5),
		mqttRegCh:     make(chan mqtt.MqttMessage, 5),
		close:         make(chan struct{}),
		keepalive:     int64(keepalive),
		orderInterval: 30,
	}
}

//SubRegMQTT 监听MQTT的注册报文回复信息
func (c *Client) SubRegMQTT() {
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
				if trData.Ignore {
					return
				}
				buffer := bytebufferpool.Get()
				defer func() {
					buffer.Reset()
					bytebufferpool.Put(buffer)
				}()
				encoder := json.NewEncoder(buffer)
				if err = encoder.Encode(msg); err != nil {
					return
				} else if err = c.Send(buffer.Bytes()); err != nil {
					return
				}
			}()
		}
	}
}

//SubMQTT 监听MQTT非注册的一般信息
func (c *Client) SubMQTT() {
	c.hub.Clients.Store(c.chargeStation.CoreID(), c)
	wp := workpool.New(1, 5).Start()
	defer wp.Stop()
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
					// 优化bytes
					buffer := bytebufferpool.Get()
					defer func() {
						buffer.Reset()
						bytebufferpool.Put(buffer)
					}()
					encoder := json.NewEncoder(buffer)
					if err = encoder.Encode(msg); err != nil {
						return
					} else if err = c.Send(buffer.Bytes()); err != nil {
						return
					}
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
	msg := make([]byte, 0, readBufferSize)
	defer func() {
		msg = nil
		_ = c.Close(err)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err = c.conn.SetReadDeadline(time.Now().Add(readWait))
	if err != nil {
		return
	}
	c.conn.SetPingHandler(func(appData string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(readWait))
		fmt.Printf("[%s]ping message received from %s\n", time.Now().Format("2006-01-02 15:04:05"), c.chargeStation.SN())
		return c.PingHandler(appData)
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
		_, msg, err = c.conn.ReadMessage()
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

			c.hub.PubMqttMsg <- mqtt.MqttMessage{
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
	//ticker := time.NewTicker(10 * time.Second)
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

func (c *Client) Hub() *lib.Hub {
	return c.hub
}

func (c *Client) RemoteAddress() string {
	return c.remoteAddress
}

func (c *Client) ChargeStation() interfaces.ChargeStation {
	return c.chargeStation
}

func (c *Client) SetChargeStation(chargeStation interfaces.ChargeStation) {
	c.chargeStation = chargeStation
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
	// c.log.Debug("ping message received", zap.String("sn", c.chargeStation.SN()))
	_ = c.conn.SetReadDeadline(time.Now().Add(readWait))
	err := c.conn.WriteControl(websocket.PongMessage, []byte(msg), time.Now().Add(writeWait))
	if err == websocket.ErrCloseSent {
		return nil
	} else if e, ok := err.(net.Error); ok && e.Temporary() {
		return nil
	}
	redisConn := redis.GetRedis()
	defer redisConn.Close()
	_, err = redisConn.Do("expire", keys.Equipment(strconv.FormatUint(c.chargeStation.CoreID(), 10)), 190)
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

func (c *Client) Conn() net.Conn {
	return c.conn.UnderlyingConn()
}

func (c *Client) SetKeepalive(keepalive int64) {
	c.keepalive = keepalive
}

func (c *Client) SetRemoteAddress(address string) {
	c.remoteAddress = address
}

func (c *Client) SetOrderInterval(interval int) {
	c.orderInterval = interval
}

func (c *Client) OrderInterval() int {
	return c.orderInterval
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}
