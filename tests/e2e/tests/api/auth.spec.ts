import { apiTest, expect } from "../../src/fixtures/test";
import { requireData } from "../../src/api/devflow-api";
import { newUserInput } from "../../src/support/data";

apiTest.describe("API｜认证", () => {
  apiTest("健康检查返回统一响应结构", async ({ api }) => {
    const result = await api.health();

    expect(result.response.status()).toBe(200);
    expect(result.body).toMatchObject({
      code: 0,
      message: "ok",
      data: { status: "ok", app: "devflow" }
    });
    expect(result.body.data?.env).toEqual(expect.any(String));
    expect(result.body.data?.env.length).toBeGreaterThan(0);
  });

  apiTest("注册得到的 token 可以读取当前用户", async ({ data }) => {
    const actor = await data.createActor();

    const result = await actor.api.me();

    expect(result.response.status()).toBe(200);
    expect(result.body.code).toBe(0);
    expect(result.body.data).toMatchObject({
      id: actor.user.id,
      username: actor.username,
      nickname: actor.nickname
    });
  });

  apiTest("重复注册返回 409，且不会覆盖原用户", async ({ api }) => {
    const input = newUserInput();
    const first = await api.register(input);
    const original = requireData(first, "首次注册");
    const second = await api.register({
      ...input,
      password: "changed123",
      nickname: `x_${input.username.slice(-8)}`
    });

    expect(first.response.status()).toBe(200);
    expect(second.response.status()).toBe(409);
    expect(second.body).toEqual({
      code: 409,
      message: "username already exists"
    });

    const login = await api.login({ username: input.username, password: input.password });
    expect(login.response.status()).toBe(200);
    expect(login.body.data?.user).toMatchObject({
      id: original.user.id,
      username: original.user.username,
      nickname: original.user.nickname
    });
  });

  apiTest("伪造 token 访问受保护接口返回 401", async ({ api }) => {
    const result = await api.as("not-a-valid-jwt").me();

    expect(result.response.status()).toBe(401);
    expect(result.body).toMatchObject({ code: 401, message: "invalid token" });
  });
});
