import type { ApiResult, CreatePostInput, Notification, Post, RegisterInput, User } from "../api/devflow-api";
import { DevFlowApi, requireData, requireSuccess } from "../api/devflow-api";
import { newUserInput, uniquePostTitle } from "./data";

export interface TestActor {
  username: string;
  password: string;
  nickname: string;
  token: string;
  user: User;
  api: DevFlowApi;
}

export interface OwnedPost {
  post: Post;
  author: TestActor;
}

interface TrackedPost {
  actor: TestActor;
  postID: number;
}

interface TrackedUserRelation {
  actor: TestActor;
  targetUserID: number;
}

export class TestData {
  private readonly posts: TrackedPost[] = [];
  private readonly likes: TrackedPost[] = [];
  private readonly favorites: TrackedPost[] = [];
  private readonly follows: TrackedUserRelation[] = [];

  constructor(private readonly api: DevFlowApi) {}

  async createActor(overrides: Partial<RegisterInput> = {}): Promise<TestActor> {
    const credentials = { ...newUserInput(), ...overrides };
    const auth = requireData(await this.api.register(credentials), "注册测试用户");
    return {
      ...credentials,
      token: auth.token,
      user: auth.user,
      api: this.api.as(auth.token)
    };
  }

  async createPost(
    author: TestActor,
    overrides: Partial<CreatePostInput> = {}
  ): Promise<OwnedPost> {
    const input: CreatePostInput = {
      title: uniquePostTitle(),
      content: "Playwright 自动化创建的测试动态",
      tags: "Playwright,TypeScript",
      ...overrides
    };
    const post = requireData(await author.api.createPost(input), "创建测试动态");
    this.trackPost(author, post.id);
    return { post, author };
  }

  async like(actor: TestActor, ownedPost: OwnedPost): Promise<void> {
    requireSuccess(await actor.api.like(ownedPost.post.id), "点赞测试动态");
    this.trackLike(actor, ownedPost.post.id);
  }

  async favorite(actor: TestActor, ownedPost: OwnedPost): Promise<void> {
    requireSuccess(await actor.api.favorite(ownedPost.post.id), "收藏测试动态");
    this.trackFavorite(actor, ownedPost.post.id);
  }

  trackPost(actor: TestActor, postID: number): void {
    this.posts.push({ actor, postID });
  }

  trackLike(actor: TestActor, postID: number): void {
    this.likes.push({ actor, postID });
  }

  trackFavorite(actor: TestActor, postID: number): void {
    this.favorites.push({ actor, postID });
  }

  trackFollow(actor: TestActor, targetUserID: number): void {
    this.follows.push({ actor, targetUserID });
  }

  async waitForNotification(
    recipient: TestActor,
    matcher: (notification: Notification) => boolean,
    timeout = 5_000
  ): Promise<Notification> {
    const deadline = Date.now() + timeout;
    while (Date.now() < deadline) {
      const page = requireData(await recipient.api.notifications(), "查询通知列表");
      const notification = page.items.find(matcher);
      if (notification) {
        return notification;
      }
      await new Promise((resolve) => setTimeout(resolve, 200));
    }
    throw new Error(`等待通知超时（${timeout}ms）`);
  }

  async cleanup(): Promise<void> {
    const failures: unknown[] = [];

    for (const item of [...this.favorites].reverse()) {
      await this.cleanupResult(
        "取消测试收藏",
        item.actor.api.unfavorite(item.postID),
        failures
      );
    }
    for (const item of [...this.likes].reverse()) {
      await this.cleanupResult(
        "取消测试点赞",
        item.actor.api.unlike(item.postID),
        failures
      );
    }
    for (const item of [...this.follows].reverse()) {
      await this.cleanupResult(
        "取消测试关注",
        item.actor.api.unfollow(item.targetUserID),
        failures
      );
    }
    for (const item of [...this.posts].reverse()) {
      await this.cleanupResult(
        "删除测试动态",
        item.actor.api.deletePost(item.postID),
        failures
      );
    }

    if (failures.length > 0) {
      throw new AggregateError(failures, `测试数据清理失败（${failures.length}项）`);
    }
  }

  private async cleanupResult(
    operation: string,
    resultPromise: Promise<ApiResult<unknown>>,
    failures: unknown[]
  ): Promise<void> {
    try {
      const result = await resultPromise;
      if (result.response.status() === 404) {
        return;
      }
      requireSuccess(result, operation);
    } catch (error) {
      console.warn(`[e2e cleanup] ${operation}`, error);
      failures.push(error);
    }
  }
}
