"""帖子链路：发布 / 详情 / 删除。

业务规则（来自 internal/handler/post.go）：
- 创建/删除需登录，否则 401
- 删除他人帖子返回 403（ErrForbidden）
- 详情返回包含 liked / favorited 字段（基于当前 viewer 状态）
"""

import pytest

from clients.post import PostClient


@pytest.mark.smoke
def test_create_post_success(registered_user):
    """已登录发帖应返回 id 和作者信息。"""
    resp = registered_user.post.create(title="hello", content="content body")

    assert resp.ok, resp.message
    assert resp.data["id"] > 0
    assert resp.data["title"] == "hello"
    assert resp.data["author_id"] == registered_user.user_id


def test_create_post_without_auth_unauthorized(anonymous_client):
    """未登录发帖必须 401。"""
    resp = PostClient(anonymous_client).create(title="x", content="y")
    assert resp.status_code == 401


def test_get_post_returns_full_fields(published_post, registered_user):
    """详情接口字段完整：含计数 + 当前用户的 liked/favorited 状态。"""
    resp = registered_user.post.get(published_post["id"])

    assert resp.ok
    data = resp.data
    assert data["id"] == published_post["id"]
    assert data["title"] == published_post["title"]
    for field in ("like_count", "comment_count", "favorite_count", "liked", "favorited"):
        assert field in data, f"详情应包含字段 {field}"
    assert data["like_count"] == 0


def test_delete_post_by_non_author_forbidden(published_post, second_user):
    """非作者删除他人帖子必须 403，验证 ErrForbidden 映射正确。"""
    resp = second_user.post.delete(published_post["id"])
    assert resp.status_code == 403, f"期望 403，实际 {resp.status_code} {resp.message}"
