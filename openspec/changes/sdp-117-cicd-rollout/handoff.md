<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-117-cicd-rollout

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
- **更新**：主對話 — 2026-11-14

---

## architect（Hikaru）— 2026-11-13 00:00

- **做了什麼**：
  - 建立變更目錄：proposal.md、design.md（資料流圖、Staging 前端時序圖、執行驗證觀察點 5 項、變異測試目標「無」）、specs delta（`ci-server`、`backend-delivery` MODIFIED；`frontend-delivery` MODIFIED 2 項、ADDED 2 項）、openapi.yaml（`paths: {}`）、tasks.md（[BE] 9 項、[FE] 0 項）。
  - 系統分析結論：Nginx 樣板與部署腳本以變數帶主機名，Staging 改為 `api.staging.` 不需改邏輯；`api-staging` 在 repo 內是環境名稱，維持不變。需要實際修改的只有文件、註解範例，以及前端 Staging 固定主機名（`vercel-deploy.sh`＋`.drone.yml`）。
  - 發現的缺口：runbook 沒寫要在 Vercel 設 `NEXT_PUBLIC_API_BASE_URL`，照現行文件上線前端會呼叫 `http://localhost:8080`；已排入 1.3。
  - 發現的外部風險：`staging.stevenshih-littleduck.com` 是 mygo 改名前的舊 Staging 主機名，主機上可能仍有 A 記錄與 certbot renewal；已排入 1.2 的切換步驟。
- **未完成**：全部實作任務。
- **風險**：
  - `.drone.yml`、`deploy/**` 命中高風險檔案樣式，整個變更為**高風險**，PR 中要等 user 回覆「合併」。
  - 所有任務的檔案都在後端 Agent 寫入 hook 放行範圍（`backend/` 與變更目錄）之外，見待確認 Q2。
- **API 契約變動**：無
- **退回**：無

### 上線後由 user 手動驗收

本次 PR 不含以下項目。合併後 user 照 `docs/deployment-runbook.md` 架設，再逐項驗收（編號對應 `openspec/changes/archive/2026-11-09-sdp-113-drone-cicd/tasks.md`）：

| SDP-113 項次 | 驗收內容 |
|------|------|
| 1.3 | 照 runbook 第 1 節在主機啟動 Drone，`docker compose ps` 兩個容器皆 running |
| 1.4 | `https://drone.stevenshih-littleduck.com` 顯示登入頁；非允許帳號登入被拒；直接連容器埠失敗 |
| 2.4 | Drone 啟用 repo、建立 secret（全部不允許 PR）；開只改文件的 PR，`gh pr checks` 看到 Drone check 且成功，GHCR 與 Vercel 沒有新部署 |
| 3.3 | 合併後 Staging 跑新映像，`https://api.staging.stevenshih-littleduck.com/health` 回 200、冒煙通過 |
| 3.4 | promote 到 production 後 `https://api.stevenshih-littleduck.com/health` 回 200；沒有 promote 時 Production 不變 |
| 3.5 | 照 runbook 第 6 節在 Staging 演練一次後端回滾 |
| 4.1 | 推送非 main 分支，Vercel 沒有新部署 |
| 4.2 | 合併後 `cd-frontend-preview` 印出 Preview 網址且冒煙通過（有 `frontend_staging_host` 時 `https://staging.stevenshih-littleduck.com` 為新版本） |
| 4.3 | promote 後 `https://stevenshih-littleduck.com` 更新；沒有 promote 時不變 |
| 5.4 | 對刻意讓 Go 測試失敗的 PR 跑 `/ship-issue`，第 7.1 步退回，`handoff.md` 有含 Drone 連結的紀錄 |
| 6.1 | 從開 PR 到 Production 完整走一次，把各階段耗時填進 `docs/cicd.md`「各階段耗時」 |

另外（本次新增的需求）：Staging 前端的 API 請求打 `api.staging.`、正式前端打 `api.`（frontend-delivery「前端依環境呼叫對應的後端」），以瀏覽器開發者工具確認。

### 待確認問題

- **Q1 前端 Staging 的綁定方式與是否公開**：CD 以 Vercel CLI 部署且 Git 自動部署已關閉，Vercel「網域綁 Preview 分支」不會套用。提案採用每次 Preview 部署後 `vercel alias set` 到 `staging.stevenshih-littleduck.com`（tasks 第 3 組，新 secret `frontend_staging_host`）。另外，Standard Protection 會讓 `staging.` 需要 Vercel 登入才看得到。請 user 決定：(a) 照提案 alias，接受需登入；(b) 照提案 alias，並把 Deployment Protection 調整為公開（Staging 前端任何人可看）；(c) 不在 CD 綁固定主機名，只在文件保留 `staging.`（刪除第 3 組與 frontend-delivery 的 ADDED 需求「Preview 綁定固定的 Staging 主機名」）。
- **Q2 後端 Agent 寫入範圍**：所有任務改的是 `docs/`、`deploy/`、`.drone.yml`，後端 Agent 的寫入 hook 只放行 `backend/` 與變更目錄，派工後會被擋。可選：(a) 暫時在主 repo `CLAUDE.md`「## agent-team 設定」的「後端子目錄」加上 `deploy/、docs/、.drone.yml`（hook 支援多路徑，完成後改回）；(b) 由主對話直接依 tasks.md 實作；(c) 其他做法。
- **Q3 PR 合併後的 Linear 狀態**：PR 合併後 Linear automation 可能把 SDP-117 轉為 Done，但 11 項實機驗收尚未做。是否要讓本 PR 不關閉 SDP-117（例如 PR 只引用不用關閉字詞，或另開 sub-issue 承接 repo 端變更）？

---

## architect（Hikaru）— 2026-11-14 17:42（重寫提案）

- **做了什麼**：
  - 依最新 `main`（SDP-157 前端自架、SDP-164 Node 24 已合併）重新分析，原地改寫 proposal.md、design.md（資料流圖、Staging 前後端互動時序圖、執行驗證觀察點 8 項、變異測試目標「無」）、specs delta（`ci-server` 不變；`backend-delivery` MODIFIED 2 項；`frontend-delivery` 改為 MODIFIED 3 項，刪除只對 Vercel 成立的 2 項 ADDED）、openapi.yaml（`paths: {}`）、tasks.md（[BE] 9 項、[FE] 0 項）。
  - 系統分析結論：前端主機名、前端主機建置、清理 mygo 舊 `staging.` 都已由 SDP-157 寫進 runbook；`.drone.yml` 所有主機名都走 secret，**本次不需改 `.drone.yml`**。剩下的工作是把後端與 Drone 的 `<網域>`、`api-staging.<網域>`、`<staging>`／`<production>`、`example.com` 佔位換成實際主機名，補 Secret 值欄與「各階段耗時」空表。
  - 發現的缺口：runbook 第 3 節前端 `.env` 的 `API_BASE_URL` 引用後端佔位，user 照抄會讓 `cd-frontend-staging` 的 runtime-config 冒煙失敗；已排入 1.4。
  - 刪除的 Vercel 內容：`vercel alias set`、`frontend_staging_host` 新 secret（SDP-157 已建立）、Protection Bypass、Vercel 環境變數 `NEXT_PUBLIC_API_BASE_URL`、`vercel-deploy.sh`／`.drone.yml` 任務。issue 原文 Vercel 項目的自架等價對照寫在 proposal.md。
- **未完成**：全部實作任務。
- **風險**：
  - `deploy/drone/.env.example`、`deploy/ci/smoke-*.sh` 命中高風險檔案樣式 `deploy/**`，整個變更為**高風險**，PR 中要等 user 回覆「合併」（實際只改註解與範本值）。
  - 文件中不得出現 `vercel` 字樣（`frontend-delivery` 規格），tasks 已註明。
- **後端 Agent 需要的寫入路徑**（user 已決定暫時放寬主 repo `CLAUDE.md` 的「後端子目錄」，完成後改回）：`docs/`、`deploy/`。本次**不需要** `.drone.yml`。實際會改的檔案：`docs/deployment-runbook.md`、`docs/cicd.md`、`deploy/drone/.env.example`、`deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh`。
- **API 契約變動**：無
- **退回**：退回 architect 第 1 次：前端改自架（SDP-157）
- **上一節待確認問題**：Q1～Q3 已結案（Q1 隨 Vercel 作廢；Q2 暫時放寬後端子目錄；Q3 合併後由 user 手動改回 In Progress），本節不再列。

### 上線後由 user 手動驗收（自架版本，取代上一節的清單）

本次 PR 不含以下項目。合併後 user 照 `docs/deployment-runbook.md` 架設，再逐項驗收（SDP-113 項次對應 `openspec/changes/archive/2026-11-09-sdp-113-drone-cicd/tasks.md`；原本針對 Vercel 的 2.4、4.1、4.2、4.3 已改寫為自架的等價項目）：

| SDP-113 項次 | 驗收內容 |
|------|------|
| 1.3 | 照 runbook 第 1 節在主機啟動 Drone，`docker compose ps` 中 drone-server、drone-runner 皆 running |
| 1.4 | `https://drone.stevenshih-littleduck.com` 顯示登入頁且憑證有效；非允許帳號登入被拒；直接連 Drone 容器埠失敗 |
| 2.4 | Drone 啟用 repo（Trusted、ignore forks）、建立 secret（全部不允許 PR）；開只改文件的 PR，`gh pr checks` 看到 `continuous-integration/drone/pr` 且成功；GHCR 沒有新的 `littleduck-api`／`littleduck-web` 映像，主機容器不變 |
| 3.3 | 合併後 `cd-backend-staging` 通過，`https://api.staging.stevenshih-littleduck.com/health` 回 200 |
| 3.4 | promote 到 production 後 `https://api.stevenshih-littleduck.com/health` 回 200；沒有 promote 時 Production 不變 |
| 3.5 | 照 runbook 第 6 節在 Staging 演練一次後端回滾 |
| 4.1 | 推送非 main 分支時 Drone 只跑 `ci-*`，沒有任何 `cd-*` pipeline；GHCR 沒有新的 `littleduck-web` 映像，`littleduck-web*` 容器不變 |
| 4.2 | 合併後 `cd-frontend-staging` 冒煙通過；`https://staging.stevenshih-littleduck.com/` 回 200，`/runtime-config.js` 的 `apiBaseUrl` 為 `https://api.staging.stevenshih-littleduck.com` |
| 4.3 | promote 後 `cd-frontend-production` 冒煙通過；`https://stevenshih-littleduck.com/` 為新版本，`/runtime-config.js` 為 `https://api.stevenshih-littleduck.com`；沒有 promote 時不變 |
| 5.4 | 對刻意讓 Go 測試失敗的 PR 跑 `/ship-issue`，第 7.1 步退回，`handoff.md` 有含 Drone 連結的紀錄 |
| 6.1 | 從開 PR 到 Production 完整走一次，把各階段耗時填進 `docs/cicd.md`「各階段耗時」 |

另外（自架後新增，非 SDP-113 項次）：

- runbook 第 3 節第 1 步：量測主機記憶體並填寫紀錄表，判準不足時先調整再上線。
- runbook 第 3 節第 2–3 步：apex 與 `staging.` 改 A 記錄（DNS only），清除 mygo 舊 `staging.` 的 A 記錄與 certbot 續約設定。
- 在瀏覽器開 Staging 與正式前端的文章列表，以開發者工具確認 API 請求分別打 `api.staging.` 與 `api.`（frontend-delivery「前端 Staging 與 Production 互相隔離」）。
- （建議）照 runbook 第 6 節在 Staging 演練一次前端回滾。
- Production 穩定後，照 runbook 第 3 節第 8 步刪除舊託管服務的專案與 Drone 裡殘留的四個舊 secret。

### 待確認問題

無。

---

## backend（Yasumi）— 2026-11-14

- **做了什麼**：
  - 第 1 組：`docs/deployment-runbook.md` 主機名對照表、第 0/1/2/5/6 節佔位換成實際主機名，第 2 節憑證改兩行明確指令，第 3 節前端 `.env` 的 `API_BASE_URL` 改實際值。
  - 第 2 組：`docs/cicd.md` 冒煙目標、Secret 清單加「值」欄、文末新增「各階段耗時」空表（耗時與日期為「待量測」）。
  - 第 3 組：`deploy/drone/.env.example` 預填 `DRONE_SERVER_HOST`；兩支 smoke 腳本只改 `#` 註解。
  - 第 4 組驗證：design.md 觀察點全部符合（佔位 grep 無輸出、`api.staging.` 計數 6、`DNS only` 2 行、render.sh 渲染計數 4、`sh -n` ok、無 vercel 字樣）；`openspec validate sdp-117-cicd-rollout --strict` 通過；`git diff --name-only origin/main` 只含 Impact 檔案與本變更目錄。
- **未完成**：無。
- **風險**：
  - 命中 `deploy/**`，整個變更為高風險。
  - 觀察點第 2 條 `git grep` 範圍含 `openspec/specs`：主規格 `backend-delivery/spec.md` 第 20 行仍有 `api-staging.<網域>`，由本變更的 delta 在歸檔時同步取代，非本次可直接改（不得直接改主規格）。歸檔後該 grep 應為空，請 integrator 於歸檔後確認。
  - 本次沒有 Go 程式碼變動，未跑 `go test`；未涉及 AutoMigrate，無 schema 變更。
- **API 契約變動**：無（openapi.yaml 維持 `paths: {}`）

---

## integrator（Teshima）— 檢查通過 2026-11-14

- **各端檢查**（repos 只有 backend，本次無 Go 程式碼變動，仍照表跑）：`gofmt -l .` 無輸出；`go vet ./...` 通過；`go test ./...` 全部 ok。前端不在 repos，執行驗證時另跑 `npm ci`、`npm run build` 成功。無 migration（未涉 AutoMigrate）。
- **規格**：`openspec validate sdp-117-cicd-rollout --strict` 通過；`openspec validate --specs` 23 passed。
- **變異測試**：不適用（本次無關鍵模組）。
- **tasks**：無未勾選項；worktree 乾淨、與 origin 一致；commit 只動 docs/、deploy/ 與變更目錄，未越界；proposal.md 只有提案 commit。`git diff --name-only` 只含 Impact 檔案；`deploy/ci/` diff 只有 `#` 註解行。
- **執行驗證（專案通用）**：後端 `/health` 回 `{"status":"ok","database":"up"}`；`/api/v1/articles` 200；前端 `/blog/a1` 200、`/` 200。已停止並清除（docker down -v、kill next），無殘留。
- **執行驗證（本次專屬，design.md 8 條）**：
  1. 佔位 grep：無輸出，通過。
  2. `api-staging.` grep：範圍含 `openspec/specs`，主規格 `backend-delivery/spec.md:20` 仍有 `api-staging.<網域>`。判定為**歸檔前可預期狀態**：變更 delta 的 MODIFIED「自動部署到 Staging」同名且已改為 `api.staging.`，歸檔 sync 會取代；docs、deploy、delta specs 部分無輸出。**歸檔後須重跑此 grep 確認為空**。不判失敗。
  3. `api.staging.` 計數 6（>=5），通過。
  4. vercel grep：無輸出，通過。
  5. `## 各階段耗時` 恰好一行（168），通過。
  6. DRONE_SERVER_HOST 計數 1，通過。
  7. `sh -n` 兩支腳本 ok。
  8. render.sh 渲染計數 4，通過。
- **截圖觀察點**：無（略過）。截圖暫存目錄：無。
- **規格免同步聲明**：無（專案無 spec-sync.yaml，不適用）。
- **風險**：命中 `deploy/**`，整個變更為高風險，PR 中要等 user 回覆「合併」；觀察點 2 於歸檔後須複查。
