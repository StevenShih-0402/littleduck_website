# remote-control Specification

## Purpose
讓使用者在本機以一個指令啟動、查詢、停止 Claude Code Remote Control server，使手機或 claude.ai/code 可以監看本 repo 的 session 並開新 session 指派任務。

## Requirements

### Requirement: 啟動 Remote Control server
啟動指令 SHALL 在背景、與終端機生命週期無關的 session 中，於 repo 根目錄以 server 模式啟動 Remote Control。session 名稱 SHALL 使用 repo 目錄名。手機開的新 session SHALL 在 repo 根目錄執行，預設權限模式為 `auto`。

#### Scenario: 首次啟動
- **WHEN** 前置檢查全數通過，且本 repo 的 server 尚未執行時，使用者執行啟動指令
- **THEN** server 在背景開始執行，指令以結束碼 0 結束，並印出接回畫面看 QR code 的方法

#### Scenario: 關掉終端機後仍持續執行
- **WHEN** 啟動成功後，使用者關閉執行啟動指令的終端機
- **THEN** server 仍在執行，手機仍看得到這個 repo 的 session

#### Scenario: 從子目錄執行
- **WHEN** 使用者在 repo 的子目錄執行啟動指令
- **THEN** server 的工作目錄仍是 repo 根目錄

### Requirement: 重複啟動不建立第二個 server
本 repo 的 server 已在執行時，啟動指令 MUST NOT 再啟動第二個 server。

#### Scenario: 已在執行時再次啟動
- **WHEN** 本 repo 的 server 已在執行，使用者再執行一次啟動指令
- **THEN** 指令告知 server 已在執行，以結束碼 0 結束，執行中的 server 數量維持一個

### Requirement: 啟動前的前置檢查
啟動指令 MUST 在啟動 server 之前檢查 Remote Control 的必要條件。任一項不符時 MUST NOT 啟動 server，MUST 以非 0 結束碼結束，並印出不符的項目與修正方法。檢查項目：`claude` 與 `tmux` 指令存在；已用 claude.ai 帳號登入；未設定 API key 類環境變數；`ANTHROPIC_BASE_URL` 未設定或指向官方 API；未設定會關閉 Remote Control 的環境變數。

#### Scenario: 缺少必要指令
- **WHEN** 系統找不到 `tmux` 或 `claude`，使用者執行啟動指令
- **THEN** 指令印出缺少的指令名稱，以非 0 結束碼結束，沒有 server 被啟動

#### Scenario: 未以 claude.ai 帳號登入
- **WHEN** `claude` 未登入，或登入方式不是 claude.ai 帳號，使用者執行啟動指令
- **THEN** 指令提示改用 claude.ai 帳號登入，以非 0 結束碼結束，沒有 server 被啟動

#### Scenario: 設定了不相容的環境變數
- **WHEN** 環境中設有 `ANTHROPIC_API_KEY`、`ANTHROPIC_AUTH_TOKEN`、`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` 或 `DISABLE_GROWTHBOOK` 任一項，或 `ANTHROPIC_BASE_URL` 指向官方 API 以外的位址，使用者執行啟動指令
- **THEN** 指令列出造成問題的變數名稱，以非 0 結束碼結束，沒有 server 被啟動

### Requirement: 查詢 server 狀態
狀態指令 SHALL 以結束碼回報本 repo 的 server 是否在執行：執行中為 0，未執行為非 0，並印出對應的說明文字。

#### Scenario: server 執行中
- **WHEN** 本 repo 的 server 在執行，使用者執行狀態指令
- **THEN** 指令印出執行中，結束碼為 0

#### Scenario: server 未執行
- **WHEN** 本 repo 的 server 沒有執行，使用者執行狀態指令
- **THEN** 指令印出未執行，結束碼非 0

### Requirement: 停止 server
停止指令 SHALL 結束本 repo 的 server 與它所在的背景 session。server 原本就沒在執行時 SHALL 視為成功。

#### Scenario: 停止執行中的 server
- **WHEN** 本 repo 的 server 在執行，使用者執行停止指令
- **THEN** server 結束，之後執行狀態指令回報未執行

#### Scenario: 停止未執行的 server
- **WHEN** 本 repo 的 server 沒有執行，使用者執行停止指令
- **THEN** 指令告知原本就未執行，以結束碼 0 結束

### Requirement: 接回 server 畫面
接回指令 SHALL 把目前終端機接到本 repo server 的背景 session，讓使用者看到 QR code 並回答首次啟動的確認提示。server 未執行時 MUST 以非 0 結束碼結束並提示先啟動。

#### Scenario: 接回執行中的 server
- **WHEN** 本 repo 的 server 在執行，使用者執行接回指令
- **THEN** 終端機顯示 server 的畫面

#### Scenario: server 未執行時接回
- **WHEN** 本 repo 的 server 沒有執行，使用者執行接回指令
- **THEN** 指令提示先執行啟動指令，結束碼非 0

### Requirement: 不支援的子命令
腳本收到不認得的子命令時 MUST 印出用法說明並以非 0 結束碼結束，不得啟動或停止任何 server。

#### Scenario: 打錯子命令
- **WHEN** 使用者執行 `remote-control.sh strat`
- **THEN** 指令印出用法說明，結束碼非 0，server 狀態不變
