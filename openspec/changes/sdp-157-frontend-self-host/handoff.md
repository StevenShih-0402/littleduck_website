<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-157-frontend-self-host

## 狀態

<!-- 就地改值，不追加。中斷後重新派工時，開發 Agent 以這裡判斷從哪裡接續；不寫 commit SHA（以 git log --grep "wip(<ISSUE-ID>): <任務編號>" 對應）。
  階段由誰寫入：proposal（architect 建立；主對話提交提案時更新「最後完成」）→ backend（後端 Agent 開工）
  → frontend（後端 Agent 交棒，或沒有 [BE] 任務時前端 Agent 開工）→ integrating（最後一個開發 Agent 交接）→ reviewing（integrator 檢查通過；主對話做 PR 前審查）
  → pr-checking（integrator 收尾重切並開 PR 後；主對話做 PR 中檢查）
  → awaiting-approval（只有高風險：主對話等待 user 回覆「合併」）→ archived（integrator 歸檔；之後由 integrator 合併）。
  integrator 檢查、PR 前審查、CI 或 PR 中檢查退回時改回被退回那一端的 backend／frontend（退回 architect 為 proposal）。
  合併隊列失敗不改階段，停下交給 user。 -->

- **階段**：awaiting-approval <!-- proposal | backend | frontend | integrating | reviewing | pr-checking | awaiting-approval | archived -->
- **最後完成**：PR 中檢查
- **下一步**：等待 user 回覆「合併」後歸檔
- **暫停原因**：高風險，等待 user 回覆合併
- **更新**：主對話 — 2026-11-14

---

## architect（Hikaru）— 2026-11-14

- **做了什麼**：建立 proposal、design（資料流圖、CD 時序圖、執行驗證觀察點、變異測試目標）、specs delta（`frontend-delivery`、`frontend-data-layer`、`ci-pipeline`、`ci-server`）、`openapi.yaml`（`paths: {}`）、tasks（[BE] 10 項、[FE] 5 項）。
  - 關鍵決策：`NEXT_PUBLIC_API_BASE_URL` 是建置期寫死，改成執行期由 `/runtime-config.js` 注入 `API_BASE_URL`，讓同一映像 Staging → promote → Production（design D2）。
  - 前端 compose 放 `deploy/web/`（[BE]），Dockerfile 放 `frontend/`（[FE]）。
- **未完成**：無（提案階段）。
- **風險**：
  - 高風險變更（`.drone.yml`、`deploy/**`、`**/Dockerfile`）。
  - `[BE]` 3.1 的 `.drone.yml` 引用 `[FE]` 4.1 才建立的 `frontend/Dockerfile`、`frontend/.trivyignore`；Yasumi 階段無法實際建映像，整合驗證在 4.5。
  - 後端 Agent 寫入 `deploy/`、`docs/`、`.drone.yml`、`README.md`、`.gitignore`、`CLAUDE.md` 依賴主 repo `CLAUDE.md` 的暫時放寬（hook 讀主 repo 那份，主對話已實測放行）。worktree 的 `CLAUDE.md` 只移除高風險樣式中的 `frontend/vercel.json`，「## agent-team 設定」不動。
- **修訂**（2026-11-14，依主對話轉達的 user 答覆）：新增 [BE] 3.4 清除根目錄 `README.md`、`.gitignore`、`CLAUDE.md` 的 Vercel 殘留；執行驗證的 vercel grep 觀察點與 frontend-delivery 場景納入這三個檔案；移除兩項待確認問題。
- **API 契約變動**：無。
- **退回**：無

---

## backend（Yasumi）— 2026-11-14

- **做了什麼**：[BE] 2.1–2.6、3.1–3.4 全部完成，每項一筆 WIP commit。
  - `deploy/ci/image-ref.sh` 接受 `backend`（預設）／`frontend`，新增 `frontend-version.sh`；`remote-deploy.sh` 加選填第 4 參數 `service`（預設 `api`，後端呼叫不需改）。
  - 新增 `deploy/web/docker-compose.{staging,prod}.yml`、`deploy/edge/web{,-staging}.conf.template`、`deploy/ci/frontend-bundle.sh`、`smoke-frontend.sh`；刪除 `vercel-deploy.sh`。
  - `.drone.yml`：`ci-frontend` 加 docker-build／trivy／cleanup；`cd-frontend-preview` 改為 `cd-frontend-staging`（映像流程）；`cd-frontend-production` 改為 bundle → deploy → smoke。
  - `docs/cicd.md`、`docs/deployment-runbook.md`、`README.md`、`.gitignore` 改寫或清除；worktree `CLAUDE.md` 只移除高風險樣式的 `frontend/vercel.json`。
- **驗證**：`image-ref.sh` 三種輸出、`sh -n`、兩份 compose `config -q`（未設映像變數時失敗）、`render.sh web web-staging`（server_name 各 2、憑證路徑正確）、`frontend-bundle.sh staging`（五個檔案齊全）、`drone lint --trusted .drone.yml` 通過、各檔 vercel grep 皆為 0。
- **未驗證（待前端完成後 4.5）**：`smoke-frontend.sh` 對真實映像的成功與失敗路徑（僅以假 HTTP 伺服器驗證過邏輯）；`.drone.yml` 引用的 `frontend/Dockerfile`、`frontend/.trivyignore` 尚未存在。
- **風險**：
  - 高風險變更（`.drone.yml`、`deploy/**`）。
  - `smoke-frontend.sh` 不跟隨轉址，`/blog` 須直接回 2xx（Next.js 預設不帶結尾斜線轉址，預期可行；4.5 會實測，若 `/blog` 回 308 需調整）。
  - runbook 為避免 `grep -i vercel` 命中，以「舊託管服務」代稱，內容仍指示 user 刪除舊專案與四個舊 secret。
- **前端注意**：compose healthcheck 打 `http://127.0.0.1:3000/runtime-config.js`，所以映像內必須有 `wget`（node:20-alpine 的 busybox 有）且服務監聽 `0.0.0.0:3000`；`/runtime-config.js` 路由須存在。映像內需能從 `API_BASE_URL` 執行期環境變數讀值。
- **API 契約變動**：無（`openapi.yaml` 維持 `paths: {}`，未改）。
- **退回**：無

---

## frontend（Fujimori）— 2026-11-14

- **做了什麼**：[FE] 4.1–4.5 全部完成。
  - `next.config.mjs` 加 `output: "standalone"`；新增 `frontend/Dockerfile`（三段式、`USER node`、`HOSTNAME=0.0.0.0`、`PORT=3000`）、`.dockerignore`、`.trivyignore`（僅說明註解）。
  - 新增 `lib/runtime-config.mjs` 與 6 項測試；`app/runtime-config.js/route.js`（`force-dynamic`、`no-store`）；`app/layout.jsx` 以 `<Script beforeInteractive>` 載入（ESLint 未報錯，不需豁免）；`lib/api.js` 每次請求取位址。
  - 刪除 `frontend/vercel.json`；`.env.example` 說明 `API_BASE_URL` 與 `NEXT_PUBLIC_API_BASE_URL` 的優先序。
- **驗證**：
  - `npm ci`、`npm run test:design`（32 項全過）、`npm run lint`（0 error、9 既有 warning）、`npm run typecheck`、`npm run build` 皆通過；`npm audit --audit-level=high` 結束碼 0（僅 moderate 的 minimatch，既有依賴）。
  - 映像（3101 埠）：`/`、`/blog`、`/blog/a1` 皆直接回 200；`/runtime-config.js` 回 `{"apiBaseUrl":"http://localhost:18080"}`；容器內 `wget` 打 `/runtime-config.js` 成功（compose healthcheck 可用）；`User` 為 `node`。
  - `smoke-frontend.sh`：位址正確時三項通過、`exit=0`；位址錯誤時 `/runtime-config.js` 失敗、`exit=1`。`/blog` 不轉址，後端的疑慮（308）不成立。
  - `sh -n` deploy 腳本無語法錯誤；vercel grep（`.drone.yml deploy docs frontend README.md .gitignore CLAUDE.md`）無輸出。
  - 未跑：`docker compose` healthcheck 實際由 Docker 判定、Trivy 掃描（需在 CI 執行）。
- **仍靠 mock**：無（`openapi.yaml` 為 `paths: {}`）。
- **規格同步**：專案無 `openspec/spec-sync.yaml`，不適用。
- **風險**：高風險（`**/Dockerfile`）。映像建置需連 Google Fonts（`next/font/google`）；本機建置成功，CI 環境需可連外。
- **API 契約變動**：無。
- **退回**：無（不需退回後端）。

---

## integrator（Teshima）— 檢查通過

- **後端（backend/）**：`go test ./...` 全 ok；`gofmt -l .` 無輸出；`go vet ./...` 通過。migration：無 model 變更，不適用。
- **前端（frontend/）**：`npm ci`；`npm run test:design` 32 通過；`npm run lint` 0 error、9 既有 warning；`npm run typecheck` 通過；`npm run build` 通過；`npm audit --audit-level=high` exit 0。
- **規格**：`openspec validate sdp-157-frontend-self-host --strict` 通過；`openspec validate --specs` 23/23。無 `spec-sync.yaml`，規格同步不適用。
- **變異測試**：略過（CLAUDE.md 未設定變異測試；目標為 `frontend/lib/runtime-config.mjs`）。
- **執行驗證（專案通用）**：`/health` 為 `{"status":"ok","database":"up"}`；`/api/v1/articles` 200；`localhost:3100/blog/a1` 200。
- **執行驗證（本次專屬）**：3100 `/runtime-config.js` 為 `apiBaseUrl:null`；映像 3101 的 `/` 200、`/runtime-config.js` 為 `http://localhost:18080`、`/blog/a1` 200、User `node`；smoke 對正確位址三項通過 exit=0，對錯誤位址 runtime-config 失敗 exit=1；`sh -n` 無錯誤；image-ref 輸出 api 為 `littleduck-api:0.1.0-<sha7>`、frontend 為 `littleduck-web:2.0.0-<sha7>`；render.sh server_name 兩份各 2；vercel grep 無輸出；`drone lint --trusted .drone.yml` ok；映像已移除、服務已停止。無截圖觀察點。
- **核對**：tasks.md 無未勾選；worktree 乾淨且與 origin 同步；BE/FE commit 無越界；proposal.md 未被開發 Agent 改動；openapi `paths: {}` 無契約變動。
- **規格免同步聲明**：無。
- **風險**：高風險（`.drone.yml`、`deploy/**`、`**/Dockerfile`）；Trivy 掃描與 compose healthcheck 只能在 CI／部署時驗證；映像建置需連 Google Fonts。
- **退回**：無

---

## integrator（Teshima）— 收尾

- **重切前 SHA（OLD）**：`a23d528feb3326e4bfcddd95e769f00fef70f0ce`
- **重切起點（FROM）**：`c4215fdb7c3b598aa744cfdb2e61c478061b09ec`（merge-base）
- **重切後 commit**：
  - `5dcf4a0` feat(ci): 前端改為映像建置並自架部署（含 OpenSpec 變更目錄）
  - `9379f98` feat(web): 前端容器化並於執行期注入 API 位址
  - `b9a3243` docs: 改寫前端部署文件並清除 Vercel 殘留
- **tree 比對**：`git diff OLD HEAD` 為空
- **PR**：https://github.com/StevenShih-0402/littleduck_website/pull/12（內文附 review.md 3 條 PR 備註，並註明主 repo CLAUDE.md 對後端子目錄的暫時放寬不在本 PR 內）
- **風險**：高風險（`.drone.yml`、`deploy/**`、`**/Dockerfile`）；Trivy 與 compose healthcheck 待 CI／部署驗證；review.md 3 條中嚴重度備註未處理。
- **收尾 HEAD**：`b9a32437ca7d7663a5f21dc96f89bb8f8c7c7a8e`
