# Tasks

> 驗收演練用，PR 不合併。只有一項 `[BE]` 任務，沒有 `[FE]` 任務。

## 1. 後端

- [x] 1.1 [BE] 新增 `backend/internal/pkg/apperror/ci_failure_drill_test.go`（`package apperror`）：`TestCIFailureDrill` 在 `os.Getenv("DRONE") != "true"` 時 `t.Skip`，否則 `t.Fatal("CI 失敗演練（SDP-117 5.4），不可合併")`；檔頭照專案慣例寫路徑與用途註解，不改其他檔案。（完成：新增演練測試，本機 SKIP、DRONE=true 時失敗）驗證（在 `backend/` 執行）：`gofmt -l .` 無輸出、`go vet ./...` 通過、`go test ./...` 通過且 `go test -v -run TestCIFailureDrill ./internal/pkg/apperror/` 顯示 `--- SKIP`、`DRONE=true go test -run TestCIFailureDrill ./...` 失敗且訊息含「CI 失敗演練（SDP-117 5.4），不可合併」

## Workflow follow-up

- 本 change 不合併、不歸檔：PR 的 Drone CI 失敗、`/ship-issue` 第 7.1 步退回並在 `handoff.md` 留下含 Drone 建置連結的紀錄後，由 user 中斷流程、關閉 PR、刪除分支，issue 改為 Canceled。
- 被退回時後端 Agent **不得**修掉或移除這個測試。
