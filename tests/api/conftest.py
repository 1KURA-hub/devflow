"""全局 pytest fixtures。

设计原则：
- 每条用例自带独立用户/帖子，避免互相污染
- 多用户场景用 registered_user / second_user，模拟真实跨用户交互
- fixture 是分层的：底层 http_client → registered_user → published_post
"""

from __future__ import annotations

import os
import sys
import time
import uuid
from dataclasses import dataclass
from typing import Iterator
from urllib.parse import urlparse

import pytest

sys.path.insert(0, os.path.dirname(__file__))

from clients.auth import AuthClient
from clients.base import BaseClient
from clients.comment import CommentClient
from clients.feed import FeedClient
from clients.follow import FollowClient
from clients.interaction import InteractionClient
from clients.notification import NotificationClient
from clients.post import PostClient
from config import ALLOW_REMOTE_TESTS, BASE_URL, DEFAULT_PASSWORD


@dataclass
class TestUser:
    username: str
    password: str
    nickname: str
    user_id: int
    token: str
    http: BaseClient
    auth: AuthClient
    post: PostClient
    interaction: InteractionClient
    follow: FollowClient
    feed: FeedClient
    notification: NotificationClient
    comment: CommentClient


@pytest.fixture(scope="session")
def server_ready():
    """会话级门禁：确认目标是明确允许写入的 DevFlow 测试环境。"""
    hostname = (urlparse(BASE_URL).hostname or "").lower()
    if hostname not in {"localhost", "127.0.0.1", "::1"} and not ALLOW_REMOTE_TESTS:
        pytest.exit(
            "拒绝向远程环境写入测试数据；确认是隔离测试环境后设置 "
            "DEVFLOW_ALLOW_REMOTE_TESTS=true"
        )

    probe = BaseClient()
    try:
        last_err = None
        for _ in range(20):
            try:
                resp = probe.get("/healthz")
                if resp.ok:
                    health = resp.data if isinstance(resp.data, dict) else {}
                    if health.get("app") != "devflow":
                        pytest.exit(f"目标不是 DevFlow：{BASE_URL}/healthz")
                    if health.get("env") != "test":
                        pytest.exit(
                            f"拒绝执行写数据接口测试：目标 APP_ENV={health.get('env', 'unknown')}，必须为 test"
                        )
                    return True
                last_err = f"HTTP {resp.status_code}, code={resp.code}, message={resp.message}"
            except Exception as e:
                last_err = e
            time.sleep(0.5)
        pytest.exit(f"devflow server 未就绪 (base_url={probe.base_url}, err={last_err})")
    finally:
        probe.close()


@pytest.fixture()
def unique_username() -> str:
    """每条用例独立用户名，避免重复注册冲突。

    用户名长度上限为 12（见 auth service 校验），这里取 u_ + 10 位 hex = 12 字符。
    """
    return f"u_{uuid.uuid4().hex[:10]}"


def _register(unique_name: str) -> TestUser:
    http = BaseClient()
    auth = AuthClient(http)
    nickname = f"n_{unique_name[-6:]}"
    resp = auth.register(unique_name, DEFAULT_PASSWORD, nickname)
    assert resp.ok, f"注册失败 {resp.status_code} {resp.message}"
    token = resp.data["token"]
    user_id = resp.data["user"]["id"]
    http.set_token(token)
    return TestUser(
        username=unique_name,
        password=DEFAULT_PASSWORD,
        nickname=nickname,
        user_id=user_id,
        token=token,
        http=http,
        auth=AuthClient(http),
        post=PostClient(http),
        interaction=InteractionClient(http),
        follow=FollowClient(http),
        feed=FeedClient(http),
        notification=NotificationClient(http),
        comment=CommentClient(http),
    )


@pytest.fixture()
def registered_user(server_ready, unique_username) -> Iterator[TestUser]:
    user = _register(unique_username)
    yield user
    user.http.close()


@pytest.fixture()
def second_user(server_ready) -> Iterator[TestUser]:
    user = _register(f"u_{uuid.uuid4().hex[:10]}")
    yield user
    user.http.close()


@pytest.fixture()
def anonymous_client(server_ready) -> Iterator[BaseClient]:
    """无 token 的客户端，用于测未授权场景。"""
    client = BaseClient()
    yield client
    client.close()


@pytest.fixture()
def published_post(registered_user) -> Iterator[dict]:
    """创建独立帖子；测试结束后尽量删除，避免污染长期测试环境。"""
    title = f"title_{uuid.uuid4().hex[:6]}"
    resp = registered_user.post.create(title=title, content="hello devflow test")
    assert resp.ok, f"发帖失败 {resp.status_code} {resp.message}"
    post = {"id": resp.data["id"], "title": title, "author": registered_user}
    yield post
    # 某些用例会主动删除帖子；清理时遇到 404 可以安全忽略。
    try:
        registered_user.post.delete(post["id"])
    except Exception:
        # 清理是best-effort，不应覆盖测试本身的失败信息。
        pass
