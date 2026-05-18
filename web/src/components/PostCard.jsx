import React, { useState } from "react";
import { Bookmark, Heart, MessageCircle } from "lucide-react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../state/auth";
import { formatDate, splitTags } from "../utils/format";

export function PostCard({ post, compact = false, rich = false }) {
  const { user } = useAuth();
  const [liked, setLiked] = useState(false);
  const [favorited, setFavorited] = useState(false);
  const [counts, setCounts] = useState({
    like_count: post.like_count,
    favorite_count: post.favorite_count,
    comment_count: post.comment_count
  });

  async function toggleLike(event) {
    event.preventDefault();
    if (!user) {
      return;
    }
    const nextLiked = !liked;
    await (nextLiked ? api.like(post.id) : api.unlike(post.id));
    setLiked(nextLiked);
    setCounts((current) => ({
      ...current,
      like_count: current.like_count + (nextLiked ? 1 : -1)
    }));
  }

  async function toggleFavorite(event) {
    event.preventDefault();
    if (!user) {
      return;
    }
    const nextFavorited = !favorited;
    await (nextFavorited ? api.favorite(post.id) : api.unfavorite(post.id));
    setFavorited(nextFavorited);
    setCounts((current) => ({
      ...current,
      favorite_count: current.favorite_count + (nextFavorited ? 1 : -1)
    }));
  }

  return (
    <article className={`post-card ${compact ? "compact" : ""}`}>
      {rich ? <div className="post-preview" aria-hidden="true" /> : null}
      <div className="post-meta">
        <Link to={`/user/${post.author_id}`}>作者 #{post.author_id}</Link>
        <time>{formatDate(post.created_at)}</time>
      </div>
      <Link className="post-title" to={`/post/${post.id}`}>
        {post.title}
      </Link>
      <p className="post-copy">{post.content}</p>
      <div className="tag-row">
        {splitTags(post.tags).map((tag) => (
          <span key={tag}>{tag}</span>
        ))}
      </div>
      <footer className="post-actions">
        <button className={liked ? "active" : ""} onClick={toggleLike} disabled={!user}>
          <Heart size={16} />
          {counts.like_count}
        </button>
        <Link to={`/post/${post.id}`}>
          <MessageCircle size={16} />
          {counts.comment_count}
        </Link>
        <button className={favorited ? "active" : ""} onClick={toggleFavorite} disabled={!user}>
          <Bookmark size={16} />
          {counts.favorite_count}
        </button>
      </footer>
    </article>
  );
}
