import { apiTest, expect } from "../../src/fixtures/test";
import { requireData } from "../../src/api/devflow-api";
import { uniquePostTitle } from "../../src/support/data";

apiTest.describe("API｜动态与关注", () => {
  apiTest("作者可以创建、查询并删除动态", async ({ data }) => {
    const actor = await data.createActor();
    const input = {
      title: uniquePostTitle("lifecycle"),
      content: "验证动态 API 生命周期",
      tags: "API,Smoke"
    };
    const { post } = await data.createPost(actor, input);

    const detail = await actor.api.getPost(post.id);
    expect(detail.response.status()).toBe(200);
    expect(detail.body.data).toMatchObject({
      id: post.id,
      author_id: actor.user.id,
      ...input,
      liked: false,
      favorited: false
    });

    const deleted = await actor.api.deletePost(post.id);
    expect(deleted.response.status()).toBe(200);
    expect(deleted.body).toEqual({ code: 0, message: "ok" });

    const missing = await actor.api.getPost(post.id);
    expect(missing.response.status()).toBe(404);
  });

  apiTest("未登录用户不能创建动态", async ({ api }) => {
    const result = await api.createPost({
      title: uniquePostTitle("anonymous"),
      content: "未登录请求",
      tags: "API"
    });

    expect(result.response.status()).toBe(401);
    expect(result.body).toMatchObject({
      code: 401,
      message: "missing authorization header"
    });
  });

  apiTest("非作者删除动态返回 403", async ({ data }) => {
    const author = await data.createActor();
    const stranger = await data.createActor();
    const ownedPost = await data.createPost(author);

    const result = await stranger.api.deletePost(ownedPost.post.id);

    expect(result.response.status()).toBe(403);
    expect(result.body).toMatchObject({ code: 403, message: "forbidden" });
    expect((await author.api.getPost(ownedPost.post.id)).response.status()).toBe(200);
  });

  apiTest("重复关注保持幂等且关注状态为 true", async ({ data }) => {
    const follower = await data.createActor();
    const target = await data.createActor();

    const first = await follower.api.follow(target.user.id);
    const second = await follower.api.follow(target.user.id);
    data.deferApiCleanup("取消重复关注用例的关注关系", () =>
      follower.api.unfollow(target.user.id)
    );

    expect(first.response.status()).toBe(200);
    expect(second.response.status()).toBe(200);
    expect(first.body.code).toBe(0);
    expect(second.body.code).toBe(0);

    const state = await follower.api.followState(target.user.id);
    expect(state.body.data).toEqual({ followed: true });
  });
});
