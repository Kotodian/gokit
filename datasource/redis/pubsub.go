package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Kotodian/gokit/sync/errgroup.v2"

	rdlib "github.com/gomodule/redigo/redis"
)

// Publish publishes message to channel.
func Publish(channel, message interface{}) (int, error) {
	c := GetRedis()
	defer c.Close()
	n, err := rdlib.Int(c.Do("PUBLISH", channel, message))
	if err != nil {
		return 0, fmt.Errorf("redis publish %s %s, err: %v", channel, message, err)
	}
	return n, nil
}

// ConsumeFunc consumes message at the channel.
// _break:0 continue _break:1 break
type ConsumeFunc func(channel string, message []byte) (_break bool, err error)

// Subscribe subscribes messages at the channels.
func SubscribeWithTimeout(ctx context.Context, timeout time.Duration, consume ConsumeFunc, channel ...string) error {
	c := GetRedis()
	psc := rdlib.PubSubConn{Conn: c}
	defer func() {
		psc.Close()
		_ = c.Close()
	}()
	if err := psc.Subscribe(rdlib.Args{}.AddFlat(channel)...); err != nil {
		return err
	}

	done := make(chan struct{}, 1)
	g := errgroup.WithTimeout(ctx, timeout)
	g.Go(func(ctx context.Context) (err error) {
		defer func() {
			_ = psc.Unsubscribe()
			close(done)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			var b bool
			_msg := psc.ReceiveWithTimeout(10 * time.Second)
			switch msg := _msg.(type) {
			//	return msg
			case error:
				err = msg
				return
			case rdlib.Message:
				b, err = consume(msg.Channel, msg.Data)
				if err != nil {
					return
				} else if !b {
					continue
				}
				return
			case rdlib.Subscription:
				if msg.Count == 0 {
					return
				}
			}
		}
	})

	//g.Go(func(ctx context.Context) (err error) {
	//	// health check
	//	tick := time.NewTicker(1 * time.Second)
	//	defer func() {
	//		fmt.Println("sub go2 exit,", err)
	//		tick.Stop()
	//	}()
	//	for {
	//		select {
	//		case <-ctx.Done():
	//			fmt.Println("ctx done 2")
	//			_ = psc.Unsubscribe()
	//			//return fmt.Errorf("timeout")
	//		case <-done:
	//			return nil
	//		case <-tick.C:
	//			fmt.Println("ping !!!!")
	//			if err = psc.Ping("a"); err != nil {
	//				return err
	//			}
	//		}
	//	}
	//})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

//
//// Subscribe subscribes messages at the channels.
//func Subscribe(ctx context.Context, consume ConsumeFunc, channel ...string) error {
//	c := GetRedis()
//	defer func() {
//		c.Close()
//	}()
//	psc := rdlib.PubSubConn{Conn: c}
//
//	if err := psc.Subscribe(rdlib.Args{}.AddFlat(channel)...); err != nil {
//		return err
//	}
//	done := make(chan error, 1)
//	// start a new goroutine to receive message
//	go func() {
//		defer psc.Close()
//		for {
//			switch msg := psc.Receive().(type) {
//			case error:
//				done <- fmt.Errorf("redis pubsub receive err: %v", msg)
//				return
//			case redis.Message:
//				if err := consume(msg.Channel, msg.Data); err != nil {
//					done <- err
//					return
//				}
//			case redis.Subscription:
//				if msg.Count == 0 {
//					// all channels are unsubscribed
//					done <- nil
//					return
//				}
//			}
//		}
//	}()
//
//	ch <- 0
//
//	// health check
//	tick := time.NewTicker(time.Minute)
//	defer tick.Stop()
//	for {
//		select {
//		case <-ctx.Done():
//			if err := psc.Unsubscribe(); err != nil {
//				return fmt.Errorf("redis pubsub unsubscribe err: %v", err)
//			}
//			return nil
//		case err := <-done:
//			return err
//		case <-tick.C:
//			if err := psc.Ping(""); err != nil {
//				return err
//			}
//		}
//	}
//
//}
