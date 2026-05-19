package handler

import (
	"net/http"
	"time"

	"devflow/internal/model"
	"devflow/internal/response"

	"github.com/gin-gonic/gin"
)

func (a *App) communityOverview(c *gin.Context) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var totalUsers int64
	var totalPosts int64
	var todayUsers int64
	var todayPosts int64

	if err := a.db.WithContext(c.Request.Context()).Model(&model.User{}).Where("status = ?", 1).Count(&totalUsers).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "overview query failed")
		return
	}
	if err := a.db.WithContext(c.Request.Context()).Model(&model.Post{}).Where("status = ?", 1).Count(&totalPosts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "overview query failed")
		return
	}
	if err := a.db.WithContext(c.Request.Context()).Model(&model.User{}).Where("status = ? AND created_at >= ?", 1, today).Count(&todayUsers).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "overview query failed")
		return
	}
	if err := a.db.WithContext(c.Request.Context()).Model(&model.Post{}).Where("status = ? AND created_at >= ?", 1, today).Count(&todayPosts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "overview query failed")
		return
	}

	response.OK(c, gin.H{
		"total_users": totalUsers,
		"total_posts": totalPosts,
		"today_users": todayUsers,
		"today_posts": todayPosts,
	})
}
