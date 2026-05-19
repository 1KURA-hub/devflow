package handler

import (
	"errors"
	"net/http"

	"devflow/internal/middleware"
	"devflow/internal/repository"
	"devflow/internal/response"
	"devflow/internal/service"

	"github.com/gin-gonic/gin"
)

func (a *App) likePost(c *gin.Context) {
	userID, postID, ok := currentUserAndPostID(c)
	if !ok {
		return
	}
	changed, err := a.interactionService.Like(c.Request.Context(), userID, postID)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	response.OK(c, gin.H{"liked": true, "changed": changed})
}

func (a *App) unlikePost(c *gin.Context) {
	userID, postID, ok := currentUserAndPostID(c)
	if !ok {
		return
	}
	changed, err := a.interactionService.Unlike(c.Request.Context(), userID, postID)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	response.OK(c, gin.H{"liked": false, "changed": changed})
}

func (a *App) favoritePost(c *gin.Context) {
	userID, postID, ok := currentUserAndPostID(c)
	if !ok {
		return
	}
	changed, err := a.interactionService.Favorite(c.Request.Context(), userID, postID)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	response.OK(c, gin.H{"favorited": true, "changed": changed})
}

func (a *App) unfavoritePost(c *gin.Context) {
	userID, postID, ok := currentUserAndPostID(c)
	if !ok {
		return
	}
	changed, err := a.interactionService.Unfavorite(c.Request.Context(), userID, postID)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	response.OK(c, gin.H{"favorited": false, "changed": changed})
}

func (a *App) listMyFavorites(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	input, err := parseListQuery(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := a.interactionService.ListMyFavorites(c.Request.Context(), userID, input)
	if err != nil {
		writeInteractionError(c, err)
		return
	}
	a.writePostList(c, result)
}

func currentUserAndPostID(c *gin.Context) (uint64, uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return 0, 0, false
	}
	postID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid post id")
		return 0, 0, false
	}
	return userID, postID, true
}

func writeInteractionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalid interaction input")
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "resource not found")
	default:
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
