import { expect, type Locator, type Page } from "@playwright/test";

export class NotificationsPage {
  constructor(private readonly page: Page) {}

  async goto(): Promise<void> {
    await this.page.goto("/notifications");
    await expect(this.page.getByRole("heading", { name: "最近发生的互动" })).toBeVisible();
  }

  notification(notificationID: number): Locator {
    return this.page.getByTestId(`notification-${notificationID}`);
  }

  async expectText(notificationID: number, text: string): Promise<void> {
    await expect(this.notification(notificationID)).toContainText(text);
  }

  async markRead(notificationID: number): Promise<void> {
    const item = this.notification(notificationID);
    await item.getByRole("button", { name: "标记已读" }).click();
    await expect(item.getByRole("button", { name: "标记已读" })).toBeHidden();
  }

  async open(notificationID: number, postID: number): Promise<void> {
    await this.notification(notificationID)
      .getByRole("button", { name: /^打开通知：/ })
      .click();
    await expect(this.page).toHaveURL(new RegExp(`/post/${postID}$`));
  }
}
