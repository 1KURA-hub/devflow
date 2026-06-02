# DevFlow 接口自动化测试

基于 `pytest + requests` 的 HTTP 接口自动化，覆盖 devflow 后端核心链路。

## 一、本地怎么跑

```bash
# 1) 启动后端依赖 + server（项目根目录）
docker compose up -d mysql redis
go run ./cmd/server          # 监听 :8080

# 2) 装测试依赖
cd tests/api
pip install -r requirements.txt

# 3) 跑全部用例
pytest

# 跑某一类
pytest tests/test_like.py
pytest -m smoke              # 只跑冒烟
pytest -m idempotent         # 只跑幂等场景
pytest -m cross              # 只跑跨链路
```

配置项（环境变量）：

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `DEVFLOW_BASE_URL` | `http://localhost:8080` | 被测 server 地址 |
| `DEVFLOW_TIMEOUT` | `10` | 单个请求超时（秒） |
| `DEVFLOW_NOTIFY_TIMEOUT` | `5` | 异步通知轮询超时 |

## 二、目录与职责

```
tests/api/
├── config.py            # 集中读环境变量
├── conftest.py          # 全局 fixtures（用户、token、帖子）
├── clients/             # 接口封装层：每个领域一个文件，只发请求不做断言
│   ├── base.py          # 统一 HTTP 调用 + 响应解析（ApiResponse）
│   ├── auth.py
│   ├── post.py
│   ├── interaction.py
│   ├── follow.py
│   ├── feed.py
│   └── notification.py
└── tests/               # 用例层：组合 fixture + client，写断言
    ├── test_auth.py     # 4 条：注册/登录/重复用户名/错密码
    ├── test_post.py     # 4 条：发帖/未授权/详情字段/非作者删帖
    ├── test_like.py     # 4 条：点赞/幂等/取消/取消未点赞
    ├── test_follow.py   # 3 条：关注/关注自己/重复关注
    ├── test_feed.py     # 3 条：latest/hot/following 冷启动降级
    └── test_cross.py    # 2 条：fan-out / 点赞触发通知
```

## 三、设计要点（面试可讲）

### 三层分离
- **用例层**只表达业务意图与断言：「点赞两次仍然 like_count=1」。
- **client 层**只封装 URL/header：换接口路径只动一个文件。
- **base 层**统一响应解析：devflow 返回结构 `{code, message, data}`，封装成 `ApiResponse`，用例同时能断言 HTTP 状态码 + 业务码 + data 字段。

### fixture 设计
- `unique_username` 每条用例独立用户名，避免互相冲突。
- `registered_user` / `second_user` 不止给 token，还预装好所有 domain client，用例直接 `user.interaction.like(...)`，可读性最高。
- `published_post` 在 `registered_user` 之上，叠出一篇可被点赞/评论的帖。

### 异步等待
- 通知链路走 MQ，本地若不开 RabbitMQ 走同步降级，CI 若开 MQ 需要轮询。
- `test_cross.py` 的 `poll_until` 兼容两种部署，避免 sleep 写死。

### 重点覆盖的"系统性契约"
| 测试 | 覆盖的设计 |
| --- | --- |
| `test_like_is_idempotent` | `INSERT ... ON CONFLICT DO NOTHING` + 唯一索引 + 事务计数 |
| `test_following_feed_cold_start_falls_back_to_latest` | 关注 feed 的冷启动降级 |
| `test_follower_can_see_followed_user_post` | 发帖 fan-out 到 inbox |
| `test_like_triggers_like_notification` | MQ + 通知幂等写入 |
| `test_duplicate_follow_known_issue` | 主动记录"重复关注返回 500"的不一致 |

## 四、面试常见问答

**Q：为什么用接口测试而不是单元测试？**  
A：devflow 的价值大部分在 service + cache + MQ 协作，纯单测难以覆盖跨层逻辑。接口测试用真实 HTTP + 真实 DB/Redis，更接近用户视角，能验证"系统作为一个整体的契约"。Go 单测我在另外补关键纯函数（hotScore、参数归一化）。

**Q：为什么分 client 层？**  
A：接口路径或 header 调整时，所有用例不用动；同一个 client 也方便在脚本里复用做数据准备。

**Q：怎么保证用例可重复执行？**  
A：用例级 fixture + 随机 username/title，每次新用户、新帖子，不依赖共享数据；本地反复跑也不会脏。

**Q：跨链路（通知、fan-out）异步怎么测？**  
A：用 `poll_until` 在限定窗口内重试断言，超时算失败。这比 sleep 更稳，也能反映"应该多久内可见"的业务期望。

**Q：测出过 bug 吗？**  
A：是的，重复关注会落到 500（写错误映射）；用例 `test_duplicate_follow_known_issue` 把现状固化下来，修复后该用例会触发回归提醒，避免悄悄变更行为。

## 五、CI

`.github/workflows/api-tests.yml` 会：
1. 起 MySQL + Redis service container
2. 编译并启动 devflow server（无 RabbitMQ，走同步降级）
3. 等 `/healthz` 就绪
4. 跑 `pytest`，产物 `report.xml` 作为 artifact
