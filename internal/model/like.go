package model

import "time"

type Like struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_likes_user_post,priority:1;index:idx_likes_user" json:"user_id"`
	PostID    uint64    `gorm:"not null;uniqueIndex:uk_likes_user_post,priority:2;index:idx_likes_post" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Like) TableName() string {
	return "likes"
}
