# DevFlow API 自动化测试

这是一个面向测试开发实习作品集的 `pytest + requests` 接口自动化子项目。它重点展示测试分层、fixture、数据隔离、核心异常、幂等性和异步结果轮询，不追求穷举所有接口。

## 分层结构

```text
tests/                         # 用例与业务断言
  ↓
conftest.py                    # 用户、认证客户端、帖子等fixture
  ↓
clients/                       # 薄client，只封装路径和参数
  ↓
BaseClient                     # Session、Bearer Token、统一响应解析
  ↓
DevFlow HTTP API
```

目录：

```text
tests/api/
├── config.py
├── conftest.py
├── pytest.ini
├── requirements.txt
├── clients/
│   ├── base.py
│   ├── auth.py
│   ├── post.py
│   ├── interaction.py
│   ├── comment.py
│   ├── follow.py
│   ├── feed.py
│   └── notification.py
└── tests/
    ├── test_auth.py
    ├── test_post.py
    ├── test_interactions.py
    └── test_social.py
```

## 覆盖范围

- 注册、登录、重复用户名、无效Token和参数化异常数据；
- 发帖、详情、未认证、作者/非作者删除权限；
- 点赞、收藏、评论及计数一致性；
- 一个重复点赞幂等用例；
- 一个并发注册代表用例；
- 关注关系、关注流和点赞通知；
- smoke、idempotent、cross 三类标记。

## 数据与环境安全

- 用户名和帖子标题使用随机后缀，测试之间不依赖执行顺序；
- 每个用户拥有独立的 `requests.Session`，fixture结束时关闭；
- `published_post` fixture会尽量删除帖子；
- 测试用户目前没有删除接口，因此CI使用临时数据库；
- 启动测试前校验 `/healthz` 的 `app=devflow`、`env=test`；
- 远程测试环境必须显式设置 `DEVFLOW_ALLOW_REMOTE_TESTS=true`，严禁指向生产环境。

## 本地运行

先启动测试环境中的DevFlow服务，然后执行：

```bash
cd tests/api
python -m pip install -r requirements.txt

# 收集用例，不发送业务请求
pytest --collect-only -q

# 最小冒烟集合
pytest -m smoke

# 幂等代表用例
pytest -m idempotent

# 跨模块场景
pytest -m cross

# 完整回归
pytest
```

环境变量：

| 变量 | 默认值 | 作用 |
| --- | --- | --- |
| `DEVFLOW_BASE_URL` | `http://localhost:8080` | 被测服务地址 |
| `DEVFLOW_TIMEOUT` | `10` | 单次请求超时秒数 |
| `DEVFLOW_NOTIFY_TIMEOUT` | `5` | 异步结果最长轮询时间 |
| `DEVFLOW_ALLOW_REMOTE_TESTS` | 未启用 | 显式允许隔离远程环境写入 |

## Allure 测试报告

`pytest.ini` 已把 `--alluredir=allure-results` 写入 addopts，因此每次运行都会在
`tests/api/allure-results/` 生成 Allure 原始数据。查看报告：

```bash
# 方式一：本地起服务并自动打开浏览器（适合演示）
allure serve allure-results

# 方式二：生成静态 HTML 后打开
allure generate allure-results --clean -o allure-report
allure open allure-report
```

Allure CLI 需要 Java 17+（macOS：`brew install allure`，或 `npm i -g allure-commandline`）。
CI 中由 workflow 下载 CLI、生成报告并作为 artifact 上传，无需手动执行。

## 关键设计说明

1. Client层不写断言，测试层保留业务期望。
2. 每次读取响应数据前先确认请求成功，使失败信息更清楚。
3. 异步通知和关注流使用限时轮询，不用固定长时间`sleep`猜测完成时间。
4. 随机数据解决并发冲突，临时数据库或安全清理解决数据残留；两者不是一回事。
5. 该项目定位为“核心接口回归”，不是全接口覆盖或性能测试框架。
