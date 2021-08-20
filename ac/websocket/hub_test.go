package websocket

import (
	"github.com/Kotodian/gokit/datasource/mqtt"
	"os"
	"testing"
)

func TestHub_Run(t *testing.T) {
	err := os.Setenv(mqtt.EnvEmqxPool, "tcp://10.43.0.11:1883")
	if err != nil {
		t.Error(err)
		return
	}
	hub := NewHub("ocpp", "2.0.1", "core_ocpp", "core_gw")
	hub.Run()
	select {}
}
