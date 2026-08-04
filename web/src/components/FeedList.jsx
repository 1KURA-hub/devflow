import React, { useMemo, useEffect, useState } from "react";
import { PostCard } from "./PostCard";

export function FeedList({
  loader,
  refreshKey = 0,
  emptyTitle,
  emptyText,
  rich = false,
  query = "",
  activeTag = "",
  fallbackItems = [],
  demoFallbackEnabled = false,
  removeOnUnfavorite = false
}) {
  const [state, setState] = useState({
    items: [],
    cursor: "",
    hasMore: false,
    loading: true,
    error: ""
  });

  useEffect(() => {
    let active = true;
    setState({ items: [], cursor: "", hasMore: false, loading: true, error: "" });
    loader({})
      .then((result) => {
        if (active) {
          setState({
            items: result.items,
            cursor: result.next_cursor || "",
            hasMore: result.has_more,
            loading: false,
            error: ""
          });
        }
      })
      .catch((error) => {
        if (active) {
          setState((current) => ({ ...current, loading: false, error: error.message }));
        }
      });
    return () => {
      active = false;
    };
  }, [loader, refreshKey]);

  const visibleItems =
    demoFallbackEnabled && state.items.length === 0 ? fallbackItems : state.items;
  const filteredItems = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const normalizedTag = activeTag.trim().toLowerCase();
    return visibleItems.filter((post) => {
      const text = `${post.title} ${post.content} ${post.tags}`.toLowerCase();
      const tagMatched = !normalizedTag || post.tags.toLowerCase().split(",").map((tag) => tag.trim()).includes(normalizedTag);
      const queryMatched = !normalizedQuery || text.includes(normalizedQuery);
      return tagMatched && queryMatched;
    });
  }, [activeTag, query, visibleItems]);

  async function loadMore() {
    const result = await loader({ cursor: state.cursor });
    setState((current) => ({
      ...current,
      items: [...current.items, ...result.items],
      cursor: result.next_cursor || "",
      hasMore: result.has_more
    }));
  }

  function removePost(postID) {
    setState((current) => ({
      ...current,
      items: current.items.filter((post) => post.id !== postID)
    }));
  }

  if (state.loading) {
    return <div className="surface state-box">正在载入动态...</div>;
  }
  if (state.error) {
    return <div className="surface state-box">{state.error}</div>;
  }
  if (visibleItems.length === 0 || filteredItems.length === 0) {
    return (
      <div className="surface empty-state">
        <h3>{visibleItems.length === 0 ? emptyTitle : "没有匹配的动态"}</h3>
        <p>{visibleItems.length === 0 ? emptyText : "换个关键词或标签试试，当前只筛选已经加载的动态。"}</p>
      </div>
    );
  }
  return (
    <div className="feed-stack">
      {filteredItems.map((post) => (
        <PostCard
          key={post.id}
          post={post}
          rich={rich}
          onDeleted={removePost}
          onUnfavorited={removeOnUnfavorite ? removePost : undefined}
        />
      ))}
      {state.hasMore ? (
        <button className="ghost-button load-more" onClick={loadMore}>
          加载更多
        </button>
      ) : null}
    </div>
  );
}
