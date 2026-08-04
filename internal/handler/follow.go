package handler

import (
	"errors"
	"net/http"
	"strconv"

	"devflow/internal/middleware"
	"devflow/internal/pagination"
	"devflow/internal/repository"
	"devflow/internal/response"
	"devflow/internal/service"

	"github.com/gin-gonic/gin"
)

func (a *App) followUser(c *gin.Context) {
	currentUserID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	targetUserID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := a.followService.Follow(c.Request.Context(), currentUserID, targetUserID); err != nil {
		writeFollowError(c, err)
		return
	}
	response.OK(c, nil)
}

func (a *App) unfollowUser(c *gin.Context) {
	currentUserID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	targetUserID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := a.followService.Unfollow(c.Request.Context(), currentUserID, targetUserID); err != nil {
		writeFollowError(c, err)
		return
	}
	response.OK(c, nil)
}

func (a *App) getFollowState(c *gin.Context) {
	currentUserID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	targetUserID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	followed, err := a.followService.IsFollowing(c.Request.Context(), currentUserID, targetUserID)
	if err != nil {
		writeFollowError(c, err)
		return
	}
	response.OK(c, gin.H{"followed": followed})
}

func (a *App) getUserProfile(c *gin.Context) {
	userID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	profile, err := a.followService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		writeFollowError(c, err)
		return
	}
	response.OK(c, profile)
}

func (a *App) listFollowingUsers(c *gin.Context) {
	userID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	input, err := parseUserListQuery(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.followService.ListFollowingUsers(c.Request.Context(), userID, input)
	if err != nil {
		writeFollowError(c, err)
		return
	}
	response.OK(c, result)
}

func (a *App) listFollowerUsers(c *gin.Context) {
	userID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	input, err := parseUserListQuery(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.followService.ListFollowerUsers(c.Request.Context(), userID, input)
	if err != nil {
		writeFollowError(c, err)
		return
	}
	response.OK(c, result)
}

func (a *App) listFollowingFeed(c *gin.Context) {
	currentUserID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	input, err := parseListQuery(c, pagination.KindChronological)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := a.followService.ListFollowingFeed(c.Request.Context(), currentUserID, input)
	if err != nil {
		writeFollowError(c, err)
		return
	}
	a.writePostList(c, result)
}

func parseUserListQuery(c *gin.Context) (service.UserListInput, error) {
	limit := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return service.UserListInput{}, errors.New("invalid limit")
		}
		limit = parsed
	}

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return service.UserListInput{}, errors.New("invalid offset")
		}
		offset = parsed
	}

	return service.UserListInput{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func writeFollowError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAlreadyFollowed):
		// 幂等：重复关注按成功处理，与重复点赞行为一致
		response.OK(c, nil)
	case errors.Is(err, service.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalid follow input")
	case errors.Is(err, service.ErrCannotFollowSelf):
		response.Error(c, http.StatusBadRequest, "cannot follow yourself")
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "resource not found")
	default:
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
