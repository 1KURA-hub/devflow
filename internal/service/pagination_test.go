package service

import (
	"testing"
	"time"

	"devflow/internal/model"
	"devflow/internal/pagination"
)

func TestBuildPostListResultEncodesIDTieBreaker(t *testing.T) {
	createdAt := time.Date(2026, time.August, 4, 15, 0, 0, 123456000, time.Local)
	result := buildPostListResult([]model.Post{
		{ID: 3, CreatedAt: createdAt},
		{ID: 2, CreatedAt: createdAt},
	}, 1)

	if !result.HasMore || result.NextCursor == "" {
		t.Fatalf("expected next cursor, got %+v", result)
	}
	cursor, err := pagination.Decode(result.NextCursor, pagination.KindChronological)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if cursor.ID != 3 || !cursor.CreatedAt().Equal(createdAt) {
		t.Fatalf("cursor lost tie-breaker: id=%d time=%s", cursor.ID, cursor.CreatedAt())
	}
}

func TestBuildHotPostListResultEncodesScoreAndID(t *testing.T) {
	posts := []model.Post{
		{ID: 9, LikeCount: 1},
		{ID: 8, LikeCount: 1},
	}
	result := buildHotPostListResult(posts, 1, map[uint64]int64{9: 7, 8: 7})

	cursor, err := pagination.Decode(result.NextCursor, pagination.KindHot)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if cursor.Score != 7 || cursor.ID != 9 {
		t.Fatalf("cursor mismatch: score=%d id=%d", cursor.Score, cursor.ID)
	}
}

func TestMissingPostIDs(t *testing.T) {
	missing := missingPostIDs(
		[]uint64{5, 4, 3},
		[]model.Post{{ID: 5}, {ID: 3}},
	)
	if len(missing) != 1 || missing[0] != 4 {
		t.Fatalf("unexpected missing IDs: %v", missing)
	}
}
