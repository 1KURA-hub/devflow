import { expect, type Page } from "@playwright/test";

export class PostDetailPage {
  constructor(private readonly page: Page) {}

  async goto(postID: number): Promise<void> {
    await this.page.goto(`/post/${postID}`);
  }

  async expectPost(title: string): Promise<void> {
    await expect(this.page.getByRole("heading", { name: title, exact: true })).toBeVisible();
  }

  async addComment(content: string): Promise<void> {
    await this.page.getByLabel("评论内容").fill(content);
    await this.page.getByRole("button", { name: "发表评论" }).click();
  }

  async expectComment(nickname: string, content: string): Promise<void> {
    const comment = this.page.locator(".comment-list article").filter({ hasText: content });
    await expect(comment).toContainText(nickname);
    await expect(comment).toContainText(content);
  }
}
