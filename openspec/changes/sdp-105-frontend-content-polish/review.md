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

## 第 1 輪 — PR 中 — 2026-11-12

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：09de519c0ea5d90ff7ed045a7dadd3eab553b681（收尾 HEAD）
- **CI**：無 CI（PR #10 間隔 30 秒查詢兩次 `statusCheckRollup` 皆為空清單，記錄略過）
- **審查指令**：
  - PR 中：`/code-review medium /home/steven/side_projects/littleduck_website/wt/sdp-105-frontend-content-polish origin/main...HEAD`
  - `/security-review`：不適用（未命中高風險檔案樣式）
  - 評定：reviewer 評定模式（已派）
- **高風險檔案**：無
- **風險等級**：中（動到 `backend/`、`frontend/` 執行期程式碼，非破壞性；未命中高風險檔案樣式；無成立的安全發現）

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | /code-review | — | — | `frontend/lib/safe-url.mjs:14` | front-content-links | `/\t/evil.com` 被判為站內。 | 不成立（重複 PR 前第 1 輪 #1） |
| 2 | BE | /code-review | — | — | `backend/internal/service/setting_service.go:64` | site-settings-api | 後端 `isValidSocialURL` 接受 `/\` 與夾控制字元的網址。 | 不成立（重複 PR 前第 1 輪 #1；目前 social 只有前端讀取，已由 `classifyHref` 擋下） |
| 3 | BE | /code-review | 低 | 中 | `backend/internal/service/setting_service.go:50-54` | site-settings-api › 社群連結網址格式驗證；design D2 | 每次存設定都驗證全部社群連結；資料庫若已有不合規舊網址，存任何設定都回 400，且訊息沒指出是哪一筆。seed 皆為 `#`（已放行），且同頁可當場修正，影響有限。建議錯誤訊息帶出第幾筆或 label。 | PR 備註 |
| 4 | FE | /code-review | — | — | `frontend/components/ui.jsx:22` | 無 | 後台編輯器預覽的連結變成可點。 | 不成立（重複 PR 前第 1 輪 #4） |
| 5 | BE | /code-review | — | — | `backend/internal/service/setting_service.go:62` | site-settings-api | 後端只接受剛好 `#`，前端接受所有 `#…`。 | 不成立（符合 delta spec 與 openapi.yaml；後半重複 PR 前第 1 輪 #1） |
| 6 | FE | /code-review | 低 | 高 | `frontend/components/admin/admin-settings.jsx:127`；`frontend/design-sync/manifest.json:57-63` | design-sync › 記錄移植時要保留的偏離；tasks 4.2 | `admin-settings.jsx` placeholder 改為「https://… 或 /路徑」偏離設計稿，但 manifest 沒登記 `deviations`；下次 `/design-pull` 可能被蓋回。只影響提示文字。 | PR 備註 |
| 7 | FE | /code-review | — | — | `frontend/components/front/front-shell.jsx:119`、`frontend/components/ui.jsx:22-28` | 無 | 建議抽出共用 `SafeLink`。 | 不成立（重用建議，無錯誤行為、未違反慣例） |
| 8 | FE | /code-review | 低 | 高 | `frontend/scripts/design-sync.mjs:94,99-102` | design-sync › 已變更的設計檔帶有偏離紀錄 | 報告中「保留偏離」與「→ targets」縮在同一層；偏離的 target 不在該項 `targets` 時沒有警告。 | PR 備註 |

- **結果**：通過
