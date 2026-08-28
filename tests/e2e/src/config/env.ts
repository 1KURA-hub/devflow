function readURL(name: string, fallback: string): string {
  const value = process.env[name] || fallback;
  const parsed = new URL(value);
  return parsed.origin;
}

const externalEnvironment = process.env.DEVFLOW_E2E_EXTERNAL === "true";
const allowExternalMutation = process.env.DEVFLOW_E2E_ALLOW_MUTATION === "true";

if (externalEnvironment) {
  const missingURLs = ["DEVFLOW_API_URL", "DEVFLOW_WEB_URL"].filter(
    (name) => !process.env[name]
  );
  if (missingURLs.length > 0) {
    throw new Error(`外部环境模式缺少必填变量：${missingURLs.join(", ")}`);
  }
  if (!allowExternalMutation) {
    throw new Error(
      "外部环境测试会创建用户和业务数据；确认目标为隔离测试环境后，请显式设置 DEVFLOW_E2E_ALLOW_MUTATION=true"
    );
  }
}

export const env = {
  apiBaseURL: readURL("DEVFLOW_API_URL", "http://127.0.0.1:8080"),
  webBaseURL: readURL("DEVFLOW_WEB_URL", "http://127.0.0.1:5173"),
  externalEnvironment,
  allowExternalMutation,
  reuseExistingServer: process.env.DEVFLOW_E2E_REUSE_SERVER === "true",
  defaultPassword: "devflow123"
};
