package service

import (
	"context"
	"errors"

	"devflow/internal/cache"
	"devflow/internal/model"
	"devflow/internal/repository"
)

var ErrCannotFollowSelf = errors.New("cannot follow self")
var ErrAlreadyFollowed = errors.New("already followed")

const followingFeedInboxRebuildLimit = 500

type FollowService struct {
	follows       *repository.FollowRepository
	users         *repository.UserRepository
	posts         *repository.PostRepository
	notifications *NotificationService
	relations     *cache.FollowRelationStore
	inboxes       *cache.FeedInboxStore
}

type UserListInput struct {
	Limit  int
	Offset int
}

type UserListResult struct {
	Items []model.User `json:"items"`
}

type UserProfileStats struct {
	Posts     int64 `json:"posts"`
	Followers int64 `json:"followers"`
	Following int64 `json:"following"`
}

type UserProfileResult struct {
	User  *model.User      `json:"user"`
	Stats UserProfileStats `json:"stats"`
}

func NewFollowService(follows *repository.FollowRepository, users *repository.UserRepository, posts *repository.PostRepository, notifications *NotificationService, relations *cache.FollowRelationStore, inboxes *cache.FeedInboxStore) *FollowService {
	return &FollowService{
		follows:       follows,
		users:         users,
		posts:         posts,
		notifications: notifications,
		relations:     relations,
		inboxes:       inboxes,
	}
}

func (s *FollowService) GetUserProfile(ctx context.Context, userID uint64) (*UserProfileResult, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	postCount, err := s.posts.CountByAuthor(ctx, userID)
	if err != nil {
		return nil, err
	}
	followerCount, err := s.follows.CountFollowers(ctx, userID)
	if err != nil {
		return nil, err
	}
	followingCount, err := s.follows.CountFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UserProfileResult{
		User: user,
		Stats: UserProfileStats{
			Posts:     postCount,
			Followers: followerCount,
			Following: followingCount,
		},
	}, nil
}

func (s *FollowService) Follow(ctx context.Context, followerID, followeeID uint64) error {
	if followerID == 0 || followeeID == 0 {
		return ErrInvalidInput
	}
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	if _, err := s.users.FindByID(ctx, followeeID); err != nil {
		return err
	}
	created, err := s.follows.Add(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if !created {
		return ErrAlreadyFollowed
	}
	logSideEffectErr("relation_add", s.relations.AddFollow(ctx, followerID, followeeID),
		"follower_id", followerID, "followee_id", followeeID)
	logSideEffectErr("inbox_backfill", s.backfillInboxAfterFollow(ctx, followerID, followeeID),
		"follower_id", followerID, "followee_id", followeeID)
	if s.notifications != nil {
		logSideEffectErr("notification_follow", s.notifications.Create(ctx, CreateNotificationInput{
			UserID:  followeeID,
			ActorID: followerID,
			Type:    NotificationFollow,
			Content: "有人关注了你",
		}), "follower_id", followerID, "followee_id", followeeID)
	}
	return nil
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uint64) error {
	if followerID == 0 || followeeID == 0 {
		return ErrInvalidInput
	}
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	deleted, err := s.follows.Remove(ctx, followerID, followeeID)
	if err != nil || !deleted {
		return err
	}
	logSideEffectErr("relation_remove", s.relations.RemoveFollow(ctx, followerID, followeeID),
		"follower_id", followerID, "followee_id", followeeID)
	logSideEffectErr("inbox_delete", s.inboxes.Delete(ctx, followerID),
		"follower_id", followerID)
	return nil
}

func (s *FollowService) IsFollowing(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	if followerID == 0 || followeeID == 0 {
		return false, ErrInvalidInput
	}
	if followerID == followeeID {
		return false, nil
	}
	if _, err := s.users.FindByID(ctx, followeeID); err != nil {
		return false, err
	}
	return s.follows.Exists(ctx, followerID, followeeID)
}

func (s *FollowService) ListFollowingUsers(ctx context.Context, userID uint64, input UserListInput) (*UserListResult, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	limit, offset := normalizeLimitOffset(input.Limit, input.Offset)
	users, err := s.follows.ListFollowingUsers(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &UserListResult{Items: users}, nil
}

func (s *FollowService) ListFollowerUsers(ctx context.Context, userID uint64, input UserListInput) (*UserListResult, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	limit, offset := normalizeLimitOffset(input.Limit, input.Offset)
	users, err := s.follows.ListFollowerUsers(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &UserListResult{Items: users}, nil
}

func (s *FollowService) ListFollowingFeed(ctx context.Context, userID uint64, input ListInput) (*PostListResult, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return nil, err
	}
	limit := normalizeLimit(input.Limit)
	followingIDs, err := s.listFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(followingIDs) == 0 {
		posts, err := s.posts.ListLatest(ctx, input.Cursor, limit+1)
		if err != nil {
			return nil, err
		}
		return buildPostListResult(posts, limit), nil
	}

	rebuildInbox := false
	// Inbox 是用于首屏加速的有限快照，不是完整历史数据源。后续页面直接
	// 使用同一个 time+id 游标查询 MySQL，避免超过快照上限的旧动态不可达。
	if input.Cursor == nil {
		if postIDs, available, err := s.inboxes.PostIDs(ctx, userID, nil, followingFeedInboxRebuildLimit+1); err == nil {
			if available {
				posts, err := s.posts.ListByIDs(ctx, postIDs)
				if err != nil {
					return nil, err
				}
				missing := missingPostIDs(postIDs, posts)
				_ = s.inboxes.RemovePosts(ctx, userID, missing)
				ordered := orderPostsByIDs(posts, postIDs)
				// 缓存中即使有已删除的残留 ID，只要仍能组成完整首屏即可使用；
				// 否则回源，避免错误地返回不足一页或 has_more=false。
				if len(ordered) > limit || len(missing) == 0 {
					return buildPostListResult(ordered, limit), nil
				}
			}
			if !available {
				rebuildInbox = true
			}
		}
	}

	queryLimit := limit + 1
	if rebuildInbox {
		queryLimit = followingFeedInboxRebuildLimit + 1
	}
	posts, err := s.posts.ListByAuthorIDs(ctx, followingIDs, input.Cursor, queryLimit)
	if err != nil {
		return nil, err
	}
	if rebuildInbox {
		_ = s.inboxes.Rebuild(ctx, userID, feedInboxItems(posts))
	}
	return buildPostListResult(posts, limit), nil
}

func missingPostIDs(requested []uint64, posts []model.Post) []uint64 {
	existing := make(map[uint64]struct{}, len(posts))
	for _, post := range posts {
		existing[post.ID] = struct{}{}
	}
	missing := make([]uint64, 0)
	for _, postID := range requested {
		if _, ok := existing[postID]; !ok {
			missing = append(missing, postID)
		}
	}
	return missing
}

func (s *FollowService) backfillInboxAfterFollow(ctx context.Context, followerID, followeeID uint64) error {
	exists, err := s.inboxes.Exists(ctx, followerID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	posts, err := s.posts.ListByAuthor(ctx, followeeID, nil, followingFeedInboxRebuildLimit)
	if err != nil {
		return err
	}
	return s.inboxes.AddPosts(ctx, followerID, feedInboxItems(posts))
}

func feedInboxItems(posts []model.Post) []cache.FeedInboxItem {
	items := make([]cache.FeedInboxItem, 0, len(posts))
	for _, post := range posts {
		items = append(items, cache.FeedInboxItem{
			PostID:    post.ID,
			CreatedAt: post.CreatedAt,
		})
	}
	return items
}

func (s *FollowService) listFollowingIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	if ids, hit, err := s.relations.FollowingIDs(ctx, userID); err == nil && hit {
		return ids, nil
	}

	ids, err := s.follows.ListFollowingIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = s.relations.SetFollowingIDs(ctx, userID, ids)
	return ids, nil
}

func normalizeLimitOffset(limit, offset int) (int, int) {
	limit = normalizeLimit(limit)
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
