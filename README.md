# DevFlow

DevFlow is a developer-focused technical activity feed platform.

Current stage:

- project design docs
- GORM models
- initial MySQL migration
- minimal Gin server skeleton
- auth register/login/me APIs

## Run

```bash
cp .env.example .env
docker compose up -d mysql
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
GET /api/feed/following
POST /api/posts/:id/like
DELETE /api/posts/:id/like
POST /api/posts/:id/favorite
DELETE /api/posts/:id/favorite
POST /api/posts/:id/comments
GET /api/posts/:id/comments
GET /api/me/favorites
```
