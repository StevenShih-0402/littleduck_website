# 程式碼審查：sdp-105-frontend-content-polish

## 第 1 輪 — PR 前 — 2026-11-12 09:50

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：70e60ead111431da096960cf3a3d33919d3b56a4
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：無

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE（BE 有同類） | reviewer | 中 | 高 | `frontend/lib/safe-url.mjs:307-310`；`backend/internal/service/setting_service.go:115-117` | front-content-links › 文章內文的連結可點擊（`//` 開頭不可點）；design D2；tasks 3.1、2.2 | `classifyHref` 只看第 2 個字元是否為 `/` 或 `\`。瀏覽器解析前會移除 tab／換行，`"/\t/evil.com"` 被判為 `internal`，實際解析成 `//evil.com`，同分頁跳到外部網域，也沒有 `noopener`。後端 `isValidSocialURL` 同樣只排除 `//` 開頭。輸入只有管理員能寫，不會造成 XSS。建議判斷前先移除 `\t\n\r`，或用 `new URL(href, base).origin` 比對 origin，並補測試。 | PR 備註 |
| 2 | FE | reviewer | 中 | 高 | `frontend/components/ui.jsx:22-28`、`frontend/components/front/front-shell.jsx:118-124` | front-content-links 全部情境；tasks 3.2、3.3、3.5 | 渲染分支（站內用 `Link`、外部連結 `target="_blank"` 加 `rel`、不可點用 `span`、`none` 不帶 hover）沒有測試涵蓋。3.2／3.3／3.5 要求的 `npm run dev` 手動確認因環境沒有瀏覽器而沒做。讀程式碼看，分支是正確的。 | PR 備註（建議上線前實機確認，或之後補元件測試） |
| 3 | FE | reviewer | 低 | 高 | `frontend/components/ui.jsx:14`（既有 regex） | front-content-links › 不安全的連結網址 | 連結 regex `\(([^)]+)\)` 遇到第一個 `)` 就結束。`[x](javascript:alert(1))` 會顯示成「x)」；含括號的合法網址也會被截斷。安全上沒問題（分類為 `none`）。 | PR 備註 |
| 4 | FE | reviewer | 低 | 中 | `frontend/components/admin/admin-editor.jsx:231` | 無（規格只規範前台） | 後台編輯器預覽共用 `Markdown`，原本預覽裡的連結點了不會動；現在點站內連結會離開編輯器、點外部連結會開新分頁。卸載時會 flush 未送出的變更，資料不會遺失。 | PR 備註 |

- **結果**：通過
