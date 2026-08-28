package cache

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"devflow/internal/pagination"

	"github.com/redis/go-redis/v9"
)

const (
	hotPostsKey      = "devflow:hot_posts"
	hotPostsReadyKey = "devflow:hot_posts:ready"
	hotPostsMaxItems = int64(1000)
)

const setHotScoreScript = `
local score = tonumber(ARGV[1])
local previous_score = redis.call("ZSCORE", KEYS[1], ARGV[2])
if redis.call("EXISTS", KEYS[2]) == 1 and redis.call("EXISTS", KEYS[1]) == 0 then
	redis.call("DEL", KEYS[2])
end
if score <= 0 then
	local removed = redis.call("ZREM", KEYS[1], ARGV[2])
	if removed > 0 then
		redis.call("DEL", KEYS[2])
	end
	return removed
end
redis.call("ZADD", KEYS[1], score, ARGV[2])
if previous_score and score < tonumber(previous_score) then
	redis.call("DEL", KEYS[2])
end
local count = redis.call("ZCARD", KEYS[1])
local max_items = tonumber(ARGV[3])
if count > max_items then
	redis.call("ZREMRANGEBYRANK", KEYS[1], 0, count - max_items - 1)
end
return 1
`

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
	return s.client.Eval(
		ctx,
		setHotScoreScript,
		[]string{hotPostsKey, hotPostsReadyKey},
		score,
		member,
		hotPostsMaxItems,
	).Err()
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
	pipe.Del(ctx, hotPostsKey, hotPostsReadyKey)
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
	pipe.Set(ctx, hotPostsReadyKey, "1", 0)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *HotPostStore) List(ctx context.Context, cursor *pagination.Cursor, limit int64) ([]HotPostItem, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}
	exists, err := s.client.Exists(ctx, hotPostsReadyKey).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	members, err := s.client.ZRevRangeWithScores(ctx, hotPostsKey, 0, -1).Result()
	if err != nil {
		return nil, false, err
	}
	items := make([]HotPostItem, 0, len(members))
	for _, member := range members {
		id, err := strconv.ParseUint(fmt.Sprint(member.Member), 10, 64)
		if err != nil {
			continue
		}
		items = append(items, HotPostItem{PostID: id, Score: int64(member.Score)})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].PostID > items[j].PostID
		}
		return items[i].Score > items[j].Score
	})

	page := make([]HotPostItem, 0, limit)
	for _, item := range items {
		if cursor != nil && (item.Score > cursor.Score || (item.Score == cursor.Score && item.PostID >= cursor.ID)) {
			continue
		}
		page = append(page, item)
		if int64(len(page)) == limit {
			break
		}
	}
	return page, true, nil
}
