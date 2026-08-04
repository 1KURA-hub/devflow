import {
  test as base,
  expect,
  type APIRequestContext
} from "@playwright/test";
import { DevFlowApi } from "../api/devflow-api";
import { DevFlowDataFactory, type TestActor } from "../api/data-factory";
import { env } from "../config/env";

interface ApiFixtures {
  apiContext: APIRequestContext;
  api: DevFlowApi;
  data: DevFlowDataFactory;
}

interface AuthenticatedFixtures {
  actor: TestActor;
}

/**
 * apiTest：不自动登录，适合接口测试和登录/注册页面测试。
 */
export const apiTest = base.extend<ApiFixtures>({
  apiContext: async ({ playwright }, use) => {
    const context = await playwright.request.newContext({
      baseURL: env.apiBaseURL,
      extraHTTPHeaders: { Accept: "application/json" }
    });
    await use(context);
    await context.dispose();
  },

  api: async ({ apiContext }, use) => {
    await use(new DevFlowApi(apiContext));
  },

  data: async ({ api }, use) => {
    const factory = new DevFlowDataFactory(api);
    await use(factory);
    await factory.cleanup();
  }
});

/**
 * test：按需创建独立用户，并通过 storageState 注入 token。
 * 使用 page 的测试天然处于登录态，同时仍保留 Playwright 的 Trace/截图能力。
 */
export const test = apiTest.extend<AuthenticatedFixtures>({
  actor: async ({ data }, use) => {
    await use(await data.createActor());
  },

  storageState: async ({ actor }, use) => {
    await use({
      cookies: [],
      origins: [
        {
          origin: env.webBaseURL,
          localStorage: [{ name: "devflow_token", value: actor.token }]
        }
      ]
    });
  }
});

export { expect };
