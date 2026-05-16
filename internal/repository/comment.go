package repository

import (
	"context"
	"time"

	"devflow/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", comment.PostID, 1).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	})
}

func (r *CommentRepository) ListByPost(ctx context.Context, postID uint64, cursor *time.Time, limit int) ([]model.Comment, error) {
	query := r.db.WithContext(ctx).
		Where("post_id = ? AND status = ?", postID, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	var comments []model.Comment
	if err := query.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}
