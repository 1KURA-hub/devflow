import React, { useState } from "react";
import { Heart, MessageCircle, Star } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../state/auth";
import { Avatar } from "./Avatar";
import { formatDate, splitTags } from "../utils/format";

export function PostCard({ post, compact = false, rich = false }) {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [liked, setLiked] = useState(false);
  const [favorited, setFavorited] = useState(false);
  const [counts, setCounts] = useState({
    like_count: post.like_count,
    favorite_count: post.favorite_count,
    comment_count: post.comment_count
  });

  async function toggleLike(event) {
    event.preventDefault();
    event.stopPropagation();
    if (!user) {
      navigate("/login");
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
    event.stopPropagation();
    if (!user) {
      navigate("/login");
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

  function openPost() {
    navigate(`/post/${post.id}`);
  }

  function stopCardOpen(event) {
    event.stopPropagation();
  }

  const author = post.author || null;
  const authorName = author?.nickname || `作者 #${post.author_id}`;

  return (
    <article className={`post-card ${compact ? "compact" : ""}`} onClick={openPost}>
      <div className="post-meta">
        <Link className="post-author" to={`/user/${post.author_id}`} onClick={stopCardOpen}>
          <Avatar user={author} label={authorName} className="tiny-avatar" />
          <span>{authorName}</span>
        </Link>
        <time>{formatDate(post.created_at)}</time>
      </div>
      {rich ? <div className="post-preview" aria-hidden="true" /> : null}
      <h3 className="post-title">{post.title}</h3>
      <p className="post-copy">{post.content}</p>
      <div className="tag-row">
        {splitTags(post.tags).map((tag) => (
          <span key={tag}>{tag}</span>
        ))}
      </div>
      <footer className="post-actions">
        <button className={liked ? "active" : ""} onClick={toggleLike}>
          <Heart size={16} fill={liked ? "currentColor" : "none"} />
          {counts.like_count}
        </button>
        <Link to={`/post/${post.id}`} onClick={stopCardOpen}>
          <MessageCircle size={16} />
          {counts.comment_count}
        </Link>
        <button className={`favorite-action ${favorited ? "active" : ""}`} onClick={toggleFavorite}>
          <Star size={16} fill={favorited ? "currentColor" : "none"} />
          {counts.favorite_count}
        </button>
      </footer>
    </article>
  );
}
