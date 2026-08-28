package repository

import (
	"context"
	"errors"

	"devflow/internal/model"
	"devflow/internal/pagination"

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

func (r *PostRepository) DeleteByID(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("id = ? AND status = ?", id, 1).
		Update("status", 0)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id uint64) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).
		Preload("Author").
		Where("status = ?", 1).
		First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) CountByAuthor(ctx context.Context, authorID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("author_id = ? AND status = ?", authorID, 1).
		Count(&count).Error
	return count, err
}

func (r *PostRepository) ListLatest(ctx context.Context, cursor *pagination.Cursor, limit int) ([]model.Post, error) {
	query := r.db.WithContext(ctx).
		Preload("Author").
		Where("status = ?", 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	query = applyChronologicalCursor(query, cursor, "created_at", "id")

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListHot(ctx context.Context, cursor *pagination.Cursor, limit int) ([]model.Post, error) {
	const scoreExpression = "(like_count * 3 + favorite_count * 5 + comment_count * 4)"
	query := r.db.WithContext(ctx).
		Preload("Author").
		Where("status = ? AND "+scoreExpression+" > 0", 1)
	if cursor != nil {
		query = query.Where(
			"("+scoreExpression+" < ?) OR ("+scoreExpression+" = ? AND id < ?)",
			cursor.Score,
			cursor.Score,
			cursor.ID,
		)
	}
	var posts []model.Post
	err := query.
		Order(scoreExpression + " DESC").
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
		Preload("Author").
		Where("id IN ? AND status = ?", ids, 1).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) ListFavoritedByUser(ctx context.Context, userID uint64, cursor *pagination.Cursor, limit int) ([]model.Post, error) {
	query := r.db.WithContext(ctx).
		Preload("Author").
		Model(&model.Post{}).
		Joins("JOIN favorites ON favorites.post_id = posts.id").
		Where("favorites.user_id = ? AND posts.status = ?", userID, 1).
		Order("posts.created_at DESC").
		Order("posts.id DESC").
		Limit(limit)
	query = applyChronologicalCursor(query, cursor, "posts.created_at", "posts.id")

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListByAuthor(ctx context.Context, authorID uint64, cursor *pagination.Cursor, limit int) ([]model.Post, error) {
	query := r.db.WithContext(ctx).
		Preload("Author").
		Where("author_id = ? AND status = ?", authorID, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	query = applyChronologicalCursor(query, cursor, "created_at", "id")

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListByAuthorIDs(ctx context.Context, authorIDs []uint64, cursor *pagination.Cursor, limit int) ([]model.Post, error) {
	if len(authorIDs) == 0 {
		return nil, nil
	}

	query := r.db.WithContext(ctx).
		Preload("Author").
		Where("author_id IN ? AND status = ?", authorIDs, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	query = applyChronologicalCursor(query, cursor, "created_at", "id")

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostRepository) ListFollowing(ctx context.Context, userID uint64, cursor *pagination.Cursor, limit int) ([]model.Post, error) {
	following := r.db.Model(&model.Follow{}).
		Select("followee_id").
		Where("follower_id = ?", userID)

	query := r.db.WithContext(ctx).
		Preload("Author").
		Where("author_id IN (?) AND status = ?", following, 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	query = applyChronologicalCursor(query, cursor, "created_at", "id")

	var posts []model.Post
	if err := query.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func applyChronologicalCursor(query *gorm.DB, cursor *pagination.Cursor, createdAtColumn, idColumn string) *gorm.DB {
	if cursor == nil {
		return query
	}
	createdAt := cursor.CreatedAt()
	if cursor.ID == 0 {
		return query.Where(createdAtColumn+" < ?", createdAt)
	}
	return query.Where(
		"("+createdAtColumn+" < ?) OR ("+createdAtColumn+" = ? AND "+idColumn+" < ?)",
		createdAt,
		createdAt,
		cursor.ID,
	)
}
