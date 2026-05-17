import React, { useState } from "react";
import { X } from "lucide-react";
import { api } from "../api/client";

export function ComposerModal({ open, onClose }) {
  const [form, setForm] = useState({ title: "", content: "", tags: "" });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  if (!open) {
    return null;
  }

  async function submit(event) {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await api.createPost(form);
      setForm({ title: "", content: "", tags: "" });
      onClose();
      window.dispatchEvent(new CustomEvent("devflow:post-created"));
    } catch (nextError) {
      setError(nextError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="发布动态">
      <form className="composer-modal" onSubmit={submit}>
        <button className="icon-command modal-close" type="button" onClick={onClose} aria-label="关闭">
          <X size={18} />
        </button>
        <div>
          <p className="eyebrow">新动态</p>
          <h2>写下今天值得分享的技术进展</h2>
        </div>
        <label>
          标题
          <input
            value={form.title}
            onChange={(event) => setForm({ ...form, title: event.target.value })}
            placeholder="例如：把通知链路改成了 RabbitMQ 异步消费"
            required
          />
        </label>
        <label>
          正文
          <textarea
            value={form.content}
            onChange={(event) => setForm({ ...form, content: event.target.value })}
            placeholder="写下问题、方案和关键权衡。"
            rows="6"
            required
          />
        </label>
        <label>
          标签
          <input
            value={form.tags}
            onChange={(event) => setForm({ ...form, tags: event.target.value })}
            placeholder="Go, Redis, RabbitMQ"
          />
        </label>
        {error ? <p className="form-error">{error}</p> : null}
        <div className="modal-actions">
          <button className="ghost-button" type="button" onClick={onClose}>
            取消
          </button>
          <button className="primary-button" disabled={submitting}>
            {submitting ? "发布中..." : "发布"}
          </button>
        </div>
      </form>
    </div>
  );
}
