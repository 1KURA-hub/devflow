"""认证接口：成功、核心异常、参数化边界和一个并发代表场景。"""

import uuid
from concurrent.futures import ThreadPoolExecutor

import pytest

from clients.auth import AuthClient
from clients.base import BaseClient
from config import DEFAULT_PASSWORD


def _fresh_name(length: int = 10) -> str:
    return uuid.uuid4().hex[:length]


@pytest.mark.smoke
def test_register_success(anonymous_client, unique_username):
    resp = AuthClient(anonymous_client).register(
        unique_username,
        DEFAULT_PASSWORD,
        "nick",
    )

    assert resp.ok, f"注册失败: {resp.status_code} {resp.message}"
    assert resp.data["token"]
    assert resp.data["user"]["id"] > 0
    assert resp.data["user"]["username"] == unique_username


def test_register_duplicate_username_conflict(anonymous_client, unique_username):
    auth = AuthClient(anonymous_client)
    first = auth.register(unique_username, DEFAULT_PASSWORD, "nick1")
    assert first.ok, f"准备用户失败: {first.status_code} {first.message}"

    second = auth.register(unique_username, DEFAULT_PASSWORD, "nick2")
    assert second.status_code == 409
    assert second.code == 409


def test_concurrent_register_same_username_has_one_success(server_ready, unique_username):
    """只保留一个并发用例，验证数据库唯一约束与错误映射。"""

    def register_once():
        http = BaseClient()
        try:
            return AuthClient(http).register(
                unique_username,
                DEFAULT_PASSWORD,
                "并发用户",
            )
        finally:
            http.close()

    with ThreadPoolExecutor(max_workers=2) as pool:
        responses = list(pool.map(lambda _: register_once(), range(2)))

    assert sorted(resp.status_code for resp in responses) == [200, 409]


@pytest.mark.smoke
def test_login_success(registered_user, anonymous_client):
    resp = AuthClient(anonymous_client).login(
        registered_user.username,
        registered_user.password,
    )

    assert resp.ok, f"登录失败: {resp.status_code} {resp.message}"
    assert resp.data["token"]
    assert resp.data["user"]["id"] == registered_user.user_id


def test_login_wrong_password_unauthorized(registered_user, anonymous_client):
    resp = AuthClient(anonymous_client).login(
        registered_user.username,
        "wrong-password",
    )

    assert resp.status_code == 401
    assert resp.code == 401
    assert resp.message == "invalid username or password"


def test_login_nonexistent_user_uses_same_error(anonymous_client):
    """不存在用户与错误密码使用相同文案，避免泄漏账号是否存在。"""
    resp = AuthClient(anonymous_client).login(_fresh_name(), DEFAULT_PASSWORD)

    assert resp.status_code == 401
    assert resp.message == "invalid username or password"


@pytest.mark.parametrize("username_length", [2, 13])
def test_register_invalid_username_length(anonymous_client, username_length):
    resp = AuthClient(anonymous_client).register(
        _fresh_name(username_length),
        DEFAULT_PASSWORD,
        "nick",
    )

    assert resp.status_code == 400


@pytest.mark.parametrize("missing_field", ["username", "password", "nickname"])
def test_register_missing_field_bad_request(anonymous_client, missing_field):
    fields = {
        "username": _fresh_name(),
        "password": DEFAULT_PASSWORD,
        "nickname": "nick",
    }
    fields.pop(missing_field)

    resp = anonymous_client.post("/api/auth/register", json=fields)
    assert resp.status_code == 400


def test_invalid_token_unauthorized(anonymous_client):
    anonymous_client.set_token("invalid.token.value")
    resp = AuthClient(anonymous_client).me()

    assert resp.status_code == 401
