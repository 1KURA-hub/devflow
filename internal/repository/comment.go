package repository

import (
	"context"

	"devflow/internal/model"
	"devflow/internal/pagination"

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
		update := tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", comment.PostID, 1).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1))
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *CommentRepository) ListByPost(ctx context.Context, postID uint64, cursor *pagination.Cursor, limit int) ([]model.Comment, error) {
	query := r.db.WithContext(ctx).
		Preload("User").
		Where("post_id = ? AND status = ?", postID, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		createdAt := cursor.CreatedAt()
		if cursor.ID == 0 {
			query = query.Where("created_at < ?", createdAt)
		} else {
			query = query.Where(
				"(created_at < ?) OR (created_at = ? AND id < ?)",
				createdAt,
				createdAt,
				cursor.ID,
			)
		}
	}

	var comments []model.Comment
	if err := query.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}
