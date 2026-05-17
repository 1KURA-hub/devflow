package cache

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const hotPostsKey = "devflow:hot_posts"

type HotPostStore struct {
	client *redis.Client
}

func NewHotPostStore(client *redis.Client) *HotPostStore {
	return &HotPostStore{client: client}
}

func (s *HotPostStore) SetScore(ctx context.Context, postID uint64, score int64) error {
	if s == nil || s.client == nil {
		return nil
	}
	member := strconv.FormatUint(postID, 10)
	if score <= 0 {
		return s.client.ZRem(ctx, hotPostsKey, member).Err()
	}
	return s.client.ZAdd(ctx, hotPostsKey, redis.Z{
		Score:  float64(score),
		Member: member,
	}).Err()
}

func (s *HotPostStore) TopPostIDs(ctx context.Context, limit int64) ([]uint64, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}

	members, err := s.client.ZRevRange(ctx, hotPostsKey, 0, limit-1).Result()
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
