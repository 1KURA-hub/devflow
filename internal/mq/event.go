package mq

import (
	"crypto/rand"
	"encoding/hex"
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

func NewEventID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(raw[:])
}
