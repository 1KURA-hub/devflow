import { test, expect } from "../../src/fixtures/test";
import { ComposerModal } from "../../src/pages/composer.modal";
import { FeedPage } from "../../src/pages/feed.page";
import { PostDetailPage } from "../../src/pages/post-detail.page";
import { uniquePostTitle } from "../../src/support/data";

test.describe("UI｜网络异常处理", () => {
  test("发布接口返回 500 时显示错误并保留草稿", async ({ page }) => {
    await page.route("**/api/posts", async (route) => {
      if (route.request().method() === "POST") {
        await route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ code: 500, message: "动态服务暂不可用" })
        });
        return;
      }
      await route.continue();
    });
    const feed = new FeedPage(page);
    const composer = new ComposerModal(page);
    const draft = {
      title: uniquePostTitle("draft"),
      content: "接口失败后不能丢失的正文",
      tags: "Network,500"
    };
    await feed.gotoLatest();
    await composer.open();
    await composer.fill(draft);

    await composer.submit();

    await expect(composer.errorMessage()).toHaveText("动态服务暂不可用");
    await composer.expectDraft(draft);
    await expect(composer.dialog.getByRole("button", { name: "发布", exact: true })).toBeEnabled();
  });

  test("动态详情接口返回 500 时页面显示可判断的错误", async ({ page, actor, data }) => {
    const ownedPost = await data.createPost(actor);
    await page.route(`**/api/posts/${ownedPost.post.id}`, async (route) => {
      if (route.request().method() === "GET") {
        await route.fulfill({
          status: 500,
          contentType: "application/json",
          body: JSON.stringify({ code: 500, message: "动态服务暂不可用" })
        });
        return;
      }
      await route.continue();
    });

    await new PostDetailPage(page).goto(ownedPost.post.id);

    await expect(page.locator(".state-box")).toHaveText("动态服务暂不可用");
  });

  test("点赞接口返回 500 时乐观更新会回滚", async ({ page, data }) => {
    const author = await data.createActor();
    const ownedPost = await data.createPost(author);
    await page.route(`**/api/posts/${ownedPost.post.id}/like`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ code: 500, message: "互动服务暂不可用" })
      });
    });
    await new PostDetailPage(page).goto(ownedPost.post.id);
    const card = new FeedPage(page).post(ownedPost.post.id);
    const likeButton = card.root.getByRole("button", { name: "点赞" });
    const initialCount = await likeButton.textContent();
    expect(initialCount).not.toBeNull();
    const dialogHandled = new Promise<void>((resolve, reject) => {
      page.once("dialog", async (dialog) => {
        try {
          expect(dialog.message()).toBe("互动服务暂不可用");
          await dialog.accept();
          resolve();
        } catch (error) {
          reject(error);
        }
      });
    });

    await likeButton.click();
    await dialogHandled;

    const rolledBackButton = card.root.getByRole("button", { name: "点赞" });
    await expect(rolledBackButton).toHaveAttribute("aria-pressed", "false");
    await expect(rolledBackButton).toHaveText(initialCount!);
  });
});
