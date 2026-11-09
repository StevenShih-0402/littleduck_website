# Tasks

> 前置：mygo_v2_anon 的 Nginx 需先能 include `/opt/edge-sites/*.conf`（SDP-114）。第 1 組可先做；第 1.4 起的對外驗證、第 3、4 組的部署驗證要等前置完成。

## 1. Drone server 與 runner（ci-server）

- [x] 1.1 新增 `deploy/drone/docker-compose.yml`（drone server＋docker runner、資源上限、`DRONE_RUNNER_CAPACITY=1`、server 不對外開埠、加入 `mygo-edge`）與 `deploy/drone/.env.example`（`DRONE_GITHUB_CLIENT_ID`／`SECRET`、`DRONE_RPC_SECRET`、`DRONE_SERVER_HOST`、`DRONE_USER_FILTER`），`docker compose -f deploy/drone/docker-compose.yml config` 通過
- [x] 1.2 新增 `deploy/edge/drone.conf.template`（`drone.<網域>` 的 80→443 轉址、ACME webroot、反向代理到 drone server），在本機以 `nginx:1.27-alpine` 執行 `nginx -t` 通過
- [ ] 1.3 撰寫 `docs/deployment-runbook.md` 的「Drone 初次建置」節：GitHub OAuth App、DNS、certbot、主機目錄、啟動與 `DRONE_USER_FILTER`、授權註記；照文件在主機啟動後 `docker compose ps` 兩個容器皆 running
  - env-blocked（2026-11-09）：runbook 第 0、1 節已寫好；主機上的實際啟動需要 GitHub OAuth App、DNS、憑證與 mygo Nginx 的 include 前置，尚未進行。
- [ ] 1.4 驗證對外存取：`https://drone.<網域>` 顯示登入頁、非允許帳號登入被拒、直接連容器埠失敗
  - env-blocked：需先完成 1.3 的主機部署。

## 2. CI pipeline（ci-pipeline）

> 2.6、2.7 是實作時發現既有弱點後追加（user 決定連 Next 16 一起升），要先做完 2.1、2.3 的「本機跑綠」才成立。

- [x] 2.6 升級後端 Go 模組清掉 Trivy 的 CRITICAL／HIGH（pgx ≥ 5.9.0、`golang.org/x/{crypto,net,text,mod}` 到修正版），`go test ./...` 通過，重建映像後 Trivy `--severity CRITICAL,HIGH --exit-code 1` 結束碼為 0
- [x] 2.7 前端升級 Next.js 16＋React 19 並 `npm audit fix`：`blog/[id]/page.jsx` 改用 `useParams()`；`npm run build`、`npm run test:design` 通過、`npm audit --audit-level=high` 結束碼為 0；以 `npm run start` 實際開啟首頁、部落格列表與一篇文章頁回 200

- [x] 2.1 新增 `.drone.yml` 的 `ci-backend` pipeline（gofmt 檢查、`go vet`、`go test ./...` 搭配 `postgres:16` service、`gosec`、`docker build`、Trivy CRITICAL/HIGH），trigger 為 `pull_request` 與 `push`/`main`；`drone lint` 通過，並以 `drone exec --pipeline ci-backend` 在本機跑綠
- [x] 2.2 新增 `backend/.trivyignore`（空檔＋登記格式說明），故意在本機分支放一個未格式化的 Go 檔跑 `drone exec`，確認 gofmt 步驟失敗並列出檔名
- [x] 2.3 新增 `ci-frontend` pipeline（`npm ci`、`npm run test:design`、`npm run build`、`npm audit --audit-level=high`），`drone exec --pipeline ci-frontend` 本機跑綠
- [ ] 2.4 在 Drone 啟用本 repo、設定「PR 不可用」的部署 secret，開一個只改文件的 PR，`gh pr checks` 列出兩個 Drone check 且皆成功、GHCR／Vercel 沒有新部署
  - env-blocked：需 Drone server 上線並設定 secret。PR 建置拿不到部署 secret 的設計已寫進 `.drone.yml` 註解與 `docs/cicd.md`。
- [x] 2.5 撰寫 `docs/cicd.md` 的 CI 節（階段與門檻表、本機 `drone exec` 用法、`.trivyignore` 登記慣例），並在 `README.md` 加連結

## 3. 後端 CD（backend-delivery）

- [x] 3.1 新增 `backend/docker-compose.staging.yml`、`backend/docker-compose.prod.yml`（`image: ${LITTLEDUCK_API_IMAGE}`、各自的 PostgreSQL 與 named volume、`env_file`、api 的 healthcheck 打 `/health`、加入 `mygo-edge`、容器名不與 mygo 衝突），`docker compose config` 兩檔皆通過
- [x] 3.2 新增 `deploy/edge/api.conf.template`、`deploy/edge/api-staging.conf.template`，`nginx -t` 通過
- [ ] 3.3 新增 `.drone.yml` 的 `cd-backend-staging` pipeline（`push`/`main`，`depends_on: ci-backend`）：推 GHCR 三個 tag → scp compose 與 edge 設定 → SSH `pull`、`up -d` → 120 秒健康輪詢 → 對 `/health` 與 `/api/v1/articles` 冒煙；合併一個文件 PR 後確認 Staging 跑新映像且冒煙通過
  - env-blocked：pipeline 已寫好、`drone lint` 通過；部署腳本已在本機驗證——`backend-bundle.sh` 產出正確部署包、`remote-deploy.sh` 在模擬主機上 10 秒內 healthy、`install-sites.sh` 在缺憑證時正確擋下、`smoke-backend.sh` 的成功／失敗／連不上三種情況都正確。實際部署需要 Hetzner、GHCR 與 secret。
- [ ] 3.4 新增 `cd-backend-production` pipeline（`event: promote`、`target: production`）：以同一個 `<version-sha>` 部署 Production → 健康輪詢 → 冒煙；promote 一次確認 `https://api.<網域>/health` 回 200，且未 promote 時 Production 不變
  - env-blocked：pipeline 已寫好；promote 前置檢查 `verify-commit-status.sh` 已對真實 commit 驗證會擋下沒有 Drone 狀態的 commit。實際 promote 需要 Drone 與主機。
- [ ] 3.5 撰寫 runbook 的「後端主機初次建置」「日常發布」「回滾」節（env 檔放置、`drone-deploy` 金鑰、GHCR 主機登入、promote 操作、指定舊 tag 回滾），照回滾步驟在 Staging 演練一次成功
  - env-blocked：runbook 第 2、5、6 節已寫好；回滾演練需要主機上的 Staging。
- [x] 3.6 在 `docs/cicd.md` 補後端 CD 階段表與映像 tag 規則

## 4. 前端 CD（frontend-delivery）

- [ ] 4.1 新增 `frontend/vercel.json`（`git.deploymentEnabled: false`），並在 runbook 寫建立 Vercel 專案、Root Directory、取得 `VERCEL_ORG_ID`／`VERCEL_PROJECT_ID`／token 的步驟；推送非 main 分支後確認 Vercel 沒有新部署
  - env-blocked：`frontend/vercel.json` 與 runbook 第 3 節已完成；「推送非 main 分支後 Vercel 沒有新部署」需要 Vercel 專案。
- [ ] 4.2 新增 `cd-frontend-preview` pipeline（`push`/`main`，`depends_on: ci-frontend`）：`vercel pull`→`vercel build`→`vercel deploy --prebuilt`，印出 Preview 網址並對首頁冒煙；合併後確認記錄有網址且冒煙通過
  - env-blocked：pipeline 與 `deploy/ci/vercel-deploy.sh` 已寫好；需要 Vercel 專案與 token。
- [ ] 4.3 新增 `cd-frontend-production` pipeline（`event: promote`、`target: production`）：`vercel build --prod`→`vercel deploy --prebuilt --prod`→對正式網域冒煙；promote 一次確認正式網域更新
  - env-blocked：同 4.2。
- [x] 4.4 在 `docs/cicd.md` 補前端 CD 節與 `vercel rollback` 回滾方式

## 5. agent-team kit 升級（agent-ci-gate）

- [x] 5.1 以 skill_repo 最新版執行 `install.sh agent-team .`，再手動合併不會被覆寫的骨架檔（`.claude/commands/ship-issue.md`、`.claude/agents/*`（含新增 `shiori.md`）、`.claude/hooks/guard-write-scope.sh`、`openspec/handoff-template.md`、新增 `openspec/review-template.md`）；`diff -r` 與 skill_repo 的 `agent-team/` 只剩本專案設定差異
- [x] 5.2 改寫 `CLAUDE.md` 的 team-hq 段為單一 repo 設定（`frontend/`、`backend/` 子目錄分工、worktree 路徑、合併方式 `--merge`、CI 等待上限 30 分鐘、高風險樣式含 `.drone.yml`／`deploy/**`／`.claude/hooks/**`／`.claude/agents/**`、風險分級、執行驗證指令），並保留「設計同步」段；以 `grep` 確認無 `<frontend-repo>` 等未填 placeholder
- [x] 5.3 驗證寫入範圍 hook：模擬 Chika 寫 `frontend/` 下的檔被擋、寫 `backend/` 下的檔放行
- [ ] 5.4 端到端驗證 CI 閘門：對一個刻意讓 Go 測試失敗的 PR 跑 `/ship-issue` 的 7.1 步驟，確認 `handoff.md` 出現含 Drone 連結的退回紀錄；修正後 CI 全綠時流程往下
  - env-blocked：需要 Drone 上線且能回報 PR check。

## 6. 整合確認

- [ ] 6.1 從開 PR → CI 綠 → 合併 → Staging／Preview → promote → Production 全程走一次，記錄各階段耗時於 `docs/cicd.md`；`openspec validate sdp-113-drone-cicd --strict` 通過
  - env-blocked：需要以上所有外部環境。`openspec validate sdp-113-drone-cicd --strict` 已通過（2026-11-09）。

## Workflow follow-up

- 完成後依 `/issue-solver` 歸檔並開 PR，code_city 的同名 change 另行歸檔。
- 前置 issue SDP-114（mygo_v2_anon Nginx include）需先合併並部署。
