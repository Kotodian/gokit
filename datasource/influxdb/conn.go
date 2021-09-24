package influxdb

import (
	"context"
	"github.com/Kotodian/gokit/boot"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"os"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	"github.com/silenceper/pool"
)

var _pool pool.Pool
var org string

func init() {
	boot.RegisterInit("influxdb connection pool", bootInit)
}

func bootInit() (err error) {
	org = os.Getenv("INFLUXDB_ORG")
	//factory 创建连接的方法
	factory := func() (interface{}, error) {
		client := influxdb.NewClient("http://influxdb:8086", os.Getenv("INFLUXDB_AUTH_TOKEN"))
		//client := influxdb.NewClient("http://10.43.0.15:8086", os.Getenv("INFLUXDB_AUTH_TOKEN"))
		_, err = client.Health(context.Background())
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	//close 关闭连接的方法
	_close := func(v interface{}) error {
		//fmt.Println("close connection", v)
		v.(influxdb.Client).Close()
		return nil
	}

	_pool, err = pool.NewChannelPool(&pool.Config{
		InitialCap: 10,
		MaxIdle:    100,
		MaxCap:     1000,
		Factory:    factory,
		Close:      _close,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 30 * time.Second,
	})
	return
}

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
	client, err := GetClient()
	if err != nil {
		return err
	}
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(org, bucket)
	p := influxdb.NewPoint(measurement, tags, fields, ts)
	err = writeAPI.WritePoint(context.Background(), p)
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
