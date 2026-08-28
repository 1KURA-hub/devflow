import React, { useState } from "react";
import { ArrowRight, LockKeyhole, UserRound } from "lucide-react";
import { Navigate, Link, useNavigate } from "react-router-dom";
import banner from "../assets/devflow-banner.jpg";
import loginIcon from "../assets/login-icon.png";
import registerIcon from "../assets/register-icon.png";
import { useAuth } from "../state/auth";
import { Brand } from "../components/Brand";

export function AuthPage({ mode }) {
  const { user, login, register } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ username: "", password: "", nickname: "" });
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  if (user) {
    return <Navigate to="/" replace />;
  }

  async function submit(event) {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      if (mode === "register") {
        await register(form);
      } else {
        await login(form);
      }
      navigate("/");
    } catch (nextError) {
      setError(nextError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="auth-shell">
      <section className="auth-story">
        <div className="auth-story-panel glass-panel">
          <Brand />
          <div>
            <p className="eyebrow">程序员技术动态社区</p>
            <h1>把每天的工程进展，沉淀成可讨论的技术动态。</h1>
          </div>
          <div className="auth-preview-card">
            <img src={banner} alt="" />
          </div>
          <div className="auth-proof-row" aria-label="核心状态">
            <span>动态发布</span>
            <span>关注流</span>
            <span>评论通知</span>
          </div>
        </div>
      </section>
      <form className="auth-form surface" onSubmit={submit}>
        <div className="auth-form-header">
          <span className="auth-form-icon">
            <img src={mode === "register" ? registerIcon : loginIcon} alt="" />
          </span>
          <h2>{mode === "register" ? "注册 DevFlow" : "登录 DevFlow"}</h2>
        </div>
        <label className="auth-field">
          用户名
          <span>
            <UserRound size={17} />
            <input
              value={form.username}
              onChange={(event) => setForm({ ...form, username: event.target.value })}
              required
            />
          </span>
        </label>
        {mode === "register" ? (
          <label className="auth-field">
            昵称
            <span>
              <UserRound size={17} />
              <input
                value={form.nickname}
                onChange={(event) => setForm({ ...form, nickname: event.target.value })}
                required
              />
            </span>
          </label>
        ) : null}
        <label className="auth-field">
          密码
          <span>
            <LockKeyhole size={17} />
            <input
              type="password"
              value={form.password}
              onChange={(event) => setForm({ ...form, password: event.target.value })}
              required
            />
          </span>
        </label>
        {error ? <p className="form-error" role="alert">{error}</p> : null}
        <button className="primary-button auth-submit" disabled={submitting}>
          {submitting ? "处理中..." : mode === "register" ? "注册" : "登录"}
          <ArrowRight size={17} />
        </button>
        <p className="muted-copy auth-switch">
          {mode === "register" ? "已有账号？" : "还没有账号？"}
          <Link to={mode === "register" ? "/login" : "/register"}>
            {mode === "register" ? "去登录" : "去注册"}
          </Link>
        </p>
      </form>
    </div>
  );
}
