<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-105-frontend-content-polish

## 狀態

<!-- 就地改值，不追加。中斷後重新派工時，開發 Agent 以這裡判斷從哪裡接續；不寫 commit SHA（以 git log --grep "wip(<ISSUE-ID>): <任務編號>" 對應）。
  階段由誰寫入：proposal（architect 建立；主對話提交提案時更新「最後完成」）→ backend（後端 Agent 開工）
  → frontend（後端 Agent 交棒，或沒有 [BE] 任務時前端 Agent 開工）→ integrating（最後一個開發 Agent 交接）→ reviewing（integrator 檢查通過；主對話做 PR 前審查）
  → pr-checking（integrator 收尾重切並開 PR 後；主對話做 PR 中檢查）
  → awaiting-approval（只有高風險：主對話等待 user 回覆「合併」）→ archived（integrator 歸檔；之後由 integrator 合併）。
  integrator 檢查、PR 前審查、CI 或 PR 中檢查退回時改回被退回那一端的 backend／frontend（退回 architect 為 proposal）。
  合併隊列失敗不改階段，停下交給 user。 -->

- **階段**：archived <!-- proposal | backend | frontend | integrating | reviewing | pr-checking | awaiting-approval | archived -->
- **最後完成**：integrator 歸檔
- **下一步**：合併
- **暫停原因**：無
- **更新**：主對話 — 2026-11-12

---

## architect（Hikaru）— 2026-11-12 08:48

- **做了什麼**：建立 proposal、design（含資料流圖與前後端時序圖、執行驗證觀察點 6 項、變異測試目標 3 項）、4 份 delta spec（新增 `front-content-links`；修改 `comments-api`、`site-settings-api`、`design-sync`）、`openapi.yaml`、`tasks.md`（`[BE]` 5 項、`[FE]` 8 項）。
  - 調查結論：`Lake.today()` 寫死已在 SDP-103 修正；仍存在的日期問題是後端留言日期用 UTC、後台新增文章年份寫死 2026。
  - 佔位內文實際在 `backend/internal/seed/seed_data.go`，seed 只在空表寫入，線上資料庫需靠 D5 的 `UpgradePlaceholders` 或後台手動更新。
- **未完成**：tasks 2.3–2.5 依待確認問題 Q1／Q2 的答覆決定保留或刪除；真實內容由 user 提供。
- **風險**：中（動到前後端執行期程式碼，未命中高風險樣式，非破壞性）。上線後線上的 `#` 社群連結會從頁尾消失，直到填入真實網址。
- **API 契約變動**：`POST /api/v1/comments` 的 `date` 改以 UTC+8 計算（schema 不變）；`PUT /api/v1/settings` 新增 400（`social[].url` 格式不合法）。非破壞性。無 schema 變更。
- **退回**：無

---

## architect（Hikaru）— 2026-11-12 09:05

- **做了什麼**：依 user 對待確認問題的答覆修訂提案（不是退回造成的修訂）。
  - Q1 文章內文：上線後由 user 在後台填寫。刪除 tasks 2.3–2.5 與執行驗證觀察點 6；design.md D5 改為「由 user 在後台填寫、seed 維持現狀」，移除 `UpgradePlaceholders`（資料流圖、Risks、Migration Plan 一併更新）。
  - Q2 社群與專案網址：上線後由 user 在後台填寫，seed 不改；proposal 的檔案範圍移除 `backend/internal/seed/**`。
  - Q3 頁尾：網址為 `#`、空或不安全的社群連結仍顯示名稱，但為不可點的文字。已修改 spec `front-content-links`（新增「網址不安全」情境）與 task 3.3。
  - 修訂後 tasks：`[BE]` 2 項（2.1、2.2）、`[FE]` 8 項；執行驗證觀察點 5 項。
- **未完成**：無；沒有待確認問題。
- **風險**：中（同上）。上線後填網址前，頁尾社群連結都是不可點的文字，user 已同意。
- **API 契約變動**：無新增（同上一節）。
- **退回**：無

---

## backend（Yasumi）— 2026-11-12

- **做了什麼**：
  - 2.1 `comment_service.Create` 以 `taipeiZone`（`time.FixedZone`，UTC+8）格式化 `date`；新增跨日測試（UTC 11/11 19:00 與 11/12 07:00 皆為 `2026 / 11 / 12`）。
  - 2.2 `setting_service.Update` 新增 `isValidSocialURL`，不合法回 400「社群連結網址格式不正確」且不呼叫 `Save`；表格測試涵蓋 7 個合法、6 個不合法值；controller swag `@Failure 400` 補說明。
- **測試**：`go test ./...` 全部通過；`gofmt -l .` 無輸出；`go vet ./...` 通過。
- **openapi.yaml**：與實作一致，無欄位差異，只把 description 的「草稿」字樣改為已定稿。`backend/docs` 的 swag 產物未重新產生（本專案無產生指令），其 400 說明可能仍是舊文字。
- **Schema / migration**：無 model 變更，無 AutoMigrate 影響，非不相容變更。
- **未完成**：無。
- **風險**：後端合法值不含 `#section`（僅精確 `#`），前端 `classifyHref` 對 `#…` 為 `none`，兩端行為一致。
- **API 契約變動**：無 schema 變動；`POST /comments` 的 `date` 改 UTC+8；`PUT /settings` 新增 400 情境。
- **退回**：無

---

## frontend（Fujimori）— 2026-11-12

- **做了什麼**：
  - 3.1 `lib/safe-url.mjs` 的 `classifyHref`（另擋 `/\host`）與測試，`test:design` 接入 `lib/*.test.mjs`。
  - 3.2／3.3 `ui.jsx` 內文連結與 `front-shell.jsx` 頁尾社群連結改依分類渲染（Link／新分頁 noopener／不可點 span），移除 `preventDefault`。
  - 3.4 後台新增文章 `y` 改取當年。3.5 設定頁 `err` 本就顯示後端 message，僅改 placeholder。
  - 4.1／4.2 `design-sync.mjs` 保留並於報告列出 `deviations`，新增 3 個測試；manifest 登記 3 筆（未動 sha256、targets、syncedVersion）。
- **檢查結果**（`frontend/`）：`npm run test:design` 20 通過；`npm run lint` 0 error（9 個既有 `no-img-element` warning）；`npm run typecheck` 通過；`npm run build` 通過；`npm audit --audit-level=high` 通過。
- **未做／限制**：3.2、3.3、3.5 要求的 `npm run dev` 瀏覽器手動確認未執行（此環境無瀏覽器）；行為由 `classifyHref` 單元測試與程式審閱涵蓋，建議 Teshima／user 在實機確認頁尾四個名稱不可點、改 `https://example.com` 後開新分頁、後台輸入 `javascript:x` 出現 400 訊息。
- **仍靠 mock**：無。
- **風險**：`test:design` 現在也跑 `lib/*.test.mjs`，名稱不再只代表設計同步（design.md 已記）。
- **規格同步**：專案無 `openspec/spec-sync.yaml`，不適用。
- **退回**：無

---

## integrator（Teshima）— 檢查通過 2026-11-12

- **後端**（`backend/`）：`go test ./...` 通過（service 以 `-count=1` 重跑通過）；`gofmt -l .` 無輸出；`go vet ./...` 通過。
- **前端**（`frontend/`）：`npm ci`；`npm run test:design` 20 通過；`npm run lint` 0 error（9 個既有 `no-img-element` warning）；`npm run typecheck` 通過；`npm run build` 通過；`npm audit --audit-level=high` exit 0（僅 8 個 moderate）。
- **規格**：`openspec validate sdp-105-frontend-content-polish --strict` 有效；`openspec validate --specs` 21 通過。
- **核對**：tasks 無未勾選；worktree 乾淨且與 origin 一致；後端 commit 只動 `backend/`、前端 commit 只動 `frontend/`；開發 Agent 未改 proposal.md；無 `spec-sync.yaml`，規格同步不適用；無 model／schema 變更；無破壞性 API 變更。
- **變異測試**：略過（CLAUDE.md 未設定變異測試）。三個目標模組（safe-url.mjs、setting_service.go、comment_service.go）皆未跑。
- **執行驗證（專案通用）**：`/health` -> `{"status":"ok","database":"up"}` 通過；`/api/v1/articles` 200 通過；`localhost:3100/blog/a1` 200 通過。
- **執行驗證（本次專屬）**：1) 留言 date `2026 / 11 / 12`，與台北日期相同，通過；2) 不安全網址 PUT 回 400「社群連結網址格式不正確」，通過；3) 被拒後 `social[0].url` 仍為 `#`，通過；4) `https://example.com` PUT 200 且讀回該值，通過；5) `.next/static/chunks` 含 `noopener noreferrer` 的檔案 7 個（>0），通過。環境無 `jq`，以 python3 解析 JSON 代替，指令語意相同。服務已停止，無殘留。
- **截圖觀察點**：無（設計未列）。
- **風險**：3.2／3.3／3.5 的瀏覽器點擊行為（內文連結、頁尾不可點文字、後台錯誤訊息顯示）未實機確認，僅由 `classifyHref` 單元測試、build 產物含 `noopener noreferrer`、後端 400 訊息佐證；`backend/docs` swag 產物未重產（400 說明可能是舊文字）；`test:design` 現在也跑 `lib/*.test.mjs`。
- **規格免同步聲明**：無。
- **退回**：無

---

## integrator（Teshima）— 收尾 2026-11-12

- **接續**：上一次收尾中斷（本地已重切、尚未推送、PR 未建立）。確認 `git diff ac2aaa6 HEAD` 為空，工作樹乾淨；審查後除 `handoff.md`、`review.md` 外無其他變動，不需重新檢查。
- **重切前的 SHA（OLD，遠端 WIP HEAD）**：`ac2aaa6318d63a163ba481ff273f49e71c78f64a`
- **重切起點（FROM）**：`176a122`（merge-base origin/main）
- **重切後的 commit**：`05833d6` fix(api) 留言日期 UTC+8；`367201e` feat(api) 社群連結驗證；`23b0bd6` feat(web) 連結可點擊；`5a88fa9` fix(web) 後台年份；`b240635` feat(web) design-sync 偏離；`09de519` docs(openspec) 提案與審查紀錄
- **tree 比對**：`git diff ac2aaa6 HEAD` 為空
- **推送**：`git push --force-with-lease=<branch>:ac2aaa6...` 成功
- **PR**：https://github.com/StevenShih-0402/littleduck_website/pull/10
- **附進 PR 的 review.md 備註**：4 條（中 2、低 2）
- **風險**：中；3.2／3.3／3.5 瀏覽器點擊行為未實機確認；`backend/docs` swag 產物未重產
- **收尾 HEAD**：`09de519c0ea5d90ff7ed045a7dadd3eab553b681`

---

## integrator（Teshima）— 歸檔 2026-11-12

- 以 `openspec archive -y` 同步規格（comments-api、design-sync、site-settings-api 各新增 requirement；新建 front-content-links）並歸檔至 `openspec/changes/archive/2026-11-12-sdp-105-frontend-content-polish/`。
- `openspec validate --specs`：22 passed, 0 failed。
- 風險等級：中，無需 user 回覆合併。
