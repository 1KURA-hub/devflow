package handler

import (
	"devflow/internal/cache"
	"devflow/internal/config"
	"devflow/internal/middleware"
	"devflow/internal/repository"
	"devflow/internal/response"
	"devflow/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config      config.Config
	DB          *gorm.DB
	RedisClient *redis.Client
}

type App struct {
	cfg                 config.Config
	db                  *gorm.DB
	authService         *service.AuthService
	postService         *service.PostService
	followService       *service.FollowService
	interactionService  *service.InteractionService
	commentService      *service.CommentService
	notificationService *service.NotificationService
}

func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Config.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	userRepo := repository.NewUserRepository(deps.DB)
	postRepo := repository.NewPostRepository(deps.DB)
	followRepo := repository.NewFollowRepository(deps.DB)
	interactionRepo := repository.NewInteractionRepository(deps.DB)
	commentRepo := repository.NewCommentRepository(deps.DB)
	notificationRepo := repository.NewNotificationRepository(deps.DB)
	notificationCounter := cache.NewNotificationCounter(deps.RedisClient)
	hotPostStore := cache.NewHotPostStore(deps.RedisClient)
	notificationService := service.NewNotificationService(notificationRepo, notificationCounter)
	app := &App{
		cfg:                 deps.Config,
		db:                  deps.DB,
		authService:         service.NewAuthService(userRepo, deps.Config.JWTSecret),
		postService:         service.NewPostService(postRepo, userRepo, hotPostStore),
		followService:       service.NewFollowService(followRepo, userRepo, postRepo, notificationService),
		interactionService:  service.NewInteractionService(interactionRepo, postRepo, userRepo, notificationService, hotPostStore),
		commentService:      service.NewCommentService(commentRepo, postRepo, userRepo, notificationService, hotPostStore),
		notificationService: notificationService,
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", app.healthz)

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
		authenticated.Use(middleware.Auth(deps.Config.JWTSecret))
		{
			authenticated.GET("/me", app.me)
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
