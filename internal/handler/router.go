package handler

import (
	"devflow/internal/config"
	"devflow/internal/middleware"
	"devflow/internal/repository"
	"devflow/internal/response"
	"devflow/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config config.Config
	DB     *gorm.DB
}

type App struct {
	cfg         config.Config
	db          *gorm.DB
	authService *service.AuthService
	postService *service.PostService
}

func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Config.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	userRepo := repository.NewUserRepository(deps.DB)
	postRepo := repository.NewPostRepository(deps.DB)
	app := &App{
		cfg:         deps.Config,
		db:          deps.DB,
		authService: service.NewAuthService(userRepo, deps.Config.JWTSecret),
		postService: service.NewPostService(postRepo, userRepo),
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
		api.GET("/users/:id/posts", app.listUserPosts)
		api.GET("/feed/latest", app.listLatestPosts)

		authenticated := api.Group("")
		authenticated.Use(middleware.Auth(deps.Config.JWTSecret))
		{
			authenticated.GET("/me", app.me)
			authenticated.POST("/posts", app.createPost)
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
