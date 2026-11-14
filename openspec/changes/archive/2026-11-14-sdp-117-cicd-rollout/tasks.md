# Tasks

> 全部任務改的是 `docs/`、`deploy/`，依主對話指示標為 `[BE]`；沒有 `[FE]` 任務。不動 `.drone.yml`、`backend/`、`frontend/`。
> 這些路徑不在 `backend/` 之內：主對話派工前會暫時放寬主 repo `CLAUDE.md` 的「後端子目錄」（需要的路徑見 handoff.md）。
> 主機名對照見 proposal.md「Why」的表格；Secret 值見 design.md D1。
> 第 1、2、3 組改不同檔案、互不依賴；第 4 組在最後。
> 寫文件時**不得出現 `vercel` 字樣**（不分大小寫），`frontend-delivery` 規格以 `git grep -il vercel` 檢查 `docs/`、`deploy/`；需要提到舊平台時沿用 runbook 的「舊託管服務」代稱。

## 1. 部署手冊（docs/deployment-runbook.md）

- [x] 1.1 [BE] 開頭第 10–16 行：刪除「`<網域>` 尚未決定（SDP-117 範圍）」與「前端的主機名已定案」兩段引言，改成一段說明加五列主機名對照表（用途／主機名／Drone secret 或設定位置）：前端 Production `frontend_host`、前端 Staging `frontend_staging_host`、Drone `deploy/drone/.env` 的 `DRONE_SERVER_HOST`、後端 Production `api_host`、後端 Staging `api_staging_host`；並註明 `api-staging`／`web-staging` 是環境名稱（目錄、容器、樣板檔名），不是主機名。驗證：在 worktree 根目錄 `sed -n '1,30p' docs/deployment-runbook.md` 人工確認五列齊全，`git grep -n '尚未決定' -- docs/deployment-runbook.md` 無輸出 — 完成：主機名對照表五列
- [x] 1.2 [BE] 第 0 節 DNS 列改為 design.md D4：`drone.stevenshih-littleduck.com`、`api.stevenshih-littleduck.com`、`api.staging.stevenshih-littleduck.com` 三筆 A 記錄指向主機 IP，Cloudflare 代理狀態 DNS only（灰雲），前端兩筆見第 3 節。驗證：`git grep -n 'DNS only' -- docs/deployment-runbook.md` 至少兩行（第 0 節與第 3 節） — 完成：DNS 三筆 A 記錄與 DNS only
- [x] 1.3 [BE] 第 1、2、5、6 節的後端與 Drone 佔位換成實際主機名：OAuth App Homepage／Callback URL、`certbot … -d drone.stevenshih-littleduck.com` 與其後說明、`DRONE_HOST=drone.stevenshih-littleduck.com deploy/edge/render.sh …`、啟用 repo 與驗證的登入網址、第 2 節環境表的「網域」欄、第 2 節第 4 步憑證改成兩行明確指令（`api.stevenshih-littleduck.com`、`api.staging.stevenshih-littleduck.com`，取代有歧義的 `-d <網域>`）、第 5 節 `https://api.staging.stevenshih-littleduck.com/health`、第 6 節後端回滾的 `curl -fsS https://api.stevenshih-littleduck.com/health`。`<主機>`（主機 IP）與 `<前一版>-<sha>` 佔位保留。驗證：`git grep -nE '<網域>|api-staging\.' -- docs/deployment-runbook.md` 只剩環境名稱（`littleduck-api-staging`、`/opt/littleduck-api-staging`），沒有 `api-staging.` 開頭的主機名與 `<網域>` — 完成：後端與 Drone 佔位換成實際主機名
- [x] 1.4 [BE] 第 3 節第 5 步前端 `.env`：Staging 改 `API_BASE_URL=https://api.staging.stevenshih-littleduck.com`、Production 改 `API_BASE_URL=https://api.stevenshih-littleduck.com`。驗證：`git grep -n 'API_BASE_URL=https://api' -- docs/deployment-runbook.md` 恰好兩行且為上述值；`git grep -c 'api\.staging\.stevenshih-littleduck\.com' -- docs/deployment-runbook.md` 至少 5；`git grep -il vercel -- docs/deployment-runbook.md` 無輸出 — 完成：前端 API_BASE_URL 改實際值

## 2. 流水線文件（docs/cicd.md）

- [x] 2.1 [BE] 後端 CD 表的冒煙目標：`https://<staging>` 改 `https://api.staging.stevenshih-littleduck.com`、`https://<production>` 改 `https://api.stevenshih-littleduck.com`（說明主機名來自 secret `api_staging_host`／`api_host`）。Secret 清單加「值」欄，依 design.md D1 填入；`hetzner_host` 寫「主機 IP，不寫進 repo」。驗證：`git grep -nE '<staging>|<production>' -- docs/cicd.md` 無輸出；`git grep -c 'stevenshih-littleduck\.com' -- docs/cicd.md` 至少 6 — 完成：冒煙目標與 Secret 值欄
- [x] 2.2 [BE] 文末新增「## 各階段耗時」段落：說明由 user 實機量測後填寫（對應 SDP-113 tasks 6.1），表格依 design.md D5 的六列，欄位「階段／起點→終點／耗時／量測日期」，耗時與日期留「待量測」。驗證：`grep -n '^## 各階段耗時' docs/cicd.md` 恰好一行；`git grep -il vercel -- docs/cicd.md` 無輸出 — 完成：新增各階段耗時空表

## 3. 部署註解與範本（deploy/）

- [x] 3.1 [BE] `deploy/drone/.env.example`：`DRONE_SERVER_HOST=` 預填 `drone.stevenshih-littleduck.com`，上方「例：drone.example.com」註解改成說明此為正式值；OAuth App 註解的 `https://<DRONE_SERVER_HOST>` 改成 `https://drone.stevenshih-littleduck.com`、`…/login`（design.md D3）。驗證：`grep -c '^DRONE_SERVER_HOST=drone\.stevenshih-littleduck\.com$' deploy/drone/.env.example` 為 `1` — 完成：.env.example 預填 DRONE_SERVER_HOST
- [x] 3.2 [BE] 只改用法註解：`deploy/ci/smoke-backend.sh` 範例改 `https://api.staging.stevenshih-littleduck.com`；`deploy/ci/smoke-frontend.sh` 範例改 `https://staging.stevenshih-littleduck.com https://api.staging.stevenshih-littleduck.com`。不改任何指令行。驗證：`git grep -n 'example\.com' -- deploy/` 無輸出；`sh -n deploy/ci/smoke-backend.sh && sh -n deploy/ci/smoke-frontend.sh && echo ok` 印出 `ok`；`git diff origin/main -- deploy/ci/` 只有 `#` 開頭的行變動 — 完成：smoke 腳本用法註解

## 4. 整合確認

- [x] 4.1 [BE] 在 worktree 根目錄跑 design.md「## 執行驗證觀察點」全部指令並符合預期；執行 `openspec validate sdp-117-cicd-rollout --strict` 通過；`git diff --name-only origin/main` 只列出 proposal.md「Impact」表中的檔案與本變更目錄。結果寫進 handoff.md — 完成：觀察點全數符合、openspec validate 通過、diff 範圍正確

## Workflow follow-up

- 歸檔時同步 `openspec/specs/ci-server`、`backend-delivery`、`frontend-delivery`。
- 合併後由 user 照 runbook 上線，並依 handoff.md「上線後由 user 手動驗收」清單逐項驗收、填 `docs/cicd.md` 各階段耗時與 runbook 記憶體量測；user 已決定合併後手動把 SDP-117 改回 In Progress，全部驗收完成才關閉。
