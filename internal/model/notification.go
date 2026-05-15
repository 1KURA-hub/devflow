package model

import "time"

type Notification struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index:idx_notifications_user_read_created,priority:1;index:idx_notifications_user_created,priority:1" json:"user_id"`
	ActorID   uint64    `gorm:"not null" json:"actor_id"`
	Type      string    `gorm:"type:varchar(32);not null" json:"type"`
	PostID    *uint64   `gorm:"index" json:"post_id,omitempty"`
	CommentID *uint64   `gorm:"index" json:"comment_id,omitempty"`
	Content   string    `gorm:"type:varchar(255);not null" json:"content"`
	IsRead    bool      `gorm:"not null;default:false;index:idx_notifications_user_read_created,priority:2" json:"is_read"`
	CreatedAt time.Time `gorm:"index:idx_notifications_user_read_created,priority:3;index:idx_notifications_user_created,priority:2" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
