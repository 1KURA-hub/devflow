package handler

import (
	"errors"
	"net/http"

	"devflow/internal/middleware"
	"devflow/internal/pagination"
	"devflow/internal/repository"
	"devflow/internal/response"
	"devflow/internal/service"

	"github.com/gin-gonic/gin"
)

type createCommentRequest struct {
	Content string `json:"content"`
}

func (a *App) createComment(c *gin.Context) {
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

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	comment, err := a.commentService.Create(c.Request.Context(), service.CreateCommentInput{
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
	})
	if err != nil {
		writeCommentError(c, err)
		return
	}
	response.OK(c, comment)
}

func (a *App) listComments(c *gin.Context) {
	postID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post id")
		return
	}
	input, err := parseListQuery(c, pagination.KindChronological)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.commentService.ListByPost(c.Request.Context(), postID, input)
	if err != nil {
		writeCommentError(c, err)
		return
	}
	response.OK(c, result)
}

func writeCommentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalid comment input")
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "resource not found")
	default:
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
