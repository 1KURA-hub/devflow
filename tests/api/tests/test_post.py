"""帖子接口：发布、详情、权限和删除。"""

import pytest

from clients.post import PostClient


@pytest.mark.smoke
def test_create_post_success(registered_user):
    resp = registered_user.post.create(title="hello", content="content body")

    assert resp.ok, f"发帖失败: {resp.status_code} {resp.message}"
    post_id = resp.data["id"]
    try:
        assert post_id > 0
        assert resp.data["title"] == "hello"
        assert resp.data["author_id"] == registered_user.user_id
    finally:
        registered_user.post.delete(post_id)


def test_create_post_without_auth_unauthorized(anonymous_client):
    resp = PostClient(anonymous_client).create(title="x", content="y")
    assert resp.status_code == 401


def test_get_post_returns_core_fields(published_post, registered_user):
    resp = registered_user.post.get(published_post["id"])

    assert resp.ok, f"查询帖子失败: {resp.status_code} {resp.message}"
    assert resp.data["id"] == published_post["id"]
    assert resp.data["title"] == published_post["title"]
    for field in ("like_count", "comment_count", "favorite_count", "liked", "favorited"):
        assert field in resp.data


def test_delete_post_by_non_author_forbidden(published_post, second_user):
    resp = second_user.post.delete(published_post["id"])
    assert resp.status_code == 403


def test_delete_post_by_author_success(published_post, registered_user):
    post_id = published_post["id"]
    deleted = registered_user.post.delete(post_id)
    assert deleted.ok, f"删帖失败: {deleted.status_code} {deleted.message}"

    detail = registered_user.post.get(post_id)
    assert detail.status_code == 404


@pytest.mark.parametrize(
    "title, content",
    [("", "non-empty content"), ("non-empty title", "")],
)
def test_create_post_empty_field_bad_request(registered_user, title, content):
    resp = registered_user.post.create(title=title, content=content)
    assert resp.status_code == 400
