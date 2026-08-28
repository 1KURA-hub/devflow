import type { APIRequestContext, APIResponse } from "@playwright/test";

export interface ApiEnvelope<T> {
  code: number;
  message: string;
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
  tags?: string;
}

export interface Post {
  id: number;
  title: string;
  content: string;
  tags: string;
  like_count: number;
  comment_count: number;
  favorite_count: number;
}

export interface PostView extends Post {
  liked: boolean;
  favorited: boolean;
}

export interface Notification {
  id: number;
  actor_id: number;
  type: string;
  post_id?: number;
  content: string;
  is_read: boolean;
}

export interface PageData<T> {
  items: T[];
}

export class DevFlowApi {
  constructor(
    private readonly request: APIRequestContext,
    private readonly apiBaseURL: string,
    private readonly token?: string
  ) {}

  as(token: string): DevFlowApi {
    return new DevFlowApi(this.request, this.apiBaseURL, token);
  }

  register(input: RegisterInput): Promise<ApiResult<AuthData>> {
    return this.call(this.request.post(this.url("/api/auth/register"), { data: input }));
  }

  createPost(input: CreatePostInput): Promise<ApiResult<Post>> {
    return this.call(
      this.request.post(this.url("/api/posts"), {
        ...this.authorized(),
        data: input
      })
    );
  }

  getPost(postID: number): Promise<ApiResult<PostView>> {
    return this.call(
      this.request.get(this.url(`/api/posts/${postID}`), this.authorized())
    );
  }

  deletePost(postID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.delete(this.url(`/api/posts/${postID}`), this.authorized())
    );
  }

  unfollow(userID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.delete(this.url(`/api/users/${userID}/follow`), this.authorized())
    );
  }

  followState(userID: number): Promise<ApiResult<{ followed: boolean }>> {
    return this.call(
      this.request.get(this.url(`/api/users/${userID}/follow-state`), this.authorized())
    );
  }

  like(postID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.post(this.url(`/api/posts/${postID}/like`), this.authorized())
    );
  }

  unlike(postID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.delete(this.url(`/api/posts/${postID}/like`), this.authorized())
    );
  }

  favorite(postID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.post(this.url(`/api/posts/${postID}/favorite`), this.authorized())
    );
  }

  unfavorite(postID: number): Promise<ApiResult<void>> {
    return this.call(
      this.request.delete(this.url(`/api/posts/${postID}/favorite`), this.authorized())
    );
  }

  favorites(limit = 50): Promise<ApiResult<PageData<PostView>>> {
    return this.call(
      this.request.get(this.url("/api/me/favorites"), {
        ...this.authorized(),
        params: { limit }
      })
    );
  }

  notifications(limit = 50): Promise<ApiResult<PageData<Notification>>> {
    return this.call(
      this.request.get(this.url("/api/notifications"), {
        ...this.authorized(),
        params: { limit }
      })
    );
  }

  private url(pathname: string): string {
    return new URL(pathname, this.apiBaseURL).toString();
  }

  private authorized(): { headers: Record<string, string> } {
    return {
      headers: this.token ? { Authorization: `Bearer ${this.token}` } : {}
    };
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
