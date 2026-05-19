package service

import (
	"context"

	"devflow/internal/cache"
	"devflow/internal/repository"
)

type InteractionService struct {
	interactions  *repository.InteractionRepository
	posts         *repository.PostRepository
	users         *repository.UserRepository
	notifications *NotificationService
	hotPosts      *cache.HotPostStore
}

func NewInteractionService(interactions *repository.InteractionRepository, posts *repository.PostRepository, users *repository.UserRepository, notifications *NotificationService, hotPosts *cache.HotPostStore) *InteractionService {
	return &InteractionService{
		interactions:  interactions,
		posts:         posts,
		users:         users,
		notifications: notifications,
		hotPosts:      hotPosts,
	}
}

func (s *InteractionService) Like(ctx context.Context, userID, postID uint64) error {
	if userID == 0 || postID == 0 {
		return ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	created, err := s.interactions.AddLike(ctx, userID, postID)
	if err != nil || !created {
		return err
	}
	_ = s.hotPosts.SetScore(ctx, postID, hotScore(post.LikeCount+1, post.FavoriteCount, post.CommentCount))
	if s.notifications == nil {
		return nil
	}
	return s.notifications.Create(ctx, CreateNotificationInput{
		UserID:  post.AuthorID,
		ActorID: userID,
		Type:    NotificationLike,
		PostID:  &postID,
		Content: "有人点赞了你的动态",
	})
}

func (s *InteractionService) Unlike(ctx context.Context, userID, postID uint64) error {
	if userID == 0 || postID == 0 {
		return ErrInvalidInput
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	deleted, err := s.interactions.RemoveLike(ctx, userID, postID)
	if err != nil || !deleted {
		return err
	}
	_ = s.hotPosts.SetScore(ctx, postID, hotScore(post.LikeCount-1, post.FavoriteCount, post.CommentCount))
	return nil
}

func (s *InteractionService) Favorite(ctx context.Context, userID, postID uint64) error {
	if userID == 0 || postID == 0 {
		return ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, userID); err != nil {
		return err
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	created, err := s.interactions.AddFavorite(ctx, userID, postID)
	if err != nil || !created {
		return err
	}
	_ = s.hotPosts.SetScore(ctx, postID, hotScore(post.LikeCount, post.FavoriteCount+1, post.CommentCount))
	if s.notifications == nil {
		return nil
	}
	return s.notifications.Create(ctx, CreateNotificationInput{
		UserID:  post.AuthorID,
		ActorID: userID,
		Type:    NotificationFavorite,
		PostID:  &postID,
		Content: "有人收藏了你的动态",
	})
}

func (s *InteractionService) Unfavorite(ctx context.Context, userID, postID uint64) error {
	if userID == 0 || postID == 0 {
		return ErrInvalidInput
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	deleted, err := s.interactions.RemoveFavorite(ctx, userID, postID)
	if err != nil || !deleted {
		return err
	}
	_ = s.hotPosts.SetScore(ctx, postID, hotScore(post.LikeCount, post.FavoriteCount-1, post.CommentCount))
	return nil
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

func (s *InteractionService) PostStates(ctx context.Context, userID uint64, postIDs []uint64) (map[uint64]bool, map[uint64]bool, error) {
	liked := make(map[uint64]bool, len(postIDs))
	favorited := make(map[uint64]bool, len(postIDs))
	if userID == 0 || len(postIDs) == 0 {
		return liked, favorited, nil
	}

	liked, err := s.interactions.LikedPostIDs(ctx, userID, postIDs)
	if err != nil {
		return nil, nil, err
	}
	favorited, err = s.interactions.FavoritedPostIDs(ctx, userID, postIDs)
	if err != nil {
		return nil, nil, err
	}
	return liked, favorited, nil
}
