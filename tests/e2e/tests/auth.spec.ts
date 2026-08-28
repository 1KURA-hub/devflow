import { unauthenticatedTest, expect } from "../src/fixtures/test";
import { AppShell } from "../src/components/app-shell";
import { AuthPage } from "../src/pages/auth.page";

unauthenticatedTest.describe("UI｜认证", () => {
  unauthenticatedTest("未登录访问收藏页会跳转登录页", async ({ page }) => {
    const authPage = new AuthPage(page);

    await page.goto("/favorites");

    await authPage.expectLoginPage();
  });

  unauthenticatedTest("用户登录成功，刷新后登录态仍然保留", async ({
    page,
    data
  }) => {
    const actor = await data.createActor();
    const authPage = new AuthPage(page);
    const appShell = new AppShell(page);
    await authPage.gotoLogin();

    await authPage.login(actor);

    await expect(page).toHaveURL(/\/$/);
    await appShell.expectLoggedIn(actor.nickname);
    await page.reload();
    await appShell.expectLoggedIn(actor.nickname);
  });

  unauthenticatedTest("错误密码登录会显示统一错误信息", async ({ page, data }) => {
    const actor = await data.createActor();
    const authPage = new AuthPage(page);
    await authPage.gotoLogin();

    await authPage.login({
      username: actor.username,
      password: "wrong123"
    });

    await expect(page).toHaveURL(/\/login$/);
    await expect(authPage.errorMessage()).toHaveText("invalid username or password");
  });
});
