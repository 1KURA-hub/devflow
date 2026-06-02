"""认证链路：注册 / 登录。

业务规则（来自 internal/handler/auth.go 的 writeAuthError 映射 +
internal/service/auth.go 的长度校验）：
- 重复用户名 -> 409 Conflict
- 错误密码 / 用户不存在 -> 401 Unauthorized（统一文案，不泄漏用户是否存在）
- 无效参数 -> 400 Bad Request
- username 长度 [3, 12]
- password 长度 [6, 16]
- nickname 长度 [3, 12]
"""

import uuid

import pytest

from clients.auth import AuthClient
from clients.base import BaseClient
from config import DEFAULT_PASSWORD


def _fresh_name(n: int = 10) -> str:
    """生成一个一定不重复、长度合规的用户名（u_ + 8 位 hex = 10 字符）。"""
    return f"u_{uuid.uuid4().hex[:n - 2]}"


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


def test_login_nonexistent_user_unauthorized(server_ready):
    """登录一个不存在的用户也应返回 401，且文案与错误密码一致，避免账号枚举。"""
    auth = AuthClient(BaseClient())
    resp = auth.login(_fresh_name(), DEFAULT_PASSWORD)

    assert resp.status_code == 401, f"期望 401，实际 {resp.status_code} {resp.message}"


# --- 注册参数边界值（等价类 + 边界值，校验规则见 service/auth.go） ---


@pytest.mark.parametrize(
    "username_len, ok",
    [
        (2, False),   # 下界外
        (3, True),    # 下界
        (12, True),   # 上界
        (13, False),  # 上界外
    ],
)
def test_register_username_length_boundary(server_ready, username_len, ok):
    """username 合法长度区间为 [3, 12]，逐个验证边界点。"""
    auth = AuthClient(BaseClient())
    username = "a" + "b" * (username_len - 1)
    resp = auth.register(username, DEFAULT_PASSWORD, "nick")

    if ok:
        assert resp.ok, f"长度 {username_len} 应注册成功: {resp.status_code} {resp.message}"
    else:
        assert resp.status_code == 400, f"长度 {username_len} 应被拒: {resp.status_code}"


@pytest.mark.parametrize(
    "password_len, ok",
    [
        (5, False),   # 下界外
        (6, True),    # 下界
        (16, True),   # 上界
        (17, False),  # 上界外
    ],
)
def test_register_password_length_boundary(server_ready, password_len, ok):
    """password 合法长度区间为 [6, 16]。"""
    auth = AuthClient(BaseClient())
    password = "p" * password_len
    resp = auth.register(_fresh_name(), password, "nick")

    if ok:
        assert resp.ok, f"密码长度 {password_len} 应成功: {resp.status_code} {resp.message}"
    else:
        assert resp.status_code == 400, f"密码长度 {password_len} 应被拒: {resp.status_code}"


@pytest.mark.parametrize(
    "nickname_len, ok",
    [
        (2, False),   # 下界外
        (3, True),    # 下界
        (12, True),   # 上界
        (13, False),  # 上界外
    ],
)
def test_register_nickname_length_boundary(server_ready, nickname_len, ok):
    """nickname 合法长度区间为 [3, 12]。"""
    auth = AuthClient(BaseClient())
    nickname = "n" * nickname_len
    resp = auth.register(_fresh_name(), DEFAULT_PASSWORD, nickname)

    if ok:
        assert resp.ok, f"昵称长度 {nickname_len} 应成功: {resp.status_code} {resp.message}"
    else:
        assert resp.status_code == 400, f"昵称长度 {nickname_len} 应被拒: {resp.status_code}"


@pytest.mark.parametrize("missing", ["username", "password", "nickname"])
def test_register_missing_field_bad_request(server_ready, missing):
    """缺失任一必填字段（空串）都应返回 400。"""
    auth = AuthClient(BaseClient())
    fields = {
        "username": _fresh_name(),
        "password": DEFAULT_PASSWORD,
        "nickname": "nick",
    }
    fields[missing] = ""
    resp = auth.register(fields["username"], fields["password"], fields["nickname"])

    assert resp.status_code == 400, f"{missing} 为空应返回 400，实际 {resp.status_code}"
