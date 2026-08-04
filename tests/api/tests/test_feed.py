"""Feed 链路：latest / hot / following 三种 feed 的基础契约。

业务规则：
- /api/feed/latest：按时间倒序返回所有公开帖
- /api/feed/hot：按 like*3 + favorite*5 + comment*4 排序（Redis ZSET 或 MySQL 回退）
- /api/feed/following：没关注任何人时 *降级* 返回 latest（FollowService.ListFollowingFeed 内置冷启动逻辑）
"""

import time
import uuid

import pytest

from clients.auth import AuthClient
from clients.base import BaseClient
from clients.post import PostClient
from config import (
    DEFAULT_PASSWORD,
    NOTIFICATION_POLL_INTERVAL,
    NOTIFICATION_POLL_TIMEOUT,
)


def _poll(predicate, *, timeout=NOTIFICATION_POLL_TIMEOUT, interval=NOTIFICATION_POLL_INTERVAL):
    deadline = time.time() + timeout
    last = None
    while time.time() < deadline:
        last = predicate()
        if last:
            return last
        time.sleep(interval)
    return last


def _register_post_client():
    """注册一个新用户，返回其 PostClient（用于构造与现有 fixture 无关的第三方用户）。"""
    http = BaseClient()
    username = f"u_{uuid.uuid4().hex[:10]}"
    nickname = f"n_{username[-6:]}"
    resp = AuthClient(http).register(username, DEFAULT_PASSWORD, nickname)
    assert resp.ok, f"注册失败 {resp.status_code} {resp.message}"
    http.set_token(resp.data["token"])
    return PostClient(http)


@pytest.mark.smoke
def test_feed_latest_contains_recent_post(published_post, registered_user):
    """刚发的帖子应出现在 latest feed 第一页中。"""
    resp = registered_user.feed.latest(limit=50)
    assert resp.ok

    ids = [item["id"] for item in resp.data["items"]]
    assert published_post["id"] in ids, "latest feed 应包含刚发布的帖子"


def test_feed_hot_returns_list_structure(registered_user):
    """hot feed 结构正确即可（实际排序依赖互动数据，不强校验顺序）。"""
    resp = registered_user.feed.hot()
    assert resp.ok
    assert "items" in resp.data
    assert isinstance(resp.data["items"], list)
    assert "has_more" in resp.data


def test_following_feed_cold_start_falls_back_to_latest(published_post, second_user):
    """新用户未关注任何人时，/feed/following 应降级为 latest，避免空 feed。

    这是 devflow 的冷启动设计：FollowService.ListFollowingFeed 检测到无 follow 关系
    时直接走 PostService.ListLatest，保证新用户首屏不空。
    """
    resp = second_user.feed.following(limit=50)
    assert resp.ok

    ids = [item["id"] for item in resp.data["items"]]
    assert published_post["id"] in ids, (
        "未关注任何人时应降级为 latest feed，但响应里看不到 published_post"
    )


def test_latest_feed_pagination_by_cursor(registered_user):
    """快速创建的数据必须能跨页完整遍历，且不重不漏。"""
    created_ids = []
    for i in range(3):
        r = registered_user.post.create(title=f"page-{i}", content="pagination test")
        assert r.ok
        created_ids.append(r.data["id"])

    seen = []
    cursor = None
    for _ in range(100):
        page = registered_user.feed.latest(limit=1, cursor=cursor)
        assert page.ok
        page_ids = [item["id"] for item in page.data["items"]]
        assert not set(page_ids).intersection(seen), "游标翻页不应返回重复帖子"
        seen.extend(page_ids)
        if set(created_ids).issubset(seen):
            break
        assert page.data["has_more"] is True, "目标帖子尚未遍历完却提前结束分页"
        cursor = page.data.get("next_cursor")
        assert cursor, "has_more=true 时必须返回 next_cursor"

    assert set(created_ids).issubset(seen), "同一时间附近创建的帖子在分页中被遗漏"


def test_hot_feed_pagination_uses_score_and_id_cursor(registered_user, second_user):
    """相同热度分的帖子必须能按稳定游标完整翻页。"""
    created_ids = []
    for i in range(3):
        created = registered_user.post.create(
            title=f"hot-page-{i}",
            content="hot pagination test",
        )
        assert created.ok
        created_ids.append(created.data["id"])
        assert second_user.interaction.like(created.data["id"]).ok

    seen = []
    cursor = None
    for _ in range(100):
        page = registered_user.feed.hot(limit=1, cursor=cursor)
        assert page.ok
        page_ids = [item["id"] for item in page.data["items"]]
        assert not set(page_ids).intersection(seen), "热门流翻页不应重复"
        seen.extend(page_ids)
        if set(created_ids).issubset(seen):
            break
        assert page.data["has_more"] is True, "热门帖子尚未遍历完却提前结束分页"
        cursor = page.data.get("next_cursor")
        assert cursor, "热门流 has_more=true 时必须返回 next_cursor"

    assert set(created_ids).issubset(seen), "相同热度分的帖子在热门流分页中被遗漏"


def test_following_feed_only_contains_followees_posts(registered_user, second_user):
    """A 关注 B 后，A 的 following feed 应含 B 的帖、不含未关注用户 C 的帖。"""
    follower = second_user
    followee = registered_user
    stranger_post = _register_post_client()

    b_post = followee.post.create(title="from-b", content="b post").data["id"]
    c_post = stranger_post.create(title="from-c", content="c post").data["id"]

    assert follower.follow.follow(followee.user_id).ok

    def _check():
        feed = follower.feed.following(limit=50)
        if not feed.ok:
            return None
        ids = [item["id"] for item in feed.data["items"]]
        return ids if b_post in ids else None

    ids = _poll(_check)
    assert ids, "following feed 未包含被关注者 B 的帖"
    assert c_post not in ids, "following feed 不应包含未关注用户 C 的帖"


def test_hot_feed_contains_liked_post(published_post, second_user):
    """B 点赞 A 的帖后，hot feed 列表中应能看到该帖。"""
    post_id = published_post["id"]
    assert second_user.interaction.like(post_id).ok

    resp = second_user.feed.hot(limit=50)
    assert resp.ok
    ids = [item["id"] for item in resp.data["items"]]
    assert post_id in ids, "点赞后 hot feed 应包含该帖"
