# DevFlow

DevFlow is a developer-focused technical activity feed platform.

Current stage:

- MySQL MVP backend is implemented
- auth register/login/me APIs
- posts, latest feed, hot feed, following feed with cold-start fallback
- follow/unfollow, likes, favorites, comments, notifications
- Redis-backed unread notification count cache
- Redis-backed following relation cache
- Redis-backed following feed inbox

## Run

```bash
cp .env.example .env
docker compose up -d mysql redis rabbitmq
go run ./cmd/server
```

The server exposes:

```text
GET /healthz
GET /api/healthz
POST /api/auth/register
POST /api/auth/login
GET /api/me
POST /api/posts
GET /api/posts/:id
GET /api/users/:id/posts
POST /api/users/:id/follow
DELETE /api/users/:id/follow
GET /api/users/:id/following
GET /api/users/:id/followers
GET /api/feed/latest
GET /api/feed/hot
GET /api/feed/following
POST /api/posts/:id/like
DELETE /api/posts/:id/like
POST /api/posts/:id/favorite
DELETE /api/posts/:id/favorite
POST /api/posts/:id/comments
GET /api/posts/:id/comments
GET /api/me/favorites
GET /api/notifications
GET /api/notifications/unread-count
POST /api/notifications/:id/read
POST /api/notifications/read-all
```
