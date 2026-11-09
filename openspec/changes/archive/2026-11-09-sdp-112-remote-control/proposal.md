# Proposal

## Why

使用者希望上班時用手機查看 Claude Code 的工作進度，也能把新任務派給本機的 session（SDP-112）。Claude Code 的 Remote Control 可以做到，但目前 repo 沒有固定的啟動方式。每次都要手動記住 server 參數，在 tmux 裡把程序掛著，還要自己處理登入、信任資料夾等前置條件。本 change 與 code_city 的同名 change（`sdp-112-remote-control`）一起處理，兩個 repo 用同一套做法。

## What Changes

- 新增啟動腳本 `.claude/scripts/remote-control.sh`，提供 `start` / `status` / `stop` / `attach` 四個子命令：
  - `start`：先做前置檢查，再於背景 tmux session `rc-<repo 名>` 啟動 `claude remote-control`（server 模式），參數固定為 `--name <repo 名> --spawn same-dir --permission-mode auto`。已經在跑就不重複啟動。
  - `status`：回報 server 是否在跑（以結束碼區分）。
  - `stop`：結束 server 與其 tmux session。
  - `attach`：接回 tmux 畫面，用來看 QR code、回答第一次啟動時的確認提示。
- 前置檢查擋下 Remote Control 已知無法運作的情況：沒有 `claude` / `tmux`、未以 claude.ai 帳號登入、設了 `ANTHROPIC_API_KEY` / `ANTHROPIC_AUTH_TOKEN`、`ANTHROPIC_BASE_URL` 指向非官方位址、設了 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` / `DISABLE_GROWTHBOOK`。
- 新增 `docs/remote-control.md`：啟動與停止、手機端操作（看進度、開新 session 派 `/ship-issue`、回答權限提示）、推播設定、WSL2 睡眠斷線等限制。README 加一段短說明並連到這份文件。
- 不改專案層 `.claude/settings.json`：`remoteControlAtStartup: true` 寫在專案層會被 Claude Code 忽略，全域自動開啟屬於使用者層設定，不在本 change 範圍。

## Capabilities

### New Capabilities
- `remote-control`：本機 Remote Control server 的啟動、狀態查詢、停止與前置檢查，讓手機可以監看並派任務給本 repo 的 session。

### Modified Capabilities
（無）

## Impact

- 新增檔案：`.claude/scripts/remote-control.sh`、`.claude/scripts/tests/remote-control.test.sh`、`docs/remote-control.md`；修改 `README.md`。
- 不動 `frontend/`、`backend/` 的程式碼，也不動 `.claude/hooks/**`、`.claude/agents/**`。
- 執行期相依：`tmux`（WSL2 已安裝 3.4）、`claude` CLI（須以 claude.ai 訂閱帳號登入）。
- 安全面：手機開的 session 預設用 `auto` 權限模式，由 auto mode 分類器判斷要不要問人。連線期間對話與工具輸出會存在 Anthropic 伺服器，檔案本身留在本機。
