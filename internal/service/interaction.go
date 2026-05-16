package service

import (
	"context"

	"devflow/internal/repository"
)

type InteractionService struct {
	interactions *repository.InteractionRepository
	posts        *repository.PostRepository
	users        *repository.UserRepository
}

func NewInteractionService(interactions *repository.InteractionRepository, posts *repository.PostRepository, users *repository.UserRepository) *InteractionService {
	return &InteractionService{
		interactions: interactions,
		posts:        posts,
		users:        users,
	}
}

func (s *InteractionService) Like(ctx context.Context, userID, postID uint64) (bool, error) {
	if userID == 0 || postID == 0 {
		return false, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return false, err
	}
	if _, err := s.posts.FindByID(ctx, postID); err != nil {
		return false, err
	}
	return s.interactions.AddLike(ctx, userID, postID)
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
	if _, err := s.posts.FindByID(ctx, postID); err != nil {
		return false, err
	}
	return s.interactions.AddFavorite(ctx, userID, postID)
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
