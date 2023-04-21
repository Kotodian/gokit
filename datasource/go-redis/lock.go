package goredis

import (
	"context"
	"time"
)

// 实现一个redis分布式锁，一定时间内没获取到锁就返回超时错误
func AcquireLock(ctx context.Context, key string, timeout time.Duration) error {
	err := rdb.SetNX(ctx, key, 0, timeout).Err()
	return err
}

func ReleaseLock(ctx context.Context, key string) error {
	err := rdb.Del(ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func TryAcquireLock(ctx context.Context, key string, timeout time.Duration) error {
	err := rdb.SetNX(ctx, key, 0, timeout).Err()
	if err != nil {
		return err
	}
	return nil
}
