import { test, expect } from "../../src/fixtures/test";
import { ComposerModal } from "../../src/pages/composer.modal";
import { FeedPage } from "../../src/pages/feed.page";
import { PostDetailPage } from "../../src/pages/post-detail.page";
import { uniquePostTitle } from "../../src/support/data";

test.describe("UI｜网络异常处理", () => {
  test("动态列表返回 500 时显示错误，不用演示数据伪装成功", async ({ page }) => {
    await page.route("**/api/feed/latest", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ code: 500, message: "动态列表暂不可用" })
      });
    });

    await new FeedPage(page).gotoLatest();

    await expect(page.locator(".feed-column .state-box")).toHaveText("动态列表暂不可用");
    await expect(page.getByText("移动端首页改成动态优先", { exact: true })).toHaveCount(0);
  });

  test("真实空动态列表显示空态，不显示演示动态", async ({ page }) => {
    await page.route("**/api/feed/latest", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          code: 0,
          message: "ok",
          data: { items: [], has_more: false }
        })
      });
    });

    await new FeedPage(page).gotoLatest();

    await expect(page.getByRole("heading", { name: "这里还很安静" })).toBeVisible();
    await expect(page.getByText("移动端首页改成动态优先", { exact: true })).toHaveCount(0);
  });

  test("当前用户接口返回 500 时保留本地 token", async ({ page, actor }) => {
    await page.route("**/api/me", async (route) => {
      await route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ code: 500, message: "用户服务暂不可用" })
      });
    });
    const responsePromise = page.waitForResponse(
      (response) => new URL(response.url()).pathname === "/api/me"
    );

    await page.goto("/");
    await responsePromise;

    await expect
      .poll(() => page.evaluate(() => localStorage.getItem("devflow_token")))
      .toBe(actor.token);
  });

  test("当前用户接口返回 401 时清除失效 token", async ({ page }) => {
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

  test("动态列表断网时显示可判断的网络错误", async ({ page }) => {
    await page.route("**/api/feed/latest", async (route) => {
      await route.abort("failed");
    });

    await new FeedPage(page).gotoLatest();

    await expect(page.locator(".feed-column .state-box")).toHaveText(
      "网络连接失败，请检查网络后重试"
    );
  });

  test("切换用户主页后，迟到的旧响应不会覆盖当前用户", async ({ page, data }) => {
    const firstUser = await data.createActor();
    const secondUser = await data.createActor();
    await data.createPost(firstUser, { title: uniquePostTitle("profile-a") });
    await data.createPost(secondUser, { title: uniquePostTitle("profile-b") });
    let releaseFirstResponse!: () => void;
    const firstResponseGate = new Promise<void>((resolve) => {
      releaseFirstResponse = resolve;
    });
    await page.route(`**/api/users/${firstUser.user.id}`, async (route) => {
      await firstResponseGate;
      await route.continue();
    });
    const firstRequest = page.waitForRequest(
      (request) => new URL(request.url()).pathname === `/api/users/${firstUser.user.id}`
    );
    await page.goto(`/user/${firstUser.user.id}`);
    await firstRequest;

    await page.evaluate((path) => {
      window.history.pushState({}, "", path);
      window.dispatchEvent(new PopStateEvent("popstate"));
    }, `/user/${secondUser.user.id}`);
    await expect(page.getByRole("heading", { name: secondUser.nickname, exact: true })).toBeVisible();
    const lateResponse = page.waitForResponse(
      (response) => new URL(response.url()).pathname === `/api/users/${firstUser.user.id}`
    );
    releaseFirstResponse();
    await lateResponse;

    await expect(page.getByRole("heading", { name: secondUser.nickname, exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { name: firstUser.nickname, exact: true })).toHaveCount(0);
  });

  test("切换动态后，迟到的评论响应不会污染当前动态", async ({ page, actor, data }) => {
    const firstPost = await data.createPost(actor, { title: uniquePostTitle("comment-a") });
    const secondPost = await data.createPost(actor, { title: uniquePostTitle("comment-b") });
    const comment = `late-comment-${Date.now()}`;
    let releaseCommentResponse!: () => void;
    const commentResponseGate = new Promise<void>((resolve) => {
      releaseCommentResponse = resolve;
    });
    await page.route(`**/api/posts/${firstPost.post.id}/comments`, async (route) => {
      if (route.request().method() === "POST") {
        await commentResponseGate;
      }
      await route.continue();
    });

    const detail = new PostDetailPage(page);
    await detail.goto(firstPost.post.id);
    await detail.expectPost(firstPost.post.title);
    const commentRequest = page.waitForRequest(
      (request) =>
        request.method() === "POST" &&
        new URL(request.url()).pathname === `/api/posts/${firstPost.post.id}/comments`
    );
    await page.getByLabel("评论内容").fill(comment);
    await page.getByRole("button", { name: "发表评论" }).click();
    await commentRequest;

    await page.evaluate((path) => {
      window.history.pushState({}, "", path);
      window.dispatchEvent(new PopStateEvent("popstate"));
    }, `/post/${secondPost.post.id}`);
    await detail.expectPost(secondPost.post.title);
    const lateResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === `/api/posts/${firstPost.post.id}/comments`
    );
    releaseCommentResponse();
    await lateResponse;

    await detail.expectPost(secondPost.post.title);
    await expect(page.locator(".comment-list").getByText(comment, { exact: true })).toHaveCount(0);
    await expect(page.getByRole("heading", { name: "0 条回应" })).toBeVisible();
    await expect(page.getByLabel("评论内容")).toHaveValue("");
  });

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
