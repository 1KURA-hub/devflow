package model

import "time"

type Favorite struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_favorites_user_post,priority:1;index:idx_favorites_user" json:"user_id"`
	PostID    uint64    `gorm:"not null;uniqueIndex:uk_favorites_user_post,priority:2;index:idx_favorites_post" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}
