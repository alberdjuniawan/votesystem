package redis

import (
	"context"
	"fmt"

	"github.com/alberdjuniawan/votesystem/internal/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("failed to instrument redis with tracing: %w", err)
	}

	return rdb, nil
}
