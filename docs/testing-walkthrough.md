# DevFlow测试框架讲解

这份文档用于帮助你理解并讲清楚这个测试项目。阅读目标不是背目录，而是掌握两条执行链：

~~~text
pytest：
测试函数 → fixture → 领域Client → BaseClient → HTTP接口 → assert

Playwright：
测试函数 → 自定义fixture → storageState/context → Page Object → Locator → expect

Midscene（按需）：
AI专用fixture → 自然语言aiAct/aiAssert → 视觉模型 → 页面操作 → API交叉断言
~~~

项目当前保留：

- 30项pytest接口测试；
- 13项Playwright UI测试；
- 1项与默认回归分开的Midscene AI复杂场景；
- 1个包含3个并行job的质量流水线。

这些内容足够展示测开实习所需要的接口、UI、数据隔离、异常验证和CI能力，同时删除了难以解释的响应竞态闸门和重复场景。

## 1. 推荐阅读顺序

不要从所有文件一起看。先沿着一条用例向下阅读。

### API主线

1. `tests/api/tests/test_post.py`；
2. `tests/api/conftest.py`；
3. `tests/api/clients/post.py`；
4. `tests/api/clients/base.py`。

### UI主线

1. `tests/e2e/tests/post-flow.spec.ts`；
2. `tests/e2e/src/fixtures/test.ts`；
3. `tests/e2e/src/pages/feed.page.ts`；
4. `tests/e2e/src/components/post-card.ts`；
5. `tests/e2e/src/support/test-data.ts`；
6. `tests/e2e/playwright.config.ts`。

### AI场景主线（可选）

1. `tests/e2e/ai/social-journey.spec.ts`；
2. `tests/e2e/src/fixtures/midscene-test.ts`；
3. `tests/e2e/src/support/test-data.ts`；
4. `tests/e2e/playwright.midscene.config.ts`。

## 2. pytest框架怎样执行

以“查询帖子核心字段”为例：

~~~python
def test_get_post_returns_core_fields(published_post, registered_user):
    resp = registered_user.post.get(published_post["id"])

    assert resp.ok
    assert resp.data["id"] == published_post["id"]
~~~

### 第一步：pytest解析函数参数

`published_post`和`registered_user`不是普通变量，它们是`conftest.py`里的fixture。

pytest看到测试函数需要它们后，会先准备依赖：

~~~text
server_ready
→ 确认服务是DevFlow且APP_ENV=test
→ unique_username
→ registered_user
→ published_post
→ 执行测试函数
~~~

### 第二步：fixture创建测试数据

`registered_user`会：

1. 生成唯一用户名；
2. 调用注册接口；
3. 保存用户ID和Token；
4. 创建带Token的领域Client；
5. 测试结束后关闭`requests.Session`。

`published_post`会：

1. 使用当前测试用户创建帖子；
2. 把帖子ID交给测试；
3. 测试结束后尽量删除帖子。

### 第三步：领域Client封装请求

测试写：

~~~python
registered_user.post.get(post_id)
~~~

`PostClient`只负责：

- API路径；
- HTTP方法；
- 请求参数。

它不写业务断言。这样同一个Client可以被成功、权限和异常测试复用。

### 第四步：BaseClient统一处理HTTP

`BaseClient`负责：

- 复用`requests.Session`；
- 添加Bearer Token；
- 拼接base URL；
- 设置超时；
- 把统一响应解析为`ApiResponse`。

测试最终得到：

~~~python
ApiResponse(
    status_code=200,
    code=0,
    message="ok",
    data={...},
)
~~~

### 第五步：断言留在测试层

测试先确认请求成功，再访问`data`：

~~~python
assert resp.ok, f"查询失败: {resp.status_code} {resp.message}"
assert resp.data["id"] == post_id
~~~

这样接口失败时，报告显示真实HTTP状态和错误信息，而不是在`None["id"]`处出现难懂的二次错误。

## 3. 为什么API测试只保留30项

原项目有78项，覆盖很广，但重复边界和多组并发场景增加了学习成本。

当前版本保留四类文件：

| 文件 | 作用 |
|---|---|
| `test_auth.py` | 注册、登录、核心异常和一个并发注册 |
| `test_post.py` | 发帖、详情、权限和删除 |
| `test_interactions.py` | 点赞、收藏、评论和一个幂等场景 |
| `test_social.py` | 关注、following feed和通知 |

设计取舍：

- 参数化仍保留，展示一份逻辑验证多组输入；
- 幂等只保留重复点赞一个代表；
- 并发只保留相同用户名并发注册一个代表；
- 跨模块只保留关注流和点赞通知；
- 不声称“全接口覆盖”。

## 4. Playwright框架怎样执行

以“点赞和收藏后页面与接口一致”为例：

~~~ts
test("点赞和收藏后，页面状态与接口数据一致", async ({
  page,
  actor,
  data,
}) => {
  const author = await data.createActor();
  const ownedPost = await data.createPost(author);

  const detail = new PostDetailPage(page);
  await detail.goto(ownedPost.post.id);

  const card = new FeedPage(page).post(ownedPost.post.id);
  await card.like();
  await card.favorite();

  const post = requireData(
    await actor.api.getPost(ownedPost.post.id),
    "查询互动后的动态",
  );
  expect(post.liked).toBe(true);
  expect(post.favorited).toBe(true);
});
~~~

### 第一步：test不是Playwright原始test

测试导入的是：

~~~ts
import { test, expect } from "../src/fixtures/test";
~~~

这个`test`在Playwright原始fixture上增加了：

- `api`；
- `data`；
- `actor`；
- 动态`storageState`。

### 第二步：fixture准备独立身份

每条已登录测试的顺序：

~~~text
检查/healthz
→ 确认app=devflow、env=test
→ TestData创建唯一actor
→ storageState把actor.token写入localStorage
→ Playwright创建当前测试的BrowserContext
→ Context创建page
→ 执行测试
→ TestData清理资源
→ Playwright销毁Context
~~~

注意：

- 每条测试都有新的Context；
- 每条测试都有唯一用户；
- 多个测试不会共享上一个测试的页面；
- storageState只是给新Context注入初始登录状态，不是再次操作登录表单。

### 第三步：Page Object组织页面

`PostDetailPage`表示帖子详情页面：

~~~ts
await detail.goto(postID);
await detail.addComment(content);
~~~

`FeedPage`表示动态列表页面：

~~~ts
await feed.gotoLatest();
const card = feed.post(postID);
~~~

Page Object保存页面级Locator和操作，不保存测试用例的最终业务结论。

### 第四步：Component Object组织复用区域

帖子卡片可能同时出现在首页、关注流、收藏页和详情页，因此单独使用`PostCard`：

~~~ts
await card.like();
await card.favorite();
~~~

组件对象不是为了增加class数量，而是为了避免在多个页面重复写同一组Locator。

当前只保留三个有明确复用价值的组件：

- `AppShell`：主导航和用户区域；
- `ComposerModal`：发布动态弹窗；
- `PostCard`：帖子卡片互动。

### 第五步：expect判断结果

项目同时使用两类断言：

普通值：

~~~ts
expect(response.status()).toBe(200);
expect(post.liked).toBe(true);
~~~

页面Locator：

~~~ts
await expect(card.root).toBeVisible();
await expect(likeButton).toHaveAttribute("aria-pressed", "true");
~~~

Locator断言会自动等待和重新查询当前DOM；普通`toBe()`不会自动等待页面变化。

## 5. TestData为什么存在

如果所有前置条件都通过UI准备：

~~~text
登录
→ 打开发布框
→ 发帖
→ 退出登录
→ 登录另一个用户
→ 再开始验证点赞
~~~

测试会非常慢，而且准备步骤失败会掩盖真正要验证的点赞功能。

所以`TestData`使用API快速准备：

- 唯一用户；
- 帖子；
- 点赞；
- 收藏。

它不是第二套接口测试。接口是否满足各种边界，由pytest负责；这里的API只服务于UI前置数据和清理。

### 清理为什么使用四个明确集合

`TestData`记录：

~~~text
posts
likes
favorites
follows
~~~

测试结束后按反向顺序清理：

~~~text
取消收藏
→ 取消点赞
→ 取消关注
→ 删除帖子
~~~

显式集合比“任意清理函数队列”更长一点，但更适合当前阶段：

- 一眼能看出管理哪些资源；
- 每一种资源如何清理都明确；
- 容易在面试时画出生命周期；
- 出错时容易定位到具体资源。

后端没有删除测试用户的接口，因此用户数据由CI临时数据库统一销毁。这是项目已知边界，不应假装已经完全清理。

## 6. 为什么有两个test

### `unauthenticatedTest`

用于：

- 登录页面；
- 错误密码；
- 未登录重定向。

它不会自动注入登录Token。

### `test`

用于：

- 发帖；
- 点赞收藏；
- 关注；
- 通知；
- 已登录网络异常。

它会为当前测试创建actor，并在Context创建前提供storageState。

这样登录测试不会错误地从“已经登录”的状态开始，业务测试也不需要每次重复操作登录表单。

## 7. 网络异常如何验证

网络测试保留四个代表场景。

### HTTP 500

~~~ts
await page.route("**/api/feed/latest", async (route) => {
  await route.fulfill({
    status: 500,
    contentType: "application/json",
    body: JSON.stringify({
      code: 500,
      message: "动态列表暂不可用",
    }),
  });
});
~~~

故障注入：浏览器请求被Playwright替换为500响应。
判断标准：页面显示后端错误信息。

### 断网

~~~ts
await route.abort("failed");
~~~

故障注入：浏览器请求在网络层失败。
判断标准：页面显示可理解的网络错误。

### 401

故障注入：当前用户接口返回401。
判断标准：应用清除失效token。

### 乐观更新回滚

故障注入：点赞接口返回500。
判断标准：

- 用户看到错误信息；
- 按钮恢复为未点赞；
- 点赞计数恢复到操作前。

删除的高级竞态测试使用Promise gate、`pushState`和`PopStateEvent`控制迟到响应。它们有价值，但不是实习作品的核心能力，所以不放在这个精简版本中。

## 8. CI怎样执行

`.github/workflows/quality.yml`包含三个并行job：

~~~text
unit-build
├── go test ./...
└── npm run build

api-tests
├── MySQL / Redis / RabbitMQ
├── 启动Go API
├── pytest → allure-results
└── 生成/上传Allure报告和失败日志

e2e
├── MySQL / Redis
├── npm run typecheck
├── Playwright Chromium
└── 生成/上传Allure报告、Playwright HTML、Trace、截图和录像
~~~

为什么UI job不启动RabbitMQ：

- UI回归只验证用户可见流程；
- 默认让后端使用同步降级路径；
- RabbitMQ真实链路由API/Go集成测试负责；
- 避免在E2E中启动但实际没有使用的服务。

## 9. 可以怎样介绍这个项目

### 60秒版本

> 我使用同一个DevFlow业务系统搭建了pytest接口自动化和TypeScript Playwright UI自动化。接口层保留30项代表性测试，通过领域Client、分层fixture和统一响应对象覆盖认证、帖子、互动、关注流及通知，并保留幂等和并发代表场景。UI层保留13项核心回归，通过Page Object、组件对象、自定义fixture和动态storageState实现独立身份，再用TestData通过API准备和清理数据。异常场景使用route.fulfill和route.abort验证500、401、断网及点赞失败回滚。我另外用Midscene增加1条按需运行的AI场景，通过自然语言完成关注、互动和通知闭环，最终用API交叉断言避免视觉假通过。CI只运行确定性回归，避免模型密钥、费用和波动影响质量门禁。

### 不要夸大的内容

不要说：

- 覆盖了所有接口；
- 验证了真实浏览器下的RabbitMQ故障；
- 用户数据能够完全清理；
- 13条UI测试支持所有浏览器；
- Midscene用例可以替代现有确定性回归；
- Midscene在任何模型上都会稳定通过；
- CI已经在云端运行通过，除非GitHub Actions确实有成功记录。

## 10. 面试前必须能回答的问题

1. 为什么pytest和Playwright都需要API Client，它们的职责有什么不同？
2. `({ page, actor, data })`分别来自哪里，执行顺序是什么？
3. storageState为什么必须在BrowserContext创建前提供？
4. 为什么Context隔离不能代替数据库测试数据隔离？
5. 为什么Page Object里保留操作，而核心业务断言留在spec？
6. 为什么`toHaveText()`比`textContent() + toBe()`更适合异步页面？
7. `route.fulfill()`和`route.abort()`分别模拟什么故障？
8. TestData为什么要在测试结束后逆序清理？
9. 为什么E2E job不启动RabbitMQ？
10. 类型检查通过、测试运行通过和测试场景设计正确有什么区别？
11. 为什么Midscene只负责页面操作，还要用API交叉断言？
12. 为什么AI用例不进入默认CI？

## 11. 面试前必须能现场修改的内容

建议练习以下五个小修改：

1. 给`AuthPage`增加注册方法；
2. 给pytest注册测试增加一组参数化边界；
3. 新增一个Page Object Locator并在spec中断言；
4. 使用`route.fulfill()`增加一个发布失败场景；
5. 给TestData新增一种资源的记录和清理。

如果这五类修改可以独立完成，这个项目才真正属于你的能力，而不只是一个能够运行的代码仓库。
