package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"devflow/internal/cache"
	"devflow/internal/model"
	"devflow/internal/repository"
)

const (
	defaultListLimit = 20
	maxListLimit     = 50
)

type PostService struct {
	posts    *repository.PostRepository
	users    *repository.UserRepository
	hotPosts *cache.HotPostStore
}

type CreatePostInput struct {
	AuthorID uint64
	Title    string
	Content  string
	Tags     string
}

type ListInput struct {
	Cursor *time.Time
	Limit  int
}

type PostListResult struct {
	Items      []model.Post `json:"items"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
}

func NewPostService(posts *repository.PostRepository, users *repository.UserRepository, hotPosts *cache.HotPostStore) *PostService {
	return &PostService{
		posts:    posts,
		users:    users,
		hotPosts: hotPosts,
	}
}

func (s *PostService) Create(ctx context.Context, input CreatePostInput) (*model.Post, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	tags := normalizeTags(input.Tags)
	if input.AuthorID == 0 || title == "" || content == "" {
		return nil, ErrInvalidInput
	}
	if utf8.RuneCountInString(title) > 120 || utf8.RuneCountInString(content) > 5000 || utf8.RuneCountInString(tags) > 255 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, input.AuthorID); err != nil {
		return nil, err
	}

	post := &model.Post{
		AuthorID: input.AuthorID,
		Title:    title,
		Content:  content,
		Tags:     tags,
		Status:   1,
	}
	if err := s.posts.Create(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *PostService) Get(ctx context.Context, id uint64) (*model.Post, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	return s.posts.FindByID(ctx, id)
}

func (s *PostService) ListLatest(ctx context.Context, input ListInput) (*PostListResult, error) {
	limit := normalizeLimit(input.Limit)
	posts, err := s.posts.ListLatest(ctx, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildPostListResult(posts, limit), nil
}

func (s *PostService) ListHot(ctx context.Context, input ListInput) (*PostListResult, error) {
	limit := normalizeLimit(input.Limit)
	if ids, available, err := s.hotPosts.TopPostIDs(ctx, int64(limit+1)); err == nil && available && len(ids) > 0 {
		posts, err := s.posts.ListByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		return buildPostListResult(orderPostsByIDs(posts, ids), limit), nil
	}

	posts, err := s.posts.ListHot(ctx, limit+1)
	if err != nil {
		return nil, err
	}
	return buildPostListResult(posts, limit), nil
}

func (s *PostService) ListByAuthor(ctx context.Context, authorID uint64, input ListInput) (*PostListResult, error) {
	if authorID == 0 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, authorID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	limit := normalizeLimit(input.Limit)
	posts, err := s.posts.ListByAuthor(ctx, authorID, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildPostListResult(posts, limit), nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func buildPostListResult(posts []model.Post, limit int) *PostListResult {
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	result := &PostListResult{
		Items:   posts,
		HasMore: hasMore,
	}
	if hasMore && len(posts) > 0 {
		result.NextCursor = posts[len(posts)-1].CreatedAt.Format(time.RFC3339Nano)
	}
	return result
}

func orderPostsByIDs(posts []model.Post, ids []uint64) []model.Post {
	postsByID := make(map[uint64]model.Post, len(posts))
	for _, post := range posts {
		postsByID[post.ID] = post
	}

	ordered := make([]model.Post, 0, len(posts))
	for _, id := range ids {
		if post, ok := postsByID[id]; ok {
			ordered = append(ordered, post)
		}
	}
	return ordered
}

func hotScore(likeCount, favoriteCount, commentCount int64) int64 {
	return likeCount*3 + favoriteCount*5 + commentCount*4
}

func normalizeTags(tags string) string {
	parts := strings.Split(tags, ",")
	normalized := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	return strings.Join(normalized, ",")
}
