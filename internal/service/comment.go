package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"devflow/internal/model"
	"devflow/internal/repository"
)

type CommentService struct {
	comments *repository.CommentRepository
	posts    *repository.PostRepository
	users    *repository.UserRepository
}

type CreateCommentInput struct {
	PostID  uint64
	UserID  uint64
	Content string
}

type CommentListResult struct {
	Items      []model.Comment `json:"items"`
	NextCursor string          `json:"next_cursor,omitempty"`
	HasMore    bool            `json:"has_more"`
}

func NewCommentService(comments *repository.CommentRepository, posts *repository.PostRepository, users *repository.UserRepository) *CommentService {
	return &CommentService{
		comments: comments,
		posts:    posts,
		users:    users,
	}
}

func (s *CommentService) Create(ctx context.Context, input CreateCommentInput) (*model.Comment, error) {
	content := strings.TrimSpace(input.Content)
	if input.PostID == 0 || input.UserID == 0 || content == "" {
		return nil, ErrInvalidInput
	}
	if utf8.RuneCountInString(content) > 1000 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, input.UserID); err != nil {
		return nil, err
	}
	if _, err := s.posts.FindByID(ctx, input.PostID); err != nil {
		return nil, err
	}

	comment := &model.Comment{
		PostID:  input.PostID,
		UserID:  input.UserID,
		Content: content,
		Status:  1,
	}
	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) ListByPost(ctx context.Context, postID uint64, input ListInput) (*CommentListResult, error) {
	if postID == 0 {
		return nil, ErrInvalidInput
	}
	if _, err := s.posts.FindByID(ctx, postID); err != nil {
		return nil, err
	}

	limit := normalizeLimit(input.Limit)
	comments, err := s.comments.ListByPost(ctx, postID, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildCommentListResult(comments, limit), nil
}

func buildCommentListResult(comments []model.Comment, limit int) *CommentListResult {
	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	result := &CommentListResult{
		Items:   comments,
		HasMore: hasMore,
	}
	if hasMore && len(comments) > 0 {
		result.NextCursor = comments[len(comments)-1].CreatedAt.Format(time.RFC3339Nano)
	}
	return result
}
