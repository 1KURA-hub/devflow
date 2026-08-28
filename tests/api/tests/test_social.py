"""关注、关注流和通知：用少量用例展示跨用户业务链路。"""

import time

import pytest

from config import NOTIFICATION_POLL_INTERVAL, NOTIFICATION_POLL_TIMEOUT


def poll_until(predicate):
    """在限定时间内轮询异步结果，避免写死长时间 sleep。"""
    deadline = time.time() + NOTIFICATION_POLL_TIMEOUT
    last_result = None
    while time.time() < deadline:
        last_result = predicate()
        if last_result:
            return last_result
        time.sleep(NOTIFICATION_POLL_INTERVAL)
    return last_result


@pytest.mark.smoke
def test_follow_updates_state_and_user_lists(registered_user, second_user):
    before = registered_user.follow.state(second_user.user_id)
    assert before.ok, f"查询关注状态失败: {before.status_code} {before.message}"
    assert before.data["followed"] is False

    followed = registered_user.follow.follow(second_user.user_id)
    assert followed.ok, f"关注失败: {followed.status_code} {followed.message}"

    after = registered_user.follow.state(second_user.user_id)
    assert after.ok, f"查询关注状态失败: {after.status_code} {after.message}"
    assert after.data["followed"] is True

    following = registered_user.follow.list_following(registered_user.user_id)
    assert following.ok, f"查询关注列表失败: {following.status_code} {following.message}"
    assert second_user.user_id in [item["id"] for item in following.data["items"]]

    followers = second_user.follow.list_followers(second_user.user_id)
    assert followers.ok, f"查询粉丝列表失败: {followers.status_code} {followers.message}"
    assert registered_user.user_id in [item["id"] for item in followers.data["items"]]


def test_follow_self_bad_request(registered_user):
    resp = registered_user.follow.follow(registered_user.user_id)
    assert resp.status_code == 400


@pytest.mark.cross
def test_following_feed_contains_followee_new_post(registered_user, second_user):
    follower = second_user
    author = registered_user
    followed = follower.follow.follow(author.user_id)
    assert followed.ok, f"关注失败: {followed.status_code} {followed.message}"

    created = author.post.create(title="feed-post", content="feed test")
    assert created.ok, f"发帖失败: {created.status_code} {created.message}"
    post_id = created.data["id"]

    def find_post():
        feed = follower.feed.following(limit=50)
        if not feed.ok:
            return None
        ids = [item["id"] for item in feed.data["items"]]
        return post_id if post_id in ids else None

    try:
        assert poll_until(find_post), "关注流未出现被关注者的新帖子"
    finally:
        author.post.delete(post_id)


@pytest.mark.cross
def test_like_creates_notification_and_can_mark_all_read(published_post, second_user):
    author = published_post["author"]
    post_id = published_post["id"]
    liked = second_user.interaction.like(post_id)
    assert liked.ok, f"点赞失败: {liked.status_code} {liked.message}"

    def find_notification():
        resp = author.notification.list()
        if not resp.ok:
            return None
        return next(
            (
                item
                for item in resp.data.get("items", [])
                if item.get("type") == "like"
                and item.get("actor_id") == second_user.user_id
                and item.get("post_id") == post_id
            ),
            None,
        )

    notification = poll_until(find_notification)
    assert notification, "超时后仍未收到点赞通知"
    assert notification["is_read"] is False

    unread = author.notification.unread_count()
    assert unread.ok, f"查询未读数失败: {unread.status_code} {unread.message}"
    assert unread.data["unread_count"] >= 1

    marked = author.notification.mark_all_read()
    assert marked.ok, f"全部已读失败: {marked.status_code} {marked.message}"

    after = author.notification.unread_count()
    assert after.ok, f"查询未读数失败: {after.status_code} {after.message}"
    assert after.data["unread_count"] == 0
