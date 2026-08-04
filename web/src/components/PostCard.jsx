import React, { useEffect, useState } from "react";
import { Heart, MessageCircle, Star, Trash2 } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../state/auth";
import { Avatar } from "./Avatar";
import { formatDate, splitTags } from "../utils/format";

export function PostCard({
  post,
  compact = false,
  rich = false,
  detail = false,
  onDeleted,
  onUnfavorited
}) {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [liked, setLiked] = useState(Boolean(post.liked));
  const [favorited, setFavorited] = useState(Boolean(post.favorited));
  const [likedChanged, setLikedChanged] = useState(false);
  const [favoritedChanged, setFavoritedChanged] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [counts, setCounts] = useState({
    like_count: post.like_count,
    favorite_count: post.favorite_count,
    comment_count: post.comment_count
  });

  useEffect(() => {
    setLiked(Boolean(post.liked));
    setFavorited(Boolean(post.favorited));
    setCounts({
      like_count: post.like_count,
      favorite_count: post.favorite_count,
      comment_count: post.comment_count
    });
  }, [post]);

  async function toggleLike(event) {
    event.preventDefault();
    event.stopPropagation();
    if (!user) {
      navigate("/login");
      return;
    }
    const nextLiked = !liked;
    setLikedChanged(true);
    setLiked(nextLiked);
    setCounts((current) => ({
      ...current,
      like_count: Math.max(0, current.like_count + (nextLiked ? 1 : -1))
    }));
    try {
      await (nextLiked ? api.like(post.id) : api.unlike(post.id));
    } catch (error) {
      setLiked(!nextLiked);
      setCounts((current) => ({
        ...current,
        like_count: Math.max(0, current.like_count + (nextLiked ? -1 : 1))
      }));
      window.alert(error.message);
    }
  }

  async function toggleFavorite(event) {
    event.preventDefault();
    event.stopPropagation();
    if (!user) {
      navigate("/login");
      return;
    }
    const nextFavorited = !favorited;
    setFavoritedChanged(true);
    setFavorited(nextFavorited);
    setCounts((current) => ({
      ...current,
      favorite_count: Math.max(0, current.favorite_count + (nextFavorited ? 1 : -1))
    }));
    try {
      await (nextFavorited ? api.favorite(post.id) : api.unfavorite(post.id));
      if (!nextFavorited) {
        onUnfavorited?.(post.id);
      }
    } catch (error) {
      setFavorited(!nextFavorited);
      setCounts((current) => ({
        ...current,
        favorite_count: Math.max(0, current.favorite_count + (nextFavorited ? -1 : 1))
      }));
      window.alert(error.message);
    }
  }

  async function deletePost(event) {
    event.preventDefault();
    event.stopPropagation();
    if (!user || user.id !== post.author_id || deleting) {
      return;
    }
    if (!window.confirm("确定删除这条动态吗？")) {
      return;
    }
    setDeleting(true);
    try {
      await api.deletePost(post.id);
      window.dispatchEvent(new Event("devflow:profile-stats-refresh"));
      onDeleted?.(post.id);
    } catch (error) {
      window.alert(error.message);
    } finally {
      setDeleting(false);
    }
  }

  function openPost() {
    if (detail) {
      return;
    }
    navigate(`/post/${post.id}`);
  }

  function stopCardOpen(event) {
    event.stopPropagation();
  }

  function stopDetailCommentNavigation(event) {
    event.preventDefault();
    event.stopPropagation();
  }

  const author = post.author || null;
  const authorName = author?.nickname || `作者 #${post.author_id}`;
  const canDelete = user?.id === post.author_id;
  const showCover = Boolean(post.cover_url) && (rich || detail);
  const showPreview = rich && !post.cover_url;

  return (
    <article
      className={`post-card ${compact ? "compact" : ""} ${detail ? "detail" : ""}`}
      data-testid={`post-card-${post.id}`}
      aria-labelledby={`post-title-${post.id}`}
      onClick={openPost}
    >
      <div className="post-meta">
        <Link className="post-author" to={`/user/${post.author_id}`} onClick={stopCardOpen}>
          <Avatar user={author} label={authorName} className="tiny-avatar" />
          <span>{authorName}</span>
        </Link>
        <time>{formatDate(post.created_at)}</time>
      </div>
      {showPreview ? <div className="post-preview" aria-hidden="true" /> : null}
      {showCover ? <img className="post-cover" src={post.cover_url} alt="" /> : null}
      <h3 id={`post-title-${post.id}`} className="post-title">{post.title}</h3>
      <p className="post-copy">{post.content}</p>
      <div className="tag-row">
        {splitTags(post.tags).map((tag) => (
          <span key={tag}>{tag}</span>
        ))}
      </div>
      <footer className="post-actions">
        <button
          className={`${liked ? "active" : ""} ${likedChanged ? "changed" : ""}`}
          aria-label={liked ? "取消点赞" : "点赞"}
          aria-pressed={liked}
          onAnimationEnd={() => setLikedChanged(false)}
          onClick={toggleLike}
        >
          <Heart size={16} fill={liked ? "currentColor" : "none"} />
          {counts.like_count}
        </button>
        <Link aria-label="查看评论" to={`/post/${post.id}`} onClick={detail ? stopDetailCommentNavigation : stopCardOpen}>
          <MessageCircle size={16} />
          {counts.comment_count}
        </Link>
        <button
          className={`favorite-action ${favorited ? "active" : ""} ${favoritedChanged ? "changed" : ""}`}
          aria-label={favorited ? "取消收藏" : "收藏"}
          aria-pressed={favorited}
          onAnimationEnd={() => setFavoritedChanged(false)}
          onClick={toggleFavorite}
        >
          <Star size={16} fill={favorited ? "currentColor" : "none"} />
          {counts.favorite_count}
        </button>
        {canDelete ? (
          <button className="delete-action" onClick={deletePost} disabled={deleting} aria-label="删除动态">
            <Trash2 size={16} />
          </button>
        ) : null}
      </footer>
    </article>
  );
}
