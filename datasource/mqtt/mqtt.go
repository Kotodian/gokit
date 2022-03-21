package mqtt

import (
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	//log "github.com/sirupsen/logrus"
)

const (
	EnvEmqxPool = "EMQX_POOL"
	EnvEmqxUser = "EMQX_USER"
	EnvEmqxPass = "EMQX_PASS"
)

type MqttMessage struct {
	Qos      byte
	Retained bool
	Topic    string
	Payload  []byte
}
type MQTTClient struct {
	mqtt mqtt.Client
}

func (mq *MQTTClient) GetMQTT() mqtt.Client {
	return mq.mqtt
}

func (mq *MQTTClient) Connect() error {
	if mq.mqtt.IsConnected() {
		return nil
	}
	var err error
	for retries := 0; retries < 3; retries++ {
		token := mq.mqtt.Connect()
		token.WaitTimeout(1 * time.Second)

		if err = token.Error(); err == nil {
			return nil
		}
	}
	return err
}

func NewMQTTClient(mqttOpts *mqtt.ClientOptions) *MQTTClient {
	// mqttOpts.SetMaxReconnectInterval(time.Second * 3)
	return &MQTTClient{
		mqtt: mqtt.NewClient(mqttOpts),
	}
}

func NewMQTTOptions(clientID string, username string, password string, onConn mqtt.OnConnectHandler, onLostConn mqtt.ConnectionLostHandler, onMsg mqtt.MessageHandler, clean bool) *mqtt.ClientOptions {
	mqttOpts := mqtt.NewClientOptions()
	mqttOpts.AddBroker(os.Getenv(EnvEmqxPool))
	mqttOpts.SetClientID(clientID)
	mqttOpts.SetUsername(username)
	mqttOpts.SetPassword(password)
	mqttOpts.SetKeepAlive(time.Second * 30)
	mqttOpts.SetPingTimeout(time.Second * 10)
	mqttOpts.SetCleanSession(true)
	mqttOpts.SetConnectionLostHandler(onLostConn)
	mqttOpts.SetOnConnectHandler(onConn)
	mqttOpts.SetDefaultPublishHandler(onMsg)
	mqttOpts.SetConnectRetry(true)
	mqttOpts.SetConnectRetryInterval(2 * time.Second)
	return mqttOpts
}

type ClientInfo struct {
	ClientID  string `json:"client_id"`
	IPAddress string `json:"ipaddress"`
	KeepAlive uint32 `json:"keepalive"`
	Port      uint32 `json:"port"`
	ProtoVer  uint32 `json:"proto_ver"`
}

func SharePrefix() string {
	return "$share/"
}

func QueuePrefix() string {
	return "$queue/"
}

func SystemPrefix() string {
	return "$SYS/"
}
