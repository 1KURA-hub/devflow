package service

import (
	"context"
	"time"

	"devflow/internal/cache"
	"devflow/internal/model"
	"devflow/internal/mq"
	"devflow/internal/repository"
)

const (
	NotificationFollow   = "follow"
	NotificationLike     = "like"
	NotificationFavorite = "favorite"
	NotificationComment  = "comment"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
	counter       *cache.NotificationCounter
	publisher     *mq.Publisher
}

type CreateNotificationInput struct {
	UserID    uint64
	ActorID   uint64
	Type      string
	PostID    *uint64
	CommentID *uint64
	Content   string
}

type NotificationListResult struct {
	Items      []model.Notification `json:"items"`
	NextCursor string               `json:"next_cursor,omitempty"`
	HasMore    bool                 `json:"has_more"`
}

func NewNotificationService(notifications *repository.NotificationRepository, counter *cache.NotificationCounter, publisher *mq.Publisher) *NotificationService {
	return &NotificationService{
		notifications: notifications,
		counter:       counter,
		publisher:     publisher,
	}
}

func (s *NotificationService) Create(ctx context.Context, input CreateNotificationInput) error {
	if input.UserID == 0 || input.ActorID == 0 || input.UserID == input.ActorID || input.Type == "" {
		return nil
	}
	content := input.Content
	if content == "" {
		content = defaultNotificationContent(input.Type)
	}
	if s.publisher != nil {
		return s.publisher.PublishNotification(ctx, mq.NotificationEvent{
			EventID:   mq.NewEventID(),
			UserID:    input.UserID,
			ActorID:   input.ActorID,
			Type:      input.Type,
			PostID:    input.PostID,
			CommentID: input.CommentID,
			Content:   content,
		})
	}
	return s.CreateNow(ctx, CreateNotificationInput{
		UserID:    input.UserID,
		ActorID:   input.ActorID,
		Type:      input.Type,
		PostID:    input.PostID,
		CommentID: input.CommentID,
		Content:   content,
	})
}

func (s *NotificationService) CreateNow(ctx context.Context, input CreateNotificationInput) error {
	if input.UserID == 0 || input.ActorID == 0 || input.UserID == input.ActorID || input.Type == "" {
		return nil
	}
	content := input.Content
	if content == "" {
		content = defaultNotificationContent(input.Type)
	}
	if err := s.notifications.Create(ctx, &model.Notification{
		UserID:    input.UserID,
		ActorID:   input.ActorID,
		Type:      input.Type,
		PostID:    input.PostID,
		CommentID: input.CommentID,
		Content:   content,
		IsRead:    false,
	}); err != nil {
		return err
	}
	_ = s.counter.IncrementIfExists(ctx, input.UserID)
	return nil
}

func (s *NotificationService) List(ctx context.Context, userID uint64, input ListInput) (*NotificationListResult, error) {
	if userID == 0 {
		return nil, ErrInvalidInput
	}
	limit := normalizeLimit(input.Limit)
	notifications, err := s.notifications.ListByUser(ctx, userID, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildNotificationListResult(notifications, limit), nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uint64) (int64, error) {
	if userID == 0 {
		return 0, ErrInvalidInput
	}
	if count, hit, err := s.counter.Get(ctx, userID); err == nil && hit {
		return count, nil
	}

	count, err := s.notifications.CountUnread(ctx, userID)
	if err != nil {
		return 0, err
	}
	_ = s.counter.Set(ctx, userID, count)
	return count, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID uint64) error {
	if userID == 0 || notificationID == 0 {
		return ErrInvalidInput
	}
	if err := s.notifications.MarkRead(ctx, userID, notificationID); err != nil {
		return err
	}
	_ = s.counter.Delete(ctx, userID)
	return nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return ErrInvalidInput
	}
	if err := s.notifications.MarkAllRead(ctx, userID); err != nil {
		return err
	}
	_ = s.counter.Delete(ctx, userID)
	return nil
}

func buildNotificationListResult(notifications []model.Notification, limit int) *NotificationListResult {
	hasMore := len(notifications) > limit
	if hasMore {
		notifications = notifications[:limit]
	}

	result := &NotificationListResult{
		Items:   notifications,
		HasMore: hasMore,
	}
	if hasMore && len(notifications) > 0 {
		result.NextCursor = notifications[len(notifications)-1].CreatedAt.Format(time.RFC3339Nano)
	}
	return result
}

func defaultNotificationContent(notificationType string) string {
	switch notificationType {
	case NotificationFollow:
		return "有人关注了你"
	case NotificationLike:
		return "有人点赞了你的动态"
	case NotificationFavorite:
		return "有人收藏了你的动态"
	case NotificationComment:
		return "有人评论了你的动态"
	default:
		return "你有一条新通知"
	}
}
