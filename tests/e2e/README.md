# DevFlow Playwright + TypeScript 自动化

这套工程是 DevFlow 的浏览器端自动化层，不替代现有的 `pytest + requests` 接口回归。

## 为什么分成两套测试

```text
pytest + requests
  └─ 深度接口回归：边界、参数化、幂等、并发和跨链路

Playwright APIRequestContext
  ├─ 核心接口冒烟
  └─ 为 UI 用例创建独立用户、动态和关系

Playwright Page
  ├─ 真实浏览器业务流程
  ├─ 页面与接口交叉校验
  └─ 500、网络异常和乐观更新回滚
```

接口层已有大量覆盖时，不应该换一种语言把全部用例再抄一遍。Playwright API 层只保留核心契约，并承担 UI 测试的数据准备。

## 目录结构

```text
tests/e2e/
├── playwright.config.ts        # 浏览器、报告、服务启动和失败证据
├── src/
│   ├── api/
│   │   ├── devflow-api.ts      # 低层 API Client，保留状态码和响应体
│   │   └── data-factory.ts     # 独立测试数据和逆序清理
│   ├── config/env.ts           # API/UI 地址
│   ├── fixtures/test.ts        # apiTest、登录用户和 storageState
│   ├── pages/                  # Page Object
│   └── support/data.ts         # 唯一用户名和动态标题
└── tests/
    ├── api/                    # 8 条核心 API 用例
    ├── ui/                     # 8 条 UI/跨用户主流程
    └── network/                # 3 条接口异常与回滚用例
```

## 本地运行

要求：Go、Docker、Node.js 22.12+。

```bash
# 1. 启动数据库和缓存
docker compose up -d --wait mysql redis

# 2. 安装前端和测试依赖
npm ci --prefix web
npm ci --prefix tests/e2e

# 3. 首次安装浏览器
cd tests/e2e
npx playwright install chromium

# 4. 执行全部测试
npm test
```

Playwright 会通过 `webServer` 自动启动：

- Go API：`http://127.0.0.1:8080`
- Vite UI：`http://127.0.0.1:5173`

Go 服务不会自动读取仓库根目录的 `.env`，所以 E2E 配置会显式传递测试环境变量，并启用 `AUTO_MIGRATE=true`。

## 常用命令

在 `tests/e2e` 下执行：

```bash
npm run test:api       # 只执行 Playwright API 冒烟
npm run test:ui        # 只执行 UI 主流程
npm run test:network   # 只执行网络异常
npm run test:headed    # 显示浏览器
npm run test:debug     # Inspector 单步调试
npm run typecheck      # TypeScript 类型检查
npm run test:report    # 打开 HTML 报告
```

## 指定其他测试环境

默认模式只连接本机临时依赖，并由 Playwright 启动 Go API 和 Vite。若明确要测试一个已经部署好的隔离环境，可设置：

```bash
DEVFLOW_E2E_EXTERNAL=true \
DEVFLOW_API_URL=http://127.0.0.1:8080 \
DEVFLOW_WEB_URL=http://127.0.0.1:5173 \
npm test
```

外部模式不会自动启动服务，而且用例会注册唯一用户并创建动态、评论和通知；不要对生产环境执行。若要在本地显式测试 RabbitMQ 链路，使用 `DEVFLOW_E2E_RABBITMQ_URL`，默认留空以验证同步降级路径。

## 两种 test 的区别

```ts
import { apiTest } from "../src/fixtures/test";
```

`apiTest` 不自动登录，适合：

- API测试；
- 登录、注册页面；
- 未登录重定向。

```ts
import { test } from "../src/fixtures/test";
```

`test` 会：

1. 使用 API 创建唯一用户；
2. 将 token 写入浏览器 `storageState`；
3. 让 `page` 打开后直接处于登录态；
4. 测试结束后执行可用的清理动作。

fixture 是按需创建的。测试函数没有声明 `actor`、`data` 或 `page` 时，不会额外制造对应数据。

## 推荐阅读顺序

第一次学习不要从配置文件逐行看，按业务链路阅读：

1. `tests/ui/post-flow.spec.ts`：先看一个完整测试想验证什么；
2. `src/pages/feed.page.ts`：看页面操作如何封装；
3. `src/fixtures/test.ts`：看用户和登录态从哪里来；
4. `src/api/data-factory.ts`：看数据如何创建和清理；
5. `src/api/devflow-api.ts`：最后看请求封装和类型；
6. `playwright.config.ts`：理解服务、浏览器和报告如何组成工程。

## 三个稳定性原则

### 1. 不使用固定账号和固定动态

每条用例创建唯一用户名和标题，避免本地重复运行、并行执行和 CI 之间相互污染。

### 2. 异步结果使用条件轮询

通知按照 `type + actor_id + post_id` 精确轮询，不读取列表第一条，也不写固定 `waitForTimeout(3000)`。

### 3. 乐观更新必须等待接口

点赞和收藏会先更新页面、再请求后端。只看到按钮变色可能是假通过，因此 Page Object 会同时等待目标接口并检查状态码。

## Mock 与真实故障的边界

`page.route()` 返回 500，验证的是“前端收到500后如何表现”，不是“后端能否抵抗数据库或Redis故障”。

```text
page.route() mock 500/断网
  → 前端异常处理测试

停止 MySQL/Redis/RabbitMQ
  → 后端容错与降级测试
```

两类测试不能混为一谈。

## 当前边界

- 后端没有测试用户删除接口，因此本地会保留唯一测试用户；CI数据库是临时的，任务结束自动销毁。
- 帖子、点赞、收藏和关注会尽量逆序清理；评论和通知目前没有删除接口。
- 首版设置 `RABBITMQ_URL=""`，通知与 Feed 使用应用现有的同步降级路径。这验证了业务结果，但不能声称验证了 RabbitMQ 链路。
- 首版只跑 Chromium。移动端、Firefox/WebKit 和真实 MQ 故障注入属于下一阶段。

## 学习练习

理解现有代码后，建议你自己完成：

1. 为收藏页增加一条 UI 用例；
2. 使用 `route.abort()` 增加真正的网络中断用例；
3. 给通知用例增加“全部已读”断言；
4. 将一条固定数据测试改写为数据驱动测试；
5. 解释为什么 CI 使用一个 worker，而不是先追求全并行。
