import React, { useEffect, useMemo, useState } from "react";
import { Bell, Settings, Star } from "lucide-react";
import { Link, useNavigate, useOutletContext } from "react-router-dom";
import { api } from "../api/client";
import { Avatar } from "../components/Avatar";
import { PostCard } from "../components/PostCard";
import { useAuth } from "../state/auth";
import { formatDate } from "../utils/format";
import devflowIcon from "../assets/devflow-icon.png";

const guestUser = {
  id: "guest",
  nickname: "Lin",
  bio: "记录开发日常、项目复盘和技术灵感。"
};
const demoUsers = [
  { id: "demo-liu", nickname: "刘超", bio: "后端架构与性能优化" },
  { id: "demo-yu", nickname: "Yu", bio: "Go / Redis / 微服务" },
  { id: "demo-chen", nickname: "Chen", bio: "云原生和工程效率" }
];
const demoPosts = [
  {
    id: "demo-post-2",
    author_id: "demo-yu",
    author: demoUsers[1],
    title: "移动端首页改成动态优先",
    content: "顶部频道只保留关注、热门、最新，底部只放动态、发布和我的。",
    cover_url: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1200&q=80",
    tags: "React,UI",
    like_count: 8,
    favorite_count: 5,
    comment_count: 1,
    liked: false,
    favorited: true,
    created_at: new Date(Date.now() - 3600 * 1000).toISOString()
  }
];
const demoNotifications = [
  {
    id: "demo-notification-1",
    type: "like",
    actor_id: "demo-liu",
    actor: demoUsers[0],
    post_id: "demo-post-2",
    post: demoPosts[0],
    content: "刘超 点赞了你的动态",
    is_read: false,
    created_at: new Date().toISOString()
  },
  {
    id: "demo-notification-2",
    type: "follow",
    actor_id: "demo-chen",
    actor: demoUsers[2],
    content: "Chen 关注了你",
    is_read: true,
    created_at: new Date(Date.now() - 7200 * 1000).toISOString()
  }
];

export function MobileMePage() {
  const { user, updateMe } = useAuth();
  const { openSettings, profileStats } = useOutletContext();
  const [activePreview, setActivePreview] = useState("");
  const [favorites, setFavorites] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [previewError, setPreviewError] = useState("");
  const [previewLoading, setPreviewLoading] = useState(false);
  const [statPreview, setStatPreview] = useState("");
  const [statItems, setStatItems] = useState([]);
  const [statError, setStatError] = useState("");
  const [statLoading, setStatLoading] = useState(false);
  const profile = user || guestUser;
  const stats = useMemo(
    () => ({
      posts: user ? (profileStats?.posts ?? 0) : demoPosts.length,
      following: user ? (profileStats?.following ?? 0) : demoUsers.length,
      followers: user ? (profileStats?.followers ?? 0) : 2
    }),
    [profileStats, user]
  );
  const statTitles = {
    following: "关注",
    followers: "粉丝",
    posts: "动态"
  };

  useEffect(() => {
    let alive = true;
    setPreviewError("");
    if (!activePreview) {
      setPreviewLoading(false);
      return undefined;
    }
    if (!user) {
      setFavorites(demoPosts);
      setNotifications(demoNotifications);
      setPreviewLoading(false);
      return undefined;
    }

    setPreviewLoading(true);
    const request = activePreview === "favorites" ? api.favorites({ limit: 3 }) : api.notifications({ limit: 4 });
    request
      .then((result) => {
        if (!alive) {
          return;
        }
        if (activePreview === "favorites") {
          setFavorites(result.items || []);
        } else {
          setNotifications(result.items || []);
        }
      })
      .catch((error) => {
        if (alive) {
          setPreviewError(error.message || "请求失败");
        }
      })
      .finally(() => {
        if (alive) {
          setPreviewLoading(false);
        }
      });

    return () => {
      alive = false;
    };
  }, [activePreview, user]);

  useEffect(() => {
    let alive = true;
    setStatError("");
    setStatItems([]);
    if (!statPreview) {
      setStatLoading(false);
      return undefined;
    }
    if (!user) {
      const demoItems =
        statPreview === "posts" ? demoPosts : statPreview === "following" ? demoUsers : demoUsers.slice(0, 2);
      setStatItems(demoItems);
      setStatLoading(false);
      return undefined;
    }

    setStatLoading(true);
    const request =
      statPreview === "posts"
        ? api.userPosts(user.id, { limit: 5 })
        : statPreview === "following"
          ? api.followingUsers(user.id, { limit: 8 })
          : api.followers(user.id, { limit: 8 });

    request
      .then((result) => {
        if (alive) {
          setStatItems(result.items || []);
        }
      })
      .catch((error) => {
        if (alive) {
          setStatError(error.message || "请求失败");
        }
      })
      .finally(() => {
        if (alive) {
          setStatLoading(false);
        }
      });

    return () => {
      alive = false;
    };
  }, [statPreview, user]);

  function showPreview(type) {
    setActivePreview((current) => (current === type ? "" : type));
  }

  function showStat(type) {
    setStatPreview(type);
  }

  async function markPreviewNotificationsRead() {
    if (user) {
      await api.markAllRead().catch(() => {});
    }
    setNotifications((current) => current.map((item) => ({ ...item, is_read: true })));
  }

  return (
    <main className="mobile-me-page">
      <header className="mobile-page-topbar">
        <div>
          <span className="mobile-brand-mark">
            <img src={devflowIcon} alt="" />
          </span>
          <strong>DevFlow</strong>
        </div>
        <button type="button" onClick={openSettings} aria-label="设置">
          <Settings size={20} />
        </button>
      </header>

      <section className="mobile-me-card">
        <div className="mobile-me-identity">
          <Avatar user={user || null} label={profile.nickname} className="avatar" />
          <div>
            <h1>{profile.nickname}</h1>
            <span>DevFlow 开发者</span>
          </div>
        </div>
        <MobileBioEditor user={user} profile={profile} updateMe={updateMe} />
        <div className="mobile-me-stats">
          <button type="button" onClick={() => showStat("following")}>
            <strong>{stats.following}</strong>
            <span>关注</span>
          </button>
          <button type="button" onClick={() => showStat("followers")}>
            <strong>{stats.followers}</strong>
            <span>粉丝</span>
          </button>
          <button type="button" onClick={() => showStat("posts")}>
            <strong>{stats.posts}</strong>
            <span>动态</span>
          </button>
        </div>
      </section>

      <section className="mobile-me-shortcuts" aria-label="我的快捷入口">
        <button
          type="button"
          className={activePreview === "favorites" ? "active" : ""}
          onClick={() => showPreview("favorites")}
        >
          <Star size={20} />
          <span>收藏</span>
        </button>
        <button
          type="button"
          className={activePreview === "notifications" ? "active" : ""}
          onClick={() => showPreview("notifications")}
        >
          <Bell size={20} />
          <span>通知</span>
        </button>
      </section>

      {activePreview ? (
        <PreviewPanel
          type={activePreview}
          loading={previewLoading}
          error={previewError}
          favorites={favorites}
          notifications={notifications}
          onMarkNotificationsRead={markPreviewNotificationsRead}
        />
      ) : null}

      {statPreview ? (
        <StatPreviewModal
          title={statTitles[statPreview]}
          type={statPreview}
          items={statItems}
          loading={statLoading}
          error={statError}
          onClose={() => setStatPreview("")}
        />
      ) : null}
    </main>
  );
}

function MobileBioEditor({ user, profile, updateMe }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(profile.bio || "");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setDraft(profile.bio || "");
  }, [profile.bio]);

  async function save(event) {
    event.preventDefault();
    if (!user || saving) {
      setEditing(false);
      return;
    }
    setSaving(true);
    try {
      await updateMe({ bio: draft });
      setEditing(false);
    } finally {
      setSaving(false);
    }
  }

  if (!editing) {
    return (
      <button className="mobile-bio-text" type="button" onClick={() => setEditing(true)}>
        {profile.bio || "还没有填写简介。"}
      </button>
    );
  }

  return (
    <form className="mobile-bio-editor" onSubmit={save}>
      <input value={draft} onChange={(event) => setDraft(event.target.value)} maxLength={255} placeholder="写一句你的技术方向" autoFocus />
      <button type="submit" disabled={saving || !user}>
        {user ? "保存" : "预览"}
      </button>
    </form>
  );
}

function StatPreviewModal({ title, type, items, loading, error, onClose }) {
  return (
    <div className="mobile-stat-backdrop" role="dialog" aria-modal="true" aria-label={title} onClick={onClose}>
      <section className="mobile-stat-modal" onClick={(event) => event.stopPropagation()}>
        <header>
          <h2>{title}</h2>
          <button type="button" onClick={onClose} aria-label="关闭">
            关闭
          </button>
        </header>
        {loading ? <p className="mobile-me-empty">加载中...</p> : null}
        {error ? <p className="mobile-me-empty">{error}</p> : null}
        {!loading && !error && type === "posts" ? <PostStatPreview items={items} /> : null}
        {!loading && !error && type !== "posts" ? <UserStatPreview items={items} /> : null}
      </section>
    </div>
  );
}

function UserStatPreview({ items }) {
  if (items.length === 0) {
    return <p className="mobile-me-empty">暂无数据。</p>;
  }

  return (
    <div className="mobile-stat-list">
      {items.map((item) => (
        <Link to={`/user/${item.id}`} key={item.id}>
          <Avatar user={item} label={item.nickname || item.username || "D"} className="tiny-avatar" />
          <span>
            <strong>{item.nickname || item.username || `用户 #${item.id}`}</strong>
            <em>{item.bio || "这个开发者还没有填写简介。"}</em>
          </span>
        </Link>
      ))}
    </div>
  );
}

function PostStatPreview({ items }) {
  if (items.length === 0) {
    return <p className="mobile-me-empty">还没有动态。</p>;
  }

  return (
    <div className="mobile-stat-posts">
      {items.map((post) => (
        <Link to={`/post/${post.id}`} key={post.id}>
          <strong>{post.title}</strong>
          <span>{post.content}</span>
        </Link>
      ))}
    </div>
  );
}

function PreviewPanel({ type, loading, error, favorites, notifications, onMarkNotificationsRead }) {
  const title = type === "favorites" ? "已收藏的动态" : "最新消息";
  const unreadCount = notifications.filter((item) => !item.is_read).length;

  return (
    <section className="mobile-me-preview">
      <header>
        <h2>{type === "notifications" ? `${title} ${unreadCount}` : title}</h2>
        {type === "notifications" ? (
          <button type="button" onClick={onMarkNotificationsRead} disabled={unreadCount === 0}>
            全部已读
          </button>
        ) : null}
      </header>
      {loading ? <p className="mobile-me-empty">加载中...</p> : null}
      {error ? <p className="mobile-me-empty">{error}</p> : null}
      {!loading && !error && type === "favorites" ? <FavoritePreview items={favorites} /> : null}
      {!loading && !error && type === "notifications" ? <NotificationPreview items={notifications} /> : null}
    </section>
  );
}

function FavoritePreview({ items }) {
  if (items.length === 0) {
    return <p className="mobile-me-empty">还没有收藏的动态。</p>;
  }

  return (
    <div className="mobile-favorite-preview">
      {items.map((post) => (
        <PostCard key={post.id} post={post} compact />
      ))}
    </div>
  );
}

function NotificationPreview({ items }) {
  const navigate = useNavigate();

  if (items.length === 0) {
    return <p className="mobile-me-empty">暂时没有消息。</p>;
  }

  function notificationText(item) {
    const actor = item.actor?.nickname || item.actor?.username || `用户 #${item.actor_id}`;
    const postTitle = item.post?.title ? `《${item.post.title}》` : "你的动态";
    if (item.type === "follow") {
      return `${actor} 关注了你`;
    }
    if (item.type === "like") {
      return `${actor} 点赞了 ${postTitle}`;
    }
    if (item.type === "favorite") {
      return `${actor} 收藏了 ${postTitle}`;
    }
    if (item.type === "comment") {
      return `${actor} 评论了 ${postTitle}`;
    }
    return item.content;
  }

  async function openNotification(item) {
    if (String(item.id).startsWith("demo-")) {
      return;
    }
    if (!item.is_read) {
      await api.markRead(item.id).catch(() => {});
    }
    if (item.post_id) {
      navigate(`/post/${item.post_id}`);
      return;
    }
    if (item.actor_id) {
      navigate(`/user/${item.actor_id}`);
    }
  }

  return (
    <div className="mobile-notification-preview">
      {items.map((item) => (
        <button type="button" key={item.id} className={item.is_read ? "read" : ""} onClick={() => openNotification(item)}>
          <Avatar user={item.actor} label={item.actor?.nickname || item.actor?.username || "D"} className="tiny-avatar" />
          <span>
            <strong>{notificationText(item)}</strong>
            <em>{formatDate(item.created_at)}</em>
          </span>
        </button>
      ))}
    </div>
  );
}
