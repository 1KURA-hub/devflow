import React, { useEffect, useState } from "react";
import { PostCard } from "./PostCard";

export function FeedList({ loader, refreshKey = 0, emptyTitle, emptyText }) {
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

  async function loadMore() {
    const result = await loader({ cursor: state.cursor });
    setState((current) => ({
      ...current,
      items: [...current.items, ...result.items],
      cursor: result.next_cursor || "",
      hasMore: result.has_more
    }));
  }

  if (state.loading) {
    return <div className="surface state-box">正在载入动态...</div>;
  }
  if (state.error) {
    return <div className="surface state-box">{state.error}</div>;
  }
  if (state.items.length === 0) {
    return (
      <div className="surface empty-state">
        <h3>{emptyTitle}</h3>
        <p>{emptyText}</p>
      </div>
    );
  }
  return (
    <div className="feed-stack">
      {state.items.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
      {state.hasMore ? (
        <button className="ghost-button load-more" onClick={loadMore}>
          加载更多
        </button>
      ) : null}
    </div>
  );
}
