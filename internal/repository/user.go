package repository

import (
	"context"
	"errors"

	"devflow/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if isDuplicateEntry(err) {
		return ErrDuplicate
	}
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListRecommended(ctx context.Context, viewerID uint64, limit int) ([]model.User, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}
	query := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if viewerID != 0 {
		query = query.Where("id <> ?", viewerID)
	}

	var users []model.User
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uint64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(updates).Error
}
