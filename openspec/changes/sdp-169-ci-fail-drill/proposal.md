# Proposal：SDP-169 CI 失敗演練測試

> 驗收演練用，**本 change 的 PR 不合併**。驗收完成後由 user 關閉 PR、刪除分支，issue 改為 Canceled。

## Why

SDP-117 第 5.4 項（對應 SDP-113 歸檔 tasks.md 的 5.4）要驗收：PR 的 Drone CI 失敗時，`/ship-issue` 第 7.1 步會判定失敗並退回後端，且 `handoff.md` 留下含 Drone 建置連結的退回紀錄。要觸發這條路徑，需要一個「本機與 integrator 檢查都通過、只有在 Drone 裡才失敗」的變更。

## What Changes

- 在 `backend/` 新增一個 Go 測試 `TestCIFailureDrill`：
  - 環境變數 `DRONE` 不是 `"true"` 時呼叫 `t.Skip`，本機與 integrator「檢查」的 `go test ./...` 會略過它，流程得以走到開 PR；
  - `DRONE` 為 `"true"`（Drone 每個 step 都會自動注入）時呼叫 `t.Fatal`，訊息為「CI 失敗演練（SDP-117 5.4），不可合併」，讓 `ci-backend` pipeline 的 `go-test` step 失敗。
- 不改其他程式碼；不動 `.drone.yml`、`frontend/`；不新增或修改 `openspec/specs`。
- 無 API 變動，`openapi.yaml` 為 `paths: {}`。非破壞性變更。

## Capabilities

### New Capabilities

無。

### Modified Capabilities

無。本 change 不改變任何系統行為的規格（只新增一個演練用測試），依 user 決定不寫 delta spec，`.openspec.yaml` 設 `skip_specs: true`。

## Impact

- **預計修改的檔案範圍**（供判斷能否與其他 issue 並行）：
  - 新增 `backend/internal/pkg/apperror/ci_failure_drill_test.go`（只有這一個程式檔）
  - `openspec/changes/sdp-169-ci-fail-drill/**`
- 不碰執行期程式碼、相依版本、資料表、路由；不命中「高風險檔案樣式」。依風險分級屬「低」（只動測試）。
- 與其他 issue 的檔案衝突機率極低，可並行。
- 副作用：此分支在 Drone 上的 `ci-backend` 必定失敗，**不可合併**；若被誤合併，`main` 的 CI 也會持續失敗（見 design.md 風險段）。
