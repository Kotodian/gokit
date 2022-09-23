package influxdb

import (
	"os"
	"testing"
	"time"
)

func TestWriteAPIBlocking(t *testing.T) {
	os.Setenv("INFLUXDB_ORG", "joyson")
	os.Setenv("INFLUXDB_AUTH_TOKEN", "CwcKwmhdhDL5vPdiKLenll5aOgqT_aPmSSkGzo1nUB5BxdFTXaAkJQRPmfd3Yrf6zoQjmAJ6UQZ8wRXPDO5lfw==")
	Init()
	for i := 0; i < 100; i++ {
		err := WriteAPIBlocking("test", "charging_info", map[string]string{
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
}

func TestQuery(t *testing.T) {
	os.Setenv("INFLUXDB_POOL", "10.43.0.15:8086")
	os.Setenv("INFLUXDB_ORG", "joyson")
	os.Setenv("INFLUXDB_AUTH_TOKEN", "CwcKwmhdhDL5vPdiKLenll5aOgqT_aPmSSkGzo1nUB5BxdFTXaAkJQRPmfd3Yrf6zoQjmAJ6UQZ8wRXPDO5lfw==")
	Init()

	result, err := Query(`from(bucket: "jxcsms-runtime")
  |> range(start: -1y)
  |> filter(fn: (r) => r["ConnectorId"] == "281536629641989")
  |> max()`)
	if err != nil {
		t.Error(err)
	} else {
		for result.Next() {
			t.Logf("%s: %f", result.Record().Field(), result.Record().Value())
		}
	}
}
