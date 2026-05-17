import React, { useCallback } from "react";
import { api } from "../api/client";
import { FeedList } from "../components/FeedList";

export function FavoritesPage() {
  const loader = useCallback((params) => api.favorites(params), []);
  return (
    <section className="main-column narrow">
      <header className="section-heading">
        <p className="eyebrow">我的收藏</p>
        <h1>稍后值得重读的内容</h1>
      </header>
      <FeedList loader={loader} emptyTitle="还没有收藏" emptyText="遇到想回看的动态，可以先收进这里。" />
    </section>
  );
}
