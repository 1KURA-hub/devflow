package repository

import (
	"context"
	"errors"
	"time"

	"devflow/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID uint64, cursor *time.Time, limit int) ([]model.Notification, error) {
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if cursor != nil {
		query = query.Where("created_at < ?", *cursor)
	}

	var notifications []model.Notification
	if err := query.Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

func (r *NotificationRepository) FindByID(ctx context.Context, userID, notificationID uint64) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		First(&notification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &notification, nil
}
