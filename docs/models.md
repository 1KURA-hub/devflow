# DevFlow Model 草案

这份文档先用于确认字段名、表名和索引。确认后再生成正式的 GORM model 和 migration。

## 1. 公共约定

### 1.1 ID 类型

建议第一版统一使用：

```go
uint64
```

原因：

- 用户、动态、评论、通知都属于业务增长数据，`uint64` 比 `uint` 更明确。
- 数据库对应 `BIGINT UNSIGNED`。
- 后续接口 JSON 返回时也更直观。

### 1.2 时间字段

建议所有主业务表统一：

```go
CreatedAt time.Time
UpdatedAt time.Time
DeletedAt gorm.DeletedAt
```

关系表如果不需要软删，可以只保留：

```go
CreatedAt time.Time
```

比如：

- follows
- likes
- favorites

### 1.3 状态字段

状态字段统一用：

```go
Status int8
```

常用约定：

```text
1 = 正常
2 = 禁用 / 隐藏
```

不建议第一版做复杂状态枚举。

## 2. User

用户表。

```go
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_users_username" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string         `gorm:"type:varchar(64);not null" json:"nickname"`
	Bio          string         `gorm:"type:varchar(255);not null;default:''" json:"bio"`
	AvatarURL    string         `gorm:"type:varchar(512);not null;default:''" json:"avatar_url"`
	Status       int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
```

建议表名：

```text
users
```

字段说明：

- `Username`：登录账号，唯一。
- `PasswordHash`：密码哈希，不返回给前端。
- `Nickname`：展示名称。
- `Bio`：个人简介。
- `AvatarURL`：头像地址，第一版可为空。
- `Status`：用户状态。

待确认：

- 是否要把 `Username` 改成 `Account`？
- 是否要增加 `Email`？
- 是否要保留 `Bio` 和 `AvatarURL`？

## 3. Post

技术动态表。

```go
type Post struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	AuthorID      uint64         `gorm:"not null;index:idx_posts_author_created,priority:1" json:"author_id"`
	Title         string         `gorm:"type:varchar(120);not null" json:"title"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	Tags          string         `gorm:"type:varchar(255);not null;default:''" json:"tags"`
	LikeCount     int64          `gorm:"not null;default:0" json:"like_count"`
	CommentCount  int64          `gorm:"not null;default:0" json:"comment_count"`
	FavoriteCount int64          `gorm:"not null;default:0" json:"favorite_count"`
	Status        int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt     time.Time      `gorm:"index:idx_posts_author_created,priority:2;index:idx_posts_created_at" json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
```

建议表名：

```text
posts
```

字段说明：

- `AuthorID`：动态作者。
- `Title`：动态标题，短标题即可。
- `Content`：动态正文。
- `Tags`：第一版用逗号分隔，例如 `Go,Redis,Docker`。
- `LikeCount`：点赞数。
- `CommentCount`：评论数。
- `FavoriteCount`：收藏数。
- `Status`：动态状态，后续可用于隐藏、删除、审核。

索引：

```text
idx_posts_author_created(author_id, created_at)
idx_posts_created_at(created_at)
```

待确认：

- 技术动态是否需要 `Title`？如果更像微博，可以去掉，只保留 `Content`。
- `Tags` 是先用字符串，还是一开始拆 `post_tags`？
- 是否预留 `CoverURL` / `ImageURLs`？第一版建议不加。

## 4. Follow

关注关系表。

```go
type Follow struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FollowerID uint64    `gorm:"not null;uniqueIndex:uk_follows_pair,priority:1;index:idx_follows_follower" json:"follower_id"`
	FolloweeID uint64    `gorm:"not null;uniqueIndex:uk_follows_pair,priority:2;index:idx_follows_followee" json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}
```

建议表名：

```text
follows
```

字段说明：

- `FollowerID`：关注别人的用户。
- `FolloweeID`：被关注的用户。

索引：

```text
uk_follows_pair(follower_id, followee_id)
idx_follows_follower(follower_id)
idx_follows_followee(followee_id)
```

待确认：

- 字段名是否使用 `FollowerID / FolloweeID`，还是改成 `UserID / TargetUserID`？

## 5. Like

点赞表。

```go
type Like struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_likes_user_post,priority:1;index:idx_likes_user" json:"user_id"`
	PostID    uint64    `gorm:"not null;uniqueIndex:uk_likes_user_post,priority:2;index:idx_likes_post" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
```

建议表名：

```text
likes
```

字段说明：

- 一个用户对一条动态只能点赞一次。
- 唯一索引用于幂等控制。

索引：

```text
uk_likes_user_post(user_id, post_id)
idx_likes_user(user_id)
idx_likes_post(post_id)
```

## 6. Favorite

收藏表。

```go
type Favorite struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_favorites_user_post,priority:1;index:idx_favorites_user" json:"user_id"`
	PostID    uint64    `gorm:"not null;uniqueIndex:uk_favorites_user_post,priority:2;index:idx_favorites_post" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
```

建议表名：

```text
favorites
```

字段说明：

- 收藏和点赞分开存。
- 第一版可以先建表，接口后做。

索引：

```text
uk_favorites_user_post(user_id, post_id)
idx_favorites_user(user_id)
idx_favorites_post(post_id)
```

## 7. Comment

评论表。

```go
type Comment struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    uint64         `gorm:"not null;index:idx_comments_post_created,priority:1" json:"post_id"`
	UserID    uint64         `gorm:"not null;index:idx_comments_user_created,priority:1" json:"user_id"`
	Content   string         `gorm:"type:varchar(1000);not null" json:"content"`
	Status    int8           `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time      `gorm:"index:idx_comments_post_created,priority:2;index:idx_comments_user_created,priority:2" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

建议表名：

```text
comments
```

字段说明：

- 第一版只做一级评论。
- 不加 `ParentID`，避免楼中楼复杂度。

索引：

```text
idx_comments_post_created(post_id, created_at)
idx_comments_user_created(user_id, created_at)
```

待确认：

- 是否提前预留 `ParentID`？第一版建议不加。

## 8. Notification

通知表。

```go
type Notification struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index:idx_notifications_user_read_created,priority:1;index:idx_notifications_user_created,priority:1" json:"user_id"`
	ActorID   uint64    `gorm:"not null" json:"actor_id"`
	Type      string    `gorm:"type:varchar(32);not null" json:"type"`
	PostID    *uint64   `gorm:"index" json:"post_id,omitempty"`
	CommentID *uint64   `gorm:"index" json:"comment_id,omitempty"`
	Content   string    `gorm:"type:varchar(255);not null" json:"content"`
	IsRead    bool      `gorm:"not null;default:false;index:idx_notifications_user_read_created,priority:2" json:"is_read"`
	CreatedAt time.Time `gorm:"index:idx_notifications_user_read_created,priority:3;index:idx_notifications_user_created,priority:2" json:"created_at"`
}
```

建议表名：

```text
notifications
```

字段说明：

- `UserID`：接收通知的人。
- `ActorID`：触发通知的人。
- `Type`：通知类型。
- `PostID`：关联动态，可为空。
- `CommentID`：关联评论，可为空。
- `Content`：通知文案快照。
- `IsRead`：是否已读。

通知类型：

```text
follow
like
comment
favorite
```

索引：

```text
idx_notifications_user_read_created(user_id, is_read, created_at)
idx_notifications_user_created(user_id, created_at)
```

待确认：

- `Type` 是否改成 `EventType`？
- `Content` 是保存文案快照，还是前端根据 type 动态拼？

## 9. 可选：FeedItem

第一版如果关注流直接查 MySQL，可以先不建 `feed_items`。

第二阶段如果做 Feed 收件箱，可以增加：

```go
type FeedItem struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;uniqueIndex:uk_feed_user_post,priority:1;index:idx_feed_user_created,priority:1" json:"user_id"`
	PostID    uint64    `gorm:"not null;uniqueIndex:uk_feed_user_post,priority:2" json:"post_id"`
	AuthorID  uint64    `gorm:"not null" json:"author_id"`
	CreatedAt time.Time `gorm:"index:idx_feed_user_created,priority:2" json:"created_at"`
}
```

建议表名：

```text
feed_items
```

作用：

- 作为数据库版 Feed 收件箱。
- 后续 Redis Feed 收件箱可以从这里恢复。

第一版建议：

```text
暂不实现 FeedItem。
```

原因：

- MySQL 关注流可以直接通过 follows + posts 查询。
- 先减少写扩散逻辑。
- 等接 RabbitMQ 分发 Feed 时再加更自然。

## 10. 第一版最终建议保留的 model

第一版直接实现：

```text
User
Post
Follow
Like
Comment
Notification
```

可以先建但不实现完整接口：

```text
Favorite
```

暂不实现：

```text
FeedItem
```

## 11. 字段命名待你确认

重点看这些：

```text
User.Username
User.Nickname
User.Bio
User.AvatarURL

Post.AuthorID
Post.Title
Post.Content
Post.Tags
Post.LikeCount
Post.CommentCount
Post.FavoriteCount

Follow.FollowerID
Follow.FolloweeID

Notification.ActorID
Notification.Type
Notification.Content
Notification.IsRead
```

确认后再生成：

```text
internal/model/*.go
migrations/001_init.sql
```
