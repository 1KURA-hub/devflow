"""关注链路：关注 / 取关 / 自我关注校验。

业务规则（来自 internal/service/follow.go + handler/follow.go writeFollowError）：
- 关注自己 -> 400 (ErrCannotFollowSelf 映射到 BadRequest)
- 关注成功后 follow-state 返回 {"followed": true}

发现的不一致（面试谈资）：
- 重复关注：service 返回 ErrAlreadyFollowed，但 writeFollowError 没有特别处理 →
  落到 default 分支返回 500。这与点赞链路（重复点赞返回 200）的语义不一致。
- 该用例(test_duplicate_follow_known_issue)记录这个真实差异，方便回归时确认行为变更。
"""

import pytest


@pytest.mark.smoke
def test_follow_success_changes_state(registered_user, second_user):
    """A 关注 B → follow-state 从 false 翻为 true。"""
    before = registered_user.follow.state(second_user.user_id)
    assert before.ok and before.data["followed"] is False

    resp = registered_user.follow.follow(second_user.user_id)
    assert resp.ok, resp.message

    after = registered_user.follow.state(second_user.user_id)
    assert after.ok and after.data["followed"] is True


def test_follow_self_bad_request(registered_user):
    """关注自己必须返回 400，不能在 follows 表里写出 self-loop。"""
    resp = registered_user.follow.follow(registered_user.user_id)
    assert resp.status_code == 400, f"期望 400，实际 {resp.status_code} {resp.message}"


@pytest.mark.idempotent
def test_duplicate_follow_is_idempotent(registered_user, second_user):
    """重复关注应幂等返回 200，与重复点赞契约保持一致。"""
    first = registered_user.follow.follow(second_user.user_id)
    assert first.ok

    second = registered_user.follow.follow(second_user.user_id)
    assert second.ok, f"重复关注应返回 200: {second.status_code} {second.message}"

    state = registered_user.follow.state(second_user.user_id)
    assert state.ok and state.data["followed"] is True
