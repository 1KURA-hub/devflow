package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    uint64         `gorm:"not null;index:idx_comments_post_created,priority:1" json:"post_id"`
	UserID    uint64         `gorm:"not null;index:idx_comments_user_created,priority:1" json:"user_id"`
	Content   string         `gorm:"type:varchar(1000);not null" json:"content"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time      `gorm:"index:idx_comments_post_created,priority:2;index:idx_comments_user_created,priority:2" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Comment) TableName() string {
	return "comments"
}
