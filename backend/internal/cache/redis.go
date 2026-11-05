package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, address string) (*redis.Client, error) {
	if address == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}

	return client, nil
}
