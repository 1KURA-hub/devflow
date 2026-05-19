package handler

import (
	"devflow/internal/middleware"
	"devflow/internal/response"

	"github.com/gin-gonic/gin"
)

func NewRouter(app *App) *gin.Engine {
	if app.cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", app.healthz)
	router.Static("/uploads", "./uploads")

	api := router.Group("/api")
	{
		api.GET("/healthz", app.healthz)
		api.POST("/auth/register", app.register)
		api.POST("/auth/login", app.login)
		api.GET("/posts/:id", app.getPost)
		api.GET("/posts/:id/comments", app.listComments)
		api.GET("/users/:id/posts", app.listUserPosts)
		api.GET("/users/:id/following", app.listFollowingUsers)
		api.GET("/users/:id/followers", app.listFollowerUsers)
		api.GET("/feed/latest", app.listLatestPosts)
		api.GET("/feed/hot", app.listHotPosts)

		authenticated := api.Group("")
		authenticated.Use(middleware.Auth(app.cfg.JWTSecret))
		{
			authenticated.GET("/me", app.me)
			authenticated.PATCH("/me", app.updateMe)
			authenticated.POST("/uploads/image", app.uploadImage)
			authenticated.POST("/posts", app.createPost)
			authenticated.POST("/users/:id/follow", app.followUser)
			authenticated.DELETE("/users/:id/follow", app.unfollowUser)
			authenticated.GET("/feed/following", app.listFollowingFeed)
			authenticated.POST("/posts/:id/like", app.likePost)
			authenticated.DELETE("/posts/:id/like", app.unlikePost)
			authenticated.POST("/posts/:id/favorite", app.favoritePost)
			authenticated.DELETE("/posts/:id/favorite", app.unfavoritePost)
			authenticated.POST("/posts/:id/comments", app.createComment)
			authenticated.GET("/me/favorites", app.listMyFavorites)
			authenticated.GET("/notifications", app.listNotifications)
			authenticated.GET("/notifications/unread-count", app.getUnreadNotificationCount)
			authenticated.POST("/notifications/read-all", app.markAllNotificationsRead)
			authenticated.POST("/notifications/:id/read", app.markNotificationRead)
		}
	}

	return router
}

func (a *App) healthz(c *gin.Context) {
	response.OK(c, gin.H{
		"status": "ok",
		"app":    "devflow",
		"env":    a.cfg.AppEnv,
	})
}
