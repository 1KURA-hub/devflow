"""认证链路：注册 / 登录。

业务规则（来自 internal/handler/auth.go 的 writeAuthError 映射）：
- 重复用户名 -> 409 Conflict
- 错误密码  -> 401 Unauthorized
- 无效参数  -> 400 Bad Request
"""

import pytest

from clients.auth import AuthClient
from clients.base import BaseClient
from config import DEFAULT_PASSWORD


@pytest.mark.smoke
def test_register_success(server_ready, unique_username):
    """注册成功应返回 token + user.id。"""
    auth = AuthClient(BaseClient())
    resp = auth.register(unique_username, DEFAULT_PASSWORD, "nick")

    assert resp.ok, resp.message
    assert resp.data["token"], "token 必须返回"
    assert resp.data["user"]["id"] > 0
    assert resp.data["user"]["username"] == unique_username


def test_register_duplicate_username_conflict(server_ready, unique_username):
    """同一用户名重复注册必须返回 409，验证用户名唯一约束生效。"""
    auth = AuthClient(BaseClient())
    first = auth.register(unique_username, DEFAULT_PASSWORD, "nick1")
    assert first.ok

    second = auth.register(unique_username, DEFAULT_PASSWORD, "nick2")
    assert second.status_code == 409, f"期望 409，实际 {second.status_code} {second.message}"


@pytest.mark.smoke
def test_login_success(registered_user):
    """已注册用户用正确密码登录应返回新 token。"""
    auth = AuthClient(BaseClient())
    resp = auth.login(registered_user.username, registered_user.password)

    assert resp.ok, resp.message
    assert resp.data["token"]
    assert resp.data["user"]["id"] == registered_user.user_id


def test_login_wrong_password_unauthorized(registered_user):
    """错误密码必须返回 401，不能泄漏 user 是否存在的信息。"""
    auth = AuthClient(BaseClient())
    resp = auth.login(registered_user.username, "wrong-password")

    assert resp.status_code == 401, f"期望 401，实际 {resp.status_code} {resp.message}"
