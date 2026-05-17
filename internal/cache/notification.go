package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const incrementIfExistsScript = `
if redis.call("EXISTS", KEYS[1]) == 1 then
	return redis.call("INCR", KEYS[1])
end
return 0
`

type NotificationCounter struct {
	client *redis.Client
}

func NewNotificationCounter(client *redis.Client) *NotificationCounter {
	return &NotificationCounter{client: client}
}

func (c *NotificationCounter) Get(ctx context.Context, userID uint64) (int64, bool, error) {
	if c == nil || c.client == nil {
		return 0, false, nil
	}

	count, err := c.client.Get(ctx, unreadNotificationKey(userID)).Int64()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return count, true, nil
}

func (c *NotificationCounter) Set(ctx context.Context, userID uint64, count int64) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Set(ctx, unreadNotificationKey(userID), count, 0).Err()
}

func (c *NotificationCounter) IncrementIfExists(ctx context.Context, userID uint64) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Eval(ctx, incrementIfExistsScript, []string{unreadNotificationKey(userID)}).Err()
}

func (c *NotificationCounter) Delete(ctx context.Context, userID uint64) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Del(ctx, unreadNotificationKey(userID)).Err()
}

func unreadNotificationKey(userID uint64) string {
	return fmt.Sprintf("devflow:notification:unread:%d", userID)
}
