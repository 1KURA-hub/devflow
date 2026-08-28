import { expect, type Locator, type Page } from "@playwright/test";

interface LoginCredentials {
  username: string;
  password: string;
}

export class AuthPage {
  constructor(private readonly page: Page) {}

  async gotoLogin(): Promise<void> {
    await this.page.goto("/login");
  }

  async login(credentials: LoginCredentials): Promise<void> {
    await this.page.getByLabel("用户名").fill(credentials.username);
    await this.page.getByLabel("密码").fill(credentials.password);
    await this.page.getByRole("button", { name: "登录", exact: true }).click();
  }

  async expectLoginPage(): Promise<void> {
    await expect(this.page).toHaveURL(/\/login$/);
    await expect(this.page.getByRole("heading", { name: "登录 DevFlow" })).toBeVisible();
    await expect(this.page.getByLabel("用户名")).toBeVisible();
    await expect(this.page.getByLabel("密码")).toBeVisible();
  }

  errorMessage(): Locator {
    return this.page.getByRole("alert");
  }
}
