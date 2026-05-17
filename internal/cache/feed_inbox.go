package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type FeedInboxStore struct {
	client *redis.Client
}

func NewFeedInboxStore(client *redis.Client) *FeedInboxStore {
	return &FeedInboxStore{client: client}
}

func (s *FeedInboxStore) AddPost(ctx context.Context, userID, postID uint64, createdAt time.Time) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.ZAdd(ctx, feedInboxKey(userID), redis.Z{
		Score:  float64(createdAt.UnixMicro()),
		Member: strconv.FormatUint(postID, 10),
	}).Err()
}

func (s *FeedInboxStore) PostIDs(ctx context.Context, userID uint64, cursor *time.Time, limit int64) ([]uint64, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}
	key := feedInboxKey(userID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	max := "+inf"
	if cursor != nil {
		max = fmt.Sprintf("(%d", cursor.UnixMicro())
	}
	members, err := s.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Max:   max,
		Min:   "-inf",
		Count: limit,
	}).Result()
	if err != nil {
		return nil, false, err
	}

	ids := make([]uint64, 0, len(members))
	for _, member := range members {
		id, err := strconv.ParseUint(member, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, true, nil
}

func (s *FeedInboxStore) Delete(ctx context.Context, userID uint64) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Del(ctx, feedInboxKey(userID)).Err()
}

func feedInboxKey(userID uint64) string {
	return fmt.Sprintf("devflow:feed:inbox:%d", userID)
}
