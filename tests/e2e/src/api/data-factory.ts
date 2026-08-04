import type {
  ApiResult,
  CreatePostInput,
  Notification,
  Post,
  RegisterInput,
  User
} from "./devflow-api";
import { DevFlowApi, requireData, requireSuccess } from "./devflow-api";
import { newUserInput, uniquePostTitle } from "../support/data";

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

type CleanupAction = () => Promise<unknown>;

export class DevFlowDataFactory {
  private readonly cleanupActions: CleanupAction[] = [];

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
    this.deferApiCleanup(
      "删除测试动态",
      () => author.api.deletePost(post.id),
      // 用例可能已经主动删除动态，404 也代表目标数据不存在。
      [200, 404]
    );
    return { post, author };
  }

  async follow(actor: TestActor, target: TestActor): Promise<void> {
    requireSuccess(await actor.api.follow(target.user.id), "关注测试用户");
    this.deferApiCleanup("取消测试关注", () => actor.api.unfollow(target.user.id));
  }

  async like(actor: TestActor, ownedPost: OwnedPost): Promise<void> {
    requireSuccess(await actor.api.like(ownedPost.post.id), "点赞测试动态");
    this.deferApiCleanup("取消测试点赞", () => actor.api.unlike(ownedPost.post.id));
  }

  async favorite(actor: TestActor, ownedPost: OwnedPost): Promise<void> {
    requireSuccess(await actor.api.favorite(ownedPost.post.id), "收藏测试动态");
    this.deferApiCleanup("取消测试收藏", () => actor.api.unfavorite(ownedPost.post.id));
  }

  async waitForNotification(
    recipient: TestActor,
    matcher: (notification: Notification) => boolean,
    timeout = 5_000
  ): Promise<Notification> {
    const deadline = Date.now() + timeout;
    while (Date.now() < deadline) {
      const page = requireData(
        await recipient.api.notifications({ limit: 50 }),
        "查询通知列表"
      );
      const matched = page.items.find(matcher);
      if (matched) {
        return matched;
      }
      await new Promise((resolve) => setTimeout(resolve, 200));
    }
    throw new Error(`等待通知超时（${timeout}ms）`);
  }

  deferApiCleanup(
    operation: string,
    action: () => Promise<ApiResult<unknown>>,
    acceptedStatuses: number[] = [200]
  ): void {
    this.cleanupActions.push(async () => {
      const result = await action();
      if (!acceptedStatuses.includes(result.response.status())) {
        throw new Error(
          `${operation}失败：HTTP ${result.response.status()}, code=${result.body.code}, message=${result.body.message}`
        );
      }
      if (result.response.status() === 200) {
        requireSuccess(result, operation);
      }
    });
  }

  async cleanup(): Promise<void> {
    for (const action of this.cleanupActions.reverse()) {
      try {
        await action();
      } catch (error) {
        // 清理失败不覆盖用例的原始结论；CI数据库本身也是临时的。
        console.warn("[e2e cleanup]", error);
      }
    }
  }
}
