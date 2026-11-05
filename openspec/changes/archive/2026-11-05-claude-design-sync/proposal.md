# Proposal: claude-design-sync

## Why

前端是從 Claude Design 一次性移植成 Next.js 的，之後改設計不會反映到 `frontend/`，也沒有方法知道兩邊差了什麼。SDP-89 要讓使用者改設計後，Claude Code 能據此更新專案。

SDP-83 原本建議用 `/design-sync`，但現況已改變，本提案改採 Artifact 拉取：

- 設計已於 2026-11-05 搬到 claude.ai，成為 Migrated Design artifact「Personal website」，不是 design-system 專案。
- standalone Claude Design 將於 2026-12-14 關閉。
- `/design-sync` 同步的是 HTML/CSS 預覽元件庫，不會改到 React 程式碼。

## What Changes

- 新增同步基準檔：在 `frontend/` 記錄設計來源、納入同步的設計檔、各檔雜湊，以及設計檔與前端檔案的對應。
- 新增差異檢查腳本：比對拉取下來的設計檔與基準，列出新增、變更、移除的檔案，並在同步完成後更新基準。
- 新增 `/design-pull` skill：讀取設計 → 比對基準 → 把差異移植到 Next.js → 驗證 build → 更新基準。
- 以目前設計建立首次基準。調查已確認現行前端與設計的程式碼與樣式對齊。
- 把 `public/assets/` 的兩張佔位圖換成設計的正式圖片，基準才會真的對齊。
- 同步為單向（設計 → 程式碼），由使用者手動觸發。

不在本提案範圍（另開 sub-issue）：後台真實登入、資料層改接 API、ECPay 串接、固定日期、社群連結與文章連結、文章內文。

## Capabilities

### New Capabilities

- `design-sync`: 偵測 claude.ai 設計與 `frontend/` 同步基準的差異，並把設計變更移植到 Next.js 程式碼。

### Modified Capabilities

（無）

## Impact

- `frontend/`：新增 `design-sync/manifest.json`、`scripts/design-sync.mjs` 與其測試，`package.json` 新增 scripts。替換 `public/assets/` 的兩張圖片。不改現有頁面與元件的程式碼。
- 根目錄：新增 `.claude/skills/design-pull/`，並更新 `CLAUDE.md`。
- 依賴：無新增套件，腳本只用 Node 內建模組。
- 後端與 API：無影響。
- 前提：同步需在已登入 claude.ai、可使用 Artifact 工具的 Claude Code session 執行。
