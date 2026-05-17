package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	followingSetPrefix = "devflow:user:following:"
	followerSetPrefix  = "devflow:user:followers:"
	followSetSentinel  = "0"
)

const mutateExistingSetScript = `
if redis.call("EXISTS", KEYS[1]) == 1 then
	return redis.call(ARGV[1], KEYS[1], ARGV[2])
end
return 0
`

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
	member := strconv.FormatUint(followeeID, 10)
	if err := s.mutateExistingSet(ctx, followingSetKey(followerID), "SADD", member); err != nil {
		return err
	}
	return s.mutateExistingSet(ctx, followerSetKey(followeeID), "SADD", strconv.FormatUint(followerID, 10))
}

func (s *FollowRelationStore) RemoveFollow(ctx context.Context, followerID, followeeID uint64) error {
	if s == nil || s.client == nil {
		return nil
	}
	member := strconv.FormatUint(followeeID, 10)
	if err := s.mutateExistingSet(ctx, followingSetKey(followerID), "SREM", member); err != nil {
		return err
	}
	return s.mutateExistingSet(ctx, followerSetKey(followeeID), "SREM", strconv.FormatUint(followerID, 10))
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
	_, err := pipe.Exec(ctx)
	return err
}

func (s *FollowRelationStore) mutateExistingSet(ctx context.Context, key, command, member string) error {
	return s.client.Eval(ctx, mutateExistingSetScript, []string{key}, command, member).Err()
}

func followingSetKey(userID uint64) string {
	return fmt.Sprintf("%s%d", followingSetPrefix, userID)
}

func followerSetKey(userID uint64) string {
	return fmt.Sprintf("%s%d", followerSetPrefix, userID)
}
