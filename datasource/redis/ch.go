package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	errgroup "github.com/Kotodian/gokit/sync/errgroup.v2"
	"github.com/gomodule/redigo/redis"
)

func SubCH(ctx context.Context, name string, out interface{}, timeout time.Duration) error {
	rd := GetRedis()
	psc := redis.PubSubConn{Conn: rd}
	defer func() {
		psc.Close()
		rd.Close()
	}()
	if err := psc.Subscribe(name); err != nil {
		return err
	}

	g := errgroup.WithTimeout(ctx, timeout)
	g.Go(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timeout")
			default:
			}
			resp := psc.ReceiveWithTimeout(timeout)
			switch resp.(type) {
			case redis.Message: //单个订阅subscribe
				return json.Unmarshal(resp.(redis.Message).Data, out)
			case error:
				return resp.(error)
			}
		}
		return nil
	})
	return g.Wait()
}

func PubCH(name string, obj interface{}) error {
	rd := GetRedis()
	defer rd.Close()
	o, _ := json.Marshal(obj)
	if _, err := rd.Do("publish", name, o); err != nil {
		return err
	}
	return nil
}
