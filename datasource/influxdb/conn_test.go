package influxdb

import (
	"os"
	"testing"
	"time"
)

func TestWriteAPIBlocking(t *testing.T) {
	os.Setenv("INFLUXDB_ORG", "joyson")
	os.Setenv("INFLUXDB_AUTH_TOKEN", "aFKnvJyfEo74Dzo8Q1GFRVobiV3vZ57khfwijnSs7EUX5gni4SiaOrW22LcSyIdlPdZA8ROHcmijfe7jMKykew==")
	err := bootInit()
	if err != nil {
		t.Error(err)
		return
	}
	err = WriteAPIBlocking("jxcsms", "charging_info", map[string]string{
		"sn":                      "T1641735211",
		"connector_serial_number": "1",
		"transaction_id":          "212123124312",
	}, map[string]interface{}{
		"start_electricity": 0.0,
		"electricity":       10.1,
		"power":             10.2,
		"current":           50,
	}, time.Now())
	if err != nil {
		t.Error(err)
		return
	}
}
