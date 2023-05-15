package api

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type TicketManager interface {
	Get(ctx context.Context) (string, error)
}

type redisTicketManager struct {
	redis redis.Cmdable
}

func NewRedisTicketManager(redis redis.Cmdable) TicketManager {
	return &redisTicketManager{redis: redis}
}

func (t *redisTicketManager) Get(ctx context.Context) (string, error) {
	return t.redis.Get(ctx, "Service:Internal:Tickets").Result()
}
