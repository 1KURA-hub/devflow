# DevFlow Playwright UI 自动化

这是 DevFlow 的浏览器自动化子项目，使用 TypeScript strict 和 Playwright Test。
完整接口回归由 `tests/api/` 下的 pytest 负责；这里的 API Client 只用于创建、校验和清理 UI 测试数据。

## 项目结构

```text
tests/e2e/
├── playwright.config.ts
├── playwright.midscene.config.ts  # AI用例的独立单worker配置
├── .env.example                   # Midscene模型环境变量示例
├── ai/
│   └── social-journey.spec.ts     # 自然语言跨用户互动闭环
├── src/
│   ├── api/devflow-api.ts          # UI测试需要的最小API Client
│   ├── components/                 # 页面内可复用组件
│   │   ├── app-shell.ts
│   │   ├── composer.modal.ts
│   │   └── post-card.ts
│   ├── pages/                      # Page Object
│   ├── fixtures/test.ts            # 自定义fixture和storageState
│   └── support/
│       ├── data.ts                 # 唯一用户名和标题
│       └── test-data.ts            # 数据创建、记录与清理
└── tests/
    ├── auth.spec.ts
    ├── post-flow.spec.ts
    ├── social-flow.spec.ts
    └── network.spec.ts
```

## 13条测试覆盖

| 文件 | 数量 | 重点 |
|---|---:|---|
| `auth.spec.ts` | 3 | 未登录跳转、登录保持、错误密码 |
| `post-flow.spec.ts` | 4 | 发布、评论、点赞收藏、取消收藏 |
| `social-flow.spec.ts` | 2 | 关注流、点赞通知 |
| `network.spec.ts` | 4 | HTTP 500、断网、401清状态、乐观更新回滚 |

默认回归只保留能代表真实能力的主流程和异常场景，没有固定sleep，也没有为了目录数量增加空封装。
此外保留1条需要模型的Midscene AI场景，它与13条默认回归分开运行。

## fixture如何工作

未登录测试使用：

```ts
import { unauthenticatedTest } from "../src/fixtures/test";
```

已登录测试使用：

```ts
import { test } from "../src/fixtures/test";
```

每条已登录测试的执行链是：

```text
检查目标APP_ENV=test
→ API创建唯一用户
→ storageState把token写入新context
→ 测试获得独立page和登录身份
→ 测试结束后清理帖子、点赞、收藏和关注
```

这里没有让多个测试共用同一个Context，也没有共用固定用户。`storageState`只是给每条测试的新Context提供初始登录状态。

## 本地运行

要求：Go、Docker、Node.js 22.12+。

```bash
docker compose up -d --wait mysql redis
npm ci --prefix web
npm ci --prefix tests/e2e
cd tests/e2e
npx playwright install chromium
npm test
```

`npm test`会先执行：

```bash
tsc --noEmit
```

类型检查通过后才运行Playwright。

常用命令：

```bash
npm run test:ui
npm run test:network
npm run test:headed
npm run test:debug
npm run test:list
npm run test:report
```

## Midscene AI复杂场景

`ai/social-journey.spec.ts`用三段自然语言执行一条跨用户业务旅程：

```text
访客关注作者
→ 在关注流打开指定动态
→ 点赞、收藏、评论
→ API校验关注与互动结果
→ 切换作者身份
→ 标记评论通知已读并打开关联动态
```

这不是将现有Page Object翻译成中文。Midscene负责观察页面并规划跨页面操作；
`TestData`和API Client仍负责数据准备、真实结果断言与清理，避免“模型说成功就算通过”。

### 配置和运行

```bash
cd tests/e2e
cp .env.example .env
# 编辑.env，填写MIDSCENE_MODEL_BASE_URL/API_KEY/NAME/FAMILY

npm run test:ai
# 或者查看浏览器操作
npm run test:ai:headed
```

Midscene会将规划、截图、操作和AI断言写入`midscene_run/report/`下的单HTML报告。
模型Key只保存在本地`.env`或CI secret中，不得提交到仓库。

### 为什么不默认运行

- `npm test`仍只执行13条确定性Playwright回归，不创建Midscene Agent。
- AI场景需要外部多模态模型，会受网络、延迟、模型版本和费用影响。
- 它用于展示自然语言复杂旅程，不替换默认质量门禁。

## 稳定性设计

- 语义定位优先：`getByRole`、`getByLabel`、`getByTestId`；
- 每条测试使用唯一用户和唯一动态标题；
- 页面操作放在Page Object/Component Object，核心断言留在spec；
- 异步结果等待明确响应或使用Web-first断言；
- `page.route()`只用于验证前端面对500、断网和401时的表现；
- 失败时保留Trace、截图和录像，同时生成HTML与Allure报告。

## Allure 测试报告

`playwright.config.ts` 已配置 `allure-playwright` reporter，每次运行都会在
`tests/e2e/allure-results/` 生成 Allure 原始数据（失败用例自动附带截图与Trace链接）。
查看报告：

```bash
npm run allure:serve   # 本地起服务并自动打开浏览器
npm run allure:report  # 生成静态 HTML（allure-report/）并打开
```

Allure CLI 需要 Java 17+；本目录的 `allure-commandline` devDependency 提供 `allure` 命令，
无需全局安装。CI 中由 workflow 生成报告并作为 artifact 上传。

## 环境安全

默认由Playwright启动本地Go API和Vite。外部环境必须显式设置：

```bash
DEVFLOW_E2E_EXTERNAL=true
DEVFLOW_E2E_ALLOW_MUTATION=true
DEVFLOW_API_URL=https://api-test.example.com
DEVFLOW_WEB_URL=https://web-test.example.com
```

fixture还会检查健康接口中的`APP_ENV=test`，拒绝向其他环境写数据。

## 当前边界

- 首版只运行Chromium；
- Midscene场景仅在提供兼容多模态模型配置时按需运行，不进入默认CI；
- 后端没有删除测试用户的接口，CI使用临时数据库，任务结束即销毁；
- 本项目不重复维护接口参数化、幂等和并发回归，这些由pytest负责；
- RabbitMQ可靠性属于后端集成测试，不通过`page.route()`伪装验证。

确定性回归推荐阅读：`tests/post-flow.spec.ts` → `pages/feed.page.ts` → `fixtures/test.ts` →
`support/test-data.ts` → `playwright.config.ts`。AI场景推荐阅读：`ai/social-journey.spec.ts` →
`fixtures/midscene-test.ts` → `playwright.midscene.config.ts`。
