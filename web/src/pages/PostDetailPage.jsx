import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { PostCard } from "../components/PostCard";
import { useAuth } from "../state/auth";
import { formatDate } from "../utils/format";

export function PostDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [post, setPost] = useState(null);
  const [comments, setComments] = useState([]);
  const [content, setContent] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([api.getPost(id), api.comments(id, {})])
      .then(([nextPost, nextComments]) => {
        setPost(nextPost);
        setComments(nextComments.items);
      })
      .catch((nextError) => setError(nextError.message));
  }, [id]);

  async function submit(event) {
    event.preventDefault();
    const comment = await api.createComment(id, { content });
    setComments((current) => [comment, ...current]);
    setContent("");
  }

  if (error) {
    return <div className="surface state-box">{error}</div>;
  }
  if (!post) {
    return <div className="surface state-box">正在载入动态...</div>;
  }

  return (
    <div className="detail-layout">
      <section className="main-column">
        <PostCard post={post} onDeleted={() => navigate("/")} />
        <section className="surface comments-panel">
          <header>
            <p className="eyebrow">评论</p>
            <h2>{comments.length} 条回应</h2>
          </header>
          {user ? (
            <form className="comment-form" onSubmit={submit}>
              <textarea
                value={content}
                onChange={(event) => setContent(event.target.value)}
                placeholder="补充你的观点"
                rows="4"
                required
              />
              <button className="primary-button">发表评论</button>
            </form>
          ) : (
            <p className="muted-copy">登录后可以参与讨论。</p>
          )}
          <div className="comment-list">
            {comments.map((comment) => (
              <article key={comment.id}>
                <div>
                  <strong>{comment.user?.nickname || comment.user?.username || `用户 #${comment.user_id}`}</strong>
                  <time>{formatDate(comment.created_at)}</time>
                </div>
                <p>{comment.content}</p>
              </article>
            ))}
          </div>
        </section>
      </section>
    </div>
  );
}
