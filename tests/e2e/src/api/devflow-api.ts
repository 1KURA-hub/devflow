import type { APIRequestContext, APIResponse } from "@playwright/test";

export interface ApiEnvelope<T> {
  code: number;
  message: string;
  // Go 的 omitempty 会让 void 成功响应没有 data 字段。
  data?: T;
}

export interface ApiResult<T> {
  response: APIResponse;
  body: ApiEnvelope<T>;
}

export interface User {
  id: number;
  username: string;
  nickname: string;
  bio: string;
  avatar_url: string;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface RegisterInput {
  username: string;
  password: string;
  nickname: string;
}

export interface AuthData {
  token: string;
  user: User;
}

export interface CreatePostInput {
  title: string;
  content: string;
  cover_url?: string;
  tags?: string;
}

export interface Post {
  id: number;
  author_id: number;
  author?: User;
  title: string;
  content: string;
  cover_url: string;
  tags: string;
  like_count: number;
  comment_count: number;
  favorite_count: number;
  status: number;
  created_at: string;
  updated_at: string;
}

export interface PostView extends Post {
  liked: boolean;
  favorited: boolean;
}

export interface Comment {
  id: number;
  post_id: number;
  user_id: number;
  user?: User;
  content: string;
  created_at: string;
}

export interface Notification {
  id: number;
  user_id: number;
  actor_id: number;
  actor?: User;
  type: "follow" | "like" | "favorite" | "comment" | string;
  post_id?: number;
  post?: Post;
  comment_id?: number;
  content: string;
  is_read: boolean;
  created_at: string;
}

export interface PageData<T> {
  items: T[];
  next_cursor?: string;
  has_more?: boolean;
}

export interface ListParams {
  cursor?: string;
  limit?: number;
}

export class DevFlowApi {
  constructor(
    private readonly request: APIRequestContext,
    private readonly token?: string
  ) {}

  as(token: string): DevFlowApi {
    return new DevFlowApi(this.request, token);
  }

  health(): Promise<ApiResult<{ status: string; app: string; env: string }>> {
    return this.call(this.request.get("/healthz"));
  }

  register(input: RegisterInput): Promise<ApiResult<AuthData>> {
    return this.call(this.request.post("/api/auth/register", { data: input }));
  }

  login(input: Pick<RegisterInput, "username" | "password">): Promise<ApiResult<AuthData>> {
    return this.call(this.request.post("/api/auth/login", { data: input }));
  }

  me(): Promise<ApiResult<User>> {
    return this.call(this.request.get("/api/me", this.authorized()));
  }

  createPost(input: CreatePostInput): Promise<ApiResult<Post>> {
    return this.call(this.request.post("/api/posts", { ...this.authorized(), data: input }));
  }

  getPost(postID: number): Promise<ApiResult<PostView>> {
    return this.call(this.request.get(`/api/posts/${postID}`, this.authorized()));
  }

  deletePost(postID: number): Promise<ApiResult<void>> {
    return this.call(this.request.delete(`/api/posts/${postID}`, this.authorized()));
  }

  follow(userID: number): Promise<ApiResult<void>> {
    return this.call(this.request.post(`/api/users/${userID}/follow`, this.authorized()));
  }

  unfollow(userID: number): Promise<ApiResult<void>> {
    return this.call(this.request.delete(`/api/users/${userID}/follow`, this.authorized()));
  }

  followState(userID: number): Promise<ApiResult<{ followed: boolean }>> {
    return this.call(this.request.get(`/api/users/${userID}/follow-state`, this.authorized()));
  }

  followingFeed(params: ListParams = {}): Promise<ApiResult<PageData<PostView>>> {
    return this.call(
      this.request.get("/api/feed/following", {
        ...this.authorized(),
        params: this.queryParams(params)
      })
    );
  }

  like(postID: number): Promise<ApiResult<void>> {
    return this.call(this.request.post(`/api/posts/${postID}/like`, this.authorized()));
  }

  unlike(postID: number): Promise<ApiResult<void>> {
    return this.call(this.request.delete(`/api/posts/${postID}/like`, this.authorized()));
  }

  favorite(postID: number): Promise<ApiResult<void>> {
    return this.call(this.request.post(`/api/posts/${postID}/favorite`, this.authorized()));
  }

  unfavorite(postID: number): Promise<ApiResult<void>> {
    return this.call(this.request.delete(`/api/posts/${postID}/favorite`, this.authorized()));
  }

  favorites(params: ListParams = {}): Promise<ApiResult<PageData<PostView>>> {
    return this.call(
      this.request.get("/api/me/favorites", {
        ...this.authorized(),
        params: this.queryParams(params)
      })
    );
  }

  createComment(postID: number, content: string): Promise<ApiResult<Comment>> {
    return this.call(
      this.request.post(`/api/posts/${postID}/comments`, {
        ...this.authorized(),
        data: { content }
      })
    );
  }

  comments(postID: number, params: ListParams = {}): Promise<ApiResult<PageData<Comment>>> {
    return this.call(
      this.request.get(`/api/posts/${postID}/comments`, { params: this.queryParams(params) })
    );
  }

  notifications(params: ListParams = {}): Promise<ApiResult<PageData<Notification>>> {
    return this.call(
      this.request.get("/api/notifications", {
        ...this.authorized(),
        params: this.queryParams(params)
      })
    );
  }

  markNotificationRead(notificationID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.post(`/api/notifications/${notificationID}/read`, this.authorized())
    );
  }

  private authorized(): { headers: Record<string, string> } {
    return {
      headers: this.token ? { Authorization: `Bearer ${this.token}` } : {}
    };
  }

  private queryParams(params: ListParams): Record<string, string | number | boolean> {
    return Object.fromEntries(
      Object.entries(params).filter((entry): entry is [string, string | number | boolean] => {
        return entry[1] !== undefined;
      })
    );
  }

  private async call<T>(requestPromise: Promise<APIResponse>): Promise<ApiResult<T>> {
    const response = await requestPromise;
    const body = (await response.json()) as ApiEnvelope<T>;
    return { response, body };
  }
}

export function requireSuccess<T>(result: ApiResult<T>, operation: string): T | undefined {
  if (result.response.status() !== 200 || result.body.code !== 0) {
    throw new Error(
      `${operation}失败：HTTP ${result.response.status()}, code=${result.body.code}, message=${result.body.message}`
    );
  }
  return result.body.data;
}

export function requireData<T>(result: ApiResult<T>, operation: string): T {
  const data = requireSuccess(result, operation);
  if (data === undefined) {
    throw new Error(`${operation}失败：成功响应缺少 data`);
  }
  return data;
}
