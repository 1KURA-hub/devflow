import "dotenv/config";

const coreVariables = [
  "MIDSCENE_MODEL_BASE_URL",
  "MIDSCENE_MODEL_NAME",
  "MIDSCENE_MODEL_FAMILY"
] as const;

function apiKeyCanBeOmitted(baseURL: string): boolean {
  return (
    baseURL.startsWith("codex://") ||
    /^http:\/\/(localhost|127\.0\.0\.1)(?::\d+)?(?:\/|$)/.test(baseURL)
  );
}

export function assertMidsceneModelConfigured(): void {
  const missing: string[] = coreVariables.filter(
    (name) => !process.env[name]?.trim()
  );
  const baseURL = process.env.MIDSCENE_MODEL_BASE_URL?.trim() || "";

  if (!apiKeyCanBeOmitted(baseURL) && !process.env.MIDSCENE_MODEL_API_KEY?.trim()) {
    missing.push("MIDSCENE_MODEL_API_KEY");
  }

  if (missing.length > 0) {
    throw new Error(
      `Midscene 模型配置不完整，缺少：${missing.join(", ")}。` +
        "请先复制 tests/e2e/.env.example 为 tests/e2e/.env 并填入兼容的多模态模型配置。"
    );
  }
}
