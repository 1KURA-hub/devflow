import { apiTest, expect } from "../../src/fixtures/test";
import { AuthPage } from "../../src/pages/auth.page";
import { AppShell } from "../../src/pages/app-shell";
import { newUserInput } from "../../src/support/data";

apiTest.describe("UI｜认证", () => {
  apiTest("未登录访问收藏页会跳转登录页", async ({ page }) => {
    const authPage = new AuthPage(page);

    await page.goto("/favorites");

    await authPage.expectLoginPage();
  });

  apiTest("用户登录成功，刷新后登录态仍然保留", async ({ page, data }) => {
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

  apiTest("用户可以通过注册页创建账号并进入首页", async ({ page }) => {
    const input = newUserInput();
    const authPage = new AuthPage(page);
    const appShell = new AppShell(page);
    await authPage.gotoRegister();

    await authPage.register(input);

    await expect(page).toHaveURL(/\/$/);
    await appShell.expectLoggedIn(input.nickname);
  });
});
