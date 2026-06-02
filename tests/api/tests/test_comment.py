"""评论链路：发评论 / 计数 / 列表 / 校验。

业务规则（来自 internal/service/comment.go + handler/comment.go）：
- 发评论需登录，content 不能为空，长度 <= 1000 个字符（utf8 RuneCount）
- 发评论成功后帖子的 comment_count + 1
- 评论会给帖子作者生成一条 type=comment 的通知
- GET /posts/:id/comments 返回 {"items": [...], "has_more": bool}
"""

import pytest


@pytest.mark.smoke
def test_create_comment_increments_count(published_post, second_user):
    """B 评论 A 的帖子 → 200，且帖子 comment_count 变为 1。"""
    post_id = published_post["id"]

    resp = second_user.comment.create(post_id, "nice post!")
    assert resp.ok, resp.message
    assert resp.data["id"] > 0
    assert resp.data["content"] == "nice post!"

    detail = second_user.post.get(post_id)
    assert detail.data["comment_count"] == 1


def test_comment_appears_in_list(published_post, second_user):
    """评论后应能在评论列表里查到这条评论。"""
    post_id = published_post["id"]
    create = second_user.comment.create(post_id, "first comment")
    assert create.ok
    comment_id = create.data["id"]

    listing = second_user.comment.list(post_id)
    assert listing.ok, listing.message
    ids = [c["id"] for c in listing.data["items"]]
    assert comment_id in ids


def test_create_empty_comment_bad_request(published_post, second_user):
    """空内容评论应返回 400。"""
    resp = second_user.comment.create(published_post["id"], "   ")
    assert resp.status_code == 400, f"空评论应 400，实际 {resp.status_code} {resp.message}"


def test_create_too_long_comment_bad_request(published_post, second_user):
    """超过 1000 字符的评论应返回 400。"""
    resp = second_user.comment.create(published_post["id"], "x" * 1001)
    assert resp.status_code == 400, f"超长评论应 400，实际 {resp.status_code} {resp.message}"


def test_create_comment_without_auth_unauthorized(published_post, anonymous_client):
    """未登录发评论必须 401。"""
    from clients.comment import CommentClient

    resp = CommentClient(anonymous_client).create(published_post["id"], "hi")
    assert resp.status_code == 401
