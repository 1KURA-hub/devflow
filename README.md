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
- RabbitMQ-backed async notification and feed workers
- React frontend for feeds, profiles, favorites, notifications, and publishing

## Run

```bash
cp .env.example .env
docker compose up -d mysql redis rabbitmq
go run ./cmd/server
```

Frontend:

```bash
cd web
npm install
npm run dev
```

## Production deployment

```bash
cp .env.prod.example .env.prod
docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build mysql redis rabbitmq app web
docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
```

The production web container serves the React build and proxies `/api/*` to the Go app.
The demo seed command creates reusable showcase data, including users, avatar URLs, posts, follow relations, likes, favorites, comments, and notifications.

## GitHub Actions CI/CD

`.github/workflows/ci-cd.yml` runs Go tests and the React production build on pull requests and pushes.
Pushes to `main` deploy to the production server over SSH.

Required repository secrets:

```text
DEPLOY_HOST=43.136.63.219
DEPLOY_PORT=22
DEPLOY_USER=root
DEPLOY_PATH=/go-course/devflow
DEPLOY_SSH_KEY=<private key with access to the server>
```

The deploy job uploads a git bundle to the server, updates the working tree from that bundle, rebuilds `app`, `web`, and `seed`, runs seed with `AUTO_MIGRATE=true`, then restarts `app` and `web`.
The server must already contain `.env.prod`; it is not stored in GitHub secrets or committed to the repository.

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
