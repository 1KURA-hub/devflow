import React, { useEffect, useState } from "react";
import {
  Bell,
  Flame,
  Home,
  Images,
  Rows3,
  Moon,
  Plus,
  RotateCcw,
  Settings,
  Star,
  SunMedium,
  Upload,
  UserRound,
  X
} from "lucide-react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../state/auth";
import { Avatar } from "./Avatar";
import { Brand } from "./Brand";
import { ComposerModal } from "./ComposerModal";

const backgroundDBName = "devflow_ui";
const backgroundStoreName = "preferences";
const backgroundRecordKey = "background_image";
const maxBackgroundSize = 8 * 1024 * 1024;

function openBackgroundDB() {
  return new Promise((resolve, reject) => {
    if (!window.indexedDB) {
      reject(new Error("当前浏览器不支持大图背景存储"));
      return;
    }

    const request = window.indexedDB.open(backgroundDBName, 1);
    request.onupgradeneeded = () => {
      request.result.createObjectStore(backgroundStoreName);
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error || new Error("背景存储打开失败"));
  });
}

function backgroundStoreAction(mode, value) {
  return openBackgroundDB().then(
    (db) =>
      new Promise((resolve, reject) => {
        const tx = db.transaction(backgroundStoreName, mode === "get" ? "readonly" : "readwrite");
        const store = tx.objectStore(backgroundStoreName);
        const request =
          mode === "get"
            ? store.get(backgroundRecordKey)
            : mode === "delete"
              ? store.delete(backgroundRecordKey)
              : store.put(value, backgroundRecordKey);

        request.onsuccess = () => resolve(request.result);
        request.onerror = () => reject(request.error || new Error("背景存储失败"));
        tx.oncomplete = () => db.close();
        tx.onerror = () => {
          db.close();
          reject(tx.error || new Error("背景存储失败"));
        };
      })
  );
}

export function AppShell() {
  const { user, updateMe, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [composerOpen, setComposerOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [celebrating, setCelebrating] = useState(false);
  const [backgroundImage, setBackgroundImage] = useState("");
  const [backgroundError, setBackgroundError] = useState("");
  const [unreadCount, setUnreadCount] = useState(0);
  const [darkTheme, setDarkTheme] = useState(false);
  const [richFeed, setRichFeed] = useState(false);
  const [profileBioDraft, setProfileBioDraft] = useState("");
  const [profileStats, setProfileStats] = useState({ posts: 0, following: 0, followers: 0 });
  const isFeedRoute = location.pathname === "/" || location.pathname === "/hot" || location.pathname === "/following";
  const isMobileMeRoute =
    location.pathname === "/me" ||
    location.pathname.startsWith("/user/") ||
    location.pathname === "/favorites" ||
    location.pathname === "/notifications";

  useEffect(() => {
    let active = true;
    let objectURL = "";

    backgroundStoreAction("get")
      .then((blob) => {
        if (!active || !blob) {
          return;
        }
        objectURL = URL.createObjectURL(blob);
        setBackgroundImage(objectURL);
      })
      .catch(() => {});

    return () => {
      active = false;
      if (objectURL) {
        URL.revokeObjectURL(objectURL);
      }
    };
  }, []);

  useEffect(() => {
    setProfileBioDraft(user?.bio || "");
  }, [user]);

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
  }, [location.pathname, user]);

  useEffect(() => {
    const refresh = () => {
      if (!user) {
        return;
      }
      Promise.all([
        api.userPosts(user.id, { limit: 50 }),
        api.followingUsers(user.id, { limit: 50 }),
        api.followers(user.id, { limit: 50 })
      ])
        .then(([posts, following, followers]) => {
          setProfileStats({
            posts: posts.items.length,
            following: following.items.length,
            followers: followers.items.length
          });
        })
        .catch(() => {});
    };
    window.addEventListener("devflow:profile-stats-refresh", refresh);
    window.addEventListener("devflow:post-created", refresh);
    return () => {
      window.removeEventListener("devflow:profile-stats-refresh", refresh);
      window.removeEventListener("devflow:post-created", refresh);
    };
  }, [user]);

  useEffect(() => {
    let timer;
    const celebrate = () => {
      setCelebrating(true);
      window.clearTimeout(timer);
      timer = window.setTimeout(() => setCelebrating(false), 1300);
    };
    window.addEventListener("devflow:celebrate", celebrate);
    return () => {
      window.clearTimeout(timer);
      window.removeEventListener("devflow:celebrate", celebrate);
    };
  }, []);

  function openComposer() {
    if (user) {
      setComposerOpen(true);
      return;
    }
    navigate("/login");
  }

  function adjustProfileStats(delta) {
    setProfileStats((current) => ({
      posts: current.posts + (delta.posts || 0),
      following: Math.max(0, current.following + (delta.following || 0)),
      followers: Math.max(0, current.followers + (delta.followers || 0))
    }));
  }

  async function changeBackground(event) {
    const file = event.target.files?.[0];
    event.target.value = "";
    setBackgroundError("");
    if (!file) {
      return;
    }
    if (!file.type.startsWith("image/")) {
      setBackgroundError("请选择图片文件");
      return;
    }
    if (file.size > maxBackgroundSize) {
      setBackgroundError("图片不能超过 8MB");
      return;
    }

    try {
      await backgroundStoreAction("put", file);
      const nextURL = URL.createObjectURL(file);
      setBackgroundImage((current) => {
        if (current.startsWith("blob:")) {
          URL.revokeObjectURL(current);
        }
        return nextURL;
      });
    } catch (error) {
      setBackgroundError(error.message || "背景保存失败，请换一张更小的图片");
    }
  }

  async function resetBackground() {
    setBackgroundImage((current) => {
      if (current.startsWith("blob:")) {
        URL.revokeObjectURL(current);
      }
      return "";
    });
    setBackgroundError("");
    try {
      await backgroundStoreAction("delete");
    } catch {
      setBackgroundError("背景重置失败，请稍后重试");
    }
  }

  async function changeAvatar(event) {
    const file = event.target.files?.[0];
    event.target.value = "";
    setBackgroundError("");
    if (!file) {
      return;
    }
    try {
      const result = await api.uploadImage(file);
      await updateMe({ avatar_url: result.url });
    } catch (error) {
      setBackgroundError(error.message);
    }
  }

  async function saveProfileBio(event) {
    event.preventDefault();
    setBackgroundError("");
    try {
      await updateMe({ bio: profileBioDraft });
    } catch (error) {
      setBackgroundError(error.message);
    }
  }

  function logoutAndClose() {
    logout();
    setSettingsOpen(false);
    navigate("/login");
  }

  return (
    <div
      className={`app-shell ${darkTheme ? "theme-dark" : ""}`}
      style={{
        ...(backgroundImage ? { "--custom-bg": `url(${backgroundImage})` } : {})
      }}
    >
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
                <Avatar user={user} />
                <div className="sidebar-profile-main">
                  <strong>{user.nickname}</strong>
                  <span>{user.bio || "还没有填写简介"}</span>
                </div>
                <button type="button" onClick={() => setSettingsOpen(true)} aria-label="设置">
                  <Settings size={16} />
                </button>
                <div className="sidebar-profile-stats">
                  <button type="button" onClick={() => navigate(`/user/${user.id}?tab=posts`)}>
                    <strong>{profileStats.posts}</strong>
                    动态
                  </button>
                  <button type="button" onClick={() => navigate(`/user/${user.id}?tab=following`)}>
                    <strong>{profileStats.following}</strong>
                    关注
                  </button>
                  <button type="button" onClick={() => navigate(`/user/${user.id}?tab=followers`)}>
                    <strong>{profileStats.followers}</strong>
                    粉丝
                  </button>
                </div>
              </>
            ) : (
              <>
                <Avatar label="D" />
                <div className="sidebar-profile-main">
                  <strong>访客模式</strong>
                  <span>登录后同步你的动态</span>
                </div>
                <button type="button" onClick={() => setSettingsOpen(true)} aria-label="设置">
                  <Settings size={16} />
                </button>
              </>
            )}
          </section>

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
            <button type="button" onClick={() => setSettingsOpen(true)} aria-label="快捷设置">
              <Settings size={18} />
            </button>
          </div>
        </aside>

        <div className="workspace">
          <main className="page-shell">
            <Outlet
              context={{
                openComposer,
                openSettings: () => setSettingsOpen(true),
                richFeed,
                profileStats,
                adjustProfileStats
              }}
            />
          </main>
        </div>

        <nav className="mobile-bottom-nav" aria-label="移动端主导航">
          <NavLink to="/" end className={isFeedRoute ? "active" : ""}>
            <Home size={20} />
            <span>动态</span>
          </NavLink>
          <button className="mobile-publish-button" type="button" onClick={openComposer} aria-label="发布动态">
            <Plus size={28} strokeWidth={2.5} />
          </button>
          <button
            type="button"
            className={isMobileMeRoute ? "active" : ""}
            onClick={() => navigate("/me")}
            aria-label="我的"
          >
            <span className="mobile-nav-icon-wrap">
              <UserRound size={20} />
              {unreadCount > 0 ? <em>{unreadCount > 99 ? "99+" : unreadCount}</em> : null}
            </span>
            <span>我的</span>
          </button>
        </nav>
      </div>
      {settingsOpen ? (
        <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="界面设置" onClick={() => setSettingsOpen(false)}>
          <section className="settings-modal" onClick={(event) => event.stopPropagation()}>
            <button className="icon-command modal-close" type="button" onClick={() => setSettingsOpen(false)} aria-label="关闭">
              <X size={18} />
            </button>
            <div>
              <p className="eyebrow">界面设置</p>
              <h2>调整首页显示方式</h2>
            </div>
            {user ? (
              <form className="setting-row bio-setting" onSubmit={saveProfileBio}>
                <span>
                  <strong>个人简介</strong>
                  <em>展示在首页和个人资料卡片里</em>
                </span>
                <input
                  value={profileBioDraft}
                  onChange={(event) => setProfileBioDraft(event.target.value)}
                  maxLength={255}
                  placeholder="写一句你的技术方向"
                />
                <button className="primary-button" type="submit">
                  保存
                </button>
              </form>
            ) : null}
            {user ? (
              <div className="setting-row background-control">
                <span>
                  <strong>个人头像</strong>
                  <em>{user.avatar_url ? "已使用自定义头像" : "上传一张头像图片"}</em>
                </span>
                <div>
                  <label className="icon-command background-upload" aria-label="上传头像">
                    <Upload size={17} />
                    <input type="file" accept="image/*" onChange={changeAvatar} />
                  </label>
                </div>
              </div>
            ) : null}
            <button className="setting-row" type="button" onClick={() => setDarkTheme((value) => !value)}>
              <span>
                <strong>主题模式</strong>
                <em>{darkTheme ? "深色玻璃背景" : "浅色玻璃背景"}</em>
              </span>
              {darkTheme ? <Moon size={19} /> : <SunMedium size={19} />}
            </button>
            <button className="setting-row" type="button" onClick={() => setRichFeed((value) => !value)}>
              <span>
                <strong>动态显示</strong>
                <em>{richFeed ? "图片卡片模式" : "紧凑列表模式"}</em>
              </span>
              {richFeed ? <Images size={19} /> : <Rows3 size={19} />}
            </button>
            <div className="setting-row background-control">
              <span>
                <strong>背景图片</strong>
                <em>{backgroundImage ? "已使用本地自定义背景" : "当前使用默认背景"}</em>
              </span>
              <div>
                <label className="icon-command background-upload" aria-label="上传背景">
                  <Upload size={17} />
                  <input type="file" accept="image/*" onChange={changeBackground} />
                </label>
                <button className="icon-command" type="button" onClick={resetBackground} aria-label="恢复默认背景">
                  <RotateCcw size={17} />
                </button>
              </div>
            </div>
            {backgroundError ? <p className="form-error">{backgroundError}</p> : null}
            {user ? (
              <button className="setting-row danger-row" type="button" onClick={logoutAndClose}>
                <span>
                  <strong>退出账号</strong>
                  <em>清除当前登录状态并返回登录页</em>
                </span>
              </button>
            ) : null}
          </section>
        </div>
      ) : null}
      {celebrating ? (
        <div className="celebration-burst" aria-hidden="true">
          {Array.from({ length: 18 }, (_, index) => (
            <span key={index} style={{ "--i": index }} />
          ))}
        </div>
      ) : null}
      {user ? <ComposerModal open={composerOpen} onClose={() => setComposerOpen(false)} /> : null}
    </div>
  );
}
