const tokenKey = "devflow_token";
const configuredTimeout = Number(import.meta.env.VITE_API_TIMEOUT_MS);
const defaultRequestTimeoutMS =
  Number.isFinite(configuredTimeout) && configuredTimeout > 0 ? configuredTimeout : 10_000;

export class ApiError extends Error {
  constructor(message, { status = 0, code = null, kind = "api", cause } = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.kind = kind;
    if (cause) {
      this.cause = cause;
    }
  }
}

export function isUnauthorizedError(error) {
  return error instanceof ApiError && error.status === 401;
}

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
  const { signal: externalSignal, timeout = defaultRequestTimeoutMS, ...fetchOptions } = options;
  const token = getStoredToken();
  const isFormData = fetchOptions.body instanceof FormData;
  const headers = {
    ...(isFormData ? {} : { "Content-Type": "application/json" }),
    ...(fetchOptions.headers || {})
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const controller = new AbortController();
  let timedOut = false;
  const abortFromExternalSignal = () => controller.abort(externalSignal?.reason);
  if (externalSignal?.aborted) {
    abortFromExternalSignal();
  } else {
    externalSignal?.addEventListener("abort", abortFromExternalSignal, { once: true });
  }
  const timeoutID = globalThis.setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeout);

  let response;
  try {
    response = await fetch(path, {
      ...fetchOptions,
      headers,
      signal: controller.signal
    });
  } catch (error) {
    if (timedOut) {
      throw new ApiError("请求超时，请稍后重试", { kind: "timeout", cause: error });
    }
    if (controller.signal.aborted) {
      throw new ApiError("请求已取消", { kind: "aborted", cause: error });
    }
    throw new ApiError("网络连接失败，请检查网络后重试", { kind: "network", cause: error });
  } finally {
    globalThis.clearTimeout(timeoutID);
    externalSignal?.removeEventListener("abort", abortFromExternalSignal);
  }

  const payload = await response.json().catch(() => null);
  if (!response.ok || payload?.code !== 0) {
    throw new ApiError(payload?.message || "请求失败", {
      status: response.status,
      code: payload?.code ?? response.status,
      kind: "http"
    });
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
  communityOverview: () => request("/api/community/overview"),
  recommendedUsers: (params) => request(`/api/recommended-users${query(params)}`),
  getPost: (id) => request(`/api/posts/${id}`),
  createPost: (body) =>
    request("/api/posts", {
      method: "POST",
      body: JSON.stringify(body)
    }),
  deletePost: (id) => request(`/api/posts/${id}`, { method: "DELETE" }),
  userProfile: (id) => request(`/api/users/${id}`),
  userPosts: (id, params) => request(`/api/users/${id}/posts${query(params)}`),
  followers: (id, params) => request(`/api/users/${id}/followers${query(params)}`),
  followingUsers: (id, params) => request(`/api/users/${id}/following${query(params)}`),
  followState: (id) => request(`/api/users/${id}/follow-state`),
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
