package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthorID      uint64         `gorm:"not null;index:idx_posts_author_created,priority:1" json:"author_id"`
	Author        *User          `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Title         string         `gorm:"type:varchar(120);not null" json:"title"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	Tags          string         `gorm:"type:varchar(255);not null;default:''" json:"tags"`
	LikeCount     int64          `gorm:"not null;default:0" json:"like_count"`
	CommentCount  int64          `gorm:"not null;default:0" json:"comment_count"`
	FavoriteCount int64          `gorm:"not null;default:0" json:"favorite_count"`
	Status        int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt     time.Time      `gorm:"index:idx_posts_author_created,priority:2;index:idx_posts_created_at" json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Post) TableName() string {
	return "posts"
}
