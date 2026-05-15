# DevFlow 设计文档

## 1. 项目定位

DevFlow 是一个面向开发者的技术动态 Feed 平台。

用户可以发布短技术动态、关注其他开发者、浏览关注流、点赞、评论，并接收互动通知。项目重点不是做长文章 CMS，也不是做复杂推荐算法，而是实现一个真实后端业务中常见的 Feed 流系统。

第一版目标：

- 跑通用户、动态、关注、Feed、点赞、评论、通知的核心业务闭环。
- 先以 MySQL 实现完整功能。
- 后续再逐步接入 Redis 和 RabbitMQ，增强性能和异步处理能力。

## 2. 第一版做什么

MVP 必做功能：

- 用户注册、登录、获取当前用户信息
- 发布技术动态
- 查看动态详情
- 关注 / 取关用户
- 查看关注流
- 点赞 / 取消点赞
- 评论动态
- 查看通知列表

第一版不做：

- 图片上传
- 推荐算法
- 全文搜索
- 私信
- 长文章编辑器
- 多级评论
- 后台管理
- 微服务拆分
- 复杂权限系统

## 3. 业务主链路

### 3.1 注册登录链路

```text
用户提交账号密码
  -> 参数校验
  -> 检查用户名是否存在
  -> 密码哈希
  -> 写入 users 表
  -> 登录时签发 JWT
  -> 前端保存 token
```

第一版只做普通账号密码登录，不接第三方 OAuth。

### 3.2 发布动态链路

```text
用户发布动态
  -> JWT 鉴权
  -> 参数校验
  -> MySQL 写入 posts 表
  -> 返回发布成功
```

第一版发布动态只写 MySQL。后续接 RabbitMQ 时，再把 Feed 分发和通知异步化。

### 3.3 关注链路

```text
用户关注作者
  -> JWT 鉴权
  -> 检查不能关注自己
  -> 检查目标用户是否存在
  -> follows 表写入关注关系
  -> 唯一索引保证重复关注幂等
  -> 返回关注成功
```

取消关注：

```text
用户取消关注
  -> JWT 鉴权
  -> 删除 follows 表中的关注关系
  -> 返回取消成功
```

### 3.4 关注流查询链路

第一版使用 MySQL 查询关注流：

```text
用户请求关注流
  -> JWT 鉴权
  -> 查询我关注的用户 ID
  -> 查询这些用户发布的 posts
  -> 按发布时间倒序
  -> 游标分页返回
```

后续 Redis 版本演进为：

```text
发布动态
  -> RabbitMQ 投递 feed 分发任务
  -> Worker 查询粉丝列表
  -> 写入每个粉丝的 Redis Feed 收件箱
  -> 查询关注流时优先读 Redis
```

### 3.5 点赞链路

```text
用户点赞动态
  -> JWT 鉴权
  -> 检查动态是否存在
  -> likes 表写入 user_id + post_id
  -> 唯一索引防重复点赞
  -> posts.like_count + 1
  -> 给作者写入通知
  -> 返回点赞成功
```

取消点赞：

```text
用户取消点赞
  -> JWT 鉴权
  -> 删除 likes 记录
  -> posts.like_count - 1
  -> 返回取消成功
```

### 3.6 评论链路

```text
用户评论动态
  -> JWT 鉴权
  -> 检查动态是否存在
  -> comments 表写入评论
  -> posts.comment_count + 1
  -> 给作者写入通知
  -> 返回评论成功
```

第一版只做一级评论。

### 3.7 通知链路

```text
点赞 / 评论 / 关注产生事件
  -> MySQL 写入 notifications 表
  -> 用户打开通知中心
  -> 查询 notifications
  -> 支持标记已读
```

第一版同步写通知。后续接 RabbitMQ 后改成异步通知。

## 4. 数据库表设计

### 4.1 users

用户表。

```sql
CREATE TABLE users (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NOT NULL,
  bio VARCHAR(255) NOT NULL DEFAULT '',
  avatar_url VARCHAR(512) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  deleted_at DATETIME NULL,
  UNIQUE KEY uk_users_username (username),
  KEY idx_users_deleted_at (deleted_at)
);
```

### 4.2 posts

动态表。

```sql
CREATE TABLE posts (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  author_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(120) NOT NULL,
  content TEXT NOT NULL,
  tags VARCHAR(255) NOT NULL DEFAULT '',
  like_count INT NOT NULL DEFAULT 0,
  comment_count INT NOT NULL DEFAULT 0,
  favorite_count INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  deleted_at DATETIME NULL,
  KEY idx_posts_author_created (author_id, created_at),
  KEY idx_posts_created_at (created_at),
  KEY idx_posts_deleted_at (deleted_at)
);
```

说明：

- `tags` 第一版先用逗号分隔字符串，例如 `Go,Redis,Docker`。
- 后续如果要做标签详情页，再拆成 `tags` 和 `post_tags` 两张表。

### 4.3 follows

关注关系表。

```sql
CREATE TABLE follows (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  follower_id BIGINT UNSIGNED NOT NULL,
  followee_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_follows_pair (follower_id, followee_id),
  KEY idx_follows_followee (followee_id),
  KEY idx_follows_follower (follower_id)
);
```

说明：

- `follower_id` 是发起关注的人。
- `followee_id` 是被关注的人。
- 唯一索引用来保证重复关注幂等。

### 4.4 likes

点赞表。

```sql
CREATE TABLE likes (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  post_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_likes_user_post (user_id, post_id),
  KEY idx_likes_post (post_id)
);
```

### 4.5 favorites

收藏表。第一版可以先建表，接口可以晚一点做。

```sql
CREATE TABLE favorites (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  post_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL,
  UNIQUE KEY uk_favorites_user_post (user_id, post_id),
  KEY idx_favorites_post (post_id)
);
```

### 4.6 comments

评论表。

```sql
CREATE TABLE comments (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  post_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  content VARCHAR(1000) NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  deleted_at DATETIME NULL,
  KEY idx_comments_post_created (post_id, created_at),
  KEY idx_comments_user_created (user_id, created_at),
  KEY idx_comments_deleted_at (deleted_at)
);
```

第一版不做楼中楼，所以不加 `parent_id`。

### 4.7 notifications

通知表。

```sql
CREATE TABLE notifications (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  actor_id BIGINT UNSIGNED NOT NULL,
  type VARCHAR(32) NOT NULL,
  post_id BIGINT UNSIGNED NULL,
  comment_id BIGINT UNSIGNED NULL,
  content VARCHAR(255) NOT NULL,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME NOT NULL,
  KEY idx_notifications_user_read_created (user_id, is_read, created_at),
  KEY idx_notifications_user_created (user_id, created_at)
);
```

通知类型：

```text
follow
like
comment
```

## 5. GORM Model 规划

第一版建议按表一一对应：

```text
internal/model/user.go
internal/model/post.go
internal/model/follow.go
internal/model/like.go
internal/model/favorite.go
internal/model/comment.go
internal/model/notification.go
```

公共字段可以先不抽象，避免过早封装。

建议 ID 用 `uint64` 或 `uint`。如果使用 GORM 默认习惯，可以先用 `uint`，后面统一即可。

## 6. 接口设计

### 6.1 Auth

```text
POST /api/auth/register
POST /api/auth/login
GET  /api/me
```

### 6.2 Posts

```text
POST /api/posts
GET  /api/posts/:id
GET  /api/users/:id/posts
```

### 6.3 Feed

```text
GET /api/feed/following
GET /api/feed/latest
```

说明：

- `following` 是关注流。
- `latest` 是全站最新动态，用于新用户没有关注时兜底。
- 热门流第一版可先不做，第二版用 Redis ZSET 实现。

### 6.4 Follow

```text
POST   /api/users/:id/follow
DELETE /api/users/:id/follow
GET    /api/users/:id/followers
GET    /api/users/:id/following
```

### 6.5 Like

```text
POST   /api/posts/:id/like
DELETE /api/posts/:id/like
```

### 6.6 Favorite

```text
POST   /api/posts/:id/favorite
DELETE /api/posts/:id/favorite
GET    /api/me/favorites
```

收藏接口可以第二阶段实现。

### 6.7 Comment

```text
POST /api/posts/:id/comments
GET  /api/posts/:id/comments
```

### 6.8 Notification

```text
GET  /api/notifications
POST /api/notifications/:id/read
POST /api/notifications/read-all
```

## 7. 分页设计

Feed 列表使用游标分页，避免深分页性能问题。

请求：

```text
GET /api/feed/following?cursor=2026-05-15T10:00:00Z&limit=20
```

第一版游标可以使用 `created_at`。

查询逻辑：

```text
如果 cursor 为空：
  查询最新 20 条

如果 cursor 不为空：
  查询 created_at < cursor 的 20 条
```

响应：

```json
{
  "items": [],
  "next_cursor": "2026-05-15T09:58:00Z",
  "has_more": true
}
```

## 8. Redis 设计，第二阶段接入

第一版不强制使用 Redis。第二阶段再加入以下 key。

### 8.1 关注关系缓存

```text
devflow:user:following:{userID}
devflow:user:followers:{userID}
```

类型：

```text
Set
```

用途：

- 快速判断是否关注
- Feed 分发时快速获取粉丝列表

### 8.2 Feed 收件箱

```text
devflow:feed:inbox:{userID}
```

类型：

```text
ZSET
```

设计：

```text
score = post.created_at Unix 时间戳
member = post_id
```

查询关注流时：

```text
ZREVRANGE devflow:feed:inbox:{userID} 0 19
```

### 8.3 热门动态

```text
devflow:hot_posts
```

类型：

```text
ZSET
```

第二阶段再做。

分数可以先简单设计为：

```text
hot_score = like_count * 3 + favorite_count * 5 + comment_count * 4
```

### 8.4 未读通知数

```text
devflow:notification:unread:{userID}
```

类型：

```text
String / Counter
```

用途：

- 前端快速展示未读通知数量。

## 9. RabbitMQ 设计，第三阶段接入

第一版同步写 MySQL。第三阶段再接 RabbitMQ。

### 9.1 交换机

```text
devflow.events
```

类型：

```text
topic
```

### 9.2 事件类型

```text
post.published
interaction.liked
interaction.commented
user.followed
```

### 9.3 队列

```text
devflow.feed.distribute
devflow.notification.create
```

### 9.4 异步 Feed 分发

```text
用户发布动态
  -> MySQL 写 posts
  -> 发布 post.published 事件
  -> feed worker 消费
  -> 查询作者粉丝列表
  -> 写入粉丝 Redis Feed 收件箱
```

### 9.5 异步通知

```text
用户点赞 / 评论 / 关注
  -> 主业务写入成功
  -> 发布 interaction / follow 事件
  -> notification worker 消费
  -> 写入 notifications 表
  -> 更新未读数
```

## 10. 幂等和一致性

第一版重点做好数据库层幂等。

### 10.1 关注幂等

```text
UNIQUE KEY uk_follows_pair (follower_id, followee_id)
```

重复关注时返回成功或明确提示已关注。

### 10.2 点赞幂等

```text
UNIQUE KEY uk_likes_user_post (user_id, post_id)
```

只有第一次点赞时增加 `like_count`。

### 10.3 收藏幂等

```text
UNIQUE KEY uk_favorites_user_post (user_id, post_id)
```

只有第一次收藏时增加 `favorite_count`。

### 10.4 计数更新

点赞、收藏、评论数量更新要和记录写入放在同一个事务里。

例如点赞：

```text
事务开始
  -> 插入 likes
  -> posts.like_count + 1
事务提交
```

如果插入 likes 因唯一索引失败，则不增加计数。

## 11. 后续亮点规划

这些不放入第一版，后续逐个加。

### 11.1 Redis 热门流

使用 Redis ZSET 维护热门动态。

```text
点赞 +3
收藏 +5
评论 +4
```

查询热门流时直接从 ZSET 取 TopN。

### 11.2 新用户冷启动

如果用户没有关注任何人，关注流返回：

```text
全站最新动态
热门作者
热门标签
```

第一版先用 `latest` 兜底。

### 11.3 标签热榜

使用 Redis ZSET 维护热门标签。

```text
devflow:hot_tags
```

发布动态时给标签加分。

### 11.4 通知聚合

把多条相同类型通知聚合成：

```text
3 人点赞了你的动态
5 人评论了你的动态
```

### 11.5 图片上传

先预留 `avatar_url`。动态图片第一版不做。

后续可以接：

```text
本地 uploads
MinIO
阿里云 OSS
```

## 12. 项目目录规划

建议目录：

```text
devflow/
  cmd/
    server/
      main.go
  internal/
    config/
    db/
    middleware/
    model/
    repository/
    service/
    handler/
    response/
    auth/
  migrations/
  docs/
    design.md
  web/
  docker-compose.yml
  README.md
```

第一版可以不强行做复杂分层，但建议保持：

```text
handler 负责 HTTP 参数和响应
service 负责业务事务
repository 负责数据库访问
model 负责 GORM 模型
```

## 13. 开发阶段

### 第一阶段：MySQL MVP

目标：业务闭环跑通。

- 初始化 Gin 项目
- 接入 MySQL + GORM
- 用户注册登录 + JWT
- 发布动态
- 关注 / 取关
- 关注流 MySQL 查询
- 点赞 / 取消点赞
- 评论
- 通知列表

### 第二阶段：Redis 优化

目标：增强 Feed 和互动性能。

- 关注关系缓存
- Feed 收件箱缓存
- 未读通知数缓存
- 热门动态 ZSET

### 第三阶段：RabbitMQ 异步化

目标：把非核心链路异步处理。

- 发布动态后异步分发 Feed
- 点赞 / 评论 / 关注后异步生成通知
- Worker 消费失败日志
- 后续再考虑重试和死信队列

### 第四阶段：前端

目标：可演示。

- 登录页
- 首页 Feed
- 发布动态弹窗
- 动态详情页
- 个人主页
- 通知中心

### 第五阶段：部署

目标：线上可访问。

- Dockerfile
- docker-compose.yml
- MySQL / Redis / RabbitMQ
- 云服务器部署
- README 写清启动方式

## 14. 面试讲法

项目一句话：

> DevFlow 是一个面向开发者的技术动态 Feed 平台，支持发布动态、关注作者、浏览关注流、点赞评论和通知。第一版用 MySQL 跑通核心业务，后续通过 Redis Feed 收件箱和 RabbitMQ 异步分发优化 Feed 查询和通知链路。

主链路讲法：

```text
发布动态
  -> JWT 鉴权
  -> 参数校验
  -> MySQL 写 posts
  -> 后续通过 MQ 异步分发到粉丝 Feed

查询关注流
  -> 查关注关系
  -> 查关注作者的动态
  -> 游标分页返回
  -> 后续优化为 Redis Feed 收件箱

点赞评论
  -> 唯一索引保证幂等
  -> 事务内更新计数
  -> 写入通知
  -> 后续通过 MQ 异步生成通知
```

设计取舍：

> 第一版没有一开始引入复杂推荐算法和图片上传，而是先把 Feed 项目的核心业务闭环做稳。Redis 和 RabbitMQ 会作为第二、第三阶段演进点，用来解决关注流查询性能、互动计数和通知异步化问题。
