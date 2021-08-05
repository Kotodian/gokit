package websocket

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Kotodian/gokit/ac/lib"
	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/gokit/sync/errgroup.v2"
	"github.com/Kotodian/gokit/workpool"
	mqttClient "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

type MqttMessage struct {
	Qos      byte
	Retained bool
	Topic    string
	Payload  []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	TopicPrefix string
	Protocol    string
	// MqttClient MQTT连接
	MqttClient *mqtt.MQTTClient

	// PubMqttMsg 发送到MQTT的信息通道
	PubMqttMsg chan MqttMessage

	// TR 协议翻译器
	TR lib.ITranslate //协议翻译器

	// ResponseFn 返回函数
	ResponseFn func(ctx context.Context, payload interface{}) ([]byte, error)

	// ResponseErrFn 返回错误的参数
	ResponseErrFn func(ctx context.Context, err error, desc ...string) []byte

	//Clients Registered clients.
	Clients sync.Map

	//RegClients 需要执行注册的客户端
	RegClients sync.Map
}

func NewHub(protocol string, topicPrefix string, username string, password string) *Hub {
	//监听MQTT信息
	hostname, _ := os.Hostname()
	mqClient := mqtt.NewMQTTClient(mqtt.NewMQTTOptions(hostname, username, password, func(c mqttClient.Client) {
		logrus.Info("mqtt connected")
	}, func(c mqttClient.Client, err error) {
		logrus.Errorf("disconnect mqtt error, err:%s", err.Error())
	}, func(c mqttClient.Client, m mqttClient.Message) {
		logrus.Warnf("got mqtt unhandled msg:%v", m)
	}))

	//connect
	if err := mqClient.Connect(); err != nil {
		panic(fmt.Sprintf("connect to mqtt error, err:%s", err.Error()))
	}

	hub := &Hub{
		MqttClient:  mqClient,
		PubMqttMsg:  make(chan MqttMessage, 1000),
		TopicPrefix: topicPrefix,
		Protocol:    protocol,
	}
	return hub
}

func (h *Hub) SetTR(tr lib.ITranslate) {
	h.TR = tr
}

func (h *Hub) SendMsgToDevice(evseID string, msg []byte) error {
	if c, ok := h.Clients.Load(evseID); ok {
		return c.(*Client).Send(msg)
	}
	return fmt.Errorf("evse_id:%s offline", evseID)
}

// CloseClient 断开连接
func (h *Hub) CloseClient(evseID string) {
	h.Clients.Delete(evseID)
}

func (h *Hub) Run() {
	_, cancelFn := context.WithCancel(context.Background())
	g := errgroup.WithCancel(context.Background())
	defer cancelFn()

	//监听注册报文
	g.Go(func(ctx context.Context) (err error) {
		token := h.MqttClient.GetMQTT().Subscribe(h.TopicPrefix+"core/"+h.Protocol+"/D/R/#", 2, func(mqc mqttClient.Client, m mqttClient.Message) {
			var err error
			topic := m.Topic()

			logEntry := logrus.WithFields(logrus.Fields{
				"topic": topic,
			})
			defer func() {
				if e := recover(); e != nil {
					err = e.(error)
				}
				if err != nil {
					logEntry.Error(err)
				}
			}()

			fmt.Println("go reg mqtt msg", fmt.Sprintf("%+v", m))

			var sn string
			//根据topic获取sn
			if topics := strings.Split(topic, "/"); len(topics) < 5 {
				err = fmt.Errorf("cannot find sn from topic")
				return
			} else {
				sn = topics[5]
			}

			fmt.Println("go reg mqtt sn", sn)

			var _client *Client
			if c, ok := h.RegClients.Load(sn); !ok {
				return
			} else {
				_client = c.(*Client)
			}

			fmt.Println("go reg mqtt client", fmt.Sprintf("%+v", _client))

			_client.MqttRegCh <- MqttMessage{
				Topic:    topic,
				Payload:  m.Payload(),
				Qos:      m.Qos(),
				Retained: m.Retained(),
			}
		})
		token.WaitTimeout(10 * time.Second)
		if err = token.Error(); err != nil {
			return fmt.Errorf("sub reg cmd chan error, err:%v", err.Error())
		}
		return
	})

	//监听下发给设备的CMD报文
	g.Go(func(ctx context.Context) (err error) {
		topics := map[string]byte{
			h.TopicPrefix + "/core/"+ h.Protocol +"/D/C/#": 2,
			h.TopicPrefix + "/core/"+ h.Protocol +"/D/M/#": 2,
		}

		token := h.MqttClient.GetMQTT().SubscribeMultiple(topics, func(mqc mqttClient.Client, m mqttClient.Message) {
			var err error
			topic := m.Topic()

			logEntry := logrus.WithFields(logrus.Fields{
				"topic": topic,
			})
			defer func() {
				if e := recover(); e != nil {
					err = e.(error)
				}
				if err != nil {
					logEntry.Error(err)
				}
			}()

			fmt.Println("go mqtt msg", fmt.Sprintf("%+v", m))

			var evseID string
			//根据topic获取sn
			if topics := strings.Split(topic, "/"); len(topics) < 5 {
				err = fmt.Errorf("cannot find sn from topic")
				return
			} else {
				evseID = topics[5]
			}
			fmt.Println("go mqtt msg sn", evseID)

			c, ok := h.Clients.Load(evseID)

			var _client *Client
			if !ok {
				return
			} else {
				_client = c.(*Client)
			}
			fmt.Println("go mqtt msg client", fmt.Sprintf("%+v", _client))

			_client.MqttMsgCh <- MqttMessage{
				Topic:    topic,
				Payload:  m.Payload(),
				Qos:      m.Qos(),
				Retained: m.Retained(),
			}
		})
		token.WaitTimeout(time.Second * 5)
		if err = token.Error(); err != nil {
			return fmt.Errorf("sub core cmd chan error, err:%v", err.Error())
		}
		return
	})

	g.Go(func(ctx context.Context) (err error) {
		n := 100
		wp := workpool.New(n, n*3).Start()
		//for i := 0; i < n; i++ {
		for {
			select {
			case <-ctx.Done():
				return nil
			case m := <-h.PubMqttMsg:
				wp.PushTaskFunc(func(w *workpool.WorkPool, args ...interface{}) workpool.Flag {
					token := h.MqttClient.GetMQTT().Publish(
						m.Topic,
						m.Qos,
						m.Retained,
						m.Payload,
					)
					token.WaitTimeout(time.Second * 3)
					if err := token.Error(); err != nil {
						logrus.Errorf("pub msg to topic:%s error, %s", m.Topic, err.Error())
					}
					return workpool.FLAG_OK
				})
			}
		}
	})
	_ = g.Wait()
}
