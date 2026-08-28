package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	followingSetPrefix = "devflow:user:following:"
	followerSetPrefix  = "devflow:user:followers:"
	followSetSentinel  = "0"
	followRelationTTL  = 5 * time.Minute
)

type FollowRelationStore struct {
	client *redis.Client
}

func NewFollowRelationStore(client *redis.Client) *FollowRelationStore {
	return &FollowRelationStore{client: client}
}

func (s *FollowRelationStore) FollowingIDs(ctx context.Context, userID uint64) ([]uint64, bool, error) {
	return s.ids(ctx, followingSetKey(userID))
}

func (s *FollowRelationStore) SetFollowingIDs(ctx context.Context, userID uint64, ids []uint64) error {
	return s.replaceIDs(ctx, followingSetKey(userID), ids)
}

func (s *FollowRelationStore) FollowerIDs(ctx context.Context, userID uint64) ([]uint64, bool, error) {
	return s.ids(ctx, followerSetKey(userID))
}

func (s *FollowRelationStore) SetFollowerIDs(ctx context.Context, userID uint64, ids []uint64) error {
	return s.replaceIDs(ctx, followerSetKey(userID), ids)
}

func (s *FollowRelationStore) AddFollow(ctx context.Context, followerID, followeeID uint64) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.invalidatePair(ctx, followerID, followeeID)
}

func (s *FollowRelationStore) RemoveFollow(ctx context.Context, followerID, followeeID uint64) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.invalidatePair(ctx, followerID, followeeID)
}

func (s *FollowRelationStore) ids(ctx context.Context, key string) ([]uint64, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	if exists == 0 {
		return nil, false, nil
	}

	members, err := s.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, false, err
	}
	ids := make([]uint64, 0, len(members))
	for _, member := range members {
		if member == followSetSentinel {
			continue
		}
		id, err := strconv.ParseUint(member, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, true, nil
}

func (s *FollowRelationStore) replaceIDs(ctx context.Context, key string, ids []uint64) error {
	if s == nil || s.client == nil {
		return nil
	}
	members := make([]any, 0, len(ids)+1)
	members = append(members, followSetSentinel)
	for _, id := range ids {
		members = append(members, strconv.FormatUint(id, 10))
	}

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, key)
	pipe.SAdd(ctx, key, members...)
	pipe.Expire(ctx, key, followRelationTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *FollowRelationStore) invalidatePair(ctx context.Context, followerID, followeeID uint64) error {
	return s.client.Del(
		ctx,
		followingSetKey(followerID),
		followerSetKey(followeeID),
	).Err()
}

func followingSetKey(userID uint64) string {
	return fmt.Sprintf("%s%d", followingSetPrefix, userID)
}

func followerSetKey(userID uint64) string {
	return fmt.Sprintf("%s%d", followerSetPrefix, userID)
}
