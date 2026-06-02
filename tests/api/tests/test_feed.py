"""Feed 链路：latest / hot / following 三种 feed 的基础契约。

业务规则：
- /api/feed/latest：按时间倒序返回所有公开帖
- /api/feed/hot：按 like*3 + favorite*5 + comment*4 排序（Redis ZSET 或 MySQL 回退）
- /api/feed/following：没关注任何人时 *降级* 返回 latest（FollowService.ListFollowingFeed 内置冷启动逻辑）
"""

import pytest


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
    """limit=1 取第一页拿到 next_cursor，再用 cursor 取下一页，两页 id 不重复。

    覆盖游标分页契约：has_more=true 时返回 next_cursor，按其翻页应拿到更旧的帖。
    """
    # 至少造 3 篇帖，保证有多页
    created_ids = []
    for i in range(3):
        r = registered_user.post.create(title=f"page-{i}", content="pagination test")
        assert r.ok
        created_ids.append(r.data["id"])

    first = registered_user.feed.latest(limit=1)
    assert first.ok
    assert first.data["has_more"] is True, "limit=1 且帖子充足时应 has_more=true"
    assert first.data.get("next_cursor"), "has_more=true 时应返回 next_cursor"
    assert len(first.data["items"]) == 1
    first_id = first.data["items"][0]["id"]

    cursor = first.data["next_cursor"]
    second = registered_user.feed.latest(limit=1, cursor=cursor)
    assert second.ok
    assert len(second.data["items"]) == 1
    second_id = second.data["items"][0]["id"]

    assert first_id != second_id, "翻页后不应再次返回上一页的同一条帖子"
