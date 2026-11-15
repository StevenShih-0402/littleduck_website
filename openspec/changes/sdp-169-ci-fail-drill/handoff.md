<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-169-ci-fail-drill

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
- **更新**：主對話 — 2026-11-15

---

## architect（Hikaru）— 2026-11-15 21:40

- **做了什麼**：
  - 建立 change `sdp-169-ci-fail-drill`：proposal、design（資料流圖、主對話／GitHub／Drone 互動時序圖、執行驗證觀察點 3 項、變異測試目標「無」）、tasks（1 項 `[BE]`、0 項 `[FE]`）、`openapi.yaml`（`paths: {}`）、`.linear.md`。
  - 測試檔落點定為 `backend/internal/pkg/apperror/ci_failure_drill_test.go`：`apperror` 是無 DB、無 `TestMain` 的純函式套件，CI 失敗時只會看到這個套件 FAIL，判讀不混淆（理由見 design.md D2）。
  - **規格處理**：依 user 決定不新增或修改 `openspec/specs`。OpenSpec 預設要求至少一份 delta spec，本 change 改在 `.openspec.yaml` 設 `skip_specs: true`（OpenSpec 官方提供、給無行為變更的 change 使用的退出機制），不捏造需求；`openspec validate --strict` 據此通過。
- **未完成**：無（提案階段）。
- **風險**：
  - **本 change 的 PR 不可合併**：合併後 `main` 的 `ci-backend` 會持續失敗。
  - 驗收預期 `/ship-issue` 第 7.1 步因 CI 失敗退回後端；**user 看到第一次退回後會中斷流程，後端 Agent 不得修掉或移除演練測試**。
  - 非破壞性變更；只動測試，風險分級「低」，不命中高風險檔案樣式；無 schema 變更。
- **API 契約變動**：無。
- **退回**：無。

---

## backend（Yasumi）— 2026-11-15

- **做了什麼**：新增 `backend/internal/pkg/apperror/ci_failure_drill_test.go`（`TestCIFailureDrill`：非 `DRONE=true` 時 Skip，否則 Fatal「CI 失敗演練（SDP-117 5.4），不可合併」）。
- **驗證**：`gofmt -l .` 無輸出；`go vet ./...` 通過；`go test ./...` 全過；`-v -run TestCIFailureDrill` 顯示 SKIP；`DRONE=true go test -run TestCIFailureDrill ./...` 失敗且含演練訊息，僅 apperror 套件 FAIL。
- **未完成**：無。
- **風險**：PR 不可合併；被退回時不得修掉或移除此測試。無 schema 變更、無 migration。
- **API 契約變動**：無（`openapi.yaml` 為 `paths: {}`，無需改動）。

---

## integrator（Teshima）— 檢查通過

- 後端（backend/）：`gofmt -l .` 無輸出；`go vet ./...` 通過；`go test -count=1 ./...` 全過（演練測試 SKIP）；`go test -v -run TestCIFailureDrill ./internal/pkg/apperror/` 顯示 `--- SKIP`；`DRONE=true go test -count=1 -run TestCIFailureDrill ./...` 僅 apperror 套件 FAIL，訊息含「CI 失敗演練（SDP-117 5.4），不可合併」。
- 前端：無 [FE] 任務、未改動，只在執行驗證時 `npm ci`、`npm run build` 通過。
- 規格：`openspec validate sdp-169-ci-fail-drill --strict` 通過（skip_specs）；`openspec validate --specs` 23 項全過。
- 變異測試：不適用（本次無關鍵模組）。
- 執行驗證（專案通用）：`/health` 回 `{"status":"ok","database":"up"}`；`/api/v1/articles` 200；`/blog/a1` 200。已停止服務並 `down -v`。
- 執行驗證（本次專屬）：design.md 三點皆通過（見上述 go test 結果）。無截圖觀察點。
- 核對：tasks 全勾；無 API／migration 變更；僅動 1 個後端測試檔（無越界）；未動 proposal.md；worktree 乾淨且與 origin 一致。
- 規格免同步聲明：無。專案無 openspec/spec-sync.yaml 規則適用。
- 風險：PR 不可合併，CI 預期失敗（本演練目的）。
