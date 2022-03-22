package mqtt

import (
	"os"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	os.Setenv("EMQX_POOL", "tcp://10.43.0.11:1883")
	options := NewMQTTOptions("lqk_test", "core_ocpp", "core_gw", nil, nil, nil, true)
	options = options.SetWill("coregw/disconnect/jx-ac-ohf", "", 2, false)
	client := NewMQTTClient(options)
	err := client.Connect()
	if err != nil {
		t.Error(err)
	}
	// client.mqtt.Publish("coregw/disconnect/jx-ac-ohf", 2, false, "Hello World")
	<-time.NewTimer(5 * time.Second).C
}
