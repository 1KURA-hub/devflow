import React, { useEffect, useState } from "react";
import { api } from "../api/client";
import { formatDate } from "../utils/format";

export function NotificationsPage() {
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

  if (error) {
    return <div className="surface state-box">{error}</div>;
  }

  return (
    <section className="main-column narrow">
      <header className="section-heading toolbar-heading">
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
            <div>
              <strong>{item.content}</strong>
              <time>{formatDate(item.created_at)}</time>
            </div>
            {!item.is_read ? (
              <button className="text-link" onClick={() => markRead(item.id)}>
                标记已读
              </button>
            ) : null}
          </article>
        ))}
      </div>
    </section>
  );
}
