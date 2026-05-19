import React, { useEffect, useState } from "react";
import {
  Bell,
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
  const [profileStats, setProfileStats] = useState({ posts: 0, following: 0, followers: 0 });

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

  useEffect(() => {
    let active = true;
    if (!user) {
      setProfileStats({ posts: 0, following: 0, followers: 0 });
      return undefined;
    }

    Promise.all([
      api.userPosts(user.id, { limit: 50 }),
      api.followingUsers(user.id, { limit: 50 }),
      api.followers(user.id, { limit: 50 })
    ])
      .then(([posts, following, followers]) => {
        if (active) {
          setProfileStats({
            posts: posts.items.length,
            following: following.items.length,
            followers: followers.items.length
          });
        }
      })
      .catch(() => {
        if (active) {
          setProfileStats({ posts: 0, following: 0, followers: 0 });
        }
      });

    return () => {
      active = false;
    };
  }, [user]);

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
                <div className="sidebar-profile-main">
                  <strong>{user.nickname}</strong>
                  <span>全栈开发工程师</span>
                </div>
                <button type="button" aria-label="设置">
                  <Settings size={16} />
                </button>
                <div className="sidebar-profile-stats">
                  <span>
                    <strong>{profileStats.posts}</strong>
                    动态
                  </span>
                  <span>
                    <strong>{profileStats.following}</strong>
                    关注
                  </span>
                  <span>
                    <strong>{profileStats.followers}</strong>
                    粉丝
                  </span>
                </div>
              </>
            ) : (
              <>
                <div className="profile-avatar">D</div>
                <div className="sidebar-profile-main">
                  <strong>访客模式</strong>
                  <span>登录后同步你的动态</span>
                </div>
                <button type="button" aria-label="设置">
                  <Settings size={16} />
                </button>
              </>
            )}
          </section>
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
