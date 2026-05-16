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
