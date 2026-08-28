import { expect, type Locator, type Page } from "@playwright/test";
import type { CreatePostInput } from "../api/devflow-api";

export class ComposerModal {
  readonly dialog: Locator;

  constructor(private readonly page: Page) {
    this.dialog = page.getByRole("dialog", { name: "发布动态" });
  }

  async open(): Promise<void> {
    await this.page.getByRole("button", { name: "创建动态" }).click();
    await expect(this.dialog).toBeVisible();
  }

  async fill(input: CreatePostInput): Promise<void> {
    await this.dialog.getByLabel("标题").fill(input.title);
    await this.dialog.getByLabel("正文").fill(input.content);
    if (input.tags !== undefined) {
      await this.dialog.getByLabel("标签").fill(input.tags);
    }
  }

  async submit(): Promise<void> {
    await this.dialog.getByRole("button", { name: "发布", exact: true }).click();
  }

  async publish(input: CreatePostInput): Promise<void> {
    await this.open();
    await this.fill(input);
    await this.submit();
    await expect(this.dialog).toBeHidden();
  }

}
