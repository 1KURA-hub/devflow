# DevFlow 测试开发作品集

DevFlow 是一个可本地运行的开发者社区应用。本仓库以同一套真实业务系统为被测对象，展示
`pytest + requests` 接口自动化与 `TypeScript + Playwright` UI 自动化如何分层协作，并用
`Midscene` 补充一条按需运行的自然语言复杂场景。重点覆盖测试隔离、异步链路、故障注入、
失败证据和持续集成。

## 项目目标

- 用接口测试深挖参数边界、权限、幂等性、并发一致性和跨组件业务契约。
- 用浏览器测试验证登录、发帖、互动、关注流和通知等关键用户旅程。
- 让每条测试可以独立、重复运行，并明确区分真实后端故障与前端故障注入。
- 在 CI 中同时执行单元/构建、API 回归和浏览器回归，并保留可定位的测试产物。

## 覆盖矩阵

| 领域 | pytest 接口层 | Playwright UI层 |
| --- | --- | --- |
| 认证 | 注册/登录、参数边界、重复用户名、同名并发注册、无效Token | 登录、错误密码、未登录重定向、刷新后保持状态 |
| 动态 | 创建/查询/删除、权限、字段校验 | 发布、详情、评论、收藏列表 |
| 互动 | 点赞幂等与取消、收藏状态与列表、评论 | 点赞/收藏状态与接口结果交叉校验 |
| 关注与Feed | 关注状态与用户列表、关注自己校验、following Feed | 关注后在关注流看到目标动态 |
| 通知 | 点赞通知、未读数、全部已读、异步通知契约 | 查看、标记已读并打开关联动态 |
| 异常处理 | 真实MySQL/Redis/RabbitMQ协作链路 | HTTP 500、401、断网、乐观更新回滚 |

## 测试架构

```text
GitHub Actions: quality.yml
├── unit-build
│   ├── Go tests
│   └── React/Vite build
├── api-tests
│   └── pytest → domain clients → HTTP API → MySQL / Redis / RabbitMQ
└── e2e
    ├── Playwright specs → Page Objects → Chromium → React UI
    ├── fixtures / TestData → 独立用户与测试数据
    └── route mock → 500 / 401 / 断网 / UI回滚
```

```text
tests/api/
├── clients/          # 按领域封装请求，不写业务断言
├── conftest.py       # 用户、Token、动态等分层fixture
└── tests/            # 接口契约与跨链路断言

tests/e2e/
├── ai/              # 1条按需运行的Midscene自然语言场景
├── src/api/          # UI数据准备所需的最小API Client
├── src/components/   # 可复用页面组件
├── src/pages/        # Page Object
├── src/fixtures/     # 测试用户、登录态和资源生命周期
├── src/support/      # 唯一数据与显式清理
└── tests/            # 4个扁平spec：认证、动态、社交、网络异常
```

## 最短运行方式

### 环境要求

- Go 1.24+
- Python 3.11+
- Node.js 22+
- Docker与Docker Compose

### pytest接口回归

先在项目根目录启动真实依赖和测试环境API：

```bash
docker compose up -d --wait mysql redis rabbitmq

APP_ENV=test \
AUTO_MIGRATE=true \
REDIS_ADDR=127.0.0.1:6379 \
RABBITMQ_URL=amqp://devflow:devflow@127.0.0.1:5672/ \
go run ./cmd/server
```

再开一个终端执行：

```bash
cd tests/api
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements.txt
python -m pytest
```

常用子集：

```bash
python -m pytest -m smoke
python -m pytest -m idempotent
python -m pytest -m cross
```

### Playwright UI回归

Playwright会自动启动Go API和Vite UI，本地只需先启动MySQL与Redis：

```bash
docker compose up -d --wait mysql redis
npm ci --prefix web
npm ci --prefix tests/e2e

cd tests/e2e
npx playwright install chromium
npm test
```

常用命令：

```bash
npm run test:ui
npm run test:network
npm run test:headed
npm run test:debug
npm run test:report
```

### Midscene自然语言复杂场景

该场景由API创建独立访客、作者和动态，Midscene再通过自然语言完成“关注作者 →
在关注流打开动态 → 点赞/收藏/评论 → 作者从通知返回动态”。业务结果由API二次
校验，测试结束后仍由`TestData`清理。

```bash
cd tests/e2e
cp .env.example .env       # 填入兼容的多模态模型配置
npm run test:ai            # 无头模式
npm run test:ai:headed     # 展示模式
```

运行后可在`tests/e2e/midscene_run/report/`查看每个AI规划、操作和断言的可视化回放。
Midscene需要调用模型服务，可能产生费用；它不在默认`npm test`和CI中运行。

更完整的环境切换与调试说明见[接口测试文档](tests/api/README.md)和
[Playwright测试文档](tests/e2e/README.md)。

### 查看测试报告

pytest与Playwright运行后会分别生成 `tests/api/allure-results/` 与
`tests/e2e/allure-results/` 原始数据，用Allure CLI查看：

```bash
cd tests/api && allure serve allure-results   # 或：cd tests/e2e && npm run allure:serve
```

Allure CLI需要Java 17+（macOS：`brew install allure`；`tests/e2e`目录已内置
`allure-commandline`，`npm ci`后可直接用`npx allure`）。CI中的静态Allure报告通过
Actions artifact下载，浏览器打开即可。

第一次阅读代码建议配合[测试框架讲解](docs/testing-walkthrough.md)，按照一条API用例和
一条UI用例的真实执行顺序理解，不需要先背完整目录。

## pytest与Playwright的职责边界

| 层次 | 负责 | 不负责 |
| --- | --- | --- |
| pytest | HTTP状态码、响应结构、边界值、权限、幂等、并发、真实跨组件链路 | 重复模拟浏览器操作 |
| Playwright | 用户可见流程、DOM状态、前后端交叉校验、前端异常反馈 | 重写完整接口回归 |
| Midscene | 用自然语言和视觉理解执行跨页面复杂旅程 | 替代稳定回归或单独证明业务数据正确 |
| Playwright APIRequestContext | 为UI用例快速创建/清理数据 | 作为第二套接口测试框架 |
| `page.route()` | 验证前端收到500、401或网络失败后的表现 | 证明后端能够抵抗真实数据库或MQ故障 |

## 核心设计取舍

### 独立测试数据

接口fixture和Playwright TestData为用例创建唯一用户、标题和关系，避免依赖执行顺序。
Playwright使用按测试创建的登录态，不共享固定账号；可删除资源在teardown中逆序清理。

### 条件等待而非固定休眠

通知与Feed存在异步可见窗口。接口层按精确条件轮询，UI层使用Playwright自动等待和
Web-first断言，不用固定`sleep`掩盖竞态。

### 可写环境安全门禁

测试启动前会检查健康接口中的应用标识和`APP_ENV=test`。对外部地址执行写数据回归时，
还需要显式开启写入授权，降低误连生产或共享环境的风险。

### 失败证据

pytest与Playwright运行后生成Allure测试报告（用例步骤、失败原因、附带的截图/Trace均可追溯）；
Playwright失败时额外保留Trace、截图和录像，CI失败时保留服务日志，便于区分环境、接口、
定位器和断言问题。

### AI用例与稳定回归分开

现有13条Playwright用例作为默认质量门禁，不调用模型。Midscene只运行1条适合AI规划的
跨用户场景，并将确定性的数据准备、结果校验和清理交给Playwright/API基础设施。

## CI

[`.github/workflows/quality.yml`](.github/workflows/quality.yml)在Pull Request、`main`分支推送和
手动触发时运行三个并行任务：

1. `unit-build`：Go单元测试和Web构建；
2. `api-tests`：使用MySQL、Redis、RabbitMQ运行pytest回归；
3. `e2e`：使用MySQL、Redis和Chromium运行TypeScript类型检查及Playwright回归。

每个测试job运行后由workflow下载Allure CLI并生成静态报告；Allure报告、Playwright HTML报告
以及失败Trace、截图和录像均作为Actions artifact上传。
作品仓库只展示质量门禁，不包含生产部署和SSH密钥操作。

## 项目亮点（简历摘要）

- 基于`pytest + requests`实现30项接口测试，以client、fixture、用例三层结构覆盖边界值、权限、幂等、并发注册和异步通知链路。
- 基于TypeScript strict与Playwright实现13项浏览器测试，使用Page Object、组件对象、自定义fixture、独立登录态和TestData覆盖认证、发帖、互动及跨用户流程。
- 使用`route.fulfill()`和`route.abort()`验证500、401、断网与乐观更新回滚，并通过UI状态与接口结果交叉校验降低假通过。
- 集成Midscene实现1条可选AI E2E场景，以自然语言执行关注、互动和通知闭环，并用API交叉断言防止视觉假通过。
- 将Go/Web构建、pytest和Playwright接入三路并行CI，生成Allure测试报告并归档Trace、截图和录像等失败证据。

## 已知边界

- 浏览器回归目前以Chromium为主，CI使用单worker换取共享测试环境下的稳定性。
- Playwright默认让通知与Feed走应用的同步降级路径；RabbitMQ真实链路由API任务覆盖，不能把UI结果表述为MQ故障验证。
- Midscene结果会受模型、网络、延迟和费用影响，因此只按需运行，不作为默认CI质量门禁。
- 后端暂未提供测试用户、评论和通知的完整删除接口；CI使用临时数据库，任务结束后销毁环境，本地长期运行需定期清理测试数据。
- 当前重点是功能、集成和UI自动化，不替代性能、安全、兼容性和混沌测试。
