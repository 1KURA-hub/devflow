function readURL(name: string, fallback: string): string {
  const value = process.env[name] || fallback;
  const parsed = new URL(value);
  return parsed.origin;
}

export const env = {
  apiBaseURL: readURL("DEVFLOW_API_URL", "http://127.0.0.1:8080"),
  webBaseURL: readURL("DEVFLOW_WEB_URL", "http://127.0.0.1:5173"),
  externalEnvironment: process.env.DEVFLOW_E2E_EXTERNAL === "true",
  defaultPassword: "devflow123"
};
