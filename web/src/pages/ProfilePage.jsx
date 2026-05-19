import React, { useEffect, useMemo, useState, useCallback } from "react";
import { Link, useParams } from "react-router-dom";
import { api } from "../api/client";
import { FeedList } from "../components/FeedList";
import { useAuth } from "../state/auth";

export function ProfilePage() {
  const { id } = useParams();
  const { user } = useAuth();
  const [stats, setStats] = useState({ posts: 0, followers: 0, following: 0 });
  const [lists, setLists] = useState({ followers: [], following: [] });
  const [activeTab, setActiveTab] = useState("posts");
  const [followed, setFollowed] = useState(false);
  const loader = useCallback((params) => api.userPosts(id, params), [id]);
  const profileUser = useMemo(() => {
    if (user && String(user.id) === id) {
      return user;
    }
    return [...lists.followers, ...lists.following].find((item) => String(item.id) === id) || null;
  }, [id, lists.followers, lists.following, user]);
  const profileName = profileUser?.nickname || `用户 #${id}`;
  const profileBio = profileUser?.bio || "记录公开的工程进展与技术动态。";

  useEffect(() => {
    setActiveTab("posts");
    Promise.all([api.userPosts(id, { limit: 50 }), api.followers(id, { limit: 50 }), api.followingUsers(id, { limit: 50 })])
      .then(([posts, followers, following]) => {
        setStats({
          posts: posts.items.length,
          followers: followers.items.length,
          following: following.items.length
        });
        setLists({
          followers: followers.items,
          following: following.items
        });
      })
      .catch(() => {
        setStats({ posts: 0, followers: 0, following: 0 });
        setLists({ followers: [], following: [] });
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
        <div className="avatar">{profileName.slice(0, 1)}</div>
        <div>
          <p className="eyebrow">开发者主页</p>
          <h1>{profileName}</h1>
          <p>{profileBio}</p>
        </div>
        <div className="profile-stats">
          <button className={activeTab === "posts" ? "active" : ""} type="button" onClick={() => setActiveTab("posts")}>
            <strong>{stats.posts}</strong> 动态
          </button>
          <button className={activeTab === "following" ? "active" : ""} type="button" onClick={() => setActiveTab("following")}>
            <strong>{stats.following}</strong> 关注
          </button>
          <button className={activeTab === "followers" ? "active" : ""} type="button" onClick={() => setActiveTab("followers")}>
            <strong>{stats.followers}</strong> 粉丝
          </button>
        </div>
        {user && String(user.id) !== id ? (
          <button className={followed ? "ghost-button" : "primary-button"} onClick={toggleFollow}>
            {followed ? "取消关注" : "关注"}
          </button>
        ) : null}
      </section>
      {activeTab === "posts" ? (
        <FeedList loader={loader} emptyTitle="还没有动态" emptyText="这个开发者暂时没有公开内容。" />
      ) : (
        <UserList
          items={activeTab === "following" ? lists.following : lists.followers}
          emptyText={activeTab === "following" ? "还没有关注任何开发者。" : "还没有粉丝。"}
        />
      )}
    </div>
  );
}

function UserList({ items, emptyText }) {
  if (items.length === 0) {
    return (
      <div className="surface empty-state">
        <h3>列表为空</h3>
        <p>{emptyText}</p>
      </div>
    );
  }

  return (
    <div className="surface user-list-panel">
      {items.map((item) => (
        <Link className="user-list-item" key={item.id} to={`/user/${item.id}`}>
          <div className="profile-avatar">{item.nickname.slice(0, 1)}</div>
          <div>
            <strong>{item.nickname}</strong>
            <span>{item.bio || "这个开发者还没有填写简介。"}</span>
          </div>
        </Link>
      ))}
    </div>
  );
}
