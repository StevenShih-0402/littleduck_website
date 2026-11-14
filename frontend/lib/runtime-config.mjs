// frontend/lib/runtime-config.mjs
// 執行期設定：API 位址由伺服器環境變數 API_BASE_URL 經 /runtime-config.js 注入，讓同一映像可在 Staging 與 Production 使用

const DEFAULT_API_BASE_URL = "http://localhost:8080";

// 去頭尾空白與結尾斜線；只接受 http／https 的絕對 URL，其他回 null
export function normalizeApiBaseUrl(value) {
  if (typeof value !== "string") return null;
  const trimmed = value.trim().replace(/\/+$/, "");
  if (!trimmed) return null;
  let url;
  try { url = new URL(trimmed); } catch { return null; }
  if (url.protocol !== "http:" && url.protocol !== "https:") return null;
  return trimmed;
}

// 產生 window.__LD_RUNTIME_CONFIG__ 腳本；JSON 內的 < > & U+2028 U+2029 轉成 \uXXXX
export function renderRuntimeConfigScript(apiBaseUrl) {
  const json = JSON.stringify({ apiBaseUrl: apiBaseUrl ?? null })
    .replace(/</g, "\\u003c")
    .replace(/>/g, "\\u003e")
    .replace(/&/g, "\\u0026")
    .replace(/\u2028/g, "\\u2028")
    .replace(/\u2029/g, "\\u2029");
  return "window.__LD_RUNTIME_CONFIG__=" + json + ";";
}

// 優先序：執行期值 → 建置期值 → 本機預設
export function resolveApiBaseUrl(runtimeValue, buildValue) {
  return normalizeApiBaseUrl(runtimeValue) || normalizeApiBaseUrl(buildValue) || DEFAULT_API_BASE_URL;
}
