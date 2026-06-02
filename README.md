# DevFlow

DevFlow 是一个面向开发者社区的动态平台后端，包含注册登录、发帖、关注、点赞/收藏/评论、通知中心，以及三种 Feed（最新、热门、关注流）。

## 技术栈

- 后端：Go、Gin、GORM
- 数据库：MySQL 8
- 缓存：Redis 7
- 消息队列：RabbitMQ 3
- 前端：React + Vite（`web/`）
- 部署：Docker / Docker Compose / GitHub Actions

## 核心能力

- 用户：注册、登录、获取/更新个人信息
- 动态：发布、详情、删除、用户动态列表
- 关系：关注/取关、关注状态、关注/粉丝列表
- 互动：点赞、收藏、评论
- Feed：
  - `latest`：按时间倒序
  - `hot`：按互动热度排序（Redis 优先，MySQL 回退）
  - `following`：基于 inbox 的关注流（支持冷启动降级）
- 通知：异步通知落库、未读数缓存、已读/全部已读

## 项目结构

```text
devflow/
├── cmd/
│   ├── server/     # HTTP 服务入口
│   ├── worker/     # MQ 消费者入口（拆分部署时使用）
│   └── seed/       # 演示数据初始化
├── internal/
│   ├── handler/    # 路由与 HTTP 处理
│   ├── service/    # 业务逻辑
│   ├── repository/ # DB 访问
│   ├── cache/      # Redis 访问封装
│   ├── mq/         # RabbitMQ 封装
│   └── worker/     # 消费逻辑
├── migrations/     # SQL 初始化脚本
├── tests/api/      # pytest 接口自动化
└── web/            # 前端
```

## 快速启动（本地开发）

### 1) 准备环境变量

```bash
cp .env.example .env
```

`.env.example` 关键配置：

- `HTTP_ADDR`：后端监听地址（默认 `:8080`）
- `MYSQL_DSN`：MySQL 连接串（默认连本机 `3307`）
- `REDIS_ADDR`：Redis 地址
- `RABBITMQ_URL`：RabbitMQ 连接串
- `JWT_SECRET`：JWT 密钥
- `DISABLE_WORKERS`：是否禁用 server 进程内 worker（默认 `false`）

### 2) 启动依赖

```bash
docker compose up -d mysql redis rabbitmq
```

### 3) 启动后端

```bash
go run ./cmd/server
```

默认端口：`http://127.0.0.1:8080`

### 4) 启动前端（可选）

```bash
cd web
npm install
npm run dev
```

## 运行模式

### 模式 A：单体模式（默认）

- 只启动 `cmd/server`
- server 进程内同时运行 HTTP + MQ 消费者
- 配置：`DISABLE_WORKERS=false`

### 模式 B：拆分 worker 模式

- `cmd/server` 仅提供 HTTP
- `cmd/worker` 独立消费 MQ
- 配置：`DISABLE_WORKERS=true`

本地示例：

```bash
# 终端 1：HTTP 服务
DISABLE_WORKERS=true go run ./cmd/server

# 终端 2：MQ 消费者
go run ./cmd/worker
```

## 生产部署（Docker Compose）

```bash
cp .env.prod.example .env.prod

docker compose --env-file .env.prod -f docker-compose.prod.yml up -d --build mysql redis rabbitmq app web
```

初始化演示数据：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml run --rm seed
```

若启用拆分 worker（`DISABLE_WORKERS=true`），可加 profile 启动 worker 服务：

```bash
docker compose --env-file .env.prod -f docker-compose.prod.yml --profile split up -d worker
```

## API 自动化测试（pytest）

测试目录：`tests/api/`

```bash
cd tests/api
pip install -r requirements.txt
export DEVFLOW_BASE_URL=http://127.0.0.1:8080
pytest -v
```

支持按 marker 运行：

```bash
pytest -m smoke
pytest -m idempotent
pytest -m cross
```

## CI / CD

- `.github/workflows/ci-cd.yml`：Go + Web 构建与部署流程
- `.github/workflows/api-tests.yml`：接口自动化回归（启动依赖并跑 pytest）

## 主要接口（摘要）

### 健康检查

- `GET /healthz`
- `GET /api/healthz`

### 认证与用户

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/me`
- `PATCH /api/me`
- `GET /api/recommended-users`

### 动态与 Feed

- `POST /api/posts`
- `GET /api/posts/:id`
- `DELETE /api/posts/:id`
- `GET /api/feed/latest`
- `GET /api/feed/hot`
- `GET /api/feed/following`

### 关系与互动

- `POST /api/users/:id/follow`
- `DELETE /api/users/:id/follow`
- `GET /api/users/:id/follow-state`
- `POST /api/posts/:id/like`
- `DELETE /api/posts/:id/like`
- `POST /api/posts/:id/favorite`
- `DELETE /api/posts/:id/favorite`
- `POST /api/posts/:id/comments`
- `GET /api/posts/:id/comments`

### 通知

- `GET /api/notifications`
- `GET /api/notifications/unread-count`
- `POST /api/notifications/:id/read`
- `POST /api/notifications/read-all`

## 常见问题

### 1) 为什么看不到通知？

- 检查 `RABBITMQ_URL` 是否配置正确
- 若未启 worker 拆分模式，确认 `cmd/server` 日志里消费者已启动
- 若启用了拆分模式，确认 `cmd/worker` 正在运行

### 2) 热门榜为什么会短暂不准？

- 热门榜优先读 Redis
- 项目内有定时重建（从 MySQL topN 回灌），用于修复缓存丢失后“半截榜”

### 3) 重复关注为什么返回成功？

- 关注接口按幂等语义处理，重复关注返回 200，便于前端重试与状态收敛
