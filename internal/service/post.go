package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"devflow/internal/cache"
	"devflow/internal/model"
	"devflow/internal/mq"
	"devflow/internal/pagination"
	"devflow/internal/repository"
)

const (
	defaultListLimit = 20
	maxListLimit     = 50
)

type PostService struct {
	posts     *repository.PostRepository
	follows   *repository.FollowRepository
	users     *repository.UserRepository
	hotPosts  *cache.HotPostStore
	relations *cache.FollowRelationStore
	inboxes   *cache.FeedInboxStore
	publisher *mq.Publisher
}

type CreatePostInput struct {
	AuthorID uint64
	Title    string
	Content  string
	CoverURL string
	Tags     string
}

type ListInput struct {
	Cursor *pagination.Cursor
	Limit  int
}

type PostListResult struct {
	Items      []model.Post `json:"items"`
	NextCursor string       `json:"next_cursor,omitempty"`
	HasMore    bool         `json:"has_more"`
}

func NewPostService(posts *repository.PostRepository, follows *repository.FollowRepository, users *repository.UserRepository, hotPosts *cache.HotPostStore, relations *cache.FollowRelationStore, inboxes *cache.FeedInboxStore, publisher *mq.Publisher) *PostService {
	return &PostService{
		posts:     posts,
		follows:   follows,
		users:     users,
		hotPosts:  hotPosts,
		relations: relations,
		inboxes:   inboxes,
		publisher: publisher,
	}
}

func (s *PostService) Create(ctx context.Context, input CreatePostInput) (*model.Post, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	coverURL := strings.TrimSpace(input.CoverURL)
	tags := normalizeTags(input.Tags)
	if input.AuthorID == 0 || title == "" || content == "" {
		return nil, ErrInvalidInput
	}
	if utf8.RuneCountInString(title) > 120 ||
		utf8.RuneCountInString(content) > 5000 ||
		utf8.RuneCountInString(coverURL) > 512 ||
		utf8.RuneCountInString(tags) > 255 {
		return nil, ErrInvalidInput
	}
	if _, err := s.users.FindByID(ctx, input.AuthorID); err != nil {
		return nil, err
	}

	post := &model.Post{
		AuthorID: input.AuthorID,
		Title:    title,
		Content:  content,
		CoverURL: coverURL,
		Tags:     tags,
		Status:   1,
	}
	if err := s.posts.Create(ctx, post); err != nil {
		return nil, err
	}
	if s.publisher != nil {
		if err := s.publisher.PublishPostPublished(ctx, mq.PostPublishedEvent{
			EventID:   mq.NewEventID(),
			PostID:    post.ID,
			AuthorID:  post.AuthorID,
			CreatedAt: post.CreatedAt,
		}); err == nil {
			return post, nil
		}
	}
	logSideEffectErr("feed_distribute_sync", s.DistributeFeedNow(ctx, post.AuthorID, post.ID, post.CreatedAt),
		"post_id", post.ID, "author_id", post.AuthorID)
	return post, nil
}

func (s *PostService) DistributeFeedNow(ctx context.Context, authorID, postID uint64, createdAt time.Time) error {
	if authorID == 0 || postID == 0 || createdAt.IsZero() {
		return ErrInvalidInput
	}
	followerIDs, err := s.listFollowerIDs(ctx, authorID)
	if err != nil {
		return err
	}
	var distributeErr error
	for _, followerID := range followerIDs {
		if err := s.inboxes.AddPost(ctx, followerID, postID, createdAt); err != nil {
			logSideEffectErr("inbox_add", err, "follower_id", followerID, "post_id", postID)
			distributeErr = errors.Join(distributeErr, err)
		}
	}
	return distributeErr
}

func (s *PostService) Get(ctx context.Context, id uint64) (*model.Post, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	return s.posts.FindByID(ctx, id)
}

func (s *PostService) Delete(ctx context.Context, userID, postID uint64) error {
	if userID == 0 || postID == 0 {
		return ErrInvalidInput
	}
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.AuthorID != userID {
		return ErrForbidden
	}
	if err := s.posts.DeleteByID(ctx, postID); err != nil {
		return err
	}
	logSideEffectErr("hot_score_delete", s.hotPosts.SetScore(ctx, postID, 0),
		"post_id", postID)
	return nil
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
	// Redis 热门榜只负责加速首屏。它是有限 topN 快照，后续页面必须按
	// score+id 游标查询 MySQL，才能继续访问快照边界之外的热门动态。
	if input.Cursor == nil {
		if hotItems, available, err := s.hotPosts.List(ctx, nil, int64(limit+1)); err == nil && available && len(hotItems) > 0 {
			ids := hotPostIDs(hotItems)
			posts, err := s.posts.ListByIDs(ctx, ids)
			if err != nil {
				return nil, err
			}
			// 缓存含已删除的残留 ID 时回源，避免首屏缺项或提前结束分页。
			if len(posts) == len(ids) {
				return buildHotPostListResult(orderPostsByIDs(posts, ids), limit, hotPostScores(hotItems)), nil
			}
		}
	}

	posts, err := s.posts.ListHot(ctx, input.Cursor, limit+1)
	if err != nil {
		return nil, err
	}
	return buildHotPostListResult(posts, limit, nil), nil
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
		result.NextCursor, _ = pagination.Encode(pagination.Chronological(
			posts[len(posts)-1].CreatedAt,
			posts[len(posts)-1].ID,
		))
	}
	return result
}

func buildHotPostListResult(posts []model.Post, limit int, scores map[uint64]int64) *PostListResult {
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}

	result := &PostListResult{
		Items:   posts,
		HasMore: hasMore,
	}
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		score := hotScore(last.LikeCount, last.FavoriteCount, last.CommentCount)
		if cachedScore, ok := scores[last.ID]; ok {
			score = cachedScore
		}
		result.NextCursor, _ = pagination.Encode(pagination.Hot(
			score,
			last.ID,
		))
	}
	return result
}

func hotPostIDs(items []cache.HotPostItem) []uint64 {
	ids := make([]uint64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.PostID)
	}
	return ids
}

func hotPostScores(items []cache.HotPostItem) map[uint64]int64 {
	scores := make(map[uint64]int64, len(items))
	for _, item := range items {
		scores[item.PostID] = item.Score
	}
	return scores
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

func (s *PostService) listFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	if ids, hit, err := s.relations.FollowerIDs(ctx, userID); err == nil && hit {
		return ids, nil
	}

	ids, err := s.follows.ListFollowerIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = s.relations.SetFollowerIDs(ctx, userID, ids)
	return ids, nil
}

func hotScore(likeCount, favoriteCount, commentCount int64) int64 {
	return likeCount*3 + favoriteCount*5 + commentCount*4
}

// HotPostsRebuildLimit 是定时重建时从 MySQL 取出的 topN 数量。
// 重建是按完整热度分排序，覆盖 Redis 中的 hot_posts ZSET。
const HotPostsRebuildLimit = 200

// RebuildHotPosts 从 MySQL 拉取 topN 帖子，按统一公式算分并整体重写 Redis 热门榜，
// 解决缓存丢失/淘汰后被零散 ZADD 填出"半截榜"的问题。
func (s *PostService) RebuildHotPosts(ctx context.Context) error {
	posts, err := s.posts.ListHot(ctx, nil, HotPostsRebuildLimit)
	if err != nil {
		return err
	}
	items := make([]cache.HotPostItem, 0, len(posts))
	for _, p := range posts {
		score := hotScore(p.LikeCount, p.FavoriteCount, p.CommentCount)
		if score <= 0 {
			continue
		}
		items = append(items, cache.HotPostItem{PostID: p.ID, Score: score})
	}
	return s.hotPosts.Rebuild(ctx, items)
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
