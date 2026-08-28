import path from "node:path";
import { defineConfig, devices } from "@playwright/test";
import { env } from "./src/config/env";

const repoRoot = path.resolve(__dirname, "../..");
const webRoot = path.join(repoRoot, "web");
const apiPort = new URL(env.apiBaseURL).port || "8080";
const webPort = new URL(env.webBaseURL).port || "5173";

const localWebServers = [
  {
    command: "go run ./cmd/server",
    cwd: repoRoot,
    url: `${env.apiBaseURL}/healthz`,
    timeout: 120_000,
    reuseExistingServer: env.reuseExistingServer,
    env: {
      ...process.env,
      APP_ENV: "test",
      HTTP_ADDR: `:${apiPort}`,
      MYSQL_DSN:
        process.env.MYSQL_DSN ||
        "devflow:devflow@tcp(127.0.0.1:3307)/devflow?charset=utf8mb4&parseTime=True&loc=Local",
      REDIS_ADDR: process.env.REDIS_ADDR || "127.0.0.1:6379",
      // 首版使用同步降级路径，避免把 MQ 可用性混入 UI 回归。
      RABBITMQ_URL: process.env.DEVFLOW_E2E_RABBITMQ_URL || "",
      AUTO_MIGRATE: "true",
      JWT_SECRET: process.env.JWT_SECRET || "devflow-e2e-secret"
    }
  },
  {
    command: `npm run dev -- --host 127.0.0.1 --port ${webPort} --strictPort`,
    cwd: webRoot,
    url: env.webBaseURL,
    timeout: 120_000,
    reuseExistingServer: env.reuseExistingServer,
    env: {
      ...process.env,
      DEVFLOW_API_PROXY_TARGET: env.apiBaseURL
    }
  }
];

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  forbidOnly: Boolean(process.env.CI),
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  timeout: 30_000,
  expect: {
    timeout: 8_000
  },
  reporter: [
    ["list"],
    ["html", { outputFolder: "playwright-report", open: "never" }],
    ["allure-playwright", { outputFolder: "allure-results" }]
  ],
  outputDir: "test-results/artifacts",
  use: {
    baseURL: env.webBaseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure"
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] }
    }
  ],
  webServer: env.externalEnvironment ? undefined : localWebServers
});
