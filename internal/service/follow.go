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
	_ = s.relations.AddFollow(ctx, followerID, followeeID)
	_ = s.backfillInboxAfterFollow(ctx, followerID, followeeID)
	if s.notifications != nil {
		return s.notifications.Create(ctx, CreateNotificationInput{
			UserID:  followeeID,
			ActorID: followerID,
			Type:    NotificationFollow,
			Content: "有人关注了你",
		})
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
	_ = s.relations.RemoveFollow(ctx, followerID, followeeID)
	_ = s.inboxes.Delete(ctx, followerID)
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
	if postIDs, available, err := s.inboxes.PostIDs(ctx, userID, input.Cursor, int64(limit+1)); err == nil {
		if available {
			posts, err := s.posts.ListByIDs(ctx, postIDs)
			if err != nil {
				return nil, err
			}
			return buildPostListResult(orderPostsByIDs(posts, postIDs), limit), nil
		}
		if !available && input.Cursor == nil {
			rebuildInbox = true
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
