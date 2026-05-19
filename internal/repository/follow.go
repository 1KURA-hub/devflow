package repository

import (
	"context"
	"errors"

	"devflow/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FollowRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Add(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.Follow{FollowerID: followerID, FolloweeID: followeeID})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *FollowRepository) Remove(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&model.Follow{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *FollowRepository) Exists(ctx context.Context, followerID, followeeID uint64) (bool, error) {
	var follow model.Follow
	err := r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		First(&follow).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *FollowRepository) CountFollowing(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *FollowRepository) ListFollowingIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Pluck("followee_id", &ids).Error
	return ids, err
}

func (r *FollowRepository) ListFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.WithContext(ctx).
		Model(&model.Follow{}).
		Where("followee_id = ?", userID).
		Pluck("follower_id", &ids).Error
	return ids, err
}

func (r *FollowRepository) ListFollowingUsers(ctx context.Context, userID uint64, limit, offset int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Joins("JOIN follows ON follows.followee_id = users.id").
		Where("follows.follower_id = ?", userID).
		Order("follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *FollowRepository) ListFollowerUsers(ctx context.Context, userID uint64, limit, offset int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.followee_id = ?", userID).
		Order("follows.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}
