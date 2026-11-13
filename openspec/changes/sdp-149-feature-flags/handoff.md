<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-149-feature-flags

## 狀態

<!-- 就地改值，不追加。中斷後重新派工時，開發 Agent 以這裡判斷從哪裡接續；不寫 commit SHA（以 git log --grep "wip(<ISSUE-ID>): <任務編號>" 對應）。
  階段由誰寫入：proposal（architect 建立；主對話提交提案時更新「最後完成」）→ backend（後端 Agent 開工）
  → frontend（後端 Agent 交棒，或沒有 [BE] 任務時前端 Agent 開工）→ integrating（最後一個開發 Agent 交接）→ reviewing（integrator 檢查通過；主對話做 PR 前審查）
  → pr-checking（integrator 收尾重切並開 PR 後；主對話做 PR 中檢查）
  → awaiting-approval（只有高風險：主對話等待 user 回覆「合併」）→ archived（integrator 歸檔；之後由 integrator 合併）。
  integrator 檢查、PR 前審查、CI 或 PR 中檢查退回時改回被退回那一端的 backend／frontend（退回 architect 為 proposal）。
  合併隊列失敗不改階段，停下交給 user。 -->

- **階段**：reviewing <!-- proposal | backend | frontend | integrating | reviewing | pr-checking | awaiting-approval | archived -->
- **最後完成**：PR 前審查
- **下一步**：Teshima 收尾
- **暫停原因**：無
- **更新**：主對話 — 2026-11-13

---

## architect（Hikaru）— 2026-11-13

- **做了什麼**：調查後端認證（單一管理員 JWT、`/admin` 後台）、網站設定模型與前台可操作區塊；建立 proposal、design（資料流圖、前後端時序圖、執行驗證觀察點 8 項、變異測試目標 3 項）、specs（新增 `feature-flags`、修改 `comments-api`）、openapi.yaml 草稿、tasks（[BE] 7 項、[FE] 4 項）。
- **未完成**：repo 根目錄 `CLAUDE.md` 的「功能開關」開發原則不在任何 Agent 寫入範圍，草稿段落放在 proposal.md「CLAUDE.md 開發原則」，交主對話處理。
- **風險**：新增 `backend/internal/model/feature_flag.go` 命中高風險樣式（資料表結構），PR 需停下等 user 回覆「合併」。schema 為相容新增（新表）。提案中的開關清單、預設值、後端把關為建議值，待 user 確認。
- **API 契約變動**：新增 `GET /api/v1/features`、`GET /api/v1/admin/features`、`PUT /api/v1/features/{key}`；`POST /api/v1/comments` 新增 403 `FEATURE_DISABLED`。非破壞性。
- **退回**：無

## architect（Hikaru）— 2026-11-13（修訂：user 回覆待確認問題）

- **做了什麼**：依 user 回覆修訂。Q1–Q3 照建議定案（三個開關、只藏表單／按鈕、既有預設啟用、新功能預設停用、`POST /comments` 停用時回 403）。Q4 改為本次一起做管理員預覽：前台以「sessionStorage 有權杖且 `GET /admin/features` 成功」判斷管理員，停用中的區塊加「停用中・預覽」標示並可試用；後端新增 `OptionalAuth`，只掛在 `POST /comments`，帶有效 JWT 時略過開關，無效或過期權杖視同訪客、不回 401（取捨見 design.md D4-1、D6、D7）。Q5 更新 proposal.md 的 CLAUDE.md 草稿，加入管理員預覽，仍交主對話寫入。tasks 改為 [BE] 8 項、[FE] 5 項；觀察點增為 10 項；變異測試目標增為 4 項。
- **未完成**：無
- **風險**：除 `backend/internal/model/**` 外，又命中 `backend/internal/middleware/**`（`OptionalAuth`），仍為高風險，PR 需跑 `/security-review`。管理員預覽送出的留言會以 pending 寫入真實資料。
- **API 契約變動**：`POST /api/v1/comments` 新增選用的 `Authorization`（有效管理員 JWT 時略過開關；無效權杖不回 401）。其餘同上一節。
- **退回**：無（依 user 回覆修訂，不是退回）

## backend（Yasumi）— 2026-11-13

- **做了什麼**：[BE] 1.1–1.8 全部完成。新增 `model.FeatureFlag`、`FeatureRepository`、`apperror.FeatureDisabled`（403 `FEATURE_DISABLED`）、`feature_service`（登記表三開關、PublicMap／List／Update／IsEnabled／EnsureDefaults）、`feature_controller` 與三條路由；`comment_service.Create(in, asAdmin)` 依 articleId 檢查 guestbook／article_comments（停用 403、讀取失敗 500、管理員略過）；`middleware.OptionalAuth`（與 `Auth` 共用 `bearerToken` 解析，只掛 `POST /comments`）；`seed.Run` 每次啟動呼叫 `EnsureDefaults`；重新產生 Swagger、更新 README。
- **測試**：`go test ./...` 全過、`gofmt -l .` 無輸出、`go vet ./...` 通過。
- **Schema 變更**：新增 `feature_flags` 表（key 主鍵 varchar 64、enabled bool not null、created_at、updated_at），為相容的新增，不動既有表。**尚未在真實 PostgreSQL 驗證 AutoMigrate**（單元測試用 mock），空資料庫與既有資料庫啟動請 Teshima 在執行驗證確認。
- **API 契約變動（openapi.yaml 定稿 vs 草稿）**：欄位與端點與草稿一致；版本改 0.1.0；`GET /admin/features`、`PUT /features/{key}`、`POST /comments` 補上實作會回的 500；`FeatureFlag.description` 範例文字改為登記表實際說明。`updated_at` 在資料庫尚無該列時為零值時間（理論上 seed 後不會發生）。`GET /features` 的 data 只含登記表內的 key。
- **給前端**：`PUT /features/{key}` 的 body 必須是布林 `enabled`；`POST /comments` 前端已自動附權杖，不需另外處理；後台清單回傳順序即登記表順序。
- **風險**：命中高風險樣式 `backend/internal/model/**`、`backend/internal/middleware/**`；`OptionalAuth` 無效權杖視同訪客不回 401。
- **規格同步**：專案無 `openspec/spec-sync.yaml`，不適用。
- **退回**：無

## frontend（Fujimori）— 2026-11-13

- **做了什麼**：[FE] 2.1–2.5 全部完成。`lib/feature-flags.mjs`（FEATURES、isFeatureEnabled 嚴格 true、gateMode）與 6 組測試；`lake-context` 另外請求 `/features`（失敗設 `{}`、不設 error）、`reloadAdmin` 讀 `/admin/features` 放 `featureList`（logout 與 401 清回 null）、提供 `adminPreview`、`setFeatureEnabled`；新增 `components/feature-gate.jsx`（含 `PreviewBadge`），`SponsorFab` 支援 `preview`；`front-shell`（SponsorFab）、`guest-page`（投書表單）、`article-page`（留言表單）套用；新增後台 `admin-features.jsx` 與「功能開關」分頁。`manifest.json` 為 front-shell、guest-page、sponsor-fab、article-page、admin-shell 登記 deviations，另為 `lake-context.jsx`（app.jsx 對應檔，tasks 未列但同樣是設計稿對應檔）補一筆。
- **檢查**：`npm run test:design` 26 項全過；`npm run lint` 0 error（9 個 `<img>` 警告為既有 baseline）；`npm run typecheck` 通過；`npm run build` 成功，`.next/static/chunks` 內「功能開關」「停用中・預覽」各有 1 個檔案。
- **未完成**：無。未實機對接後端（未啟動服務），串接以 openapi.yaml 為準，請 Teshima 在執行驗證時確認管理員預覽流程。
- **仍靠 mock**：無。
- **實作備註**：`SponsorFab` 的預覽小標放在浮動按鈕上方（fixed）與視窗標題列；`reloadAdmin` 的 `/admin/features` 為獨立請求，失敗（含 401）只讓 `featureList` 維持 null，不寫 error。
- **契約問題**：無，無需退回。
- **規格同步**：專案無 `openspec/spec-sync.yaml`，不適用。

## integrator（Teshima）— 檢查通過

- **後端**（backend/）：`go test -count=1 ./internal/...` 全綠；`gofmt -l .` 無輸出；`go vet ./...` 無輸出。
- **前端**（frontend/）：`npm ci`；`npm run test:design` 26/26 通過；`npm run lint` 0 error、9 warning（皆為 no-img-element）；`npm run typecheck` 通過；`npm run build` 成功；`npm audit --audit-level=high` exit 0（僅 8 個 moderate）。
- **規格**：`openspec validate sdp-149-feature-flags --strict` 通過；`openspec validate --specs` 22/22 通過。無 `spec-sync.yaml`，規格同步不適用。
- **核對**：tasks.md 無未勾選項；worktree 乾淨且與 origin 一致；BE／FE commit 無越界；proposal.md 未被開發 Agent 修改；openapi.yaml 三條路由與 FEATURE_DISABLED 與實作一致。
- **migration（GORM AutoMigrate）**：空資料庫啟動成功（6 個 model、seed 與功能開關補齊）；既有資料庫重啟，保留 sponsor=false 未被覆寫；DROP TABLE feature_flags 後重啟（模擬舊 schema 升級）自動重建並補齊預設。相容的新增表，非不相容 schema 變更。
- **變異測試**：略過（CLAUDE.md 未設定變異測試）。
- **執行驗證（專案通用）**：health `{"status":"ok","database":"up"}` 通過；`/api/v1/articles` 200 通過；`/blog/a1` 200 通過。
- **執行驗證（本次專屬 10 點）**：1 預設三開關皆 true 通過；2 未授權 PUT 401 通過；3 管理員停用 guestbook 通過；4 訪客張貼 403 FEATURE_DISABLED 通過；5 管理員權杖張貼 201 通過；6 無效權杖 403 通過；7 文章留言 201 通過；8 未知 key 404 通過；9 還原 200、admin 清單 3 筆通過；10 build chunks 含「功能開關」與「停用中・預覽」各 1 個檔案通過。無截圖觀察點。服務已停止、埠已釋放。
- **風險**：命中高風險樣式（model、middleware、database 相關），PR 需等 user 回覆「合併」；根目錄 CLAUDE.md 由主對話另 commit。畫面層行為（隱藏區塊、預覽標示）僅由單元測試與審查涵蓋，未實機截圖。
- **規格免同步聲明**：無。
- **退回**：無
