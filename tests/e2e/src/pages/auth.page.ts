import { expect, type Page } from "@playwright/test";
import type { RegisterInput } from "../api/devflow-api";

export class AuthPage {
  constructor(private readonly page: Page) {}

  async gotoLogin(): Promise<void> {
    await this.page.goto("/login");
  }

  async gotoRegister(): Promise<void> {
    await this.page.goto("/register");
  }

  async login(credentials: Pick<RegisterInput, "username" | "password">): Promise<void> {
    await this.page.getByLabel("用户名").fill(credentials.username);
    await this.page.getByLabel("密码").fill(credentials.password);
    await this.page.getByRole("button", { name: "登录", exact: true }).click();
  }

  async register(input: RegisterInput): Promise<void> {
    await this.page.getByLabel("用户名").fill(input.username);
    await this.page.getByLabel("昵称").fill(input.nickname);
    await this.page.getByLabel("密码").fill(input.password);
    await this.page.getByRole("button", { name: "注册", exact: true }).click();
  }

  async expectLoginPage(): Promise<void> {
    await expect(this.page).toHaveURL(/\/login$/);
    await expect(this.page.getByRole("heading", { name: "登录 DevFlow" })).toBeVisible();
    await expect(this.page.getByLabel("用户名")).toBeVisible();
    await expect(this.page.getByLabel("密码")).toBeVisible();
  }

  errorMessage() {
    return this.page.getByRole("alert");
  }
}
