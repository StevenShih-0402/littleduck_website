# 程式碼審查：sdp-187-footer-contact

## 第 1 輪 — PR 前 — 2026-11-25 23:35

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：0101cba0d98fa9f6ed53946c0b4a5dba14ade0d4
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：無

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | reviewer | 中 | 高 | frontend/components/front/front-shell.jsx:127-131 | 頁尾顯示聯絡 Email、兩種主題都顯示聯絡 Email › 全部 Scenario | 規格情境沒有自動化測試（前端尚無元件測試框架）；日後移除 `normal-case`、改掉 `href` 或把 Email 移進條件分支，測試仍會通過，只能靠打包產物 grep 觀察點確認 | PR 備註 |

- **結果**：通過

## 第 2 輪 — PR 中 — 2026-11-25 23:39

- **範圍**：origin/main...HEAD（PR 中第 1 輪）
- **已審到**：09c763fa49749a1b65ae8ecf30e88284f11e8c2f（收尾 HEAD）
- **審查指令**：
  - PR 中：CI 無 CI（PR #23 兩次查詢 statusCheckRollup 皆為空，間隔 30 秒）；`/code-review medium /home/steven/side_projects/littleduck_website/wt/sdp-187-footer-contact origin/main...HEAD`；/security-review 不適用；reviewer 評定模式（已派）
- **高風險檔案**：無
- **風險等級**（只有 PR 中）：中 — 動到 `frontend/` 執行期程式碼（非破壞性），未命中高風險檔案樣式

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | /code-review | 中 | 高 | frontend/components/front/front-shell.jsx:44-61 | 頁尾顯示聯絡 Email › 一般瀏覽者在前台頁面看到 Email | settings 載入中或載入失敗時整頁提早 return，沒有頁尾；SSR 初始 HTML 只有「載入中…」，後端掛掉或不跑 JS 的檢查工具看不到 Email。design.md Non-Goals／Risks 已列為既有行為、不在本次處理，但與審核目的有落差 | PR 備註 |
| 2 | FE | /code-review | 低 | 高 | frontend/components/front/front-shell.jsx:130 | 頁尾顯示聯絡 Email › 點擊 Email | `.hover-rule` 底線只在 hover 出現，觸控裝置上 Email 看不出是連結；與既有社群連結一致，規格只要求可點擊 | PR 備註 |
| 3 | FE | /code-review | — | — | frontend/components/front/front-shell.jsx:128 | — | 規格情境無自動化測試 | 不成立：與第 1 輪 #1 重複 |
| 4 | FE | /code-review | — | — | frontend/data/site.js:12 | 頁尾顯示聯絡 Email | 明文公開個人 Email 可能被爬取 | 不成立：公開明文是需求本身，design.md Risks 已決定不混淆 |

- **結果**：通過
