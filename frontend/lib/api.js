// frontend/lib/api.js
// 後端 API client：fetch 封裝、JWT token 保存與附帶、統一 {code,message,data} envelope 解析、401 自動登出

// base URL 來自環境變數，預設本機後端 8080 埠
const BASE_URL = (process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080").replace(/\/$/, "");
const API_PREFIX = "/api/v1";

const TOKEN_KEY = "lake_token";

// ── token 保存（sessionStorage，隨分頁關閉失效；主題偏好另存 localStorage） ──
export function getToken() {
  if (typeof window === "undefined") return null;
  try { return sessionStorage.getItem(TOKEN_KEY); } catch { return null; }
}
export function setToken(token) {
  if (typeof window === "undefined") return;
  try { sessionStorage.setItem(TOKEN_KEY, token); } catch {}
}
export function clearToken() {
  if (typeof window === "undefined") return;
  try { sessionStorage.removeItem(TOKEN_KEY); } catch {}
}

// 401 時由 context 註冊的 handler 強制登出
let onUnauthorized = null;
export function setUnauthorizedHandler(fn) { onUnauthorized = fn; }

// 核心 fetch：附帶 token、JSON headers、解析 envelope、錯誤丟 Error、401 清 token 並通知
async function request(method, path, body) {
  const headers = { "Accept": "application/json" };
  const token = getToken();
  if (token) headers["Authorization"] = "Bearer " + token;
  const init = { method, headers };
  if (body !== undefined && body !== null) {
    headers["Content-Type"] = "application/json";
    init.body = JSON.stringify(body);
  }

  const res = await fetch(BASE_URL + API_PREFIX + path, init);

  if (res.status === 401) {
    clearToken();
    if (onUnauthorized) onUnauthorized();
  }

  // 204 無內容（如瀏覽回報）
  if (res.status === 204) return null;

  let envelope = null;
  try { envelope = await res.json(); } catch { envelope = null; }

  if (!res.ok) {
    const message = (envelope && envelope.message) || ("請求失敗（" + res.status + "）");
    throw new Error(message);
  }
  // 讀取 envelope 的 data 欄位
  return envelope ? envelope.data : null;
}

export function apiGet(path) { return request("GET", path); }
export function apiPost(path, body) { return request("POST", path, body); }
export function apiPut(path, body) { return request("PUT", path, body); }
export function apiDelete(path) { return request("DELETE", path); }
