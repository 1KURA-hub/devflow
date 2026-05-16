package service

import (
	"context"

	"devflow/internal/repository"
)

type InteractionService struct {
	interactions  *repository.InteractionRepository
	posts         *repository.PostRepository
	users         *repository.UserRepository
	notifications *NotificationService
}

func NewInteractionService(interactions *repository.InteractionRepository, posts *repository.PostRepository, users *repository.UserRepository, notifications *NotificationService) *InteractionService {
	return &InteractionService{
		interactions:  interactions,
		posts:         posts,
		users:         users,
		notifications: notifications,
	}
}

func (s *InteractionService) Like(ctx context.Context, userID, postID uint64) (bool, error) {
	if userID == 0 || postID == 0 {
		return false, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return false, err
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return false, err
	}
	changed, err := s.interactions.AddLike(ctx, userID, postID)
	if err != nil || !changed || s.notifications == nil {
		return changed, err
	}
	return changed, s.notifications.Create(ctx, CreateNotificationInput{
		UserID:  post.AuthorID,
		ActorID: userID,
		Type:    NotificationLike,
		PostID:  &postID,
		Content: "有人点赞了你的动态",
	})
}

func (s *InteractionService) Unlike(ctx context.Context, userID, postID uint64) (bool, error) {
	if userID == 0 || postID == 0 {
		return false, ErrInvalidInput
	}
	if _, err := s.posts.FindByID(ctx, postID); err != nil {
		return false, err
	}
	return s.interactions.RemoveLike(ctx, userID, postID)
}

func (s *InteractionService) Favorite(ctx context.Context, userID, postID uint64) (bool, error) {
	if userID == 0 || postID == 0 {
		return false, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return false, err
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return false, err
	}
	changed, err := s.interactions.AddFavorite(ctx, userID, postID)
	if err != nil || !changed || s.notifications == nil {
		return changed, err
	}
	return changed, s.notifications.Create(ctx, CreateNotificationInput{
		UserID:  post.AuthorID,
		ActorID: userID,
		Type:    NotificationFavorite,
		PostID:  &postID,
		Content: "有人收藏了你的动态",
	})
}

func (s *InteractionService) Unfavorite(ctx context.Context, userID, postID uint64) (bool, error) {
	if userID == 0 || postID == 0 {
		return false, ErrInvalidInput
	}
	if _, err := s.posts.FindByID(ctx, postID); err != nil {
		return false, err
	}
	return s.interactions.RemoveFavorite(ctx, userID, postID)
}

func (s *InteractionService) ListMyFavorites(ctx context.Context, userID uint64, input ListInput) (*PostListResult, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	limit := normalizeLimit(input.Limit)
	posts, err := s.posts.ListFavoritedByUser(ctx, userID, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildPostListResult(posts, limit), nil
}
