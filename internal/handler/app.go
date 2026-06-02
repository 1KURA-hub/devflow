package handler

import (
	"context"
	"log"
	"time"

	"devflow/internal/cache"
	"devflow/internal/config"
	"devflow/internal/mq"
	"devflow/internal/repository"
	"devflow/internal/service"
	"devflow/internal/worker"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// hotPostsRebuildInterval 是定时重建热门榜的周期。
// 取值是“偶尔脏一点也能接受 + 不给 DB 太大压力”的折中。
const hotPostsRebuildInterval = 5 * time.Minute

type Dependencies struct {
	Config      config.Config
	DB          *gorm.DB
	RedisClient *redis.Client
	Broker      *mq.Broker
	// EnableWorkers 控制是否在当前进程内启动 MQ 消费者。
	// API 进程通常设为 false，由独立的 worker 进程消费；保留 true 兼容单体部署。
	EnableWorkers bool
	// Ctx 用于驱动后台 goroutine（worker、ticker）退出，由调用方传入可取消的 ctx。
	Ctx context.Context
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
	ctx := deps.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

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

	if deps.EnableWorkers {
		if err := worker.StartNotificationConsumer(ctx, deps.Broker, notificationService); err != nil {
			panic(err)
		}
		if err := worker.StartFeedConsumer(ctx, deps.Broker, app.postService); err != nil {
			panic(err)
		}
		if deps.RedisClient != nil {
			startHotPostsRebuilder(ctx, app.postService)
		}
	}

	return app
}

// startHotPostsRebuilder 启动一个后台 goroutine，每隔 hotPostsRebuildInterval
// 从 MySQL 整体重建 Redis 热门榜，避免缓存丢失后被零散互动写入造成的"半截榜"。
// 启动时立即跑一次，达到热启动的效果。
func startHotPostsRebuilder(ctx context.Context, posts *service.PostService) {
	go func() {
		if err := posts.RebuildHotPosts(ctx); err != nil {
			log.Printf("hot_posts rebuild on startup failed: %v", err)
		}
		ticker := time.NewTicker(hotPostsRebuildInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := posts.RebuildHotPosts(ctx); err != nil {
					log.Printf("hot_posts rebuild failed: %v", err)
				}
			}
		}
	}()
}
