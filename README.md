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
```
