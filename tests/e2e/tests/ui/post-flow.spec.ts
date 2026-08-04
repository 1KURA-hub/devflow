import { test, expect } from "../../src/fixtures/test";
import { requireData } from "../../src/api/devflow-api";
import { FeedPage } from "../../src/pages/feed.page";
import { PostDetailPage } from "../../src/pages/post-detail.page";
import { uniquePostTitle } from "../../src/support/data";

test.describe("UI｜动态主流程", () => {
  test("登录用户可以通过页面发布动态", async ({ page, actor, data }) => {
    const feed = new FeedPage(page);
    const input = {
      title: uniquePostTitle("ui"),
      content: "通过 Playwright 页面发布的动态",
      tags: "Playwright,UI"
    };
    await feed.gotoLatest();
    const responsePromise = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === "/api/posts"
    );

    await feed.composer.publish(input);
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    const envelope = (await response.json()) as { data: { id: number } };
    const postID = envelope.data.id;
    data.deferApiCleanup(
      "删除页面发布的测试动态",
      () => actor.api.deletePost(postID),
      [200, 404]
    );

    await feed.expectPost(postID, input);
  });

  test("登录用户可以在动态详情页发表评论", async ({ page, actor, data }) => {
    const ownedPost = await data.createPost(actor, {
      title: uniquePostTitle("comment"),
      content: "等待评论的动态"
    });
    const detail = new PostDetailPage(page);
    const comment = `评论_${Date.now().toString(36)}`;
    await detail.goto(ownedPost.post.id);
    await detail.expectPost(ownedPost.post.title);

    await detail.addComment(comment);

    await detail.expectComment(actor.nickname, comment);
    await expect(page.getByRole("heading", { name: "1 条回应" })).toBeVisible();
    await new FeedPage(page).post(ownedPost.post.id).expectCommentCount(1);
    await expect(page.getByLabel("评论内容")).toHaveValue("");
  });

  test("点赞和收藏后，页面状态与接口数据一致", async ({ page, actor, data }) => {
    const author = await data.createActor();
    const ownedPost = await data.createPost(author, {
      title: uniquePostTitle("interaction")
    });
    data.deferApiCleanup("取消页面产生的点赞", () => actor.api.unlike(ownedPost.post.id));
    data.deferApiCleanup("取消页面产生的收藏", () => actor.api.unfavorite(ownedPost.post.id));
    const detail = new PostDetailPage(page);
    await detail.goto(ownedPost.post.id);
    await detail.expectPost(ownedPost.post.title);
    const card = new FeedPage(page).post(ownedPost.post.id);

    await card.like();
    await card.favorite();

    const post = requireData(await actor.api.getPost(ownedPost.post.id), "查询互动后的动态");
    expect(post.liked).toBe(true);
    expect(post.favorited).toBe(true);
    expect(post.like_count).toBe(1);
    expect(post.favorite_count).toBe(1);
    await card.expectInteractionCounts(post.like_count, post.favorite_count);
  });

  test("在收藏页取消收藏后，动态会从列表移除", async ({ page, actor, data }) => {
    const author = await data.createActor();
    const ownedPost = await data.createPost(author, {
      title: uniquePostTitle("unfavorite")
    });
    await data.favorite(actor, ownedPost);
    const feed = new FeedPage(page);
    await page.goto("/favorites");
    await feed.expectPost(ownedPost.post.id, ownedPost.post);
    const card = feed.post(ownedPost.post.id);

    await card.root.getByRole("button", { name: "取消收藏" }).click();

    await expect(card.root).toBeHidden();
    const favorites = requireData(await actor.api.favorites({ limit: 50 }), "查询取消后的收藏列表");
    expect(favorites.items.some((post) => post.id === ownedPost.post.id)).toBe(false);
  });
});
