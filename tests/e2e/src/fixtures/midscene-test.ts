import type { PlayWrightAiFixtureType } from "@midscene/web/playwright";
import { PlaywrightAiFixture } from "@midscene/web/playwright";
import { test as playwrightTest, expect } from "./test";

/**
 * Midscene 只扩展 AI 专用例，原有 13 条 Playwright 回归不会创建 Agent 或调用模型。
 */
export const test = playwrightTest.extend<PlayWrightAiFixtureType>(
  PlaywrightAiFixture({
    waitForNetworkIdleTimeout: 2_000,
    replanningCycleLimit: 12,
    waitAfterAction: 500
  })
);

export { expect };
