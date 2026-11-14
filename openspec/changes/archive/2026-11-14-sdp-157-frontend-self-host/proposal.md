# Proposal

## Why

網站之後要接贊助金流（SDP-104 綠界 ECPay），有金流就不符合 Vercel Hobby 方案的「非商業用途」限制。前端改成跟後端一樣自架在 Hetzner 主機（mygo_v2_anon 使用的 CX23），停用 Vercel；SDP-117（CI/CD 上線）要等本變更完成，才能依自架版本的文件重寫提案。

## What Changes

**前端程式與映像（`frontend/`）**

- `next.config.mjs` 加 `output: "standalone"`；新增 `frontend/Dockerfile`（multi-stage，`node:20-alpine`，非 root 執行 `node server.js`，埠 3000）、`frontend/.dockerignore`、`frontend/.trivyignore`（空清單，格式同 `backend/.trivyignore`）。
- **後端位址改在執行期決定**：新增 `/runtime-config.js` 路由，回傳 `window.__LD_RUNTIME_CONFIG__={"apiBaseUrl":…}`，值取自容器環境變數 `API_BASE_URL`；root layout 以 `beforeInteractive` 載入。`lib/api.js` 每次請求依「執行期設定 → 建置期 `NEXT_PUBLIC_API_BASE_URL` → `http://localhost:8080`」決定位址。這樣同一個映像能在 Staging 驗證後原封不動 promote 到 Production（見 design D2）。
- 移除 `frontend/vercel.json`；`.env.example` 補 `API_BASE_URL` 說明。

**部署（`deploy/`、`.drone.yml`、`docs/`）**

- 新增 `deploy/web/docker-compose.staging.yml`、`docker-compose.prod.yml`：容器 `littleduck-web-staging`／`littleduck-web`，接 `mygo-edge`，不對外開埠，設記憶體上限與 healthcheck。
- 新增 Nginx 樣板 `deploy/edge/web.conf.template`（`${FRONTEND_HOST}`，正式 `stevenshih-littleduck.com`）、`web-staging.conf.template`（`${FRONTEND_STAGING_HOST}`，Staging `staging.stevenshih-littleduck.com`），沿用 `/opt/edge-sites/*.conf` 的載入方式。
- 部署腳本：`image-ref.sh` 接受 `backend|frontend`（預設 `backend`）；新增 `frontend-version.sh`、`frontend-bundle.sh`、`smoke-frontend.sh`；`remote-deploy.sh` 加選填的第 4 個參數 `<service>`（預設 `api`）。後端既有呼叫方式不變。
- `.drone.yml`：
  - `ci-frontend` 加 `docker-build`、`trivy`、`cleanup`，與 `ci-backend` 對齊；
  - `cd-frontend-preview` 改名為 `cd-frontend-staging`：建置映像 → trivy → 推 GHCR（`ghcr.io/stevenshih-0402/littleduck-web`）→ 部署到 Staging → 冒煙；
  - `cd-frontend-production`：verify-build → 部署 Staging 驗證過的同一個 `<version>-<sha>` 映像 → 冒煙。
- 刪除 `deploy/ci/vercel-deploy.sh`；Drone secret 移除 `vercel_token`、`vercel_org_id`、`vercel_project_id`、`vercel_bypass_secret`，新增 `frontend_staging_host`。
- `docs/cicd.md`、`docs/deployment-runbook.md` 改成自架版本：前端主機初次建置（目錄、`.env`、憑證、DNS、清理 mygo 舊 `staging.` 主機名）、主機記憶體評估、日常發布與回滾。

**BREAKING（部署層面，非 API）**：Drone pipeline `cd-frontend-preview` 改名為 `cd-frontend-staging`，前端不再部署到 Vercel；上線前 user 要照 runbook 完成主機設定與 secret。HTTP API 沒有任何變動。

## Non-goals

- 實機架設（DNS、certbot 核發、`/opt/littleduck-web*` 目錄與 `.env`、Drone secret 建立、刪除 Vercel 專案）由 user 照 runbook 手動完成。
- 不處理 `www.` 子網域與後端 CORS 收斂（目前 `Access-Control-Allow-Origin: *`，新前端主機名可直接呼叫）。
- 不把 `docs/` 裡後端的 `<網域>`、`api-staging.<網域>` 佔位換成實際主機名，那是 SDP-117 的範圍；本變更新增的前端段落沿用同樣的佔位寫法，並在主機名對照表列出 issue 指定的實際值。
- 不改任何後端程式碼與 HTTP 端點。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `frontend-delivery`：由「Vercel Preview → Production」改為「GHCR 映像 → Hetzner Staging → promote 同一映像到 Production」，新增映像產出、Staging 部署與冒煙、回滾等需求。
- `frontend-data-layer`：新增「後端位址於執行期決定」需求。
- `ci-pipeline`：前端 CI 門檻加入映像建置與弱點掃描；「CI 不部署」場景移除 Vercel。
- `ci-server`：部署用祕密清單移除 Vercel token。

## Impact

預計修改的檔案（供判斷能否與其他 issue 並行）：

| 端 | 檔案 | 改動 |
|----|------|------|
| BE | `.drone.yml` | `ci-frontend` 加映像建置與掃描；`cd-frontend-preview` → `cd-frontend-staging`；`cd-frontend-production` 改自架 |
| BE | `deploy/ci/image-ref.sh`、`remote-deploy.sh` | 參數化（後端預設值不變） |
| BE | `deploy/ci/frontend-version.sh`、`frontend-bundle.sh`、`smoke-frontend.sh` | 新增 |
| BE | `deploy/ci/vercel-deploy.sh` | 刪除 |
| BE | `deploy/edge/web.conf.template`、`web-staging.conf.template` | 新增；`render.sh` 只改用法註解 |
| BE | `deploy/web/docker-compose.staging.yml`、`docker-compose.prod.yml` | 新增 |
| BE | `docs/cicd.md`、`docs/deployment-runbook.md` | 改成自架版本 |
| BE | 根目錄 `README.md`、`.gitignore`、`CLAUDE.md` | 清除 Vercel 殘留：README CI/CD 段落改自架描述、移除 `.vercel/`、高風險樣式表移除 `frontend/vercel.json`（`CLAUDE.md` 其餘段落不動） |
| FE | `frontend/next.config.mjs`、`frontend/Dockerfile`、`frontend/.dockerignore`、`frontend/.trivyignore` | standalone 輸出與映像 |
| FE | `frontend/lib/runtime-config.mjs`、`frontend/lib/runtime-config.test.mjs`、`frontend/app/runtime-config.js/route.js`、`frontend/app/layout.jsx`、`frontend/lib/api.js` | 執行期決定後端位址 |
| FE | `frontend/vercel.json`（刪除）、`frontend/.env.example` | 移除 Vercel |
| — | `openspec/changes/sdp-157-frontend-self-host/**` | 本變更目錄；歸檔時同步 `frontend-delivery`、`frontend-data-layer`、`ci-pipeline`、`ci-server` |

- 根目錄 `README.md`、`.gitignore`、`CLAUDE.md` 由後端 Agent 處理（user 已同意暫時放寬主 repo 的後端寫入範圍，放寬本身不進 PR）。
- 命中「高風險檔案樣式」：`.drone.yml`、`deploy/**`、`**/Dockerfile`。整個變更判定為**高風險**，PR 中要等 user 回覆「合併」。
- `openapi.yaml` 為 `paths: {}`；`/runtime-config.js` 是前端站台自己的路由，不是後端 API。
- 與其他改 `.drone.yml`、`deploy/**`、`docs/cicd.md`、`docs/deployment-runbook.md`、`frontend/lib/api.js`、`frontend/app/layout.jsx`、`frontend/next.config.mjs` 的 issue 不宜並行（SDP-117 本來就被本 issue 擋住）；與只動 `backend/` 程式碼或其他前端頁面的 issue 可並行。
