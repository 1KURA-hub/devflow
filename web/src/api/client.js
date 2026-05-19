const tokenKey = "devflow_token";

export function getStoredToken() {
  return localStorage.getItem(tokenKey);
}

export function setStoredToken(token) {
  if (token) {
    localStorage.setItem(tokenKey, token);
  } else {
    localStorage.removeItem(tokenKey);
  }
}

async function request(path, options = {}) {
  const token = getStoredToken();
  const isFormData = options.body instanceof FormData;
  const headers = {
    ...(isFormData ? {} : { "Content-Type": "application/json" }),
    ...(options.headers || {})
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(path, {
    ...options,
    headers
  });
  const payload = await response.json().catch(() => null);
  if (!response.ok || payload?.code !== 0) {
    throw new Error(payload?.message || "请求失败");
  }
  return payload.data;
}

export const api = {
  register: (body) =>
    request("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(body)
    }),
  login: (body) =>
    request("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(body)
    }),
  me: () => request("/api/me"),
  updateMe: (body) =>
    request("/api/me", {
      method: "PATCH",
      body: JSON.stringify(body)
    }),
  uploadImage: (file) => {
    const form = new FormData();
    form.append("file", file);
    return request("/api/uploads/image", {
      method: "POST",
      body: form
    });
  },
  latestFeed: (params) => request(`/api/feed/latest${query(params)}`),
  hotFeed: (params) => request(`/api/feed/hot${query(params)}`),
  followingFeed: (params) => request(`/api/feed/following${query(params)}`),
  getPost: (id) => request(`/api/posts/${id}`),
  createPost: (body) =>
    request("/api/posts", {
      method: "POST",
      body: JSON.stringify(body)
    }),
  userPosts: (id, params) => request(`/api/users/${id}/posts${query(params)}`),
  followers: (id, params) => request(`/api/users/${id}/followers${query(params)}`),
  followingUsers: (id, params) => request(`/api/users/${id}/following${query(params)}`),
  follow: (id) => request(`/api/users/${id}/follow`, { method: "POST" }),
  unfollow: (id) => request(`/api/users/${id}/follow`, { method: "DELETE" }),
  like: (id) => request(`/api/posts/${id}/like`, { method: "POST" }),
  unlike: (id) => request(`/api/posts/${id}/like`, { method: "DELETE" }),
  favorite: (id) => request(`/api/posts/${id}/favorite`, { method: "POST" }),
  unfavorite: (id) => request(`/api/posts/${id}/favorite`, { method: "DELETE" }),
  comments: (id, params) => request(`/api/posts/${id}/comments${query(params)}`),
  createComment: (id, body) =>
    request(`/api/posts/${id}/comments`, {
      method: "POST",
      body: JSON.stringify(body)
    }),
  favorites: (params) => request(`/api/me/favorites${query(params)}`),
  notifications: (params) => request(`/api/notifications${query(params)}`),
  unreadCount: () => request("/api/notifications/unread-count"),
  markRead: (id) => request(`/api/notifications/${id}/read`, { method: "POST" }),
  markAllRead: () => request("/api/notifications/read-all", { method: "POST" })
};

function query(params = {}) {
  const entries = Object.entries(params).filter(([, value]) => value !== undefined && value !== null && value !== "");
  if (entries.length === 0) {
    return "";
  }
  return `?${new URLSearchParams(entries).toString()}`;
}
