package influxdb

import (
	"context"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"os"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/silenceper/pool"
)

var _pool pool.Pool
var org string
var token string
var url string

// Init Deprecated
func Init() {
	org = os.Getenv("INFLUXDB_ORG")
	token = os.Getenv("INFLUXDB_AUTH_TOKEN")
	url = "http://" + os.Getenv("INFLUXDB_POOL")
}

// GetClient Deprecated
func GetClient() (influxdb.Client, error) {
	c, err := _pool.Get()
	if err != nil {
		return nil, err
	}
	return c.(influxdb.Client), nil
}

func CloseClient(v interface{}) error {
	return _pool.Put(v)
}

func WriteAPIBlocking(bucket, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	client := influxdb.NewClient(url, token)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(org, bucket)
	p := influxdb.NewPoint(measurement, tags, fields, ts)
	err := writeAPI.WritePoint(context.Background(), p)
	return err
}

func Query(query string) (*api.QueryTableResult, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	queryAPI := client.QueryAPI(org)
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func WriteAPI(bucket, measurement string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	client := influxdb.NewClient("http://influxdb:8086", token)

	writeAPI := client.WriteAPI(org, bucket)
	p := influxdb.NewPoint(measurement, tags, fields, ts)
	writeAPI.WritePoint(p)

	writeAPI.Flush()
	client.Close()
}
