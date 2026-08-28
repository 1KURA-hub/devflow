import "dotenv/config";
import { defineConfig, devices } from "@playwright/test";
import baseConfig from "./playwright.config";

export default defineConfig({
  ...baseConfig,
  testDir: "./ai",
  fullyParallel: false,
  retries: 0,
  workers: 1,
  timeout: 120_000,
  expect: {
    timeout: 10_000
  },
  reporter: [
    ["list"],
    [
      "@midscene/web/playwright-reporter",
      { type: "merged", outputFormat: "single-html" }
    ]
  ],
  outputDir: "test-results/ai-artifacts",
  projects: [
    {
      name: "chromium-ai",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1440, height: 900 }
      }
    }
  ]
});
