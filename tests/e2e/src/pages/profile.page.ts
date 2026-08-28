import { expect, type Page } from "@playwright/test";

export class ProfilePage {
  constructor(private readonly page: Page) {}

  async goto(userID: number): Promise<void> {
    await this.page.goto(`/user/${userID}`);
  }

  async follow(userID: number): Promise<void> {
    const profile = this.page.getByRole("region", { name: "开发者资料" });
    const responsePromise = this.page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === `/api/users/${userID}/follow`
    );
    await profile.getByRole("button", { name: /^关注(?:\s|$)/ }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    await expect(profile.getByRole("button", { name: /^取消关注(?:\s|$)/ })).toBeVisible();
  }
}
