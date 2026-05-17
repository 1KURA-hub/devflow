package repository

import (
	"context"
	"errors"
	"time"

	"devflow/internal/model"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *PostRepository) FindByID(ctx context.Context, id uint64) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) ListLatest(ctx context.Context, cursor *time.Time, limit int) ([]model.Post, error) {
	query := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListHot(ctx context.Context, limit int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).
		Where("status = ? AND (like_count * 3 + favorite_count * 5 + comment_count * 4) > 0", 1).
		Order("(like_count * 3 + favorite_count * 5 + comment_count * 4) DESC").
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) ListByIDs(ctx context.Context, ids []uint64) ([]model.Post, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var posts []model.Post
	err := r.db.WithContext(ctx).
		Where("id IN ? AND status = ?", ids, 1).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) ListFavoritedByUser(ctx context.Context, userID uint64, cursor *time.Time, limit int) ([]model.Post, error) {
	query := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Joins("JOIN favorites ON favorites.post_id = posts.id").
		Where("favorites.user_id = ? AND posts.status = ?", userID, 1).
		Order("posts.created_at DESC").
		Order("posts.id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("posts.created_at < ?", *cursor)
	}

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListByAuthor(ctx context.Context, authorID uint64, cursor *time.Time, limit int) ([]model.Post, error) {
	query := r.db.WithContext(ctx).
		Where("author_id = ? AND status = ?", authorID, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListByAuthorIDs(ctx context.Context, authorIDs []uint64, cursor *time.Time, limit int) ([]model.Post, error) {
	if len(authorIDs) == 0 {
		return nil, nil
	}

	query := r.db.WithContext(ctx).
		Where("author_id IN ? AND status = ?", authorIDs, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListFollowing(ctx context.Context, userID uint64, cursor *time.Time, limit int) ([]model.Post, error) {
	following := r.db.Model(&model.Follow{}).
		Select("followee_id").
		Where("follower_id = ?", userID)

	query := r.db.WithContext(ctx).
		Where("author_id IN (?) AND status = ?", following, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}
