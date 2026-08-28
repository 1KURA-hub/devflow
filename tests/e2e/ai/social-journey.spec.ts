import { requireData } from "../src/api/devflow-api";
import { assertMidsceneModelConfigured } from "../src/config/midscene-env";
import { test, expect } from "../src/fixtures/midscene-test";
import { uniquePostTitle } from "../src/support/data";

test.describe("AI｜Midscene 跨用户互动闭环", () => {
  test.beforeAll(() => {
    assertMidsceneModelConfigured();
  });

  test("访客关注并互动后，作者可从通知返回关联动态", async ({
    page,
    actor,
    data,
    aiAct,
    aiAssert,
    recordToReport
  }) => {
    const author = await data.createActor();
    const ownedPost = await data.createPost(author, {
      title: uniquePostTitle("midscene"),
      content: "这是一条用于 Midscene 跨用户业务闭环的动态。",
      tags: "Midscene,AI E2E"
    });
    const comment = `Midscene 自然语言评论_${Date.now().toString(36)}`;

    // 先记录可能被 AI 创建的关系，即使中途失败 teardown 也会尝试清理。
    data.trackFollow(actor, author.user.id);
    data.trackLike(actor, ownedPost.post.id);
    data.trackFavorite(actor, ownedPost.post.id);

    await page.goto(`/user/${author.user.id}`);
    await expect(page.getByRole("region", { name: "开发者资料" })).toBeVisible();

    await aiAct(
      `请完成一段跨页面操作：
1. 在开发者资料区域点击用于关注当前作者的按钮，不要点资料统计里的“关注”标签。
2. 通过主导航进入“关注”动态流。
3. 找到标题为“${ownedPost.post.title}”的动态并打开详情。
完成后停留在这条动态的详情页。`
    );

    await expect(page).toHaveURL(new RegExp(`/post/${ownedPost.post.id}$`));
    await aiAct(
      `在当前动态详情页完成全部互动：
1. 点赞这条动态。
2. 收藏这条动态。
3. 在“评论内容”输入框填写“${comment}”并发表评论。
看到新评论出现在回应列表后结束。`
    );
    await aiAssert(
      `当前是标题为“${ownedPost.post.title}”的动态详情页，回应列表中显示了评论“${comment}”。`
    );

    const followState = requireData(
      await actor.api.followState(author.user.id),
      "校验 Midscene 关注结果"
    );
    const postAfterInteraction = requireData(
      await actor.api.getPost(ownedPost.post.id),
      "校验 Midscene 互动结果"
    );
    expect(followState.followed).toBe(true);
    expect(postAfterInteraction.liked).toBe(true);
    expect(postAfterInteraction.favorited).toBe(true);
    expect(postAfterInteraction.like_count).toBe(1);
    expect(postAfterInteraction.favorite_count).toBe(1);
    expect(postAfterInteraction.comment_count).toBe(1);

    const commentNotification = await data.waitForNotification(
      author,
      (item) =>
        item.type === "comment" &&
        item.actor_id === actor.user.id &&
        item.post_id === ownedPost.post.id
    );
    await recordToReport("访客互动完成", {
      content: `API 已确认关注、点赞、收藏和评论生效，评论通知 ID=${commentNotification.id}。`
    });

    // 测试基础设施直接切换作者 token，避免把重复登录表单纳入模型调用。
    await page.evaluate((token) => {
      localStorage.setItem("devflow_token", token);
    }, author.token);
    await page.goto("/notifications");
    await expect(page.getByRole("heading", { name: "最近发生的互动" })).toBeVisible();

    await aiAct(
      `在通知中心找到“${actor.nickname}”对《${ownedPost.post.title}》的评论通知。
先将这条通知标记为已读，再打开这条通知关联的动态，最终停留在动态详情页。`
    );
    await expect(page).toHaveURL(new RegExp(`/post/${ownedPost.post.id}$`));
    await aiAssert(
      `当前显示《${ownedPost.post.title}》的动态详情，并且评论区可以看到“${comment}”。`
    );

    const notificationsAfterRead = requireData(
      await author.api.notifications(),
      "校验 Midscene 通知已读结果"
    );
    expect(
      notificationsAfterRead.items.find(
        (item) => item.id === commentNotification.id
      )?.is_read
    ).toBe(true);
    await recordToReport("跨用户闭环完成", {
      content: "作者已从评论通知返回关联动态，且 API 确认通知已读。"
    });
  });
});
