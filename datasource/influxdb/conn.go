package influxdb

import (
	"time"

	"github.com/Kotodian/gokit/boot"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/silenceper/pool"
)

var _pool pool.Pool

func init() {
	boot.RegisterInit("influxdb connection pool", bootInit)
}

func bootInit() (err error) {
	//factory 创建连接的方法
	factory := func() (interface{}, error) {
		//fmt.Println("new connection")
		return client.NewHTTPClient(client.HTTPConfig{
			Addr:               "http://influxdb:8086",
			InsecureSkipVerify: true,
			Timeout:            time.Second * 300,
			// Username: username,
			// Password: password,
		})
	}

	//ping := func(v interface{}) error { return nil }

	//close 关闭连接的方法
	_close := func(v interface{}) error {
		//fmt.Println("close connection", v)
		return v.(client.Client).Close()
	}

	_pool, err = pool.NewChannelPool(&pool.Config{
		InitialCap: 10,
		MaxIdle:    100,
		MaxCap:     1000,
		Factory:    factory,
		Close:      _close,
		//Ping:       ping,
		//连接最大空闲时间，超过该时间的连接 将会关闭，可避免空闲时连接EOF，自动失效的问题
		IdleTimeout: 30 * time.Second,
	})
	return
}

// queryDB convenience function to query the database
func QueryDBWithClient(clnt client.Client, q client.Query) (res []client.Result, err error) {
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func GetClient() (client.Client, error) {
	c, err := _pool.Get()
	if err != nil {
		return nil, err
	}
	return c.(client.Client), nil
}

func CloseClient(v interface{}) error {
	return _pool.Put(v)
}

func QueryDB(q client.Query) error {
	c, err := GetClient()
	if err != nil {
		return err
	}
	defer CloseClient(c)
	if _, err := QueryDBWithClient(c, q); err != nil {
		return err
	}
	return nil
}

func SaveWithNS(db string, name string, tags map[string]string, fields map[string]interface{}, t time.Time, retention ...string) (err error) {
	// Create a new HTTPClient
	var c client.Client
	c, err = GetClient()
	if err != nil {
		// log.Fatal(err)
		return err
	}
	defer CloseClient(c)

	r := "week"
	if len(retention) > 0 {
		r = retention[0]
	}
	// Create a new point batch
	var bp client.BatchPoints

	bp, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:        db,
		Precision:       "ns",
		RetentionPolicy: r,
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	// tags := map[string]string{"cpu": "cpu-total"}
	// fields := map[string]interface{}{
	// 	"idle":   10.1,
	// 	"system": 53.3,
	// 	"user":   46.6,
	// }
	//defer func() {
	//	fmt.Println("\n\n", name, tags, fields, t, r, err, "\n\n")
	//}()

	var pt *client.Point
	pt, err = client.NewPoint(name, tags, fields, t)
	if err != nil {
		// log.Fatal(err)
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err = c.Write(bp); err != nil {
		// log.Fatal(err)
		return err
	}

	return nil
	// Close client resources
	// if err := c.Close(); err != nil {
	// 	log.Fatal(err)
	// }
}

func Save(db string, name string, tags map[string]string, fields map[string]interface{}, t time.Time, retention ...string) (err error) {
	// Create a new HTTPClient
	var c client.Client
	c, err = GetClient()
	if err != nil {
		// log.Fatal(err)
		return err
	}
	defer CloseClient(c)

	r := "week"
	if len(retention) > 0 {
		r = retention[0]
	}
	// Create a new point batch
	var bp client.BatchPoints

	bp, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:        db,
		Precision:       "s",
		RetentionPolicy: r,
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	// tags := map[string]string{"cpu": "cpu-total"}
	// fields := map[string]interface{}{
	// 	"idle":   10.1,
	// 	"system": 53.3,
	// 	"user":   46.6,
	// }
	//defer func() {
	//	fmt.Println("\n\n", name, tags, fields, t, r, err, "\n\n")
	//}()

	var pt *client.Point
	pt, err = client.NewPoint(name, tags, fields, t)
	if err != nil {
		// log.Fatal(err)
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err = c.Write(bp); err != nil {
		// log.Fatal(err)
		return err
	}

	return nil
	// Close client resources
	// if err := c.Close(); err != nil {
	// 	log.Fatal(err)
	// }
}
