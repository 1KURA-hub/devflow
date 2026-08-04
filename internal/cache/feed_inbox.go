package cache

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"devflow/internal/pagination"

	"github.com/redis/go-redis/v9"
)

const (
	feedInboxMaxItems = int64(1000)
	feedInboxTTL      = 7 * 24 * time.Hour
)

const addPostsIfInboxExistsScript = `
if redis.call("EXISTS", KEYS[2]) == 1 then
	local max_items = tonumber(ARGV[1])
	local ttl_seconds = tonumber(ARGV[2])
	for i = 3, #ARGV, 2 do
		redis.call("ZADD", KEYS[1], ARGV[i], ARGV[i + 1])
	end
	local count = redis.call("ZCARD", KEYS[1])
	if count > max_items then
		redis.call("ZREMRANGEBYRANK", KEYS[1], 0, count - max_items - 1)
	end
	redis.call("EXPIRE", KEYS[1], ttl_seconds)
	redis.call("EXPIRE", KEYS[2], ttl_seconds)
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
	exists, err := s.client.Exists(ctx, feedInboxReadyKey(userID)).Result()
	return exists > 0, err
}

func (s *FeedInboxStore) AddPosts(ctx context.Context, userID uint64, items []FeedInboxItem) error {
	if s == nil || s.client == nil || len(items) == 0 {
		return nil
	}
	args := make([]any, 0, len(items)*2+2)
	args = append(args, feedInboxMaxItems, int64(feedInboxTTL/time.Second))
	for _, item := range items {
		if item.PostID == 0 || item.CreatedAt.IsZero() {
			continue
		}
		args = append(args, item.CreatedAt.UnixMicro(), strconv.FormatUint(item.PostID, 10))
	}
	if len(args) == 2 {
		return nil
	}
	return s.client.Eval(
		ctx,
		addPostsIfInboxExistsScript,
		[]string{feedInboxKey(userID), feedInboxReadyKey(userID)},
		args...,
	).Err()
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
		if int64(len(members)) > feedInboxMaxItems {
			pipe.ZRemRangeByRank(ctx, key, 0, int64(len(members))-feedInboxMaxItems-1)
		}
		pipe.Expire(ctx, key, feedInboxTTL)
	}
	pipe.Set(ctx, feedInboxReadyKey(userID), "1", feedInboxTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *FeedInboxStore) PostIDs(ctx context.Context, userID uint64, cursor *pagination.Cursor, limit int64) ([]uint64, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}
	key := feedInboxKey(userID)
	exists, err := s.client.Exists(ctx, feedInboxReadyKey(userID)).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	members, err := s.client.ZRevRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, false, err
	}

	items := make([]FeedInboxItem, 0, len(members))
	for _, member := range members {
		id, err := strconv.ParseUint(fmt.Sprint(member.Member), 10, 64)
		if err != nil {
			continue
		}
		items = append(items, FeedInboxItem{
			PostID:    id,
			CreatedAt: time.UnixMicro(int64(member.Score)),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].PostID > items[j].PostID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	ids := make([]uint64, 0, limit)
	for _, item := range items {
		if cursor != nil {
			cursorTime := cursor.CreatedAt()
			if item.CreatedAt.After(cursorTime) || (item.CreatedAt.Equal(cursorTime) && item.PostID >= cursor.ID) {
				continue
			}
		}
		ids = append(ids, item.PostID)
		if int64(len(ids)) == limit {
			break
		}
	}
	return ids, true, nil
}

func (s *FeedInboxStore) Delete(ctx context.Context, userID uint64) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Del(ctx, feedInboxKey(userID), feedInboxReadyKey(userID)).Err()
}

func (s *FeedInboxStore) RemovePosts(ctx context.Context, userID uint64, postIDs []uint64) error {
	if s == nil || s.client == nil || len(postIDs) == 0 {
		return nil
	}
	members := make([]any, 0, len(postIDs))
	for _, postID := range postIDs {
		if postID != 0 {
			members = append(members, strconv.FormatUint(postID, 10))
		}
	}
	if len(members) == 0 {
		return nil
	}
	return s.client.ZRem(ctx, feedInboxKey(userID), members...).Err()
}

func feedInboxKey(userID uint64) string {
	return fmt.Sprintf("devflow:feed:inbox:%d", userID)
}

func feedInboxReadyKey(userID uint64) string {
	return fmt.Sprintf("devflow:feed:inbox-ready:%d", userID)
}
