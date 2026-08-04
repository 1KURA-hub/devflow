import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Code2, Filter, Image, Link2, Plus, Search, Vote } from "lucide-react";
import { Link, NavLink, useOutletContext } from "react-router-dom";
import { api } from "../api/client";
import { Avatar } from "../components/Avatar";
import { FeedList } from "../components/FeedList";
import { useAuth } from "../state/auth";
import devflowIcon from "../assets/devflow-icon.png";

const demoMode = import.meta.env.VITE_DEMO_MODE === "true";
const topics = ["Go", "Redis", "微服务", "Docker", "K8s", "并发编程", "AI", "RAG", "设计模式", "Linux", "数据库", "DevOps"];
const feedTabs = [
  { to: "/", label: "最新", end: true },
  { to: "/hot", label: "热门" },
  { to: "/following", label: "关注" }
];
const mobileFeedTabs = [
  { to: "/following", label: "关注" },
  { to: "/hot", label: "热门" },
  { to: "/", label: "最新", end: true }
];
const demoAuthors = [
  { id: 901, nickname: "刘超", bio: "后端架构与性能优化" },
  { id: 902, nickname: "Yu", bio: "Go / Redis / 微服务" },
  { id: 903, nickname: "Chen", bio: "云原生和工程效率" }
];
const demoFeedPosts = [
  {
    id: "demo-feed-2",
    author_id: demoAuthors[1].id,
    author: demoAuthors[1],
    title: "移动端首页改成动态优先",
    content: "顶部频道只保留关注、热门、最新，底部只放动态、发布和我的。信息密度要够，但按钮不能挤。",
    cover_url: "https://images.unsplash.com/photo-1515879218367-8466d910aaa4?auto=format&fit=crop&w=1200&q=80",
    tags: "React,UI,DevOps",
    like_count: 12,
    favorite_count: 5,
    comment_count: 2,
    liked: true,
    favorited: false,
    created_at: new Date(Date.now() - 48 * 60 * 1000).toISOString()
  },
  {
    id: "demo-feed-3",
    author_id: demoAuthors[2].id,
    author: demoAuthors[2],
    title: "通知页要能说明谁做了什么",
    content: "点赞、收藏、评论跳动态详情，关注跳用户主页。列表里直接展示 actor 和 post，比固定文案清楚很多。",
    cover_url: "https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=1200&q=80",
    tags: "产品设计,React",
    like_count: 9,
    favorite_count: 4,
    comment_count: 1,
    liked: false,
    favorited: true,
    created_at: new Date(Date.now() - 2 * 3600 * 1000).toISOString()
  }
];

export function FeedPage({ mode }) {
  const { user, updateMe } = useAuth();
  const { openComposer, richFeed, profileStats, adjustProfileStats } = useOutletContext();
  const [refreshKey, setRefreshKey] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTag, setActiveTag] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [mobileSearchOpen, setMobileSearchOpen] = useState(false);
  const [showAllTopics, setShowAllTopics] = useState(false);
  const [followPage, setFollowPage] = useState(0);
  const [followed, setFollowed] = useState({});
  const [recommendedUsers, setRecommendedUsers] = useState([]);
  const [overview, setOverview] = useState({
    total_users: 0,
    total_posts: 0,
    today_users: 0,
    today_posts: 0
  });
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

  useEffect(() => {
    let active = true;
    api
      .communityOverview()
      .then((result) => {
        if (active) {
          setOverview(result);
        }
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, [refreshKey]);

  useEffect(() => {
    let active = true;
    api
      .recommendedUsers({ limit: 15 })
      .then((result) => {
        if (active) {
          setRecommendedUsers(result.items || []);
        }
      })
      .catch(() => {
        if (active) {
          setRecommendedUsers([]);
        }
      });
    return () => {
      active = false;
    };
  }, [user]);

  const visibleTopics = showAllTopics ? topics : topics.slice(0, 8);
  const suggestedFollows = useMemo(
    () => {
      const users = recommendedUsers.length ? recommendedUsers : demoMode ? demoAuthors : [];
      return Array.from({ length: Math.min(3, users.length) }, (_, index) => users[(followPage * 3 + index) % users.length]);
    },
    [followPage, recommendedUsers]
  );

  function applyTopic(topic) {
    setActiveTag((current) => (current === topic ? "" : topic));
    setFiltersOpen(true);
  }

  return (
    <div className="dashboard-layout">
      <section className="glass-panel feed-frame">
        <div className="feed-column">
          <header className="mobile-feed-header">
            <div className="mobile-page-topbar">
              <div>
                <span className="mobile-brand-mark">
                  <img src={devflowIcon} alt="" />
                </span>
                <strong>DevFlow</strong>
              </div>
              <button type="button" onClick={() => setMobileSearchOpen((value) => !value)} aria-label="搜索">
                <Search size={20} />
              </button>
            </div>
            <nav aria-label="移动端动态频道">
              {mobileFeedTabs.map((tab) => (
                <NavLink key={tab.to} to={tab.to} end={tab.end} className={({ isActive }) => (isActive ? "active" : "")}>
                  {tab.label}
                </NavLink>
              ))}
            </nav>
          </header>

          {mobileSearchOpen ? (
            <label className="mobile-search-panel inset-panel">
              <Search size={17} />
              <input
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
                placeholder="搜索内容、标签或用户"
              />
            </label>
          ) : null}

          <div className="composer-strip inset-panel">
            <Avatar user={user} label="D" className="mini-avatar" />
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
            fallbackItems={demoFeedPosts}
            demoFallbackEnabled={demoMode}
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
            <Avatar user={user} label="L" className="profile-avatar large" />
            <div>
              <strong>{user ? user.nickname : "Lin"}</strong>
              {user ? (
                <InlineBioEditor user={user} updateMe={updateMe} />
              ) : (
                <span>登录后同步你的动态</span>
              )}
            </div>
            <dl>
              <div>
                <dt>动态</dt>
                <dd>
                  <Link to={user ? `/user/${user.id}?tab=posts` : "/login"}>{profileStats?.posts ?? 0}</Link>
                </dd>
              </div>
              <div>
                <dt>关注</dt>
                <dd>
                  <Link to={user ? `/user/${user.id}?tab=following` : "/login"}>{profileStats?.following ?? 0}</Link>
                </dd>
              </div>
              <div>
                <dt>粉丝</dt>
                <dd>
                  <Link to={user ? `/user/${user.id}?tab=followers` : "/login"}>{profileStats?.followers ?? 0}</Link>
                </dd>
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
                  <Avatar user={follow} label={follow.nickname || follow.username || "D"} className="tiny-avatar" />
                  <div>
                    <strong>{follow.nickname || follow.username || `用户 #${follow.id}`}</strong>
                    <span>{follow.bio || "这个开发者还没有填写简介。"}</span>
                  </div>
                  <button
                    className={followed[follow.id] ? "active" : ""}
                    type="button"
                    onClick={async () => {
                      const nextFollowed = !followed[follow.id];
                      if (user && typeof follow.id === "number") {
                        try {
                          await (nextFollowed ? api.follow(follow.id) : api.unfollow(follow.id));
                        } catch {
                          return;
                        }
                      }
                      setFollowed((current) => ({ ...current, [follow.id]: nextFollowed }));
                      adjustProfileStats?.({ following: nextFollowed ? 1 : -1 });
                    }}
                  >
                    {followed[follow.id] ? "已关注" : "关注"}
                  </button>
                </article>
              ))}
            </div>
          </section>

          <section className="glass-panel rail-card overview-card">
            <header>
              <h2>社区概览</h2>
            </header>
            <div>
              <strong>{overview.total_users}</strong>
              <span>社区用户</span>
            </div>
            <div>
              <strong>{overview.today_posts}</strong>
              <span>今日新增动态</span>
            </div>
            <div>
              <strong>{overview.total_posts}</strong>
              <span>动态总数</span>
            </div>
            <div>
              <strong>{overview.today_users}</strong>
              <span>今日新用户</span>
            </div>
          </section>
        </div>
      </aside>
    </div>
  );
}

function InlineBioEditor({ user, updateMe }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(user.bio || "");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setDraft(user.bio || "");
  }, [user.bio]);

  async function save(event) {
    event.preventDefault();
    if (saving) {
      return;
    }
    setSaving(true);
    try {
      await updateMe({ bio: draft });
      setEditing(false);
    } finally {
      setSaving(false);
    }
  }

  if (!editing) {
    return (
      <button className="inline-bio-text" type="button" onClick={() => setEditing(true)}>
        {user.bio || "还没有填写简介"}
      </button>
    );
  }

  return (
    <form className="inline-bio-editor" onSubmit={save}>
      <input value={draft} onChange={(event) => setDraft(event.target.value)} maxLength={255} placeholder="写一句你的技术方向" autoFocus />
      <button type="submit" disabled={saving}>
        保存
      </button>
    </form>
  );
}
