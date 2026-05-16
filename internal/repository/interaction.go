package repository

import (
	"context"
	"errors"

	"devflow/internal/model"

	"gorm.io/gorm"
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
		var like model.Like
		err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(&model.Like{UserID: userID, PostID: postID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", postID, 1).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			return err
		}
		created = true
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
		var favorite model.Favorite
		err := tx.Where("user_id = ? AND post_id = ?", userID, postID).First(&favorite).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(&model.Favorite{UserID: userID, PostID: postID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Post{}).
			Where("id = ? AND status = ?", postID, 1).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
			return err
		}
		created = true
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
