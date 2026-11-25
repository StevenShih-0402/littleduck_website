# Tasks：前台頁尾加上聯絡 Email（SDP-187）

本次只動前端，沒有 `[BE]` 任務；後端 Agent 不派工，前端 Agent 直接開工。
以下任務都在 `frontend/` 執行，依序進行（1.1 → 1.2 → 1.3 → 1.4），不可平行：1.2 依賴 1.1 的常數，1.3 記錄 1.1、1.2 的偏離。

## 1. 前端

- [x] 1.1 [FE] 在 `frontend/data/site.js` 的 `SITE` 新增 `contactEmail: "steven0210423@gmail.com"`（design.md D1）。驗證：在 `frontend/` 跑 `npm run lint`、`npm run typecheck` 通過。 — 新增 SITE.contactEmail
- [x] 1.2 [FE] 在 `frontend/components/front/front-shell.jsx` 頁尾加入聯絡區塊：「Contact」標示＋`<a href={"mailto:" + SITE.contactEmail}>`，Email 套 `normal-case` 以小寫顯示、沿用 `hover-rule` 與主題 class，不加 `FeatureGate`、不開新分頁；更新檔頭註解（design.md D2、D3）。驗證：在 `frontend/` 跑 `npm run lint`、`npm run typecheck`、`npm run build` 通過；`npm run dev` 或 `next start` 後開 `/` 與 `/blog`，分別在 MAG、TERM 主題確認頁尾顯示小寫 `steven0210423@gmail.com` 且連結為 `mailto:`，並在 handoff.md 記錄確認結果。 — 頁尾新增 Contact mailto 連結（lint／typecheck／build 通過；無瀏覽器，主題顯示以程式碼審視確認，見 handoff）
- [x] 1.3 [FE] 在 `frontend/design-sync/manifest.json` 記錄偏離（design.md D4）：`project/data.js` 新增 `deviations`（`data/site.js`），`project/v2/front-home.jsx` 的 `deviations` 追加一筆（`components/front/front-shell.jsx`），兩筆 `issue` 皆為 `SDP-187`；不改 `sha256`、`targets`、`syncedVersion`。驗證：在 `frontend/` 跑 `npm run test:design` 通過，且 `node -e "JSON.parse(require('fs').readFileSync('design-sync/manifest.json','utf8'))"` 無錯誤。 — manifest 記錄兩筆 SDP-187 偏離，test:design 通過
- [x] 1.4 [FE] 完整檢查：在 `frontend/` 依 CLAUDE.md「檢查指令」跑 `npm run test:design`、`npm run lint`、`npm run typecheck`、`npm run build`、`npm audit --audit-level=high`；build 後確認 `grep -rlF 'steven0210423@gmail.com' .next/static | wc -l` 大於 0。驗證：以上全部通過，結果寫進 handoff.md。 — 全部檢查通過
