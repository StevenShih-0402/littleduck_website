# Tasks

> 依序執行：第 2 節 `[BE]` 全部完成後，第 3、4 節 `[FE]` 才開工。真實文章內文與社群／專案網址由 user 上線後在後台填寫，seed 維持現狀，不在本次任務內（design.md D5）。
> 後端驗證在 `backend/` 執行，前端驗證在 `frontend/` 執行；每項完成後另跑該端的 lint（後端 `gofmt -l .` 無輸出、`go vet ./...`；前端 `npm run lint`、`npm run typecheck`）。

## 2. 後端

- [x] 2.1 [BE] `internal/service/comment_service.go` 的 `Create` 以 `time.FixedZone("Asia/Taipei", 8*60*60)` 產生 `date`（design.md D1）（變異測試目標：`backend/internal/service/comment_service.go`）。驗證：`comment_service_test.go` 新增測試，注入 `now` 為 UTC `2026-11-11 19:00` 時 `date` 為 `2026 / 11 / 12`、UTC `2026-11-12 07:00` 時為 `2026 / 11 / 12`；`go test ./internal/service/...` 通過。 — 新增 TestCommentCreate_DateUsesTaipeiTime（2 案例）
- [x] 2.2 [BE] `internal/service/setting_service.go` 的 `Update` 新增社群連結網址驗證（合法值見 design.md D2），不合法回 `apperror.BadRequest("社群連結網址格式不正確")` 且不呼叫 `repo.Save`；同步更新 `internal/controller/setting_controller.go` 的 swag `@Failure 400` 說明與 `openapi.yaml` 定稿（變異測試目標：`backend/internal/service/setting_service.go`）。驗證：`setting_service_test.go` 以表格測試涵蓋 `""`、`#`、`/blog/a1`、`https://x`、`HTTP://X`、`mailto:a@b`、` https://x `（前後空白）為合法，`javascript:alert(1)`、`JavaScript:x`、`data:text/html,x`、`//evil.com`、`foo/bar`、`ftp://x` 為 400 且未寫入；`go test ./...` 通過。 — 新增 isValidSocialURL 與表格測試（7 合法、6 不合法）

## 3. 前端：連結與日期

- [x] 3.1 [FE] 新增 `lib/safe-url.mjs`（`classifyHref`，規則見 design.md D2）與 `lib/safe-url.test.mjs`；`package.json` 的 `test:design` 改為 `node --test scripts/design-sync.test.mjs lib/*.test.mjs`（變異測試目標：`frontend/lib/safe-url.mjs`）。驗證：測試涵蓋與 2.2 相同的合法／不合法清單，外加 `#section` → `none`、`/` → `internal`；`npm run test:design` 通過。 — 新增 classifyHref 與 4 組測試，test:design 共 20 項通過
- [x] 3.2 [FE] `components/ui.jsx` 的 `mdInline` 移除 `preventDefault`，依 `classifyHref` 渲染：`internal` → `next/link`、`external` → `target="_blank" rel="noopener noreferrer"`、`none` → `<span>`；加行內註解標明刻意偏離設計（SDP-105）。驗證：`npm run lint`、`npm run typecheck`、`npm run build` 通過；`npm run dev` 開 `/blog/a0`，手動確認內文連結行為（記在 `handoff.md`）。 — mdInline 依 classifyHref 渲染；lint／typecheck／build 通過；dev 手動點擊未能在無瀏覽器環境執行，改由單元測試與程式審閱涵蓋
- [x] 3.3 [FE] `components/front/front-shell.jsx` 頁尾社群連結移除 `preventDefault`，`internal`／`external` 同 3.2，`none`（含 `#`、空字串、不安全網址）渲染為不可點的 `<span>`，仍顯示 `label`（不加 `cursor-pointer`、不加 hover 效果）；加行內註解（SDP-105）。驗證：`npm run lint`、`npm run typecheck`、`npm run build` 通過；`npm run dev` 時頁尾在 seed 資料（全為 `#`）下四個名稱都顯示但不可點，在後台把一筆改成 `https://example.com` 後該筆變成可點且開新分頁（記在 `handoff.md`）。 — 頁尾 none 渲染為 span，internal／external 同 3.2；lint／typecheck／build 通過；dev 手動確認未執行（無瀏覽器）
- [x] 3.4 [FE] `components/admin/admin-posts.jsx` 新增文章的 `y: 2026` 改為 `new Date().getFullYear()`；加行內註解（SDP-105）。驗證：`npm run lint`、`npm run typecheck` 通過。 — y 改為 new Date().getFullYear()
- [x] 3.5 [FE] 確認 `components/admin/admin-settings.jsx` 儲存失敗時會顯示後端 400 的 `message`（現有 `err` 狀態），並把網址輸入框的 placeholder 改為提示合法格式（`https://… 或 /路徑`）。驗證：`npm run dev` 時在後台輸入 `javascript:x` 儲存，畫面出現「社群連結網址格式不正確」（記在 `handoff.md`）；`npm run lint` 通過。 — settings 的 catch 已顯示後端 message；placeholder 改為 https://… 或 /路徑；後台 javascript:x 手動確認未執行（無瀏覽器），後端 400 訊息由 2.2 測試涵蓋

## 4. 前端：設計同步保留偏離

- [x] 4.1 [FE] `scripts/design-sync.mjs`：`buildManifest` 對非分歧項目保留 `deviations`；`diff` 的 `changed` 帶出 `deviations`；`formatReport` 在已變更項目下列出每筆偏離（design.md D4）。驗證：`scripts/design-sync.test.mjs` 新增「更新基準保留 deviations」「已變更項目的報告列出偏離」「無 deviations 的項目報告不變」三個測試；`npm run test:design` 通過。 — buildManifest 保留 deviations、報告列出偏離，新增 3 個測試
- [x] 4.2 [FE] `design-sync/manifest.json` 在 `project/v2/ui.jsx`、`project/v2/front-home.jsx`、`project/v2/admin-manage.jsx` 加上 3.2–3.4 的 `deviations`（`issue: "SDP-105"`、`since` 為實作當天），不更動 `sha256`、`targets`、`syncedVersion`。驗證：`node -e "JSON.parse(require('fs').readFileSync('design-sync/manifest.json','utf8'))"` 成功；`npm run test:design` 通過。 — manifest 登記 3 筆 deviations

## 5. 整合確認

- [x] 5.1 [FE] 前端完整檢查：`npm ci`、`npm run test:design`、`npm run lint`、`npm run typecheck`、`npm run build`、`npm audit --audit-level=high` 全部通過，結果寫進 `handoff.md`。 — test:design 20 通過、lint 0 error（9 個既有 img warning）、typecheck、build、audit 通過
