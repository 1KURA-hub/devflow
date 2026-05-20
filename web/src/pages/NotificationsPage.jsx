import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { Avatar } from "../components/Avatar";
import { formatDate } from "../utils/format";

export function NotificationsPage() {
  const navigate = useNavigate();
  const [items, setItems] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .notifications({})
      .then((result) => setItems(result.items))
      .catch((nextError) => setError(nextError.message));
  }, []);

  async function markRead(id) {
    await api.markRead(id);
    setItems((current) => current.map((item) => (item.id === id ? { ...item, is_read: true } : item)));
  }

  async function markAll() {
    await api.markAllRead();
    setItems((current) => current.map((item) => ({ ...item, is_read: true })));
  }

  async function openNotification(item) {
    if (!item.is_read) {
      await api.markRead(item.id);
      setItems((current) => current.map((next) => (next.id === item.id ? { ...next, is_read: true } : next)));
    }
    if (item.post_id) {
      navigate(`/post/${item.post_id}`);
      return;
    }
    if (item.actor_id) {
      navigate(`/user/${item.actor_id}`);
    }
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

  if (error) {
    return <div className="surface state-box">{error}</div>;
  }

  return (
    <section className="main-column narrow">
      <header className="section-heading toolbar-heading notification-heading">
        <div>
          <p className="eyebrow">通知中心</p>
          <h1>最近发生的互动</h1>
        </div>
        <button className="ghost-button" onClick={markAll}>
          全部已读
        </button>
      </header>
      <div className="surface notification-list">
        {items.length === 0 ? <p className="muted-copy">暂时没有通知。</p> : null}
        {items.map((item) => (
          <article key={item.id} className={item.is_read ? "read" : ""}>
            <button className="notification-link" type="button" onClick={() => openNotification(item)}>
              <Avatar user={item.actor} label={item.actor?.nickname || item.actor?.username || "D"} className="tiny-avatar" />
              <div>
                <strong>{notificationText(item)}</strong>
                <span>{item.post?.title || item.content}</span>
                <time>{formatDate(item.created_at)}</time>
              </div>
            </button>
            <div className="notification-actions">
              {!item.is_read ? (
                <button className="text-link" onClick={() => markRead(item.id)}>
                  标记已读
                </button>
              ) : null}
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}
