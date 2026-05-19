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

func (a *App) listNotifications(c *gin.Context) {
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

	result, err := a.notificationService.List(c.Request.Context(), userID, input)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	response.OK(c, result)
}

func (a *App) getUnreadNotificationCount(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := a.notificationService.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		writeNotificationError(c, err)
		return
	}
	response.OK(c, gin.H{"unread_count": count})
}

func (a *App) markNotificationRead(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	notificationID, err := parseUint64Param(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := a.notificationService.MarkRead(c.Request.Context(), userID, notificationID); err != nil {
		writeNotificationError(c, err)
		return
	}
	response.OK(c, nil)
}

func (a *App) markAllNotificationsRead(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := a.notificationService.MarkAllRead(c.Request.Context(), userID); err != nil {
		writeNotificationError(c, err)
		return
	}
	response.OK(c, nil)
}

func writeNotificationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		response.Error(c, http.StatusBadRequest, "invalid notification input")
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "notification not found")
	default:
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
