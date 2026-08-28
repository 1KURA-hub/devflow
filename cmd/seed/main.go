package main

import (
	"errors"
	"log"
	"strings"
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
	CoverURL       string
	Tags           string
	CreatedAgo     time.Duration
}

func main() {
	cfg := config.Load()
	if err := ensureSeedAllowed(cfg.AppEnv); err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := seed(database); err != nil {
		log.Fatalf("seed database: %v", err)
	}
	log.Println("demo data seeded")
}

func ensureSeedAllowed(appEnv string) error {
	if strings.EqualFold(strings.TrimSpace(appEnv), "prod") {
		return errors.New("refusing to seed data in production")
	}
	return nil
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
		{Username: "mysqljian", Nickname: "MySQL 简", Bio: "索引、事务和慢查询复盘", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=mysqljian"},
		{Username: "cachechen", Nickname: "缓存陈", Bio: "Redis、缓存一致性和热点治理", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=cachechen"},
		{Username: "mqzhou", Nickname: "消息队列周", Bio: "RabbitMQ 与异步架构实践", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=mqzhou"},
		{Username: "dockeryang", Nickname: "Docker 杨", Bio: "容器化部署和 Compose 运维", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=dockeryang"},
		{Username: "k8swang", Nickname: "K8s 王", Bio: "云原生发布与可观测性", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=k8swang"},
		{Username: "apizhao", Nickname: "接口赵", Bio: "REST API、幂等和错误码设计", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=apizhao"},
		{Username: "srehe", Nickname: "SRE 何", Bio: "线上稳定性、告警和容量规划", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=srehe"},
		{Username: "securityluo", Nickname: "安全罗", Bio: "认证鉴权和后端安全边界", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=securityluo"},
		{Username: "ragwu", Nickname: "RAG 吴", Bio: "检索增强和工程化落地", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=ragwu"},
		{Username: "perfma", Nickname: "性能马", Bio: "压测、Profiling 和性能优化", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=perfma"},
		{Username: "archsun", Nickname: "架构孙", Bio: "分层架构和演进式重构", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=archsun"},
		{Username: "obsqian", Nickname: "观测钱", Bio: "日志、指标、Tracing 三件套", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=obsqian"},
		{Username: "jobxu", Nickname: "任务徐", Bio: "定时任务、异步任务和补偿机制", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=jobxu"},
		{Username: "gatewaygao", Nickname: "网关高", Bio: "API 网关和流量治理", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=gatewaygao"},
		{Username: "testtang", Nickname: "测试唐", Bio: "后端测试、契约测试和回归验证", AvatarURL: "https://api.dicebear.com/9.x/notionists/svg?seed=testtang"},
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

	if err := database.Model(&model.Post{}).
		Where("title = ?", "Redis 热 key 排查的三个层次").
		Update("status", 0).Error; err != nil {
		return err
	}

	posts := []seedPost{
		{AuthorUsername: "gonight", Title: "Go 服务里我最常用的并发模式", Content: "业务 Go 服务里最常用的并发模式其实很少：工作池、errgroup、带缓冲 channel。工作池适合把外部调用限流，errgroup 适合聚合多个可取消任务，缓冲 channel 适合削峰但不能替代队列。设计时我会先定取消链路，再定并发数，最后才写 goroutine。否则一旦请求超时，后台任务还在跑，问题会非常隐蔽。", CoverURL: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1200&q=80", Tags: "Go,并发编程,后端", CreatedAgo: 5 * time.Hour},
		{AuthorUsername: "codingnav", Title: "Docker Compose 也能写得很工程化", Content: "很多小项目上线失败不是因为架构复杂，而是 Compose 写得太随意。生产配置里至少要有 env_file、命名 volume、healthcheck、明确端口暴露和依赖健康条件。数据库、Redis、RabbitMQ 这类状态服务不要依赖容器生命周期保存数据。把这些基础项写完整，小团队上线和回滚时会少掉很多临时排障。", CoverURL: "https://images.unsplash.com/photo-1605745341112-85968b19335b?auto=format&fit=crop&w=1200&q=80", Tags: "Docker,DevOps", CreatedAgo: 8 * time.Hour},
		{AuthorUsername: "melon", Title: "复杂页面别急着加动效", Content: "复杂页面的高级感通常不是动效堆出来的，而是信息层级、滚动职责和间距稳定。主内容、侧边栏、弹窗如果各自承担不同滚动区域，就要让滚动条出现在用户预期的位置。按钮动效只在用户触发时出现，刷新或数据回填不应该播放动画。先把这些基本关系做好，视觉自然会安静很多。", CoverURL: "https://images.unsplash.com/photo-1497366754035-f200968a6e72?auto=format&fit=crop&w=1200&q=80", Tags: "前端,UI,体验", CreatedAgo: 12 * time.Hour},
		{AuthorUsername: "lin", Title: "Feed 流项目最值得讲的不是列表接口", Content: "Feed 流项目面试时不要只讲分页列表。真正能展开的是关注关系如何维护、写扩散还是读扩散、热点内容如何排序、缓存和数据库如何保持一致。小规模可以实时查询，大规模要把写入、分发、通知拆开。接口只是表层，真正体现后端能力的是数据流和失败补偿怎么设计。", CoverURL: "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80", Tags: "系统设计,Feed,Redis", CreatedAgo: 18 * time.Hour},
		{AuthorUsername: "liuchao", Title: "什么时候该把同步通知改成异步", Content: "当一个写接口开始承担越来越多副作用时，就该考虑异步通知。比如点赞不仅要写 like 表，还要更新计数、写通知、刷新热榜、推送给作者。同步做会让主链路变慢，也更容易部分失败。更稳妥的做法是主链路只保证核心关系写成功，后续通知和分发通过消息队列消费，并提供幂等处理。", CoverURL: "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=1200&q=80", Tags: "RabbitMQ,架构设计", CreatedAgo: 28 * time.Hour},
		{AuthorUsername: "mysqljian", Title: "MySQL 慢查询先看执行计划还是索引", Content: "慢查询排查我会先拿真实 SQL、参数和数据量，再看 explain。只看有没有索引很容易误判，因为优化器可能因为选择性差、函数包裹、隐式转换或排序回表放弃索引。正确流程是看 rows、type、extra，再结合业务访问模式决定联合索引顺序。索引不是越多越好，写入成本和维护成本也要一起算。", CoverURL: "https://images.unsplash.com/photo-1544383835-bda2bc66a55d?auto=format&fit=crop&w=1200&q=80", Tags: "MySQL,索引,后端", CreatedAgo: 32 * time.Hour},
		{AuthorUsername: "cachechen", Title: "缓存一致性不是一句先删缓存能解决", Content: "缓存一致性要先区分读多写少、强一致和最终一致场景。简单业务可以更新数据库后删除缓存，但要考虑删除失败、并发回填和热点 key 击穿。更稳的方案是延迟双删、消息补偿、版本号或逻辑过期。不要把所有场景都套成一种模式，先明确用户能接受多长时间的不一致。", CoverURL: "https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=1200&q=80", Tags: "Redis,缓存一致性,架构设计", CreatedAgo: 36 * time.Hour},
		{AuthorUsername: "mqzhou", Title: "RabbitMQ 消费幂等比重试更重要", Content: "消息系统里重试只是兜底，真正决定稳定性的是消费幂等。消费者可能因为超时、网络抖动或进程重启处理同一条消息多次。如果业务写入没有唯一键、状态机或去重表，重试会把数据越修越乱。设计消息时要把 event_id、业务唯一键和失败记录一起考虑，保证重复消费不会改变最终结果。", CoverURL: "https://images.unsplash.com/photo-1451187580459-43490279c0fa?auto=format&fit=crop&w=1200&q=80", Tags: "RabbitMQ,幂等,后端", CreatedAgo: 40 * time.Hour},
		{AuthorUsername: "apizhao", Title: "接口错误码别只返回请求失败", Content: "后端接口错误码要服务前端判断，而不是让所有异常都变成请求失败。参数错误、未登录、无权限、资源不存在、重复操作和系统错误应该有明确区分。响应体可以保持统一结构，但 message 要给用户看得懂，code 要给前端可判断。这样前端交互会简单很多，也不会因为猜测后端状态写一堆脆弱逻辑。", CoverURL: "https://images.unsplash.com/photo-1555066931-4365d14bab8c?auto=format&fit=crop&w=1200&q=80", Tags: "API,错误码,后端", CreatedAgo: 44 * time.Hour},
		{AuthorUsername: "srehe", Title: "线上告警要从用户影响反推", Content: "告警不是指标越多越好。先定义用户真正感知的问题：请求失败率、核心接口延迟、队列积压、数据库连接耗尽。再把底层指标作为定位辅助，而不是全部打到手机上。一个好的告警应该能回答三件事：影响什么用户、可能是什么原因、第一步该看哪里。否则告警只会制造噪音。", CoverURL: "https://images.unsplash.com/photo-1504384308090-c894fdcc538d?auto=format&fit=crop&w=1200&q=80", Tags: "SRE,监控,稳定性", CreatedAgo: 48 * time.Hour},
		{AuthorUsername: "k8swang", Title: "K8s 发布失败常见在探针配置", Content: "很多 K8s 发布问题不是镜像坏了，而是探针写得不合理。startupProbe 太短会让慢启动服务反复重启，readinessProbe 太宽松会把未准备好的实例接入流量，livenessProbe 太激进会在短暂 GC 或下游抖动时杀进程。发布前应该根据真实启动时间和依赖检查拆开三类探针。", CoverURL: "https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&w=1200&q=80", Tags: "K8s,发布,云原生", CreatedAgo: 52 * time.Hour},
		{AuthorUsername: "securityluo", Title: "JWT 鉴权别忽略退出和刷新", Content: "JWT 最大的优点是无状态，最大的麻烦也是无状态。只做登录签发 token 很简单，但退出、封禁、刷新、权限变更都需要额外设计。小项目可以用短过期时间和刷新 token，大一点就要考虑黑名单或 token version。鉴权不是中间件里 parse 一下就结束，关键是生命周期。", CoverURL: "https://images.unsplash.com/photo-1563986768609-322da13575f3?auto=format&fit=crop&w=1200&q=80", Tags: "JWT,安全,后端", CreatedAgo: 56 * time.Hour},
		{AuthorUsername: "perfma", Title: "压测前先确认瓶颈假设", Content: "压测不是把 QPS 打上去看哪里炸。开始前要先写下瓶颈假设：数据库连接、慢 SQL、外部接口、锁竞争、GC、网络带宽还是队列消费能力。压测过程中只改一个变量，并记录资源曲线。否则得到的只是一个数字，无法指导优化。真正有价值的是瓶颈定位和优化前后的对比。", CoverURL: "https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&w=1200&q=80", Tags: "压测,性能优化,Go", CreatedAgo: 60 * time.Hour},
		{AuthorUsername: "archsun", Title: "重构服务层时先稳定返回值", Content: "服务层重构最怕一边改数据流一边改接口语义。比如点赞、收藏、关注这类操作，前端真正需要的是操作成功后的状态，而不是理解后端内部是否创建了关系。后端内部可以用唯一索引和事务保证并发安全，对外接口要保持简单。先稳定边界，再整理实现，风险会小很多。", CoverURL: "https://images.unsplash.com/photo-1552664730-d307ca884978?auto=format&fit=crop&w=1200&q=80", Tags: "重构,服务层,Go", CreatedAgo: 64 * time.Hour},
	}

	postIDs := make(map[string]uint64, len(posts))
	for _, item := range posts {
		post := model.Post{
			AuthorID:  userIDs[item.AuthorUsername],
			Title:     item.Title,
			Content:   item.Content,
			CoverURL:  item.CoverURL,
			Tags:      item.Tags,
			Status:    1,
			CreatedAt: now.Add(-item.CreatedAgo),
			UpdatedAt: now.Add(-item.CreatedAgo),
		}
		if err := database.Where("author_id = ? AND title = ?", post.AuthorID, post.Title).
			Assign(model.Post{Content: post.Content, CoverURL: post.CoverURL, Tags: post.Tags, Status: post.Status}).
			FirstOrCreate(&post).Error; err != nil {
			return err
		}
		postIDs[item.Title] = post.ID
	}

	follows := [][2]string{
		{"lin", "liuchao"},
		{"lin", "codingnav"},
		{"lin", "gonight"},
		{"lin", "mysqljian"},
		{"lin", "cachechen"},
		{"lin", "mqzhou"},
		{"codingnav", "liuchao"},
		{"gonight", "liuchao"},
		{"melon", "gonight"},
		{"mysqljian", "lin"},
		{"cachechen", "lin"},
		{"mqzhou", "lin"},
		{"apizhao", "lin"},
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
		{"lin", "什么时候该把同步通知改成异步"},
		{"codingnav", "什么时候该把同步通知改成异步"},
		{"gonight", "什么时候该把同步通知改成异步"},
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
		{"lin", "什么时候该把同步通知改成异步"},
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
		{User: "lin", Post: "什么时候该把同步通知改成异步", Content: "把通知拆出去之后，主链路确实清楚很多。"},
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

	postID := posts["什么时候该把同步通知改成异步"]
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
