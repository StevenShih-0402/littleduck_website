# Proposal

## Why

littleduck_website 目前沒有任何 CI/CD（repo 內沒有 `.drone.yml`、`.github/workflows/`、`Jenkinsfile`），後端只有 Dockerfile 與本機用的 docker-compose，前端沒有部署設定（SDP-113）。agent-team 的 `/ship-issue` 在「PR 中」第一步會等 PR 的 CI check，但本 repo 沒有 check 可等。另外本 repo 裝的 agent-team kit 還是四角色、兩個人工檢查點的舊版，根本沒有 CI 等待步驟。本 change 與 code_city 的同名 change（`sdp-113-drone-cicd`）一起處理：兩專案都把 CI 與 CD 分開、流程內容參考 mygo_v2_anon 的 Jenkins 流水線；本 repo 以自架 Drone 為主要工具，code_city 因需要 macOS 建置改用 GitHub Actions。

## What Changes

- **Drone server 與 runner**：在 Hetzner CX23（mygo_v2_anon 所在主機）以 docker compose 架設 Drone server（`drone/drone:2`）與 Docker runner（`drone/drone-runner-docker:1`），用 GitHub OAuth App 連結，對外網址為 `drone.<網域>` 子網域。設定檔放在本 repo 的 `deploy/drone/`。
- **CI（`.drone.yml` 的 `ci-backend`／`ci-frontend` pipeline）**：PR 與推送到 `main` 時觸發，結果以 GitHub commit status 回報到 PR（`gh pr checks` 看得到）。
  - 後端：`gofmt` 檢查 → `go vet` → `go test ./...`（搭配臨時 PostgreSQL service）→ `gosec` 安全掃描（對應 mygo 的 Bandit）→ `docker build` → Trivy 映像掃描（CRITICAL／HIGH 即失敗）。
  - 前端：`npm ci` → `npm run test:design` → `npm run build` → `npm audit --audit-level=high`。
  - CI 不推送映像、不部署。
- **CD（`cd-backend`／`cd-frontend` pipeline）**：只在 `main` 收到推送、或 Drone promote 時觸發，與 CI pipeline 分開。
  - 後端：建置並推送 GHCR 映像（`<version>-<sha>`、`<version>`、`latest`）→ SSH 部署到 Hetzner Staging → 健康檢查 → 冒煙測試 → 在 Drone 介面按 promote（`target: production`）視為人工核准 → SSH 部署到 Production → Production 冒煙測試。
  - 前端：`vercel build` → 部署 Vercel Preview（視為 Staging）→ 冒煙測試 → promote（`target: production`）→ `vercel deploy --prebuilt --prod` → Production 冒煙測試。關閉 Vercel 自己的 Git 自動部署，部署只由 Drone 觸發。
- **後端正式環境的 compose**：新增 `backend/docker-compose.prod.yml`、`backend/docker-compose.staging.yml`（api ＋ PostgreSQL 16，Staging 與 Production 各自獨立的 DB 與 env 檔），加入主機上既有的 `mygo-edge` 網路，由 mygo_v2_anon 的 Nginx 反向代理。
- **Nginx 站台設定**：新增 `deploy/edge/` 下的 server block 樣板（`api`、`api-staging`、`drone` 三個子網域）。**前置條件**：mygo_v2_anon 的 Nginx 要能額外載入主機上 `/opt/edge-sites/*.conf`，這是 mygo_v2_anon 的變更，不在本 change 範圍。
- **升級 agent-team kit**：依 skill_repo（littleduck-skills）最新版升級本 repo 的 team-hq 骨架，改成和 code_city 一樣的「單一 repo」設定，補上合併方式、CI 等待上限、高風險檔案樣式（含 `.drone.yml`、`deploy/**`）、風險分級與執行驗證段落，讓 `/ship-issue` 的「7.1 CI」等得到 Drone 的 check。
- **清掉既有弱點，讓 CI 門檻一上線就是綠的**（實作時發現，user 於 2026-11-09 決定納入）：現行 `main` 上 Trivy 掃出 pgx v5.6.0（CRITICAL）與 `golang.org/x/{crypto,net,text,mod}`（HIGH），npm audit 掃出 next 14.2.35（CRITICAL）及數個 high。
  - 後端：升級上述 Go 模組到已修正版本（皆為非破壞性的 minor／patch 升級）。
  - 前端：**Next.js 14 → 16、React 18 → 19**（major 升級），並以 `npm audit fix` 修掉其餘非破壞性的弱點；`app/(front)/blog/[id]/page.jsx` 改用 `useParams()` 讀路由參數（Next 15 起 `params` 改為 Promise）。
- **文件**：新增 `docs/cicd.md`（流水線階段與門檻）與 `docs/deployment-runbook.md`（Drone／主機初次建置、日常發布、回滾）。

## Capabilities

### New Capabilities
- `ci-server`：自架 Drone server 與 runner 的部署、GitHub 連結、祕密管理與對外存取。
- `ci-pipeline`：PR 與 `main` 推送時的前後端 CI 階段、失敗門檻，以及回報到 GitHub PR 的 check。
- `backend-delivery`：後端映像推送、Hetzner Staging／Production 部署、人工核准與冒煙測試、回滾。
- `frontend-delivery`：前端 Vercel Preview／Production 部署、人工核准與冒煙測試。
- `agent-ci-gate`：agent-team 的 `/ship-issue` 等待 Drone CI 結果並據以退回或往下的行為，以及本 repo 的 team-hq 設定。

### Modified Capabilities
（無。`backend-deployment` 的兩項需求——Docker 映像可建置執行、docker-compose 本機環境——行為不變，本 change 只新增正式環境用的 compose 檔。）

## Impact

- 實作時補充（2026-11-09）：
  - 新增 `deploy/ci/*.sh`：流水線的部署、冒煙與 promote 前置檢查腳本，可在本機單獨測試。
  - `backend/docker-compose.yml` 的主機埠改為可用 `API_PORT` 覆寫，預設仍是 8080。
  - `backend/cmd/server/main.go` 只做 `gofmt`，CI 的格式門檻上線前它本來就不合格。
  - 新增 `frontend/vercel.json`。
- 新增：`.drone.yml`、`deploy/drone/`、`deploy/edge/`、`backend/docker-compose.prod.yml`、`backend/docker-compose.staging.yml`、`backend/.trivyignore`、`docs/cicd.md`、`docs/deployment-runbook.md`。
- 修改：`CLAUDE.md`（team-hq 設定）、`.claude/commands/ship-issue.md`、`.claude/agents/*`、`.claude/skills/issue-*`（kit 升級）、`openspec/handoff-template.md`、新增 `openspec/review-template.md`、`README.md`。
- 相依升級：`backend/go.mod`／`go.sum`（Go 模組）、`frontend/package.json`／`package-lock.json`（Next 16、React 19）、`frontend/app/(front)/blog/[id]/page.jsx`。前後端業務邏輯不變，也沒有 API 變更。
- 外部系統：Hetzner CX23（多跑 Drone server、runner 與 littleduck 的 api／db 容器）、GitHub（OAuth App、webhook、GHCR）、Vercel（專案與 token）、DNS（`drone.`、`api.`、`api-staging.` 子網域）。
- 相依：mygo_v2_anon 的 Nginx 必須能載入 `/opt/edge-sites/*.conf`（SDP-114）。code_city 的同名 change 不使用 Drone，兩者沒有實作上的相依。
