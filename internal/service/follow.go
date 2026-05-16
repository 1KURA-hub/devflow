package service

import (
	"context"
	"errors"

	"devflow/internal/model"
	"devflow/internal/repository"
)

var ErrCannotFollowSelf = errors.New("cannot follow self")

type FollowService struct {
	follows       *repository.FollowRepository
	users         *repository.UserRepository
	posts         *repository.PostRepository
	notifications *NotificationService
}

type UserListInput struct {
	Limit  int
	Offset int
}

type UserListResult struct {
	Items []model.User `json:"items"`
}

func NewFollowService(follows *repository.FollowRepository, users *repository.UserRepository, posts *repository.PostRepository, notifications *NotificationService) *FollowService {
	return &FollowService{
		follows:       follows,
		users:         users,
		posts:         posts,
		notifications: notifications,
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
	exists, err := s.follows.Exists(ctx, followerID, followeeID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if err := s.follows.Create(ctx, &model.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}); err != nil {
		return err
	}
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
	if err := s.follows.Delete(ctx, followerID, followeeID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
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
	limit := normalizeLimit(input.Limit)
	posts, err := s.posts.ListFollowing(ctx, userID, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildPostListResult(posts, limit), nil
}

func normalizeLimitOffset(limit, offset int) (int, int) {
	limit = normalizeLimit(limit)
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
