# Tasks

> 本 change 只有開發工具與文件，沒有後端 API，全部任務標 `[FE]`，由 Yuma 實作，Chika 不用接。腳本與 code_city 的同名 change 共用同一份內容（code_city 先實作時，這裡直接複製過來，再跑一次本 repo 的測試）。

## 1. 測試骨架（TDD：先寫測試）

- [x] 1.1 [FE] 新增 `.claude/scripts/tests/remote-control.test.sh`：在暫存目錄建立假的 `claude`、`tmux`（記錄收到的參數，用旗標檔模擬 session 存在與否、`auth status` 的回傳），並提供 AAA 風格的 assert 函式。驗證方式：此時執行測試會因為腳本不存在而失敗（紅燈）
  - 結果：腳本與測試直接複製自 code_city 的同名 change（該處已依 TDD 先紅燈後實作），本 repo 不另行重跑紅燈階段。

## 2. 子命令與前置檢查

- [x] 2.1 [FE] 寫 `start` 的測試案例：首次啟動送出 `new-session -d -s rc-<repo> -c <根目錄>`，指令內含 `--name <repo> --spawn same-dir --permission-mode auto`；從子目錄執行時 `-c` 仍是根目錄；已在執行時不再呼叫 `new-session`，結束碼為 0。驗證方式：測試紅燈
- [x] 2.2 [FE] 寫前置檢查的測試案例：缺 `tmux`、缺 `claude`、`authMethod` 不是 `claude.ai`、設了 `ANTHROPIC_API_KEY` / `ANTHROPIC_AUTH_TOKEN` / `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC` / `DISABLE_GROWTHBOOK`、`ANTHROPIC_BASE_URL` 指向非官方位址。每一種都要驗證結束碼非 0、輸出有列出問題、沒有呼叫 `new-session`；同時有多個問題時要全部列出。驗證方式：測試紅燈
- [x] 2.3 [FE] 寫 `status` / `stop` / `attach` / 未知子命令的測試案例（spec 的每個 Scenario 各一個）。驗證方式：測試紅燈
- [x] 2.4 [FE] 實作 `.claude/scripts/remote-control.sh`（檔頭兩行註解、`set -euo pipefail`、可執行權限）：`resolve_root`、`preflight`、`start`、`status`、`stop`、`attach`、`usage`。不帶子命令時等同 `start`；tmux 直接以 claude 作為 session 指令，claude 結束時 session 一起消失（見 design D3）。驗證方式：`bash .claude/scripts/tests/remote-control.test.sh` 全部通過，`shellcheck` 檢查兩支腳本沒有警告（沒裝 shellcheck 就在報告註明略過）
  - 結果：`bash .claude/scripts/tests/remote-control.test.sh` 22 項全數通過；本機未安裝 shellcheck，改以 `bash -n` 檢查兩支腳本語法並通過，shellcheck 略過。
  - 偏離：腳本沒有在 tmux 指令後接 `; exit`，而是直接以 claude 作為 session 指令（與 code_city 修訂後的 2.4 一致）。tmux 預設在指令結束時關閉 session，效果與 design D3 相同。

## 3. 文件

- [x] 3.1 [FE] 新增 `docs/remote-control.md`：前置條件（claude.ai 訂閱登入、Team／Enterprise 需管理員開啟）、首次啟動步驟（`start` → `attach` 回答確認提示並掃 QR code → `Ctrl-b d` 離開）、日常操作（`status` / `stop`）、手機端操作（看進度、開新 session 輸入 `/ship-issue SDP-xxx`、回答權限提示、切換權限模式）、建議手動開啟的使用者層設定（`/config` 的推播選項）、限制（Windows 睡眠或離線約 10 分鐘後 server 會結束、`claude remote-control --continue` 可在約 4 小時內接回、對話會存在 Anthropic 伺服器）。驗證方式：文件裡的每個指令都在本機實際跑過一次
  - 結果：內容依 code_city 版本改為本 repo 名（`rc-littleduck_website`、`littleduck_website`）。`status`（含子目錄）／未執行時的 `stop`、`attach`／未知子命令都實際跑過；`start` 會連到 Anthropic，Agent 執行時被權限分類器擋下，`attach` 需要互動終端機，這兩項待 user 跑。
- [x] 3.2 [FE] README 新增「Remote Control」小節：一行啟動指令，並連到 `docs/remote-control.md`。驗證方式：連結路徑存在
  - 結果：README「快速開始 › Remote Control」小節連到 `docs/remote-control.md`，檔案存在。

## 4. 整合驗證

- [ ] 4.1 [FE] 實機驗證：執行 `start`，`tmux ls` 看得到 `rc-littleduck_website`；手機 Claude app 看得到名為 `littleduck_website` 的 session，並能開新 session；執行 `stop` 後 `status` 結束碼非 0。驗證方式：在 handoff.md 記錄各步驟的輸出（手機端由 user 確認，Agent 無法操作時在報告註明待 user 驗證）
  - **env-blocked**：Agent 執行 `start` 時被 Claude Code auto mode 分類器擋下（判定為建立對外連線通道），手機端也需要 user 操作，整項待 user 實機驗證，所以不打勾。
- [x] 4.2 [FE] 確認沒有影響網站程式碼。驗證方式：`git diff --stat origin/main` 只列出 `.claude/scripts/**`、`docs/remote-control.md`、`README.md` 與本 change 目錄；`code_city` 與本 repo 的 `remote-control.sh` 用 `diff` 比對沒有差異
  - 結果（2026-11-09）：`git status` 只有 `README.md`、`.claude/scripts/**`、`docs/remote-control.md` 與本 change 目錄（`frontend/.claude/`、`backend/.claude/`、`sdp-113-drone-cicd/` 在本 change 之前就是未追蹤檔案，與本 change 無關）；與 code_city 的 `remote-control.sh`、測試腳本 `diff` 皆無差異。

## Workflow follow-up

- 與 code_city 的 `sdp-112-remote-control` 一起歸檔，確認兩邊的腳本內容一致（`diff` 兩支腳本沒有差異）。
