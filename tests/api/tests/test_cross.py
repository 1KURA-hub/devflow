"""跨链路用例：验证多模块协作的端到端契约。

这两条是 devflow 最值得测的"系统性"场景：
1. 关注后 fan-out：B 关注 A 后，A 发的帖应进 B 的 following feed（验证 inbox 写扩散链路）
2. 点赞触发通知：B 点赞 A 的帖，A 应收到 like 类型通知（验证通知异步链路）

异步说明：
- 通知走 MQ，本地若无 RabbitMQ 会同步降级；CI 中若开 MQ 则需要轮询等待。
- 用 poll_until() 兼容两种部署。
"""

import time

import pytest

from config import NOTIFICATION_POLL_INTERVAL, NOTIFICATION_POLL_TIMEOUT


def poll_until(predicate, *, timeout: float, interval: float):
    """轮询直到 predicate 返回真值或超时。返回最后一次结果。"""
    deadline = time.time() + timeout
    last = None
    while time.time() < deadline:
        last = predicate()
        if last:
            return last
        time.sleep(interval)
    return last


@pytest.mark.cross
def test_follower_can_see_followed_user_post(registered_user, second_user):
    """B 关注 A → A 发帖 → B 的 following feed 能看到这条帖。

    覆盖链路：follow.Add → feed_inbox 回填 → post.create 触发 fan-out → feed/following 命中 inbox。
    """
    follow_resp = second_user.follow.follow(registered_user.user_id)
    assert follow_resp.ok

    create = registered_user.post.create(title="cross-chain", content="fan-out test")
    assert create.ok
    post_id = create.data["id"]

    def _check():
        feed = second_user.feed.following(limit=50)
        if not feed.ok:
            return False
        ids = [item["id"] for item in feed.data["items"]]
        return post_id in ids

    matched = poll_until(
        _check,
        timeout=NOTIFICATION_POLL_TIMEOUT,
        interval=NOTIFICATION_POLL_INTERVAL,
    )
    assert matched, "B 应在 following feed 中看到 A 新发的帖（fan-out 失败或 inbox 未命中）"


@pytest.mark.cross
def test_like_triggers_like_notification(published_post, second_user):
    """B 点赞 A 的帖子 → A 应收到 type=like 的通知。

    覆盖链路：interaction.Like → notification.Create → MQ 或同步落库 → notifications.List。
    """
    author = published_post["author"]
    like_resp = second_user.interaction.like(published_post["id"])
    assert like_resp.ok

    def _check():
        resp = author.notification.list()
        if not resp.ok:
            return None
        for item in resp.data.get("items", []):
            if (
                item.get("type") == "like"
                and item.get("actor_id") == second_user.user_id
                and item.get("post_id") == published_post["id"]
            ):
                return item
        return None

    found = poll_until(
        _check,
        timeout=NOTIFICATION_POLL_TIMEOUT,
        interval=NOTIFICATION_POLL_INTERVAL,
    )
    assert found, "未在超时窗口内看到 like 通知，可能 MQ 未消费或链路断开"


def test_post_supports_like_favorite_comment_together(published_post, second_user):
    """同一帖被点赞 + 收藏 + 评论后，三个计数都应为 1。"""
    post_id = published_post["id"]
    assert second_user.interaction.like(post_id).ok
    assert second_user.interaction.favorite(post_id).ok
    assert second_user.comment.create(post_id, "nice").ok

    detail = second_user.post.get(post_id)
    assert detail.data["like_count"] == 1
    assert detail.data["favorite_count"] == 1
    assert detail.data["comment_count"] == 1


@pytest.mark.cross
def test_follow_triggers_follow_notification(registered_user, second_user):
    """B 关注 A → A 应收到 type=follow 的通知。"""
    author = registered_user
    follower = second_user
    assert follower.follow.follow(author.user_id).ok

    def _check():
        resp = author.notification.list()
        if not resp.ok:
            return None
        for item in resp.data.get("items", []):
            if (
                item.get("type") == "follow"
                and item.get("actor_id") == follower.user_id
            ):
                return item
        return None

    found = poll_until(
        _check,
        timeout=NOTIFICATION_POLL_TIMEOUT,
        interval=NOTIFICATION_POLL_INTERVAL,
    )
    assert found, "未在超时窗口内看到 follow 通知"


@pytest.mark.cross
def test_comment_triggers_comment_notification(published_post, second_user):
    """B 评论 A 的帖 → A 应收到 type=comment 的通知。"""
    author = published_post["author"]
    assert second_user.comment.create(published_post["id"], "nice").ok

    def _check():
        resp = author.notification.list()
        if not resp.ok:
            return None
        for item in resp.data.get("items", []):
            if (
                item.get("type") == "comment"
                and item.get("actor_id") == second_user.user_id
                and item.get("post_id") == published_post["id"]
            ):
                return item
        return None

    found = poll_until(
        _check,
        timeout=NOTIFICATION_POLL_TIMEOUT,
        interval=NOTIFICATION_POLL_INTERVAL,
    )
    assert found, "未在超时窗口内看到 comment 通知"


@pytest.mark.cross
def test_follow_backfills_followee_existing_posts(registered_user, second_user):
    """A 关注 B 时，B 已有的旧帖应能在 A 的 following feed 中出现（inbox 回填）。"""
    author = registered_user
    follower = second_user

    old = author.post.create(title="old", content="existed before follow")
    assert old.ok
    old_id = old.data["id"]

    assert follower.follow.follow(author.user_id).ok

    def _check():
        feed = follower.feed.following(limit=50)
        if not feed.ok:
            return False
        return old_id in [item["id"] for item in feed.data["items"]]

    matched = poll_until(
        _check,
        timeout=NOTIFICATION_POLL_TIMEOUT,
        interval=NOTIFICATION_POLL_INTERVAL,
    )
    assert matched, "关注后未看到被关注者的旧帖（inbox 回填可能失败）"
