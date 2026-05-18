import React, { useEffect, useState } from "react";
import {
  Bell,
  Bookmark,
  DraftingCompass,
  Flame,
  Home,
  Rows3,
  MoonStar,
  Plus,
  Search,
  Settings2,
  SunMedium,
  UserRound
} from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../state/auth";
import { Brand } from "./Brand";
import { ComposerModal } from "./ComposerModal";

export function AppShell() {
  const { user } = useAuth();
  const location = useLocation();
  const [composerOpen, setComposerOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  const [darkTheme, setDarkTheme] = useState(false);
  const [richFeed, setRichFeed] = useState(false);

  useEffect(() => {
    let active = true;
    if (!user) {
      setUnreadCount(0);
      return undefined;
    }
    api
      .unreadCount()
      .then((result) => {
        if (active) {
          setUnreadCount(result.unread_count);
        }
      })
      .catch(() => {
        if (active) {
          setUnreadCount(0);
        }
      });
    return () => {
      active = false;
    };
  }, [location.pathname, user]);

  return (
    <div className={`app-shell ${darkTheme ? "theme-dark" : ""}`}>
      <div className="ambient-overlay" />
      <div className="desktop-shell">
        <aside className="glass-panel sidebar">
          <Brand />
          <nav className="sidebar-nav" aria-label="主导航">
            <NavLink to="/">
              <Home size={18} />
              首页
            </NavLink>
            <NavLink to="/hot">
              <Flame size={18} />
              热门
            </NavLink>
            {user ? (
              <NavLink to="/following">
                <UserRound size={18} />
                关注
              </NavLink>
            ) : null}
            <NavLink to="/favorites">
              <Bookmark size={18} />
              收藏
            </NavLink>
            <NavLink className="badge-nav" to="/notifications">
              <Bell size={18} />
              通知
              {unreadCount > 0 ? <span>{unreadCount}</span> : null}
            </NavLink>
            {user ? (
              <NavLink to={`/user/${user.id}`}>
                <UserRound size={18} />
                我的
              </NavLink>
            ) : null}
          </nav>

          <section className="sidebar-profile">
            {user ? (
              <>
                <div className="profile-avatar">{user.nickname.slice(0, 1)}</div>
                <div>
                  <strong>{user.nickname}</strong>
                  <span>全栈开发工程师</span>
                </div>
              </>
            ) : (
              <>
                <div className="profile-avatar">D</div>
                <div>
                  <strong>访客模式</strong>
                  <span>登录后同步你的动态</span>
                </div>
              </>
            )}
          </section>

          <div className="sidebar-links">
            <button type="button">
              <DraftingCompass size={17} />
              草稿箱
              <span>3</span>
            </button>
            <button type="button">
              <Settings2 size={17} />
              设置
            </button>
          </div>
        </aside>

        <div className="workspace">
          <div className="floating-tools">
            <label className="glass-pill search-pill">
              <Search size={17} />
              <input placeholder="搜索内容、标签或用户" />
            </label>
            <button className="glass-pill add-pill" type="button" onClick={() => setComposerOpen(true)} aria-label="创建动态">
              <Plus size={20} />
            </button>
          </div>

          <main className="page-shell">
            <Outlet context={{ openComposer: () => setComposerOpen(true), richFeed }} />
          </main>
        </div>

        <div className="sidebar-utilities glass-pill" aria-label="快捷设置">
          <button type="button" onClick={() => setDarkTheme((value) => !value)} aria-label="切换主题">
            {darkTheme ? <MoonStar size={18} /> : <SunMedium size={18} />}
          </button>
          <button
            type="button"
            className={richFeed ? "active" : ""}
            onClick={() => setRichFeed((value) => !value)}
            aria-label="切换动态显示模式"
          >
            <Rows3 size={18} />
          </button>
          <button type="button" aria-label="快捷设置">
            <Settings2 size={18} />
          </button>
        </div>
      </div>
      {user ? <ComposerModal open={composerOpen} onClose={() => setComposerOpen(false)} /> : null}
    </div>
  );
}
