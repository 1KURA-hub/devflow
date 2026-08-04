import { expect, type Page } from "@playwright/test";
import type { CreatePostInput } from "../api/devflow-api";
import { ComposerModal } from "./composer.modal";
import { PostCard } from "./post-card";

export class FeedPage {
  readonly composer: ComposerModal;

  constructor(private readonly page: Page) {
    this.composer = new ComposerModal(page);
  }

  async gotoLatest(): Promise<void> {
    await this.page.goto("/");
    await expect(this.page.getByRole("navigation", { name: "主导航" })).toBeVisible();
  }

  post(postID: number): PostCard {
    return new PostCard(this.page, this.page.getByTestId(`post-card-${postID}`), postID);
  }

  async expectPost(postID: number, input?: Partial<CreatePostInput>): Promise<void> {
    const card = this.page.getByTestId(`post-card-${postID}`);
    await expect(card).toBeVisible();
    if (input?.title) {
      await expect(card.getByRole("heading", { name: input.title, exact: true })).toBeVisible();
    }
    if (input?.content) {
      await expect(card).toContainText(input.content);
    }
    if (input?.tags) {
      for (const tag of input.tags.split(",").map((item) => item.trim())) {
        await expect(card.getByText(tag, { exact: true })).toBeVisible();
      }
    }
  }
}
