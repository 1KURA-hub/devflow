package worker

import (
	"testing"
	"time"

	"devflow/internal/mq"
)

func TestValidatePostPublishedEvent(t *testing.T) {
	valid := mq.PostPublishedEvent{
		EventID:   "event-1",
		PostID:    10,
		AuthorID:  20,
		CreatedAt: time.Now(),
	}
	if err := validatePostPublishedEvent(valid); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}

	invalid := valid
	invalid.PostID = 0
	if err := validatePostPublishedEvent(invalid); err == nil {
		t.Fatal("event without post_id should be rejected")
	}
}

func TestValidateNotificationEvent(t *testing.T) {
	postID := uint64(10)
	commentID := uint64(30)
	valid := []mq.NotificationEvent{
		{EventID: "follow-1", UserID: 1, ActorID: 2, Type: "follow"},
		{EventID: "like-1", UserID: 1, ActorID: 2, Type: "like", PostID: &postID},
		{EventID: "favorite-1", UserID: 1, ActorID: 2, Type: "favorite", PostID: &postID},
		{EventID: "comment-1", UserID: 1, ActorID: 2, Type: "comment", PostID: &postID, CommentID: &commentID},
	}
	for _, event := range valid {
		if err := validateNotificationEvent(event); err != nil {
			t.Fatalf("valid %s event rejected: %v", event.Type, err)
		}
	}

	invalid := []mq.NotificationEvent{
		{UserID: 1, ActorID: 2, Type: "follow"},
		{EventID: "same-user", UserID: 1, ActorID: 1, Type: "follow"},
		{EventID: "missing-post", UserID: 1, ActorID: 2, Type: "like"},
		{EventID: "missing-comment", UserID: 1, ActorID: 2, Type: "comment", PostID: &postID},
		{EventID: "unknown", UserID: 1, ActorID: 2, Type: "unknown"},
	}
	for _, event := range invalid {
		if err := validateNotificationEvent(event); err == nil {
			t.Fatalf("invalid event should be rejected: %+v", event)
		}
	}
}
