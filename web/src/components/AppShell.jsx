import React, { useEffect, useState } from "react";
import {
  Bell,
  DraftingCompass,
  Flame,
  Home,
  Images,
  Rows3,
  Moon,
  Settings,
  Star,
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
            <NavLink to="/following">
              <UserRound size={18} />
              关注
            </NavLink>
            <NavLink to="/favorites">
              <Star size={18} />
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
            {user ? (
              <button type="button">
                <DraftingCompass size={17} />
                草稿箱
                <span>3</span>
              </button>
            ) : null}
            <button type="button">
              <Settings size={17} />
              设置
            </button>
          </div>
        </aside>

        <div className="workspace">
          <main className="page-shell">
            <Outlet context={{ openComposer: () => setComposerOpen(true), richFeed }} />
          </main>
        </div>

        <div className="sidebar-utilities glass-pill" aria-label="快捷设置">
          <button type="button" onClick={() => setDarkTheme((value) => !value)} aria-label="切换主题">
            {darkTheme ? <Moon size={18} /> : <SunMedium size={18} />}
          </button>
          <button
            type="button"
            className={richFeed ? "active" : ""}
            onClick={() => setRichFeed((value) => !value)}
            aria-label="切换动态显示模式"
          >
            {richFeed ? <Images size={18} /> : <Rows3 size={18} />}
          </button>
          <button type="button" aria-label="快捷设置">
            <Settings size={18} />
          </button>
        </div>
      </div>
      {user ? <ComposerModal open={composerOpen} onClose={() => setComposerOpen(false)} /> : null}
    </div>
  );
}
