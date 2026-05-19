import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Code2, Filter, Image, Link2, Plus, Search, Vote } from "lucide-react";
import { NavLink, useOutletContext } from "react-router-dom";
import { api } from "../api/client";
import { FeedList } from "../components/FeedList";
import { useAuth } from "../state/auth";

const topics = ["Go", "Redis", "微服务", "Docker", "K8s", "并发编程", "AI", "RAG", "设计模式", "Linux", "数据库", "DevOps"];
const follows = [
  { id: "tech-liu", name: "刘超的技术博客", bio: "架构与性能优化" },
  { id: "nav", name: "编程导航", bio: "优质技术资源分享" },
  { id: "go-night", name: "Go 夜读", bio: "Go 工程实践" },
  { id: "redis-lab", name: "Redis 实战派", bio: "缓存与数据结构" },
  { id: "cloud-notes", name: "云原生笔记", bio: "Docker / K8s" },
  { id: "ai-dev", name: "AI 工程化", bio: "RAG 与智能应用" }
];
const feedTabs = [
  { to: "/", label: "最新", end: true },
  { to: "/hot", label: "热门" },
  { to: "/following", label: "关注" }
];

export function FeedPage({ mode }) {
  const { user } = useAuth();
  const { openComposer, richFeed, profileStats } = useOutletContext();
  const [refreshKey, setRefreshKey] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTag, setActiveTag] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [showAllTopics, setShowAllTopics] = useState(false);
  const [followPage, setFollowPage] = useState(0);
  const [followed, setFollowed] = useState({});
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

  const visibleTopics = showAllTopics ? topics : topics.slice(0, 8);
  const suggestedFollows = useMemo(
    () => Array.from({ length: 3 }, (_, index) => follows[(followPage * 3 + index) % follows.length]),
    [followPage]
  );

  function applyTopic(topic) {
    setActiveTag((current) => (current === topic ? "" : topic));
    setFiltersOpen(true);
  }

  return (
    <div className="dashboard-layout">
      <section className="glass-panel feed-frame">
        <div className="feed-column">
          <div className="composer-strip inset-panel">
            <span className="mini-avatar">{user ? user.nickname.slice(0, 1) : "D"}</span>
            <button className="composer-prompt" type="button" onClick={openComposer}>
              分享你的技术见解、经验或有趣的想法...
            </button>
            <div>
              <button type="button" onClick={openComposer}>
                <Image size={15} />
                图片
              </button>
              <button type="button" onClick={openComposer}>
                <Code2 size={15} />
                代码
              </button>
              <button type="button" onClick={openComposer}>
                <Link2 size={15} />
                链接
              </button>
              <button type="button" onClick={openComposer}>
                <Vote size={15} />
                投票
              </button>
            </div>
          </div>

          <div className="feed-toolbar inset-panel">
            <div>
              {feedTabs.map((tab) => (
                <NavLink key={tab.to} to={tab.to} end={tab.end} className={({ isActive }) => (isActive ? "active" : "")}>
                  {tab.label}
                </NavLink>
              ))}
            </div>
            <button type="button" onClick={() => setFiltersOpen((value) => !value)} className={filtersOpen ? "active" : ""}>
              <Filter size={15} />
              筛选
            </button>
          </div>

          {filtersOpen ? (
            <div className="filter-panel inset-panel">
              <button className={!activeTag ? "active" : ""} type="button" onClick={() => setActiveTag("")}>
                全部
              </button>
              {topics.map((topic) => (
                <button
                  key={topic}
                  className={activeTag === topic ? "active" : ""}
                  type="button"
                  onClick={() => applyTopic(topic)}
                >
                  {topic}
                </button>
              ))}
            </div>
          ) : null}

          <FeedList
            loader={loader}
            refreshKey={refreshKey}
            emptyTitle="这里还很安静"
            emptyText="第一条值得讨论的技术动态，通常就从今天开始。"
            rich={richFeed}
            query={searchQuery}
            activeTag={activeTag}
          />
        </div>
      </section>

      <aside className="dashboard-rail">
        <div className="rail-top">
          <div className="rail-tools">
            <label className="glass-pill search-pill">
              <Search size={17} />
              <input
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
                placeholder="搜索内容、标签或用户"
              />
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
                <dd>{profileStats?.posts ?? 0}</dd>
              </div>
              <div>
                <dt>关注</dt>
                <dd>{profileStats?.following ?? 0}</dd>
              </div>
              <div>
                <dt>粉丝</dt>
                <dd>{profileStats?.followers ?? 0}</dd>
              </div>
            </dl>
          </section>

          <section className="glass-panel rail-card">
            <header>
              <h2>热门标签</h2>
              <button type="button" onClick={() => setShowAllTopics((value) => !value)}>
                {showAllTopics ? "收起" : "更多"}
              </button>
            </header>
            <div className="topic-grid">
              {visibleTopics.map((topic) => (
                <button className={activeTag === topic ? "active" : ""} type="button" key={topic} onClick={() => applyTopic(topic)}>
                  {topic}
                </button>
              ))}
            </div>
          </section>

          <section className="glass-panel rail-card">
            <header>
              <h2>推荐关注</h2>
              <button type="button" onClick={() => setFollowPage((value) => value + 1)}>
                换一换
              </button>
            </header>
            <div className="follow-list">
              {suggestedFollows.map((follow) => (
                <article className="follow-item" key={follow.id}>
                  <div className="tiny-avatar">{follow.name.slice(0, 1)}</div>
                  <div>
                    <strong>{follow.name}</strong>
                    <span>{follow.bio}</span>
                  </div>
                  <button
                    className={followed[follow.id] ? "active" : ""}
                    type="button"
                    onClick={() => setFollowed((current) => ({ ...current, [follow.id]: !current[follow.id] }))}
                  >
                    {followed[follow.id] ? "已关注" : "关注"}
                  </button>
                </article>
              ))}
            </div>
          </section>
        </div>

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
