package mq

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

const (
	RoutingKeyPostPublished        = "post.published"
	RoutingKeyInteractionLiked     = "interaction.liked"
	RoutingKeyInteractionFavorited = "interaction.favorited"
	RoutingKeyInteractionCommented = "interaction.commented"
	RoutingKeyUserFollowed         = "user.followed"
)

type NotificationEvent struct {
	EventID   string  `json:"event_id"`
	UserID    uint64  `json:"user_id"`
	ActorID   uint64  `json:"actor_id"`
	Type      string  `json:"type"`
	PostID    *uint64 `json:"post_id,omitempty"`
	CommentID *uint64 `json:"comment_id,omitempty"`
	Content   string  `json:"content"`
}

type PostPublishedEvent struct {
	EventID   string    `json:"event_id"`
	PostID    uint64    `json:"post_id"`
	AuthorID  uint64    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewEventID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(raw[:])
}
