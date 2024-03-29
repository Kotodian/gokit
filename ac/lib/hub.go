package lib

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource"

	"github.com/Kotodian/gokit/datasource/mqtt"
	"github.com/Kotodian/gokit/sync/errgroup.v2"
	"github.com/Kotodian/gokit/workpool"
	mqttClient "github.com/eclipse/paho.mqtt.golang"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	Hostname string
	// Protocol 协议
	Protocol string
	// ProtocolVersion 协议版本
	ProtocolVersion string
	// MqttClient MQTT连接
	MqttClient *mqtt.MQTTClient

	// PubMqttMsg 发送到MQTT的信息通道
	PubMqttMsg chan mqtt.MqttMessage

	// TR 协议翻译, 把协议的内容发到core-gw
	TR ITranslate //协议翻译器

	// CommandFn 发送函数
	CommandFn func(ctx context.Context, payload interface{}) ([]byte, error)
	// ResponseFn 返回函数
	ResponseFn func(ctx context.Context, payload interface{}) ([]byte, error)

	// ResponseErrFn 返回错误的参数
	ResponseErrFn func(ctx context.Context, err error, desc ...string) []byte

	//Clients Registered clients.
	Clients sync.Map

	//RegClients 需要执行注册的客户端
	RegClients sync.Map
	// Encrypt 加密报文
	Encrypt Encrypt
}

func NewHub(protocol string, protocolVersion, username string, password string) *Hub {
	//监听MQTT信息
	hostname, _ := os.Hostname()
	options := mqtt.NewMQTTOptions(hostname, username, password, func(c mqttClient.Client) {
		// logrus.Info("mqtt connected")
	}, func(c mqttClient.Client, err error) {
		// logrus.Errorf("disconnect mqtt error, err:%s", err.Error())
		c.Publish("coregw/disconnect/"+hostname, 2, false, "")
	}, func(c mqttClient.Client, m mqttClient.Message) {
		// logrus.Warnf("got mqtt unhandled msg:%v", m)
	}, false)
	// 设置遗愿消息
	options = options.SetWill("coregw/disconnect/"+hostname, "", 2, false)
	mqClient := mqtt.NewMQTTClient(options)
	//connect
	if err := mqClient.Connect(); err != nil {
		panic(fmt.Sprintf("connect to mqtt error, err:%s", err.Error()))
	}

	hub := &Hub{
		Hostname:        hostname,
		MqttClient:      mqClient,
		PubMqttMsg:      make(chan mqtt.MqttMessage, 1000),
		Protocol:        protocol,
		ProtocolVersion: protocolVersion,
	}
	return hub
}

func (h *Hub) SetTR(tr ITranslate) {
	h.TR = tr
}

func (h *Hub) SetEncrypt(encrypt Encrypt) {
	h.Encrypt = encrypt
}

func (h *Hub) SendMsgToDevice(evse interface{}, msg []byte) error {
	if c, ok := h.Clients.Load(evse); ok {
		return c.(ClientInterface).Send(msg)
	}
	return fmt.Errorf("sn:%s offline", evse)
}

// CloseClient 断开连接
func (h *Hub) CloseClient(evse interface{}) {
	h.Clients.Delete(evse)
}

func (h *Hub) Run() {
	_, cancelFn := context.WithCancel(context.Background())
	g := errgroup.WithCancel(context.Background())
	defer cancelFn()
	topicPrefix := h.Hostname + "/"
	topicEnd := "/#"
	//监听注册报文
	g.Go(func(ctx context.Context) (err error) {
		token := h.MqttClient.GetMQTT().Subscribe(topicPrefix+"register"+topicEnd, 2, func(mqc mqttClient.Client, m mqttClient.Message) {
			topic := m.Topic()
			_, coreID := getCoreIDFromTopic(m.Topic())

			var _client ClientInterface
			if c, ok := h.RegClients.Load(coreID); !ok {
				return
			} else {
				_client = c.(ClientInterface)
			}

			fmt.Println("go reg mqtt client", fmt.Sprintf("%+v", _client))

			_client.PublishReg(mqtt.MqttMessage{
				Topic:    topic,
				Payload:  m.Payload(),
				Qos:      m.Qos(),
				Retained: m.Retained(),
			})
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
			topicPrefix + "command" + topicEnd:   2,
			topicPrefix + "telemetry" + topicEnd: 2,
		}

		token := h.MqttClient.GetMQTT().SubscribeMultiple(topics, func(mqc mqttClient.Client, m mqttClient.Message) {
			var err error
			topic := m.Topic()

			defer func() {
				if err != nil {

				}
			}()

			//fmt.Println("go mqtt msg", fmt.Sprintf("%+v", m))

			coregw, coreID := getCoreIDFromTopic(topic)

			c, ok := h.Clients.Load(coreID)

			var _client ClientInterface
			if !ok {
				fmt.Printf("chargingStation:%d offline", coreID)
				return
			} else {
				_client = c.(ClientInterface)
			}
			if coregw != "" {
				_client.SetCoregw(coregw)
			}
			//fmt.Println("go mqtt msg client", fmt.Sprintf("%+v", _client))

			_client.Publish(mqtt.MqttMessage{
				Topic:    topic,
				Payload:  m.Payload(),
				Qos:      m.Qos(),
				Retained: m.Retained(),
			})
		})
		token.WaitTimeout(time.Second * 5)
		if err = token.Error(); err != nil {
			return fmt.Errorf("sub core cmd chan error, err:%v", err.Error())
		}
		return
	})

	g.Go(func(ctx context.Context) (err error) {
		n := 10
		wp := workpool.New(n, n*3).Start()
		defer wp.Stop()
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
						// logrus.Errorf("pub msg to topic:%s error, %s", m.Topic, err.Error())
					}
					return workpool.FLAG_OK
				})
			}
		}
	})
	// 监听踢掉设备的报文
	g.Go(func(ctx context.Context) error {
		token := h.MqttClient.GetMQTT().Subscribe(topicPrefix+"kick"+topicEnd, 2, func(mqc mqttClient.Client, message mqttClient.Message) {

			//根据topic获取sn
			var coreID uint64

			_, coreID = getCoreIDFromTopic(message.Topic())

			c, ok := h.Clients.Load(coreID)

			var _client ClientInterface
			if !ok {
				// logrus.Warnf("kick client not found, topic:%s", message.Topic())
				return
			} else {
				_client = c.(ClientInterface)
				_client.Close(nil)
			}
		})
		token.WaitTimeout(10 * time.Second)
		if err := token.Error(); err != nil {
			return fmt.Errorf("sub reg cmd chan error, err:%v", err.Error())
		}
		return nil
	})
	_ = g.Wait()
}

func getCoreIDFromTopic(topic string) (coregw string, coreID uint64) {
	//根据topic获取sn
	topics := strings.Split(topic, "/")
	length := len(topics)
	lastIndex, secondLastIndex := length-1, length-2
	temp, _ := datasource.ParseUUID(topics[lastIndex])
	coreID = temp.Uint64()
	if topics[secondLastIndex] != "command" && topics[secondLastIndex] != "telemetry" {
		coregw = topics[secondLastIndex]
	}
	return
}
