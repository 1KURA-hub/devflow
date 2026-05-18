import React, { useCallback, useEffect, useState } from "react";
import { Code2, Filter, Image, Link2, Plus, Search, Vote } from "lucide-react";
import { useOutletContext } from "react-router-dom";
import { api } from "../api/client";
import { FeedList } from "../components/FeedList";
import { useAuth } from "../state/auth";

const topics = ["Go", "Redis", "微服务", "Docker", "K8s", "并发编程", "AI", "RAG", "设计模式", "Linux", "数据库", "DevOps"];

export function FeedPage({ mode }) {
  const { user } = useAuth();
  const { openComposer, richFeed } = useOutletContext();
  const [refreshKey, setRefreshKey] = useState(0);
  const loader = useCallback(
    (params) => {
      if (mode === "hot") {
        return api.hotFeed(params);
      }
      if (mode === "following") {
        return api.followingFeed(params);
      }
      return api.latestFeed(params);
    },
    [mode]
  );

  useEffect(() => {
    const refresh = () => setRefreshKey((value) => value + 1);
    window.addEventListener("devflow:post-created", refresh);
    return () => window.removeEventListener("devflow:post-created", refresh);
  }, []);

  return (
    <div className="dashboard-layout">
      <section className="glass-panel feed-frame">
        <div className="feed-column">
          <button className="composer-strip inset-panel" type="button" onClick={openComposer}>
            <span className="mini-avatar">{user ? user.nickname.slice(0, 1) : "D"}</span>
            <span>分享你的技术见解、经验或有趣的想法...</span>
            <div>
              <em>
                <Image size={15} />
                图片
              </em>
              <em>
                <Code2 size={15} />
                代码
              </em>
              <em>
                <Link2 size={15} />
                链接
              </em>
              <em>
                <Vote size={15} />
                投票
              </em>
            </div>
          </button>

          <div className="feed-toolbar inset-panel">
            <div>
              <span className={mode === "latest" ? "active" : ""}>最新</span>
              <span className={mode === "hot" ? "active" : ""}>热门</span>
              <span className={mode === "following" ? "active" : ""}>关注</span>
            </div>
            <button type="button">
              <Filter size={15} />
              筛选
            </button>
          </div>

          <FeedList
            loader={loader}
            refreshKey={refreshKey}
            emptyTitle="这里还很安静"
            emptyText="第一条值得讨论的技术动态，通常就从今天开始。"
            rich={richFeed}
          />
        </div>
      </section>

      <aside className="dashboard-rail">
        <div className="rail-tools">
          <label className="glass-pill search-pill">
            <Search size={17} />
            <input placeholder="搜索内容、标签或用户" />
          </label>
          <button className="glass-pill add-pill" type="button" onClick={openComposer} aria-label="创建动态">
            <Plus size={18} />
          </button>
        </div>

        <section className="glass-panel user-summary">
          <div className="profile-avatar large">{user ? user.nickname.slice(0, 1) : "L"}</div>
          <div>
            <strong>{user ? user.nickname : "Lin"}</strong>
            <span>全栈开发工程师</span>
          </div>
          <dl>
            <div>
              <dt>动态</dt>
              <dd>231</dd>
            </div>
            <div>
              <dt>关注</dt>
              <dd>86</dd>
            </div>
            <div>
              <dt>粉丝</dt>
              <dd>1,324</dd>
            </div>
          </dl>
        </section>

        <section className="glass-panel rail-card">
          <header>
            <h2>热门标签</h2>
            <button type="button">更多</button>
          </header>
          <div className="topic-grid">
            {topics.map((topic) => (
              <span key={topic}>{topic}</span>
            ))}
          </div>
        </section>

        <section className="glass-panel rail-card overview-card">
          <header>
            <h2>社区概览</h2>
          </header>
          <div>
            <strong>2,342</strong>
            <span>今日活跃用户</span>
          </div>
          <div>
            <strong>156</strong>
            <span>今日新增动态</span>
          </div>
        </section>
      </aside>
    </div>
  );
}
