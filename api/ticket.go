package api

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type TicketManager interface {
	Get(ctx context.Context) (string, error)
}

type redisTicketManager struct {
	redis redis.UniversalClient
}

func NewRedisTicketManager(redis redis.UniversalClient) TicketManager {
	return &redisTicketManager{redis: redis}
}

func (t *redisTicketManager) Get(ctx context.Context) (string, error) {
	return t.redis.Get(ctx, "Service:Internal:Tickets").Result()
}
