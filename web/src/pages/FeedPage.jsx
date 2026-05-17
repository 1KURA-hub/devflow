import React, { useCallback, useEffect, useState } from "react";
import { api } from "../api/client";
import { FeedList } from "../components/FeedList";
import banner from "../assets/devflow-banner.jpg";
import { useAuth } from "../state/auth";

export function FeedPage({ mode }) {
  const { user } = useAuth();
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

  const copy = {
    latest: {
      eyebrow: "最新动态",
      title: "今天，社区里正在推进什么",
      body: "记录设计取舍、踩坑经验和刚刚完成的改动。"
    },
    hot: {
      eyebrow: "热门动态",
      title: "被更多开发者回应的内容",
      body: "按点赞、收藏和评论热度汇总。"
    },
    following: {
      eyebrow: "关注流",
      title: user ? `${user.nickname} 的订阅流` : "关注流",
      body: "只看你关注的开发者；还没有关注时，会自动回退到最新动态。"
    }
  }[mode];

  return (
    <div className="home-layout">
      <section className="main-column">
        <header className="editorial-banner">
          <img src={banner} alt="" />
          <div>
            <p className="eyebrow">{copy.eyebrow}</p>
            <h1>{copy.title}</h1>
            <p>{copy.body}</p>
          </div>
        </header>
        <FeedList
          loader={loader}
          refreshKey={refreshKey}
          emptyTitle="这里还很安静"
          emptyText="第一条值得讨论的技术动态，通常就从今天开始。"
        />
      </section>
      <aside className="side-rail">
        <section className="surface side-panel">
          <p className="eyebrow">DevFlow</p>
          <h2>把工程进展讲清楚</h2>
          <p>短动态、明确上下文、可追踪互动。更像开发者自己的工作流记录。</p>
        </section>
        <section className="surface side-panel metric-panel">
          <div>
            <strong>实时</strong>
            <span>动态流</span>
          </div>
          <div>
            <strong>异步</strong>
            <span>通知</span>
          </div>
          <div>
            <strong>缓存</strong>
            <span>收件箱</span>
          </div>
        </section>
      </aside>
    </div>
  );
}
