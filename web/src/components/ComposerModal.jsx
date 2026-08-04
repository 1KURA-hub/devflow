import React, { useState } from "react";
import { ImagePlus, X } from "lucide-react";
import { api } from "../api/client";

export function ComposerModal({ open, onClose }) {
  const [form, setForm] = useState({ title: "", content: "", cover_url: "", tags: "" });
  const [uploadingCover, setUploadingCover] = useState(false);
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
      setForm({ title: "", content: "", cover_url: "", tags: "" });
      onClose();
      window.dispatchEvent(new CustomEvent("devflow:post-created"));
      window.dispatchEvent(new CustomEvent("devflow:celebrate"));
    } catch (nextError) {
      setError(nextError.message);
    } finally {
      setSubmitting(false);
    }
  }

  async function changeCover(event) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) {
      return;
    }
    setUploadingCover(true);
    setError("");
    try {
      const result = await api.uploadImage(file);
      setForm((current) => ({ ...current, cover_url: result.url }));
    } catch (nextError) {
      setError(nextError.message);
    } finally {
      setUploadingCover(false);
    }
  }

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="发布动态" onClick={onClose}>
      <form className="composer-modal" onSubmit={submit} onClick={(event) => event.stopPropagation()}>
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
        <div className="cover-field">
          <span>封面</span>
          {form.cover_url ? (
            <div className="cover-preview">
              <img src={form.cover_url} alt="动态封面预览" />
              <button type="button" onClick={() => setForm({ ...form, cover_url: "" })}>
                移除
              </button>
            </div>
          ) : (
            <label className="cover-upload">
              <ImagePlus size={18} />
              {uploadingCover ? "上传中..." : "选择封面"}
              <input type="file" accept="image/*" onChange={changeCover} disabled={uploadingCover} />
            </label>
          )}
        </div>
        <label>
          标签
          <input
            value={form.tags}
            onChange={(event) => setForm({ ...form, tags: event.target.value })}
            placeholder="Go, Redis, RabbitMQ"
          />
        </label>
        {error ? <p className="form-error" role="alert">{error}</p> : null}
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
