package redis

import (
	"context"
	"fmt"
	errgroup "github.com/Kotodian/gokit/sync/errgroup.v2"
	rdlib "github.com/gomodule/redigo/redis"
	"time"
)

//func NewLock(key string) Lock {
//	return Lock{key: key}
//}
//
//type Lock struct {
//	key string
//	ver int64
//}
//
//func (l *Lock) Lock(timeout int) (locked bool, err error) {
//	if locked, l.ver, err = TryLock(l.key, timeout); err != nil {
//		return
//	}
//	return
//}
//
//func (l *Lock) BlockLock(timeout int) (err error) {
//	var locked bool
//	if locked, l.ver, err = TryLock(l.key, timeout, true); err != nil {
//		return
//	} else if !locked {
//		err = errors.New("获取订单状态锁失败")
//		return
//	}
//	return
//}
//
//func (l Lock) Unlock() error {
//	return Unlock(l.key, l.ver)
//}

// TryLock
// key 键名
// timeout 获取锁多长时间超时，单位秒
// block 是否堵塞等待，默认为false：获取不了锁就返回错误, true：则会堵塞等待，直到获取到锁或超时
//参考 https://huoding.com/2015/09/14/463
func TryLock(conn rdlib.Conn, key string, timeout int, block ...bool) (locked bool, val int64, err error) {
	var ret string
	var b bool
	if len(block) > 0 {
		b = block[0]
	}
	if b {
		g := errgroup.WithTimeout(context.TODO(), time.Duration(timeout)*time.Second)
		g.Go(func(ctx context.Context) error {
			t := time.NewTicker(500 * time.Millisecond)
			firstExec := make(chan struct{}, 1)
			defer func() {
				close(firstExec)
				t.Stop()
			}()
			firstExec <- struct{}{}

			for {
				select {
				case <-ctx.Done():
					return fmt.Errorf("timeout")
				case <-firstExec:
				case <-t.C:
				}
				val = time.Now().UnixNano()
				ret, err = rdlib.String(conn.Do("set", key, val, "ex", timeout, "nx"))
				if err != nil {
					if err == rdlib.ErrNil {
						continue
					}
					return err
				} else if ret == "OK" {
					locked = true
					return nil
				}
			}
		})
		err = g.Wait()
		return
		//err =
	} else {
		val = time.Now().UnixNano()
		if ret, err = rdlib.String(conn.Do("set", key, val, "ex", timeout, "nx")); err != nil {
			if err == rdlib.ErrNil {
				err = fmt.Errorf("get lock failed")
			}
			return
		} else if ret == "OK" {
			locked = true
			return
		}
	}
	return
}

// Unlock 解锁
func Unlock(rd rdlib.Conn, key string, val int64) error {
	v, err := rdlib.Int64(rd.Do("get", key))
	if err != nil {
		return err
	} else if v == val {
		rd.Do("del", key)
	} else {
		return fmt.Errorf("unlock fail, version not match")
	}
	return nil
}
