"""点赞链路：核心验证 *幂等性* + *计数一致性*。

业务规则（来自 internal/service/interaction.go + repository/interaction.go）：
- AddLike 用 INSERT ... ON CONFLICT DO NOTHING + 唯一约束 (user_id, post_id)
- 事务内同时 INSERT likes 和 like_count + 1
- 重复点赞：RowsAffected = 0 → service 返回 nil → HTTP 200 + like_count 不变
- 取消未点赞：删除 0 行 → service 返回 nil → 200

面试卖点：
这套用例直接验证了 devflow 点赞接口的契约——重复请求安全、计数不会双加。
是测试人员"知道在哪里测"的典型示例。
"""

import pytest


@pytest.mark.smoke
def test_like_increments_count(published_post, second_user):
    """B 点赞 A 的帖子 → 200 且 like_count = 1。"""
    resp = second_user.interaction.like(published_post["id"])
    assert resp.ok, resp.message

    detail = second_user.post.get(published_post["id"])
    assert detail.ok
    assert detail.data["like_count"] == 1
    assert detail.data["liked"] is True


@pytest.mark.idempotent
def test_like_is_idempotent(published_post, second_user):
    """重复点赞两次 → 都返回 200，like_count 仍是 1。

    这是 devflow 的核心契约：依赖 (user_id, post_id) 唯一索引天然防重，
    重复请求是合法语义（"确保已点赞"），而非客户端错误，所以返回 200 而不是 4xx。
    """
    post_id = published_post["id"]

    first = second_user.interaction.like(post_id)
    second = second_user.interaction.like(post_id)

    assert first.ok and second.ok, f"重复点赞应都为 200: {first.message} / {second.message}"

    detail = second_user.post.get(post_id)
    assert detail.data["like_count"] == 1, "重复点赞不能让 like_count 涨成 2"


def test_unlike_decrements_count(published_post, second_user):
    """取消点赞应使 like_count 回到 0 且 liked=False。"""
    post_id = published_post["id"]
    second_user.interaction.like(post_id)

    resp = second_user.interaction.unlike(post_id)
    assert resp.ok

    detail = second_user.post.get(post_id)
    assert detail.data["like_count"] == 0
    assert detail.data["liked"] is False


@pytest.mark.idempotent
def test_unlike_never_liked_is_idempotent(published_post, second_user):
    """取消一个没点过的赞应安静返回 200，不报错。"""
    resp = second_user.interaction.unlike(published_post["id"])
    assert resp.ok, f"取消未点赞应幂等: {resp.status_code} {resp.message}"


def test_like_nonexistent_post_not_found(second_user):
    """对不存在的 post 点赞应返回 404。"""
    resp = second_user.interaction.like(9_999_999_999)
    assert resp.status_code == 404, f"期望 404，实际 {resp.status_code} {resp.message}"


def test_like_after_unlike_recounts(published_post, second_user):
    """点赞 → 取消 → 再点赞，最终 like_count 应回到 1。"""
    post_id = published_post["id"]
    assert second_user.interaction.like(post_id).ok
    assert second_user.interaction.unlike(post_id).ok
    assert second_user.interaction.like(post_id).ok

    detail = second_user.post.get(post_id)
    assert detail.data["like_count"] == 1
    assert detail.data["liked"] is True


@pytest.mark.idempotent
def test_concurrent_like_counts_once(published_post, second_user):
    """同一用户并发点赞同一帖子多次，最终 like_count 必须等于 1。

    这是对 (user_id, post_id) 唯一索引 + ON CONFLICT DO NOTHING 在并发下的压测：
    无论多少请求同时打进来，计数都不能被双加。
    """
    from concurrent.futures import ThreadPoolExecutor

    post_id = published_post["id"]

    # 每个线程用独立 client（共享同一 token），避免 requests.Session 非线程安全问题
    def _like_once():
        client = second_user.http.clone()
        client.set_token(second_user.token)
        from clients.interaction import InteractionClient

        return InteractionClient(client).like(post_id)

    with ThreadPoolExecutor(max_workers=10) as pool:
        results = list(pool.map(lambda _: _like_once(), range(10)))

    assert all(r.status_code in (200, 409) for r in results), (
        "并发点赞每个请求都应被安全处理（200 幂等成功）"
    )

    detail = second_user.post.get(post_id)
    assert detail.data["like_count"] == 1, (
        f"并发点赞后 like_count 必须为 1，实际 {detail.data['like_count']}（计数被双加说明存在并发缺陷）"
    )
