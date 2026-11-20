# Proposal

來源：Linear SDP-183「littleduck_website：CI/CD 改用 Woodpecker CI 並簡化流程」（上層 SDP-175「簡化 CI/CD 流程」在本專案的範圍）。

## Why

本 repo 的 CI/CD 跑在自架的 Drone 上。Drone 被 Harness 收購後只剩維護性更新，長期使用有風險。Woodpecker CI 是從 Drone 0.8 分出來、仍在積極開發的專案（2026-11 最新穩定版 v3.20.1），YAML 寫法與 Drone 接近，遷移成本低。issue 同時要求把流程對齊理想模型：

- CI：PR → CI 檢查 → 其餘檢查 → 合併到 `main`
- CD：用 `main` 建出映像 → 部署 Staging → 人工在 Staging 確認 → 同意後部署 Production

逐階段比對現行 `.drone.yml` 後，主要多出來的步驟有三個。一是合併到 `main` 後，`ci-backend` 與 `ci-frontend` 對同一份程式再整套跑一次。二是 `cd-*-staging` 接著又建一次映像、再掃一次。三是前端 CI 的 `npm run build` 與緊接的 `docker build` 把 `next build` 跑了兩次。對照細節見 design.md「Context」。

## What Changes

- **CI 伺服器改為 Woodpecker**：在同一台 Hetzner 主機以 docker compose 跑 Woodpecker server 與 agent（v3.20.1），透過新子網域 `ci.stevenshih-littleduck.com` 的 HTTPS 對外，用 GitHub OAuth 登入，只允許 repo 擁有者。Drone 的 server、runner、站台、compose 與 `.env` 範本在切換穩定後移除。
- **流水線改寫為 `.woodpecker/*.yaml`**：沿用現有 6 條流水線的名稱與內容，語法改成 Woodpecker 的寫法（`CI_*` 變數、步驟層級 volume、`evaluate`、`deployment` 事件）。部署腳本 `deploy/ci/*.sh` 照常使用，只改變數名稱。
- **簡化流程**：
  - CI 只在 PR 上跑，推送到 `main` 時不再重跑。`main` 的品質由「PR 必須通過 CI 且分支是最新」保證，做法是 GitHub branch protection 加上 `/ship-issue` 的合併隊列。
  - 推送到 `main` 時直接跑 CD Staging：建映像、Trivy 掃描、推送、部署、冒煙。每個 commit 只建一次要上線的映像。
  - 前端 CI 拿掉獨立的 `npm run build`。正式建置改由 `docker build` 內的 `next build` 負責，lint 與型別檢查仍排在它前面。
  - Production 照舊不重建映像，直接用 Staging 驗證過的 `<version>-<sha>`。
- **人工核准改用 Woodpecker 的 deployment 事件**：Staging 通過後，user 在 Woodpecker 的 pipeline 頁按 Deploy（Target 填 `production`），或執行 `woodpecker-cli pipeline deploy`。Production 部署前會先確認該 commit 的 push 建置全部成功，取代 Drone 的 promote。
- **GitHub 上的 check 名稱改變（BREAKING，影響 agent 流程與 branch protection）**：`continuous-integration/drone/pr` 改為兩個 check，`ci/woodpecker/pr/ci-backend` 與 `ci/woodpecker/pr/ci-frontend`。push 建置的 context 改為 `ci/woodpecker/push/<workflow>`。
- **安全設定對齊**：fork 來的 PR 要人工核准才會跑（取代 Drone 的 ignore forks）。repo 的 Trusted 只開 volumes（掛 docker.sock 與 Trivy 快取需要）。部署用 secret 只開放 `push`、`deployment` 事件，PR 拿不到。
- **與 SDP-181 的關係**：SDP-181（Staging 身分驗證）先在 Drone 上合併。本變更把它在 `.drone.yml` 加的內容一併搬到 Woodpecker，包含兩個 Staging secret（`staging_basic_auth`、`staging_gate_token`）與冒煙參數。順序與搬移方式見 design.md 決策 8。
- 更新 `docs/cicd.md`、`docs/deployment-runbook.md`、`README.md`、`CLAUDE.md`（檢查指令說明、PR 中 CI 名稱、高風險檔案樣式），以及 compose 檔裡提到 Drone 的註解。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `ci-server`：CI 伺服器由 Drone 改為 Woodpecker。對外子網域、登入限制、事件觸發、祕密保護的需求改寫，新增「fork PR 需核准」與「Trusted 只開 volumes」。
- `ci-pipeline`：CI 只在 PR 執行，推送到 `main` 不再重跑。前端 CI 的正式建置改由映像建置負責，lint 與型別檢查仍排在建置之前。
- `backend-delivery`：映像推送不再以「同一次建置的 CI 通過」為前提。Production 核准改為 Woodpecker 的 deployment 事件（target `production`），部署前驗證該 commit 的 push 建置全部成功。
- `frontend-delivery`：「部署只由 Drone 觸發」改為由 Woodpecker 觸發，Vercel 檢查範圍改為 `.woodpecker/`。Production 核准同上。映像推送不再以同一次建置的前端 CI 為前提。
- `agent-ci-gate`：等待與退回的對象改為 Woodpecker 的 PR check，高風險樣式由 `.drone.yml` 改為 `.woodpecker/**`。
- `frontend-project-structure`：「前端統一使用 Node 24 LTS」中 CI 的檢查對象由 `.drone.yml` 的 `ci-frontend` 改為 `.woodpecker/ci-frontend.yaml`。

## Impact

- 新增：`.woodpecker/ci-backend.yaml`、`ci-frontend.yaml`、`cd-backend-staging.yaml`、`cd-frontend-staging.yaml`、`cd-backend-production.yaml`、`cd-frontend-production.yaml`；`deploy/woodpecker/docker-compose.yml`、`deploy/woodpecker/.env.example`；`deploy/edge/woodpecker.conf.template`
- 修改：`deploy/ci/image-ref.sh`（commit sha 改讀 `CI_COMMIT_SHA`）、`deploy/ci/verify-commit-status.sh`（改查 `ci/woodpecker/push/*`）、`deploy/edge/render.sh`（用法註解）；`docs/cicd.md`、`docs/deployment-runbook.md`、`README.md`、`CLAUDE.md`、`.gitignore` 註解、四個 compose 檔的註解
- 移除：`.drone.yml`（本 PR）；`deploy/drone/`、`deploy/edge/drone.conf.template`（觀察期後另開小 PR，見 tasks 的 Workflow follow-up）
- 主機與外部設定（user 手動）：DNS 與憑證（`ci.` 子網域）、GitHub OAuth App、Woodpecker 的 `.env`、repo 啟用與設定（Trusted volumes、Allow deployments、fork 核准）、重建 10 個 secret、新的部署金鑰、GitHub branch protection、停用 Drone 並移除它的 webhook
- 命中高風險檔案樣式（`.drone.yml`、`.woodpecker/**`、`deploy/**`），合併前要等 user 看過 diff
- 不改 `backend/`、`frontend/` 的執行期程式碼，Staging 與 Production 的容器、`.env`、站台設定都不變
