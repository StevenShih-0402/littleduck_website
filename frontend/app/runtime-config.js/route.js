// frontend/app/runtime-config.js/route.js
// 執行期設定腳本：把伺服器環境變數 API_BASE_URL 注入 window.__LD_RUNTIME_CONFIG__，不在建置時預先算好

import { normalizeApiBaseUrl, renderRuntimeConfigScript } from "@/lib/runtime-config.mjs";

export const dynamic = "force-dynamic";

export function GET() {
  const body = renderRuntimeConfigScript(normalizeApiBaseUrl(process.env.API_BASE_URL));
  return new Response(body, {
    headers: {
      "Content-Type": "text/javascript; charset=utf-8",
      "Cache-Control": "no-store",
    },
  });
}
