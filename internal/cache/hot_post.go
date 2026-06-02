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

// HotPostItem 是 Rebuild 的输入：post_id + 当前热度分数。
type HotPostItem struct {
	PostID uint64
	Score  int64
}

// Rebuild 用从 DB 算出的 topN 整体替换热门榜，
// 解决 ZSET 丢失/淘汰后被零散互动写入造成的"半截榜"问题。
func (s *HotPostStore) Rebuild(ctx context.Context, items []HotPostItem) error {
	if s == nil || s.client == nil {
		return nil
	}
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, hotPostsKey)
	members := make([]redis.Z, 0, len(items))
	for _, item := range items {
		if item.PostID == 0 || item.Score <= 0 {
			continue
		}
		members = append(members, redis.Z{
			Score:  float64(item.Score),
			Member: strconv.FormatUint(item.PostID, 10),
		})
	}
	if len(members) > 0 {
		pipe.ZAdd(ctx, hotPostsKey, members...)
	}
	_, err := pipe.Exec(ctx)
	return err
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
