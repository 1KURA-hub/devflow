package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const addPostsIfInboxExistsScript = `
if redis.call("EXISTS", KEYS[1]) == 1 then
	for i = 1, #ARGV, 2 do
		redis.call("ZADD", KEYS[1], ARGV[i], ARGV[i + 1])
	end
	return 1
end
return 0
`

type FeedInboxItem struct {
	PostID    uint64
	CreatedAt time.Time
}

type FeedInboxStore struct {
	client *redis.Client
}

func NewFeedInboxStore(client *redis.Client) *FeedInboxStore {
	return &FeedInboxStore{client: client}
}

func (s *FeedInboxStore) AddPost(ctx context.Context, userID, postID uint64, createdAt time.Time) error {
	return s.AddPosts(ctx, userID, []FeedInboxItem{{PostID: postID, CreatedAt: createdAt}})
}

func (s *FeedInboxStore) Exists(ctx context.Context, userID uint64) (bool, error) {
	if s == nil || s.client == nil {
		return false, nil
	}
	exists, err := s.client.Exists(ctx, feedInboxKey(userID)).Result()
	return exists > 0, err
}

func (s *FeedInboxStore) AddPosts(ctx context.Context, userID uint64, items []FeedInboxItem) error {
	if s == nil || s.client == nil || len(items) == 0 {
		return nil
	}
	args := make([]any, 0, len(items)*2)
	for _, item := range items {
		if item.PostID == 0 || item.CreatedAt.IsZero() {
			continue
		}
		args = append(args, item.CreatedAt.UnixMicro(), strconv.FormatUint(item.PostID, 10))
	}
	if len(args) == 0 {
		return nil
	}
	return s.client.Eval(ctx, addPostsIfInboxExistsScript, []string{feedInboxKey(userID)}, args...).Err()
}

func (s *FeedInboxStore) Rebuild(ctx context.Context, userID uint64, items []FeedInboxItem) error {
	if s == nil || s.client == nil {
		return nil
	}
	key := feedInboxKey(userID)
	members := make([]redis.Z, 0, len(items))
	for _, item := range items {
		if item.PostID == 0 || item.CreatedAt.IsZero() {
			continue
		}
		members = append(members, redis.Z{
			Score:  float64(item.CreatedAt.UnixMicro()),
			Member: strconv.FormatUint(item.PostID, 10),
		})
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, key)
	if len(members) > 0 {
		pipe.ZAdd(ctx, key, members...)
	}
	_, err := pipe.Exec(ctx)
	return err
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
