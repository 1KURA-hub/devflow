import {
  test as base,
  expect,
  type APIRequestContext
} from "@playwright/test";
import { DevFlowApi } from "../api/devflow-api";
import { env } from "../config/env";
import { TestData, type TestActor } from "../support/test-data";

interface SupportFixtures {
  api: DevFlowApi;
  data: TestData;
}

interface AuthenticatedFixtures {
  actor: TestActor;
}

interface HealthEnvelope {
  code: number;
  data?: {
    app?: string;
    env?: string;
  };
}

async function assertSafeTestTarget(request: APIRequestContext): Promise<void> {
  const healthURL = new URL("/healthz", env.apiBaseURL).toString();
  const response = await request.get(healthURL);
  let body: HealthEnvelope | undefined;
  try {
    body = (await response.json()) as HealthEnvelope;
  } catch {
    // 由下面的统一错误报告状态码和目标地址。
  }

  if (!response.ok() || body?.code !== 0 || body.data?.app !== "devflow") {
    throw new Error(`E2E目标健康检查失败：${healthURL} 返回 HTTP ${response.status()}`);
  }
  if (body.data.env !== "test") {
    throw new Error(
      `拒绝写入非测试环境：APP_ENV=${body.data.env ?? "unknown"}`
    );
  }
}

/** 不自动登录，供登录和未登录跳转测试使用。 */
export const unauthenticatedTest = base.extend<SupportFixtures>({
  api: [
    async ({ request }, use) => {
      await assertSafeTestTarget(request);
      await use(new DevFlowApi(request, env.apiBaseURL));
    },
    { auto: true }
  ],

  data: async ({ api }, use) => {
    const data = new TestData(api);
    await use(data);
    await data.cleanup();
  }
});

/** 每条已登录测试创建独立用户，并通过storageState注入token。 */
export const test = unauthenticatedTest.extend<AuthenticatedFixtures>({
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
