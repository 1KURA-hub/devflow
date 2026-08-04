package worker

import (
	"errors"
	"fmt"

	"devflow/internal/mq"
)

func validatePostPublishedEvent(event mq.PostPublishedEvent) error {
	switch {
	case event.EventID == "":
		return errors.New("event_id is required")
	case event.AuthorID == 0:
		return errors.New("author_id is required")
	case event.PostID == 0:
		return errors.New("post_id is required")
	case event.CreatedAt.IsZero():
		return errors.New("created_at is required")
	default:
		return nil
	}
}

func validateNotificationEvent(event mq.NotificationEvent) error {
	if event.EventID == "" {
		return errors.New("event_id is required")
	}
	if event.UserID == 0 || event.ActorID == 0 {
		return errors.New("user_id and actor_id are required")
	}
	if event.UserID == event.ActorID {
		return errors.New("notification recipient and actor must differ")
	}

	switch event.Type {
	case "follow":
		return nil
	case "like", "favorite":
		if !validOptionalID(event.PostID) {
			return fmt.Errorf("post_id is required for %s notification", event.Type)
		}
		return nil
	case "comment":
		if !validOptionalID(event.PostID) || !validOptionalID(event.CommentID) {
			return errors.New("post_id and comment_id are required for comment notification")
		}
		return nil
	default:
		return fmt.Errorf("unsupported notification type %q", event.Type)
	}
}

func validOptionalID(id *uint64) bool {
	return id != nil && *id > 0
}
