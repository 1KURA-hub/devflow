"""点赞、收藏和评论：保留每类最能说明业务契约的用例。"""

import pytest

from clients.comment import CommentClient


@pytest.mark.smoke
def test_like_increments_count(published_post, second_user):
    post_id = published_post["id"]
    liked = second_user.interaction.like(post_id)
    assert liked.ok, f"点赞失败: {liked.status_code} {liked.message}"

    detail = second_user.post.get(post_id)
    assert detail.ok, f"查询帖子失败: {detail.status_code} {detail.message}"
    assert detail.data["like_count"] == 1
    assert detail.data["liked"] is True


@pytest.mark.idempotent
def test_like_is_idempotent(published_post, second_user):
    """重复点赞都返回成功，但计数只能增加一次。"""
    post_id = published_post["id"]
    first = second_user.interaction.like(post_id)
    second = second_user.interaction.like(post_id)
    assert first.ok, f"首次点赞失败: {first.status_code} {first.message}"
    assert second.ok, f"重复点赞失败: {second.status_code} {second.message}"

    detail = second_user.post.get(post_id)
    assert detail.ok, f"查询帖子失败: {detail.status_code} {detail.message}"
    assert detail.data["like_count"] == 1


def test_unlike_decrements_count(published_post, second_user):
    post_id = published_post["id"]
    liked = second_user.interaction.like(post_id)
    assert liked.ok, f"准备点赞失败: {liked.status_code} {liked.message}"

    unliked = second_user.interaction.unlike(post_id)
    assert unliked.ok, f"取消点赞失败: {unliked.status_code} {unliked.message}"

    detail = second_user.post.get(post_id)
    assert detail.ok, f"查询帖子失败: {detail.status_code} {detail.message}"
    assert detail.data["like_count"] == 0
    assert detail.data["liked"] is False


def test_favorite_updates_detail_and_list(published_post, second_user):
    post_id = published_post["id"]
    favorited = second_user.interaction.favorite(post_id)
    assert favorited.ok, f"收藏失败: {favorited.status_code} {favorited.message}"

    detail = second_user.post.get(post_id)
    assert detail.ok, f"查询帖子失败: {detail.status_code} {detail.message}"
    assert detail.data["favorite_count"] == 1
    assert detail.data["favorited"] is True

    favorites = second_user.interaction.my_favorites()
    assert favorites.ok, f"查询收藏列表失败: {favorites.status_code} {favorites.message}"
    assert post_id in [item["id"] for item in favorites.data["items"]]


@pytest.mark.smoke
def test_comment_updates_count_and_appears_in_list(published_post, second_user):
    post_id = published_post["id"]
    created = second_user.comment.create(post_id, "nice post")
    assert created.ok, f"评论失败: {created.status_code} {created.message}"

    detail = second_user.post.get(post_id)
    assert detail.ok, f"查询帖子失败: {detail.status_code} {detail.message}"
    assert detail.data["comment_count"] == 1

    listing = second_user.comment.list(post_id)
    assert listing.ok, f"查询评论失败: {listing.status_code} {listing.message}"
    assert created.data["id"] in [item["id"] for item in listing.data["items"]]


def test_create_empty_comment_bad_request(published_post, second_user):
    resp = second_user.comment.create(published_post["id"], "   ")
    assert resp.status_code == 400


def test_create_comment_without_auth_unauthorized(published_post, anonymous_client):
    resp = CommentClient(anonymous_client).create(published_post["id"], "hello")
    assert resp.status_code == 401
