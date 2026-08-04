import { expect, type Page } from "@playwright/test";

export class AppShell {
  readonly mainNavigation;

  constructor(private readonly page: Page) {
    this.mainNavigation = page.getByRole("navigation", { name: "主导航" });
  }

  async openFollowing(): Promise<void> {
    await this.mainNavigation.getByRole("link", { name: "关注", exact: true }).click();
    await expect(this.page).toHaveURL(/\/following$/);
  }

  async openFavorites(): Promise<void> {
    await this.mainNavigation.getByRole("link", { name: "收藏", exact: true }).click();
    await expect(this.page).toHaveURL(/\/favorites$/);
  }

  async openNotifications(): Promise<void> {
    await this.mainNavigation.getByRole("link", { name: /通知/ }).click();
    await expect(this.page).toHaveURL(/\/notifications$/);
  }

  async expectLoggedIn(nickname: string): Promise<void> {
    await expect(this.page.locator(".sidebar-profile").getByText(nickname, { exact: true })).toBeVisible();
    await expect(this.mainNavigation.getByRole("link", { name: "我的" })).toBeVisible();
  }
}
