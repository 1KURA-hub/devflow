import React, { useEffect, useState, useCallback } from "react";
import { useParams } from "react-router-dom";
import { api } from "../api/client";
import { FeedList } from "../components/FeedList";
import { useAuth } from "../state/auth";

export function ProfilePage() {
  const { id } = useParams();
  const { user } = useAuth();
  const [stats, setStats] = useState({ followers: 0, following: 0 });
  const [followed, setFollowed] = useState(false);
  const loader = useCallback((params) => api.userPosts(id, params), [id]);

  useEffect(() => {
    Promise.all([api.followers(id, { limit: 50 }), api.followingUsers(id, { limit: 50 })])
      .then(([followers, following]) => {
        setStats({
          followers: followers.items.length,
          following: following.items.length
        });
      })
      .catch(() => {
        setStats({ followers: 0, following: 0 });
      });
  }, [id]);

  async function toggleFollow() {
    if (followed) {
      await api.unfollow(id);
    } else {
      await api.follow(id);
    }
    setFollowed((value) => !value);
    setStats((current) => ({
      ...current,
      followers: current.followers + (followed ? -1 : 1)
    }));
  }

  return (
    <div className="profile-layout">
      <section className="surface profile-header">
        <div className="avatar">#{id}</div>
        <div>
          <p className="eyebrow">开发者主页</p>
          <h1>用户 #{id}</h1>
          <p>记录公开的工程进展与技术动态。</p>
        </div>
        <div className="profile-stats">
          <span>
            <strong>{stats.followers}</strong> 粉丝
          </span>
          <span>
            <strong>{stats.following}</strong> 关注
          </span>
        </div>
        {user && String(user.id) !== id ? (
          <button className={followed ? "ghost-button" : "primary-button"} onClick={toggleFollow}>
            {followed ? "取消关注" : "关注"}
          </button>
        ) : null}
      </section>
      <FeedList loader={loader} emptyTitle="还没有动态" emptyText="这个开发者暂时没有公开内容。" />
    </div>
  );
}
