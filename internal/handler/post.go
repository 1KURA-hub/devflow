package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	tokenauth "devflow/internal/auth"
	"devflow/internal/middleware"
	"devflow/internal/model"
	"devflow/internal/repository"
	"devflow/internal/response"
	"devflow/internal/service"

	"github.com/gin-gonic/gin"
)

type createPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"`
}

type postResponse struct {
	model.Post
	Liked     bool `json:"liked"`
	Favorited bool `json:"favorited"`
}

type postListResponse struct {
	Items      []postResponse `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

func (a *App) createPost(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := a.postService.Create(c.Request.Context(), service.CreatePostInput{
		AuthorID: userID,
		Title:    req.Title,
		Content:  req.Content,
		Tags:     req.Tags,
	})
	if err != nil {
		writePostError(c, err)
		return
	}
	response.OK(c, post)
}

func (a *App) getPost(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := a.postService.Get(c.Request.Context(), id)
	if err != nil {
		writePostError(c, err)
		return
	}
	response.OK(c, a.postResponse(c, *post))
}

func (a *App) listLatestPosts(c *gin.Context) {
	input, err := parseListQuery(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.postService.ListLatest(c.Request.Context(), input)
	if err != nil {
		writePostError(c, err)
		return
	}
	a.writePostList(c, result)
}

func (a *App) listHotPosts(c *gin.Context) {
	input, err := parseListQuery(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.postService.ListHot(c.Request.Context(), input)
	if err != nil {
		writePostError(c, err)
		return
	}
	a.writePostList(c, result)
}

func (a *App) listUserPosts(c *gin.Context) {
	userID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	input, err := parseListQuery(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.postService.ListByAuthor(c.Request.Context(), userID, input)
	if err != nil {
		writePostError(c, err)
		return
	}
	a.writePostList(c, result)
}

func (a *App) deletePost(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	postID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post id")
		return
	}
	if err := a.postService.Delete(c.Request.Context(), userID, postID); err != nil {
		writePostError(c, err)
		return
	}
	response.OK(c, nil)
}

func (a *App) writePostList(c *gin.Context, result *service.PostListResult) {
	response.OK(c, postListResponse{
		Items:      a.postResponses(c, result.Items),
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	})
}

func (a *App) postResponse(c *gin.Context, post model.Post) postResponse {
	return a.postResponses(c, []model.Post{post})[0]
}

func (a *App) postResponses(c *gin.Context, posts []model.Post) []postResponse {
	items := make([]postResponse, 0, len(posts))
	if len(posts) == 0 {
		return items
	}

	postIDs := make([]uint64, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}
	userID := a.viewerUserID(c)
	liked, favorited, err := a.interactionService.PostStates(c.Request.Context(), userID, postIDs)
	if err != nil {
		liked = map[uint64]bool{}
		favorited = map[uint64]bool{}
	}

	for _, post := range posts {
		items = append(items, postResponse{
			Post:      post,
			Liked:     liked[post.ID],
			Favorited: favorited[post.ID],
		})
	}
	return items
}

func (a *App) viewerUserID(c *gin.Context) uint64 {
	if userID, ok := middleware.CurrentUserID(c); ok {
		return userID
	}
	header := c.GetHeader("Authorization")
	if header == "" {
		return 0
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return 0
	}
	claims, err := tokenauth.ParseToken(parts[1], a.cfg.JWTSecret)
	if err != nil {
		return 0
	}
	return claims.UserID
}

func parseUint64Param(c *gin.Context, name string) (uint64, error) {
	value := c.Param(name)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func parseListQuery(c *gin.Context) (service.ListInput, error) {
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return service.ListInput{}, errors.New("invalid limit")
		}
		limit = parsed
	}

	var cursor *time.Time
	if raw := c.Query("cursor"); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return service.ListInput{}, errors.New("invalid cursor")
		}
		cursor = &parsed
	}

	return service.ListInput{
		Cursor: cursor,
		Limit:  limit,
	}, nil
}

func writePostError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalid post input")
	case errors.Is(err, service.ErrForbidden):
		response.Error(c, http.StatusForbidden, "forbidden")
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "resource not found")
	default:
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
