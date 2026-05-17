import React, { useState } from "react";
import { Navigate, Link, useNavigate } from "react-router-dom";
import banner from "../assets/devflow-banner.jpg";
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
        <Brand />
        <div>
          <p className="eyebrow">程序员技术动态社区</p>
          <h1>把零散的工程进展，沉淀成可讨论的动态。</h1>
        </div>
        <img src={banner} alt="" />
      </section>
      <form className="auth-form surface" onSubmit={submit}>
        <div>
          <p className="eyebrow">{mode === "register" ? "创建账号" : "欢迎回来"}</p>
          <h2>{mode === "register" ? "注册 DevFlow" : "登录 DevFlow"}</h2>
        </div>
        <label>
          用户名
          <input
            value={form.username}
            onChange={(event) => setForm({ ...form, username: event.target.value })}
            required
          />
        </label>
        {mode === "register" ? (
          <label>
            昵称
            <input
              value={form.nickname}
              onChange={(event) => setForm({ ...form, nickname: event.target.value })}
              required
            />
          </label>
        ) : null}
        <label>
          密码
          <input
            type="password"
            value={form.password}
            onChange={(event) => setForm({ ...form, password: event.target.value })}
            required
          />
        </label>
        {error ? <p className="form-error">{error}</p> : null}
        <button className="primary-button" disabled={submitting}>
          {submitting ? "处理中..." : mode === "register" ? "注册" : "登录"}
        </button>
        <p className="muted-copy">
          {mode === "register" ? "已有账号？" : "还没有账号？"}
          <Link to={mode === "register" ? "/login" : "/register"}>
            {mode === "register" ? "去登录" : "去注册"}
          </Link>
        </p>
      </form>
    </div>
  );
}
