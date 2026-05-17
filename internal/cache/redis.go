package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func OpenRedis(ctx context.Context, addr string) (*redis.Client, error) {
	if addr == "" {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
