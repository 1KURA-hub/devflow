import React, { useEffect, useRef, useState } from "react";
import { ArrowLeft } from "lucide-react";
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
  const [loadedPostID, setLoadedPostID] = useState("");
  const [commentSubmitting, setCommentSubmitting] = useState(false);
  const [commentError, setCommentError] = useState("");
  const viewVersionRef = useRef(0);

  useEffect(() => {
    let active = true;
    viewVersionRef.current += 1;
    const viewVersion = viewVersionRef.current;
    setPost(null);
    setComments([]);
    setContent("");
    setError("");
    setLoadedPostID("");
    setCommentSubmitting(false);
    setCommentError("");
    Promise.all([api.getPost(id), api.comments(id, {})])
      .then(([nextPost, nextComments]) => {
        if (!active) {
          return;
        }
        setPost(nextPost);
        setComments(nextComments.items);
        setLoadedPostID(id);
      })
      .catch((nextError) => {
        if (active) {
          setError(nextError.message);
          setLoadedPostID(id);
        }
      });
    return () => {
      active = false;
      if (viewVersionRef.current === viewVersion) {
        viewVersionRef.current += 1;
      }
    };
  }, [id]);

  async function submit(event) {
    event.preventDefault();
    if (commentSubmitting) {
      return;
    }
    const submittedPostID = id;
    const submittedViewVersion = viewVersionRef.current;
    const submittedContent = content;
    const isCurrentView = () => viewVersionRef.current === submittedViewVersion;
    setCommentSubmitting(true);
    setCommentError("");
    try {
      const comment = await api.createComment(submittedPostID, { content: submittedContent });
      if (!isCurrentView()) {
        return;
      }
      setComments((current) => [comment, ...current]);
      setPost((current) =>
        current ? { ...current, comment_count: current.comment_count + 1 } : current
      );
      setContent("");
    } catch (nextError) {
      if (isCurrentView()) {
        setCommentError(nextError.message);
      }
    } finally {
      if (isCurrentView()) {
        setCommentSubmitting(false);
      }
    }
  }

  if (loadedPostID === id && error) {
    return <div className="surface state-box">{error}</div>;
  }
  if (loadedPostID !== id || !post) {
    return <div className="surface state-box">正在载入动态...</div>;
  }

  return (
    <div className="detail-layout">
      <section className="main-column">
        <button className="mobile-detail-back" type="button" onClick={() => navigate(-1)}>
          <ArrowLeft size={18} />
          返回
        </button>
        <PostCard post={post} detail onDeleted={() => navigate("/")} />
        <section className="surface comments-panel">
          <header>
            <p className="eyebrow">评论</p>
            <h2>{comments.length} 条回应</h2>
          </header>
          {user ? (
            <form className="comment-form" onSubmit={submit}>
              <textarea
                aria-label="评论内容"
                value={content}
                onChange={(event) => setContent(event.target.value)}
                placeholder="补充你的观点"
                rows="4"
                required
              />
              {commentError ? <p className="form-error" role="alert">{commentError}</p> : null}
              <button className="primary-button" disabled={commentSubmitting}>
                {commentSubmitting ? "发表中..." : "发表评论"}
              </button>
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
