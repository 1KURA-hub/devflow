package main

import (
	"log"
	"time"

	"devflow/internal/auth"
	"devflow/internal/config"
	"devflow/internal/db"
	"devflow/internal/model"

	"gorm.io/gorm"
)

type seedUser struct {
	Username  string
	Nickname  string
	Bio       string
	AvatarURL string
}

type seedPost struct {
	AuthorUsername string
	Title          string
	Content        string
	Tags           string
	CreatedAgo     time.Duration
}

func main() {
	cfg := config.Load()
	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := seed(database); err != nil {
		log.Fatalf("seed database: %v", err)
	}
	log.Println("demo data seeded")
}

func seed(database *gorm.DB) error {
	now := time.Now()
	passwordHash, err := auth.HashPassword("devflow123")
	if err != nil {
		return err
	}

	users := []seedUser{
		{Username: "lin", Nickname: "Lin", Bio: "全栈开发工程师", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=lin"},
		{Username: "liuchao", Nickname: "刘超的技术博客", Bio: "后端与系统设计分享", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=liuchao"},
		{Username: "codingnav", Nickname: "编程导航", Bio: "整理实用开发资源", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=codingnav"},
		{Username: "gonight", Nickname: "Go 夜读", Bio: "Go 语言与工程实践", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=gonight"},
		{Username: "melon", Nickname: "前端西瓜哥", Bio: "前端性能与交互体验", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=melon"},
	}

	userIDs := make(map[string]uint64, len(users))
	for _, item := range users {
		user := model.User{
			Username:     item.Username,
			PasswordHash: passwordHash,
			Nickname:     item.Nickname,
			Bio:          item.Bio,
			AvatarURL:    item.AvatarURL,
			Status:       1,
		}
		if err := database.Where("username = ?", item.Username).
			Assign(model.User{
				Nickname:  item.Nickname,
				Bio:       item.Bio,
				AvatarURL: item.AvatarURL,
				Status:    1,
			}).
			FirstOrCreate(&user).Error; err != nil {
			return err
		}
		userIDs[item.Username] = user.ID
	}

	posts := []seedPost{
		{AuthorUsername: "liuchao", Title: "Redis 热 key 排查的三个层次", Content: "先看访问分布，再看对象粒度，最后再决定是拆 key、加本地缓存，还是做异步削峰。", Tags: "Redis,缓存,性能优化", CreatedAgo: 2 * time.Hour},
		{AuthorUsername: "gonight", Title: "Go 服务里我最常用的并发模式", Content: "工作池、errgroup、带缓冲 channel 这三种模式足够覆盖大多数业务服务，关键是先把取消链路设计清楚。", Tags: "Go,并发编程,后端", CreatedAgo: 5 * time.Hour},
		{AuthorUsername: "codingnav", Title: "Docker Compose 也能写得很工程化", Content: "把 healthcheck、命名 volume、env_file 和依赖顺序写完整，小项目上线时会省掉很多排障时间。", Tags: "Docker,DevOps", CreatedAgo: 8 * time.Hour},
		{AuthorUsername: "melon", Title: "复杂页面别急着加动效", Content: "先把信息层级、滚动职责和对齐关系做好，交互自然会显得高级很多。", Tags: "前端,UI,体验", CreatedAgo: 12 * time.Hour},
		{AuthorUsername: "lin", Title: "Feed 流项目最值得讲的不是列表接口", Content: "真正能展开的是关注关系、推拉结合、缓存一致性和异步分发，接口只是表层。", Tags: "系统设计,Feed,Redis", CreatedAgo: 18 * time.Hour},
		{AuthorUsername: "liuchao", Title: "什么时候该把同步通知改成异步", Content: "当写路径已经承担过多副作用时，就该把通知和分发拆出去，否则主链路会越来越脆。", Tags: "RabbitMQ,架构设计", CreatedAgo: 28 * time.Hour},
	}

	postIDs := make(map[string]uint64, len(posts))
	for _, item := range posts {
		post := model.Post{
			AuthorID:  userIDs[item.AuthorUsername],
			Title:     item.Title,
			Content:   item.Content,
			Tags:      item.Tags,
			Status:    1,
			CreatedAt: now.Add(-item.CreatedAgo),
			UpdatedAt: now.Add(-item.CreatedAgo),
		}
		if err := database.Where("author_id = ? AND title = ?", post.AuthorID, post.Title).
			Attrs(post).
			FirstOrCreate(&post).Error; err != nil {
			return err
		}
		postIDs[item.Title] = post.ID
	}

	follows := [][2]string{
		{"lin", "liuchao"},
		{"lin", "codingnav"},
		{"lin", "gonight"},
		{"codingnav", "liuchao"},
		{"gonight", "liuchao"},
		{"melon", "gonight"},
	}
	for _, pair := range follows {
		follow := model.Follow{
			FollowerID: userIDs[pair[0]],
			FolloweeID: userIDs[pair[1]],
			CreatedAt:  now.Add(-24 * time.Hour),
		}
		if err := database.Where("follower_id = ? AND followee_id = ?", follow.FollowerID, follow.FolloweeID).
			Attrs(follow).
			FirstOrCreate(&follow).Error; err != nil {
			return err
		}
	}

	if err := seedInteractions(database, now, userIDs, postIDs); err != nil {
		return err
	}
	return refreshPostCounts(database)
}

func seedInteractions(database *gorm.DB, now time.Time, users map[string]uint64, posts map[string]uint64) error {
	likePairs := [][2]string{
		{"lin", "Redis 热 key 排查的三个层次"},
		{"codingnav", "Redis 热 key 排查的三个层次"},
		{"gonight", "Redis 热 key 排查的三个层次"},
		{"lin", "Go 服务里我最常用的并发模式"},
		{"liuchao", "Docker Compose 也能写得很工程化"},
		{"melon", "Feed 流项目最值得讲的不是列表接口"},
	}
	for _, pair := range likePairs {
		like := model.Like{UserID: users[pair[0]], PostID: posts[pair[1]], CreatedAt: now.Add(-3 * time.Hour)}
		if err := database.Where("user_id = ? AND post_id = ?", like.UserID, like.PostID).
			Attrs(like).
			FirstOrCreate(&like).Error; err != nil {
			return err
		}
	}

	favoritePairs := [][2]string{
		{"lin", "Redis 热 key 排查的三个层次"},
		{"lin", "Go 服务里我最常用的并发模式"},
		{"melon", "Docker Compose 也能写得很工程化"},
	}
	for _, pair := range favoritePairs {
		favorite := model.Favorite{UserID: users[pair[0]], PostID: posts[pair[1]], CreatedAt: now.Add(-2 * time.Hour)}
		if err := database.Where("user_id = ? AND post_id = ?", favorite.UserID, favorite.PostID).
			Attrs(favorite).
			FirstOrCreate(&favorite).Error; err != nil {
			return err
		}
	}

	comments := []struct {
		User    string
		Post    string
		Content string
	}{
		{User: "lin", Post: "Redis 热 key 排查的三个层次", Content: "先量化访问分布这点很实用。"},
		{User: "codingnav", Post: "Go 服务里我最常用的并发模式", Content: "取消链路确实比 goroutine 数量更容易出问题。"},
		{User: "melon", Post: "Feed 流项目最值得讲的不是列表接口", Content: "这个角度很适合面试展开。"},
	}
	for _, item := range comments {
		comment := model.Comment{
			UserID:    users[item.User],
			PostID:    posts[item.Post],
			Content:   item.Content,
			Status:    1,
			CreatedAt: now.Add(-90 * time.Minute),
			UpdatedAt: now.Add(-90 * time.Minute),
		}
		if err := database.Where("user_id = ? AND post_id = ? AND content = ?", comment.UserID, comment.PostID, comment.Content).
			Attrs(comment).
			FirstOrCreate(&comment).Error; err != nil {
			return err
		}
	}

	postID := posts["Redis 热 key 排查的三个层次"]
	notifications := []model.Notification{
		{UserID: users["liuchao"], ActorID: users["lin"], Type: "like", PostID: &postID, Content: "Lin 赞了你的动态", CreatedAt: now.Add(-70 * time.Minute)},
		{UserID: users["liuchao"], ActorID: users["lin"], Type: "comment", PostID: &postID, Content: "Lin 评论了你的动态", CreatedAt: now.Add(-60 * time.Minute)},
	}
	for _, notification := range notifications {
		if err := database.Where("user_id = ? AND actor_id = ? AND type = ? AND content = ?", notification.UserID, notification.ActorID, notification.Type, notification.Content).
			Attrs(notification).
			FirstOrCreate(&notification).Error; err != nil {
			return err
		}
	}
	return nil
}

func refreshPostCounts(database *gorm.DB) error {
	return database.Exec(`
		UPDATE posts
		SET
			like_count = (SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id),
			favorite_count = (SELECT COUNT(*) FROM favorites WHERE favorites.post_id = posts.id),
			comment_count = (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id AND comments.deleted_at IS NULL)
	`).Error
}
