import React, { useEffect, useMemo, useState, useCallback } from "react";
import { Bell, Star, UserRound } from "lucide-react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { Avatar } from "../components/Avatar";
import { FeedList } from "../components/FeedList";
import { useAuth } from "../state/auth";

export function ProfilePage() {
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const { user } = useAuth();
  const [stats, setStats] = useState({ posts: 0, followers: 0, following: 0 });
  const [lists, setLists] = useState({ followers: [], following: [] });
  const [activeTab, setActiveTab] = useState(searchParams.get("tab") || "posts");
  const [loadedProfile, setLoadedProfile] = useState(null);
  const [followed, setFollowed] = useState(false);
  const [followSubmitting, setFollowSubmitting] = useState(false);
  const [loadedProfileID, setLoadedProfileID] = useState("");
  const currentUserID = user?.id;
  const loader = useCallback((params) => api.userPosts(id, params), [id]);
  const profileUser = useMemo(() => {
    if (user && String(user.id) === id) {
      return user;
    }
    if (loadedProfileID !== id) {
      return null;
    }
    return loadedProfile;
  }, [id, loadedProfile, loadedProfileID, user]);
  const profileName = profileUser?.nickname || `用户 #${id}`;
  const profileBio = profileUser?.bio || "记录公开的工程进展与技术动态。";
  const isOwnProfile = user && String(user.id) === id;

  useEffect(() => {
    setActiveTab(searchParams.get("tab") || "posts");
  }, [searchParams]);

  useEffect(() => {
    let active = true;
    setLoadedProfileID("");
    setLoadedProfile(null);
    setStats({ posts: 0, followers: 0, following: 0 });
    setLists({ followers: [], following: [] });
    setFollowed(false);

    api.userProfile(id)
      .then((profile) => {
        if (!active) {
          return;
        }
        setLoadedProfile(profile.user);
        setStats(profile.stats);
        setLoadedProfileID(id);
      })
      .catch(() => {
        if (!active) {
          return;
        }
        setLoadedProfile(null);
        setStats({ posts: 0, followers: 0, following: 0 });
      });

    const followStateRequest =
      currentUserID && String(currentUserID) !== id
        ? api.followState(id).catch(() => ({ followed: false }))
        : Promise.resolve({ followed: false });
    Promise.all([
      api.followers(id, { limit: 50 }),
      api.followingUsers(id, { limit: 50 }),
      followStateRequest
    ])
      .then(([followers, following, followState]) => {
        if (!active) {
          return;
        }
        setLists({
          followers: followers.items,
          following: following.items
        });
        setFollowed(Boolean(followState.followed));
      })
      .catch(() => {
        if (!active) {
          return;
        }
        setLists({ followers: [], following: [] });
        setFollowed(false);
      });
    return () => {
      active = false;
    };
  }, [currentUserID, id]);

  async function toggleFollow() {
    if (followSubmitting) {
      return;
    }
    const nextFollowed = !followed;
    setFollowSubmitting(true);
    try {
      if (nextFollowed) {
        await api.follow(id);
      } else {
        await api.unfollow(id);
      }
      setFollowed(nextFollowed);
      setStats((current) => ({
        ...current,
        followers: Math.max(0, current.followers + (nextFollowed ? 1 : -1))
      }));
      window.dispatchEvent(new CustomEvent("devflow:profile-stats-refresh"));
    } finally {
      setFollowSubmitting(false);
    }
  }

  return (
    <div className={`profile-layout ${isOwnProfile ? "own-profile" : ""}`}>
      {isOwnProfile ? (
        <section className="surface mobile-profile-hub">
          <div className="mobile-profile-card">
            <Avatar user={profileUser} label={profileName} className="avatar" />
            <div>
              <p className="eyebrow">我的</p>
              <h1>{profileName}</h1>
              <p>{profileBio}</p>
            </div>
          </div>
          <div className="mobile-profile-stats">
            <button className={activeTab === "posts" ? "active" : ""} type="button" onClick={() => setActiveTab("posts")}>
              <strong>{stats.posts}</strong>
              <span>动态</span>
            </button>
            <button className={activeTab === "following" ? "active" : ""} type="button" onClick={() => setActiveTab("following")}>
              <strong>{stats.following}</strong>
              <span>关注</span>
            </button>
            <button className={activeTab === "followers" ? "active" : ""} type="button" onClick={() => setActiveTab("followers")}>
              <strong>{stats.followers}</strong>
              <span>粉丝</span>
            </button>
          </div>
          <div className="mobile-profile-actions">
            <button type="button" onClick={() => setActiveTab("following")}>
              <UserRound size={18} />
              关注
            </button>
            <Link to="/favorites">
              <Star size={18} />
              收藏
            </Link>
            <Link to="/notifications">
              <Bell size={18} />
              通知
            </Link>
          </div>
        </section>
      ) : null}
      <section className="surface profile-header" aria-label="开发者资料">
        <Avatar user={profileUser} label={profileName} className="avatar" />
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
        {user && String(user.id) !== id && loadedProfileID === id ? (
          <button
            className={followed ? "ghost-button" : "primary-button"}
            aria-label={`${followed ? "取消关注" : "关注"} ${profileName}`}
            onClick={toggleFollow}
            disabled={followSubmitting}
          >
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
          <Avatar user={item} className="profile-avatar" />
          <div>
            <strong>{item.nickname}</strong>
            <span>{item.bio || "这个开发者还没有填写简介。"}</span>
          </div>
        </Link>
      ))}
    </div>
  );
}
