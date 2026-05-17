import React, { useEffect, useState } from "react";
import { Bell, Bookmark, Flame, Home, LogOut, PenLine, UserRound } from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../state/auth";
import { Brand } from "./Brand";
import { ComposerModal } from "./ComposerModal";

export function AppShell() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const [composerOpen, setComposerOpen] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);

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
    <div className="app-shell">
      <header className="topbar">
        <Brand />
        <nav className="main-nav" aria-label="主导航">
          <NavLink to="/">
            <Home size={17} />
            最新
          </NavLink>
          <NavLink to="/hot">
            <Flame size={17} />
            热门
          </NavLink>
          {user ? (
            <NavLink to="/following">
              <UserRound size={17} />
              关注
            </NavLink>
          ) : null}
        </nav>
        <div className="top-actions">
          {user ? (
            <>
              <button className="icon-command" onClick={() => setComposerOpen(true)} aria-label="发布动态">
                <PenLine size={18} />
              </button>
              <NavLink className="icon-command" to="/favorites" aria-label="我的收藏">
                <Bookmark size={18} />
              </NavLink>
              <NavLink className="icon-command badge-wrap" to="/notifications" aria-label="通知">
                <Bell size={18} />
                {unreadCount > 0 ? <span>{unreadCount}</span> : null}
              </NavLink>
              <Link className="user-chip" to={`/user/${user.id}`}>
                {user.nickname}
              </Link>
              <button className="icon-command" onClick={logout} aria-label="退出登录">
                <LogOut size={18} />
              </button>
            </>
          ) : (
            <>
              <Link className="text-link" to="/login">
                登录
              </Link>
              <Link className="primary-button compact" to="/register">
                注册
              </Link>
            </>
          )}
        </div>
      </header>
      <main className="page-shell">
        <Outlet context={{ openComposer: () => setComposerOpen(true) }} />
      </main>
      {user ? <ComposerModal open={composerOpen} onClose={() => setComposerOpen(false)} /> : null}
    </div>
  );
}
