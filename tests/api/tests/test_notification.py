"""通知中心：未读数 + 标记已读。

业务规则（来自 internal/handler/notification.go + service/notification.go）：
- 点赞/评论会给被操作者生成一条通知；通知可能走 MQ 异步，需要轮询等待
- GET /notifications/unread-count 返回 {"unread_count": n}
- POST /notifications/:id/read 把单条置为已读，unread_count 相应减少
- POST /notifications/read-all 清空未读

异步说明：本地无 MQ 时同步落库；CI 开 MQ 时需轮询，故统一用 poll_until。
"""

import time

import pytest

from config import NOTIFICATION_POLL_INTERVAL, NOTIFICATION_POLL_TIMEOUT


def poll_until(predicate, *, timeout=NOTIFICATION_POLL_TIMEOUT, interval=NOTIFICATION_POLL_INTERVAL):
    deadline = time.time() + timeout
    last = None
    while time.time() < deadline:
        last = predicate()
        if last:
            return last
        time.sleep(interval)
    return last


def _unread(user) -> int:
    resp = user.notification.unread_count()
    assert resp.ok, resp.message
    return resp.data["unread_count"]


@pytest.mark.cross
def test_unread_count_increases_after_like(published_post, second_user):
    """B 点赞 A 的帖 → A 的未读通知数应至少变为 1。"""
    author = published_post["author"]
    assert _unread(author) == 0, "初始未读数应为 0"

    assert second_user.interaction.like(published_post["id"]).ok

    got = poll_until(lambda: _unread(author) >= 1)
    assert got, "点赞后作者的未读通知数应 >= 1"


@pytest.mark.cross
def test_mark_one_read_decrements_unread(published_post, second_user):
    """标记单条通知已读后，未读数应减少。"""
    author = published_post["author"]
    assert second_user.interaction.like(published_post["id"]).ok

    # 等通知落库并拿到通知 id
    notif = poll_until(lambda: _latest_notification(author))
    assert notif, "未在超时内拿到点赞通知"
    before = _unread(author)
    assert before >= 1

    resp = author.notification.mark_read(notif["id"])
    assert resp.ok, resp.message

    after = _unread(author)
    assert after == before - 1, f"标记一条已读后未读数应减 1：{before} -> {after}"


@pytest.mark.cross
def test_mark_all_read_clears_unread(published_post, second_user):
    """全部已读后未读数应归零。"""
    author = published_post["author"]
    assert second_user.interaction.like(published_post["id"]).ok
    assert poll_until(lambda: _unread(author) >= 1), "点赞后应有未读"

    resp = author.notification.mark_all_read()
    assert resp.ok, resp.message

    assert _unread(author) == 0, "read-all 后未读数应为 0"


def _latest_notification(user):
    resp = user.notification.list()
    if not resp.ok:
        return None
    items = resp.data.get("items", [])
    return items[0] if items else None
