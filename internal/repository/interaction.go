package repository

import (
	"context"

	"devflow/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InteractionRepository struct {
	db *gorm.DB
}

func NewInteractionRepository(db *gorm.DB) *InteractionRepository {
	return &InteractionRepository{db: db}
}

func (r *InteractionRepository) AddLike(ctx context.Context, userID, postID uint64) (bool, error) {
	created := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&model.Like{UserID: userID, PostID: postID})
		if result.Error != nil {
			return result.Error
		}
		created = result.RowsAffected > 0
		if !created {
			return nil
		}
		update := tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", postID, 1).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
	return created, err
}

func (r *InteractionRepository) RemoveLike(ctx context.Context, userID, postID uint64) (bool, error) {
	deleted := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Like{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		if err := tx.Model(&model.Post{}).
			Where("id = ? AND status = ? AND like_count > 0", postID, 1).
			UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (r *InteractionRepository) AddFavorite(ctx context.Context, userID, postID uint64) (bool, error) {
	created := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&model.Favorite{UserID: userID, PostID: postID})
		if result.Error != nil {
			return result.Error
		}
		created = result.RowsAffected > 0
		if !created {
			return nil
		}
		update := tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", postID, 1).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1))
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
	return created, err
}

func (r *InteractionRepository) RemoveFavorite(ctx context.Context, userID, postID uint64) (bool, error) {
	deleted := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Favorite{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		if err := tx.Model(&model.Post{}).
			Where("id = ? AND status = ? AND favorite_count > 0", postID, 1).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
			return err
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (r *InteractionRepository) LikedPostIDs(ctx context.Context, userID uint64, postIDs []uint64) (map[uint64]bool, error) {
	result := make(map[uint64]bool, len(postIDs))
	if userID == 0 || len(postIDs) == 0 {
		return result, nil
	}

	var likes []model.Like
	if err := r.db.WithContext(ctx).
		Select("post_id").
		Where("user_id = ? AND post_id IN ?", userID, postIDs).
		Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, like := range likes {
		result[like.PostID] = true
	}
	return result, nil
}

func (r *InteractionRepository) FavoritedPostIDs(ctx context.Context, userID uint64, postIDs []uint64) (map[uint64]bool, error) {
	result := make(map[uint64]bool, len(postIDs))
	if userID == 0 || len(postIDs) == 0 {
		return result, nil
	}

	var favorites []model.Favorite
	if err := r.db.WithContext(ctx).
		Select("post_id").
		Where("user_id = ? AND post_id IN ?", userID, postIDs).
		Find(&favorites).Error; err != nil {
		return nil, err
	}
	for _, favorite := range favorites {
		result[favorite.PostID] = true
	}
	return result, nil
}
