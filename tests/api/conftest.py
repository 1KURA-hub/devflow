"""全局 pytest fixtures。

设计原则：
- 每条用例自带独立用户/帖子，避免互相污染
- 多用户场景用 second_user / second_client，模拟真实跨用户交互
- fixture 是分层的：底层 http_client → registered_user → published_post
"""

from __future__ import annotations

import os
import sys
import time
import uuid
from dataclasses import dataclass

import pytest

sys.path.insert(0, os.path.dirname(__file__))

from clients.auth import AuthClient
from clients.base import BaseClient
from clients.feed import FeedClient
from clients.follow import FollowClient
from clients.interaction import InteractionClient
from clients.notification import NotificationClient
from clients.post import PostClient
from config import DEFAULT_PASSWORD


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


def _build_user_clients(http: BaseClient) -> dict:
    return {
        "auth": AuthClient(http),
        "post": PostClient(http),
        "interaction": InteractionClient(http),
        "follow": FollowClient(http),
        "feed": FeedClient(http),
        "notification": NotificationClient(http),
    }


@pytest.fixture(scope="session")
def server_ready():
    """会话级冒烟：先 ping /healthz 确认 server 可达，避免后续 N 条用例集体超时。"""
    probe = BaseClient()
    last_err = None
    for _ in range(20):
        try:
            resp = probe.get("/healthz")
            if resp.status_code == 200:
                return True
        except Exception as e:
            last_err = e
        time.sleep(0.5)
    pytest.exit(f"devflow server 未就绪 (base_url={probe.base_url}, err={last_err})")


@pytest.fixture()
def unique_username() -> str:
    """每条用例独立用户名，避免重复注册冲突。"""
    return f"u_{uuid.uuid4().hex[:12]}"


def _register(unique_name: str) -> TestUser:
    http = BaseClient()
    auth = AuthClient(http)
    nickname = f"nick_{unique_name[-6:]}"
    resp = auth.register(unique_name, DEFAULT_PASSWORD, nickname)
    assert resp.ok, f"注册失败 {resp.status_code} {resp.message}"
    token = resp.data["token"]
    user_id = resp.data["user"]["id"]
    http.set_token(token)
    clients = _build_user_clients(http)
    return TestUser(
        username=unique_name,
        password=DEFAULT_PASSWORD,
        nickname=nickname,
        user_id=user_id,
        token=token,
        http=http,
        **clients,
    )


@pytest.fixture()
def registered_user(server_ready, unique_username) -> TestUser:
    return _register(unique_username)


@pytest.fixture()
def second_user(server_ready) -> TestUser:
    return _register(f"u_{uuid.uuid4().hex[:12]}")


@pytest.fixture()
def anonymous_client(server_ready) -> BaseClient:
    """无 token 的客户端，用于测未授权场景。"""
    return BaseClient()


@pytest.fixture()
def published_post(registered_user) -> dict:
    """registered_user 发一篇帖子，返回 {"id", "title", "author"}。"""
    title = f"title_{uuid.uuid4().hex[:6]}"
    resp = registered_user.post.create(title=title, content="hello devflow test")
    assert resp.ok, f"发帖失败 {resp.status_code} {resp.message}"
    return {"id": resp.data["id"], "title": title, "author": registered_user}
