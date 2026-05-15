package model

import "time"

type Follow struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FollowerID uint64    `gorm:"not null;uniqueIndex:uk_follows_pair,priority:1;index:idx_follows_follower" json:"follower_id"`
	FolloweeID uint64    `gorm:"not null;uniqueIndex:uk_follows_pair,priority:2;index:idx_follows_followee" json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Follow) TableName() string {
	return "follows"
}
