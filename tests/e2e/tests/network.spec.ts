import { test, expect } from "../src/fixtures/test";
import { FeedPage } from "../src/pages/feed.page";
import { PostDetailPage } from "../src/pages/post-detail.page";

test.describe("UI｜网络异常", () => {
  test("动态列表返回500时显示后端错误", async ({ page }) => {
    await page.route("**/api/feed/latest", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ code: 500, message: "动态列表暂不可用" })
      });
    });

    await new FeedPage(page).gotoLatest();

    await expect(page.locator(".feed-column .state-box")).toHaveText(
      "动态列表暂不可用"
    );
  });

  test("动态列表断网时显示网络错误", async ({ page }) => {
    await page.route("**/api/feed/latest", async (route) => {
      await route.abort("failed");
    });

    await new FeedPage(page).gotoLatest();

    await expect(page.locator(".feed-column .state-box")).toHaveText(
      "网络连接失败，请检查网络后重试"
    );
  });

  test("当前用户接口返回401时清除失效token", async ({ page }) => {
    await page.route("**/api/me", async (route) => {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ code: 401, message: "unauthorized" })
      });
    });
    const responsePromise = page.waitForResponse(
      (response) => new URL(response.url()).pathname === "/api/me"
    );

    await page.goto("/");
    await responsePromise;

    await expect
      .poll(() => page.evaluate(() => localStorage.getItem("devflow_token")))
      .toBeNull();
  });

  test("点赞接口返回500时乐观更新会回滚", async ({ page, data }) => {
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
    if (initialCount === null) {
      throw new Error("点赞按钮缺少初始计数");
    }
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
    await expect(rolledBackButton).toHaveText(initialCount);
  });
});
