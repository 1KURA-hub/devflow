"""收藏链路：与点赞对称的核心互动，验证 *幂等性* + *计数一致性*。

业务规则（来自 internal/service/interaction.go + repository）：
- 收藏用 INSERT ... ON CONFLICT DO NOTHING + 唯一约束 (user_id, post_id)
- 事务内同时写 favorites 和 favorite_count + 1
- 重复收藏：RowsAffected = 0 → 200，count 不变
- 取消未收藏：删 0 行 → 200（幂等）
- 收藏接口返回 data=nil，计数与状态通过 GET /posts/:id 的 favorite_count / favorited 验证
"""

import pytest


@pytest.mark.smoke
def test_favorite_increments_count(published_post, second_user):
    """B 收藏 A 的帖子 → 200 且 favorite_count = 1、favorited = True。"""
    resp = second_user.interaction.favorite(published_post["id"])
    assert resp.ok, resp.message

    detail = second_user.post.get(published_post["id"])
    assert detail.ok
    assert detail.data["favorite_count"] == 1
    assert detail.data["favorited"] is True


@pytest.mark.idempotent
def test_favorite_is_idempotent(published_post, second_user):
    """重复收藏两次 → 都返回 200，favorite_count 仍是 1。"""
    post_id = published_post["id"]

    first = second_user.interaction.favorite(post_id)
    second = second_user.interaction.favorite(post_id)

    assert first.ok and second.ok, f"重复收藏应都为 200: {first.message} / {second.message}"

    detail = second_user.post.get(post_id)
    assert detail.data["favorite_count"] == 1, "重复收藏不能让 favorite_count 涨成 2"


def test_unfavorite_decrements_count(published_post, second_user):
    """取消收藏应使 favorite_count 回到 0 且 favorited = False。"""
    post_id = published_post["id"]
    second_user.interaction.favorite(post_id)

    resp = second_user.interaction.unfavorite(post_id)
    assert resp.ok

    detail = second_user.post.get(post_id)
    assert detail.data["favorite_count"] == 0
    assert detail.data["favorited"] is False


@pytest.mark.idempotent
def test_unfavorite_never_favorited_is_idempotent(published_post, second_user):
    """取消一个没收藏过的帖应安静返回 200，不报错。"""
    resp = second_user.interaction.unfavorite(published_post["id"])
    assert resp.ok, f"取消未收藏应幂等: {resp.status_code} {resp.message}"


def test_favorite_appears_in_my_favorites(published_post, second_user):
    """收藏后该帖应出现在 /me/favorites 列表中。"""
    post_id = published_post["id"]
    second_user.interaction.favorite(post_id)

    resp = second_user.interaction.my_favorites()
    assert resp.ok, resp.message
    ids = [item["id"] for item in resp.data["items"]]
    assert post_id in ids, "收藏的帖子应出现在我的收藏列表"
