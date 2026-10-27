type Environment = Record<string, string | undefined>;

export function apiProxy(
  environment: Environment = {},
  fileEnvironment: Environment = {},
) {
  const macro = environment.MACRO_API_URL?.trim() ||
    fileEnvironment.MACRO_API_URL?.trim() || "http://127.0.0.1:3000";
  const stock = environment.STOCK_API_URL?.trim() ||
    fileEnvironment.STOCK_API_URL?.trim() || "http://127.0.0.1:8081";
  return {
    "/api/macro": { target: macro, changeOrigin: true },
    "/api": { target: stock, changeOrigin: true },
  };
}
