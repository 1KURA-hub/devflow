import { expect, type Locator, type Page } from "@playwright/test";

export class PostCard {
  constructor(
    private readonly page: Page,
    readonly root: Locator,
    readonly postID: number
  ) {}

  async open(): Promise<void> {
    await this.root.getByRole("heading").click();
    await expect(this.page).toHaveURL(new RegExp(`/post/${this.postID}$`));
  }

  async like(): Promise<void> {
    const responsePromise = this.page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === `/api/posts/${this.postID}/like`
    );
    await this.root.getByRole("button", { name: "点赞" }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    await expect(this.root.getByRole("button", { name: "取消点赞" })).toHaveAttribute(
      "aria-pressed",
      "true"
    );
  }

  async favorite(): Promise<void> {
    const responsePromise = this.page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        new URL(response.url()).pathname === `/api/posts/${this.postID}/favorite`
    );
    await this.root.getByRole("button", { name: "收藏" }).click();
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    await expect(this.root.getByRole("button", { name: "取消收藏" })).toHaveAttribute(
      "aria-pressed",
      "true"
    );
  }

  async expectInteractionCounts(likeCount: number, favoriteCount: number): Promise<void> {
    await expect(this.root.getByRole("button", { name: "取消点赞" })).toContainText(
      String(likeCount)
    );
    await expect(this.root.getByRole("button", { name: "取消收藏" })).toContainText(
      String(favoriteCount)
    );
  }

  async expectCommentCount(commentCount: number): Promise<void> {
    await expect(this.root.getByRole("link", { name: "查看评论" })).toContainText(
      String(commentCount)
    );
  }
}
