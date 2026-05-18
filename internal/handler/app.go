package handler

import (
	"context"

	"devflow/internal/cache"
	"devflow/internal/config"
	"devflow/internal/mq"
	"devflow/internal/repository"
	"devflow/internal/service"
	"devflow/internal/worker"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config      config.Config
	DB          *gorm.DB
	RedisClient *redis.Client
	Broker      *mq.Broker
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

func NewApp(deps Dependencies) *App {
	userRepo := repository.NewUserRepository(deps.DB)
	postRepo := repository.NewPostRepository(deps.DB)
	followRepo := repository.NewFollowRepository(deps.DB)
	interactionRepo := repository.NewInteractionRepository(deps.DB)
	commentRepo := repository.NewCommentRepository(deps.DB)
	notificationRepo := repository.NewNotificationRepository(deps.DB)

	notificationCounter := cache.NewNotificationCounter(deps.RedisClient)
	hotPostStore := cache.NewHotPostStore(deps.RedisClient)
	followRelationStore := cache.NewFollowRelationStore(deps.RedisClient)
	feedInboxStore := cache.NewFeedInboxStore(deps.RedisClient)
	eventPublisher := mq.NewPublisher(deps.Broker)

	notificationService := service.NewNotificationService(notificationRepo, notificationCounter, eventPublisher)
	app := &App{
		cfg:                 deps.Config,
		db:                  deps.DB,
		authService:         service.NewAuthService(userRepo, deps.Config.JWTSecret),
		postService:         service.NewPostService(postRepo, followRepo, userRepo, hotPostStore, followRelationStore, feedInboxStore, eventPublisher),
		followService:       service.NewFollowService(followRepo, userRepo, postRepo, notificationService, followRelationStore, feedInboxStore),
		interactionService:  service.NewInteractionService(interactionRepo, postRepo, userRepo, notificationService, hotPostStore),
		commentService:      service.NewCommentService(commentRepo, postRepo, userRepo, notificationService, hotPostStore),
		notificationService: notificationService,
	}

	if err := worker.StartNotificationConsumer(context.Background(), deps.Broker, notificationService); err != nil {
		panic(err)
	}
	if err := worker.StartFeedConsumer(context.Background(), deps.Broker, app.postService); err != nil {
		panic(err)
	}

	return app
}
