package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Kotodian/gokit/ac/lib"
	"github.com/Kotodian/gokit/datasource"
	"github.com/Kotodian/gokit/workpool"
	pCharger "github.com/Kotodian/protocol/golang/hardware/charger"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	EvseID                  string          //设备SN
	Hub                     *Hub            //中间件
	conn                    *websocket.Conn //socket连接
	send                    chan []byte     //发送消息的管道
	sendPing                chan struct{}
	close                   chan struct{}    //退出的通知
	PongWait                time.Duration    //pong
	PingPeriod              time.Duration    //ping的周期
	Log                     *logrus.Entry    //日志
	HBTime                  time.Time        //心跳时间
	PfEvseID                datasource.UUID  //平台ID
	once                    sync.Once        //主要处理关闭通道
	Lock                    sync.RWMutex     //加锁，一次只能同步一个报文，减少并发
	SyncOnline              bool             //是否已同步在线，如果未同步在线就不回复其他报文
	Data                    sync.Map         //存储的上下文数据
	ClientOfflineNotifyFunc func(err error)  // 网络断开同步到core的函数
	MqttRegCh               chan MqttMessage //注册信息
	MqttMsgCh               chan MqttMessage //返回或下发的信息
	//subTopics               []string        //监听的MQTT路径
}

func (c *Client) Send(msg []byte) (err error) {
	_log := c.Log.Dup()
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			_log.Error(err)
		}
	}()
	c.send <- msg
	_log.Info("<- ", string(msg))
	return
}

func (c *Client) Close(err error) error {
	fmt.Println("关闭连接 1", c.PfEvseID, c.EvseID)
	c.once.Do(func() {
		c.Hub.Clients.Delete(c.PfEvseID.String())
		c.Hub.RegClients.Delete(c.EvseID)
		fmt.Println("关闭连接 2", c.PfEvseID, c.EvseID)
		//c.Hub.MqttClient.GetMQTT().Unsubscribe(c.subTopics...)
		_ = c.conn.Close()
		c.conn = nil
		close(c.send)
		close(c.close)
		close(c.MqttRegCh)
		close(c.MqttMsgCh)
		c.ClientOfflineNotifyFunc(err)
	})
	return nil
}

// NewClient
// 连接客户端管理类
func NewClient(evseID string, hub *Hub, conn *websocket.Conn, pongWait time.Duration) *Client {
	_log := logrus.WithFields(logrus.Fields{
		"evse_id": evseID,
	})
	pingPeriod := (pongWait * 9) / 10
	return &Client{
		Log:        _log,
		EvseID:     evseID,
		Hub:        hub,
		HBTime:     time.Now().Local(),
		conn:       conn,
		PingPeriod: pingPeriod,
		PongWait:   pongWait + 10*time.Second,
		send:       make(chan []byte, 5),
		sendPing:   make(chan struct{}, 1),
		MqttMsgCh:  make(chan MqttMessage, 5),
		MqttRegCh:  make(chan MqttMessage, 5),
		close:      make(chan struct{}, 0),
	}
}

//SubRegMQTT 监听MQTT的注册报文回复信息
func (c *Client) SubRegMQTT() {
	c.Hub.RegClients.Store(c.EvseID, c)
	//if c.PfEvseID == 0 {
	for {
		fmt.Println("------------> loop reg msg start", c.EvseID)
		select {
		case <-c.close:
			fmt.Println("------------> loop reg msg end", c.EvseID)
			return
		case m := <-c.MqttRegCh:
			func() {
				var apdu pCharger.APDU
				var err error
				topic := m.Topic
				if err = proto.Unmarshal(m.Payload, &apdu); err != nil {
					logrus.Errorf("decode register conf error, err:%v topic:%s", err.Error(), topic)
					return
				}

				trData := &lib.TRData{
					APDU: &apdu,
				}
				ctx := context.WithValue(context.TODO(), "client", c)
				ctx = context.WithValue(ctx, "log", logrus.WithFields(logrus.Fields{
					"sn": c.EvseID,
				}))
				ctx = context.WithValue(ctx, "trData", trData)

				_log := logrus.WithFields(logrus.Fields{
					"evse_id":    c.EvseID,
					"pf_evse_id": c.PfEvseID,
					"topic":      topic,
					"from":       "core",
					"action":     pCharger.MessageID_name[int32(apdu.MessageId)],
				})
				ctx = context.WithValue(ctx, "log", _log)

				defer func() {
					//如果没有错误就转发到设备上，否则只记录日志
					if err != nil {
						_log.Error(err)
					}
				}()

				var msg interface{}
				//var f lib.FromAPDUFunc
				if msg, err = c.Hub.TR.FromAPDU(ctx, &apdu); err != nil {
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
				} else if err = c.Hub.SendMsgToDevice(c.EvseID, bMsg); err != nil {
					logrus.Errorf("send msg error:%s", err.Error())
					return
				}
			}()
			//logrus.Infof("reg resp:%s, evse_id:%s", string(bMsg), c.EvseID)
		}
	}
	//}
	return
}

//SubMQTT 监听MQTT非注册的一般信息
func (c *Client) SubMQTT() {
	c.Hub.Clients.Store(c.PfEvseID.String(), c)
	wp := workpool.New(1, 5).Start()
	for {
		select {
		case <-c.close:
			return
		case m := <-c.MqttMsgCh:
			if len(m.Payload) == 0 {
				logrus.Warnf("get empty payload, topic:%s, ignore", m.Topic)
				break
			}
			var apdu pCharger.APDU
			if err := proto.Unmarshal(m.Payload, &apdu); err != nil {
				logrus.Errorf("decode cmd or measure msg error, err:%s topic:%s", err.Error(), m.Topic)
				break
			}
			wp.PushTask(workpool.Task{
				F: func(w *workpool.WorkPool, args ...interface{}) (flag workpool.Flag) {
					flag = workpool.FLAG_OK
					topic := m.Topic

					apdu := args[0].(pCharger.APDU)
					trData := &lib.TRData{
						APDU:  &apdu,
						Topic: topic,
					}
					ctx := context.WithValue(context.TODO(), "client", c)
					ctx = context.WithValue(ctx, "log", logrus.WithFields(logrus.Fields{
						"sn": c.EvseID,
					}))
					ctx = context.WithValue(ctx, "trData", trData)

					_log := logrus.WithFields(logrus.Fields{
						"evse_id":    c.EvseID,
						"pf_evse_id": c.PfEvseID,
						"topic":      topic,
						"from":       "core",
						"action":     pCharger.MessageID_name[int32(apdu.MessageId)],
					})
					var msg interface{}

					var err error
					defer func() {
						//如果没有错误就转发到设备上，否则写日志，回复到平台的错误日志有FromAPDU实现了
						if err != nil {
							_log.Errorf(err.Error())
							if trData.Ignore == false && (int32(apdu.MessageId)>>7 == 0 || apdu.MessageId == pCharger.MessageID_ID_MessageError) {
								if apdu.MessageId != pCharger.MessageID_ID_MessageError {
									apdu.MessageId = pCharger.MessageID_ID_MessageError
									apdu.Payload, _ = proto.Marshal(&pCharger.MessageError{
										Error:       pCharger.ErrorCode_EC_GenericError,
										Description: err.Error(),
									})
								}
								apduEncoded, _ := proto.Marshal(&apdu)
								pubMqttMsg := MqttMessage{
									Topic:    strings.Replace(trData.Topic, "/D/", "/U/", 1),
									Qos:      2,
									Retained: false,
									Payload:  apduEncoded,
								}
								c.Hub.PubMqttMsg <- pubMqttMsg
							}
						}
					}()

					if msg, err = c.Hub.TR.FromAPDU(ctx, &apdu); err != nil {
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
					} else if err = c.Hub.SendMsgToDevice(c.PfEvseID.String(), bMsg); err != nil {
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
	_ = c.conn.SetReadDeadline(time.Now().Add(c.PongWait))
	c.conn.SetPongHandler(func(s string) error {
		c.Log.Infoln("pong")
		_ = c.conn.SetReadDeadline(time.Now().Add(c.PongWait))
		return nil
	})
	for {
		select {
		case <-c.close:
			return
		default:
		}
		ctx := context.WithValue(context.TODO(), "client", c)
		ctx = context.WithValue(ctx, "log", logrus.WithFields(logrus.Fields{
			"sn": c.EvseID,
		}))
		if c.conn == nil {
			return
		}

		var msg []byte
		_, msg, err = c.conn.ReadMessage()
		if err != nil {
			c.Log.Errorf("error: %v", err)
			break
			//if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			//}
			//c.Hub.Unregister <- c
			//c.ReplyError(ctx, err)
			//c.Log.Error(err.Error())
			//log.Printf("error: %v", err)
			//break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))

		go func(ctx context.Context, msg []byte) {
			trData := &lib.TRData{}
			ctx = context.WithValue(ctx, "trData", trData)
			//_log := c.log.WithFields(log.Fields{
			//	"from": "device",
			//	"action"
			//})
			var err error
			defer func() {
				//如果发生了错误，都回复给设备，否则发送到平台
				if err != nil {
					c.ReplyError(ctx, err)
				}
			}()

			//acCTX.Data["ac"] = c.Hub
			//acCTX.Data["ctx"] = ctx
			//var f lib.ToAPDUFunc
			var payload proto.Message
			if payload, err = c.Hub.TR.ToAPDU(ctx, msg); err != nil {
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
				sendTopic = "core/"+ c.Hub.Protocol + "/U/C/" + c.PfEvseID.String()
				sendQos = 2
			} else {
				sendTopic = "core/"+ c.Hub.Protocol + "/U/M/" + c.PfEvseID.String()
				sendQos = 0
			}

			c.Hub.PubMqttMsg <- MqttMessage{
				Topic:    sendTopic,
				Qos:      sendQos,
				Retained: false,
				Payload:  toCoreMSG,
			}
		}(ctx, msg)
	}
}

func (c *Client) Reply(ctx context.Context, payload interface{}) {
	_log := c.Log.Dup()
	defer func() {
		if e := recover(); e != nil {
			_log.Error(e)
		}
	}()
	resp, err := c.Hub.ResponseFn(ctx, payload)
	if err != nil {
		_log.Error(err)
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
	defer func() {
		if e := recover(); e != nil {
			_log := c.Log.Dup()
			_log.Error(e)
		}
	}()
	b := c.Hub.ResponseErrFn(ctx, err, desc...)
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
	ticker := time.NewTicker(c.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Close(err)
	}()
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
				c.Log.Error(err)
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
				c.Log.Error(err)
				return
			}
		case <-ticker.C:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Log.Error(err)
				return
			}
		case <-c.sendPing:
			if c.conn == nil {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Log.Error(err)
				return
			}
		}
	}
}

func (c *Client) SendPing() {
	c.sendPing <- struct{}{}
}
