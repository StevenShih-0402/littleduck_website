# Tasks

本次沒有 `[FE]` 任務，也不改 `backend/` 的 Go 程式碼。以下 `[BE]` 任務改的是 `.woodpecker/`、`deploy/`、`.drone.yml`、`docs/`、`CLAUDE.md`、`README.md`、compose 註解。做法同 SDP-168、SDP-181：後端 Agent 的寫入範圍要暫時放寬到這些路徑。不需跑 `go test`／`gofmt`，各項驗證寫在任務內，指令在 worktree 根目錄執行。

共用工具指令：
- Woodpecker lint：`docker run --rm -v "$PWD:$PWD" -w "$PWD" woodpeckerci/woodpecker-cli:v3.20.1 lint <檔案或 .woodpecker/>`
- shellcheck：`docker run --rm -v "$PWD:/mnt" -w /mnt koalaman/shellcheck:v0.10.0 -s sh -S warning <檔案…>`

## 1. 前置

- [x] 1.1 [BE] 確認 SDP-181 已合併並歸檔（design 決策 8）：`openspec list` 不再列出 `sdp-181-staging-auth`、`git log origin/main --oneline -- .drone.yml` 有 SDP-181 的 commit、`grep -n "staging_basic_auth\|staging_gate_token" .drone.yml` 有命中。任一不符就停下回報 user，不要開始後續任務。user 若決定讓 SDP-183 先做，改以當時 `main` 的 `.drone.yml` 為準，並在 `handoff.md` 記錄「SDP-181 的 tasks 3.3 要改為修改 `.woodpecker/cd-*-staging.yaml`」— 前置順序確認

## 2. Woodpecker server 部署檔

- [x] 2.1 [BE] 新增 `deploy/woodpecker/docker-compose.yml` 與 `deploy/woodpecker/.env.example`（design 決策 1）：server、agent 映像釘 `v3.20.1`；container_name 為 `littleduck-woodpecker-server`、`littleduck-woodpecker-agent`；server 接 `default` 與外部網路 `mygo-edge`，agent 只接 `default`；兩者都不設 `ports`；資源上限同 Drone（server 0.5 CPU／256M、agent 0.25 CPU／128M）；agent 設 `WOODPECKER_MAX_WORKFLOWS: "1"` 並掛 docker.sock；`.env.example` 列出決策 1 的全部欄位，每欄附註解（`WOODPECKER_AGENT_SECRET` 用 `openssl rand -hex 32` 產生，`OPEN=false`）。開頭註解比照 `deploy/drone/docker-compose.yml`。驗證：`cp deploy/woodpecker/.env.example /tmp/wp.env` 並填入假值後，`docker compose -f deploy/woodpecker/docker-compose.yml --env-file /tmp/wp.env config` 成功、輸出中沒有 `ports:`；`grep -c "v3.20.1" deploy/woodpecker/docker-compose.yml` 為 2 — Woodpecker 部署檔
- [x] 2.2 [BE] 新增 `deploy/edge/woodpecker.conf.template`（比照 `drone.conf.template`：80 埠轉 443 加 ACME webroot；443 用 `${WOODPECKER_SITE_HOST}` 的憑證；`resolver 127.0.0.11`、`set $woodpecker_upstream littleduck-woodpecker-server:8000`、`proxy_buffering off`、`proxy_read_timeout 3600s`、`proxy_http_version 1.1`），並更新 `deploy/edge/render.sh` 用法註解的變數清單。驗證：`WOODPECKER_SITE_HOST=ci.example sh deploy/edge/render.sh /tmp/ld-wp woodpecker` 成功，產出的 `littleduck-woodpecker.conf` 不含 `${`，且 `$host`、`$woodpecker_upstream` 保留；`sh deploy/edge/test-install-sites.sh` 仍印 `ALL PASS` — Woodpecker 站台樣板
- [x] 2.3 [BE] 改寫 `docs/deployment-runbook.md` 第 1 節為「Woodpecker 初次建置」（design Migration Plan 上線前 1–6 步）。內容包含：DNS 與憑證、OAuth App（callback `/authorize`）、`/opt/woodpecker` 與 `.env`、渲染並安裝站台、啟用 repo 與 repo 設定（Trusted 只開 volumes、Allow deployments、Require approval = forks、timeout 60 分鐘）、`woodpecker-cli` 的安裝與 token 設定（`WOODPECKER_SERVER`、`WOODPECKER_TOKEN`）、記憶體量測判準（`available` < 1.4 GB 時先停 Drone runner）、授權與風險說明、容器名稱。開頭的主機名對照表把 Drone 那一列換成 `ci.stevenshih-littleduck.com`。驗證：`grep -n "ci.stevenshih-littleduck.com/authorize\|WOODPECKER_AGENT_SECRET\|Allow deployments\|volumes" docs/deployment-runbook.md` 命中上述段落 — runbook：Woodpecker 初次建置

## 3. CI workflow

- [x] 3.1 [BE] 新增 `.woodpecker/ci-backend.yaml`（design 決策 2、3）：`when: event: pull_request, branch: main`；步驟與現行 `ci-backend` 相同（gofmt、go-vet、wait-postgres、go-test、gosec、docker-build、trivy、cleanup），`services.postgres` 照舊；`GOMODCACHE` 指向 workspace 內的目錄並由 go-vet、go-test、gosec 共用；docker.sock 與 `/var/cache/woodpecker-trivy:/root/.cache` 用步驟層級 `volumes`；映像 tag 用 `ci-${CI_PIPELINE_NUMBER}`；cleanup 的 `when.status` 為 `[success, failure]`。驗證：Woodpecker lint 無 error；`woodpecker-cli exec --repo-trusted-volumes --pipeline-event pull_request .woodpecker/ci-backend.yaml`（以 woodpecker-cli 容器執行時，比照 `docs/cicd.md` 掛 docker.sock 與相同絕對路徑）成功；故意加一個未 `gofmt` 的檔案再跑一次，gofmt 步驟失敗並列出檔名（不 commit 暫時修改）— 後端 CI workflow
- [x] 3.2 [BE] 新增 `.woodpecker/ci-frontend.yaml`：觸發同 3.1；步驟為 install、lint、typecheck、test-design、audit、docker-build、trivy、cleanup，**不含**獨立的 `npm run build`（design 決策 3）；所有 Node 步驟用 `node:24-alpine`。驗證：Woodpecker lint 無 error；`woodpecker-cli exec` 成功，記錄中 `next build` 只出現在 docker-build 步驟；跑完依 `docs/cicd.md` 的指令把 `frontend/node_modules`、`.next` 的擁有者改回來 — 前端 CI workflow
- [x] 3.3 [BE] 更新 `docs/cicd.md` 的開頭、「觸發規則」、「CI」三段：說明改用 Woodpecker，CI 只在 PR 跑，check 名稱為 `ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend`；前端階段表移除 `build` 列，並註明正式建置在 docker-build 內；「在本機跑 CI」改用 `woodpecker-cli exec`；Trivy 快取路徑改為 `/var/cache/woodpecker-trivy`；新增一小段「為什麼 push 不重跑 CI」（branch protection 加合併隊列）。驗證：`grep -n "continuous-integration/drone" docs/cicd.md` 無結果；`grep -n "ci/woodpecker/pr/ci-backend" docs/cicd.md` 有結果 — cicd.md 的 CI 段落同步

## 4. CD workflow 與部署腳本

- [x] 4.1 [BE] 改 `deploy/ci/image-ref.sh`：commit sha 改讀 `${CI_COMMIT_SHA:-${DRONE_COMMIT_SHA:-$(git rev-parse HEAD)}}`，並更新註解。驗證：`CI_COMMIT_SHA=abcdef1234 sh deploy/ci/image-ref.sh` 輸出結尾為 `-abcdef1`；`DRONE_COMMIT_SHA=1234567aaa sh deploy/ci/image-ref.sh` 結尾為 `-1234567`；兩者都不設時結尾為 `git rev-parse --short=7 HEAD`；shellcheck 無輸出 — 映像 tag 讀 Woodpecker 變數
- [x] 4.2 [BE] 新增 `.woodpecker/cd-backend-staging.yaml`：`when: event: push, branch: main`，不設 `depends_on`；步驟對照**合併 SDP-181 後**的 `.drone.yml` `cd-backend-staging` 逐一搬移（build、trivy、push-ghcr、bundle、upload、deploy、smoke、cleanup），包含 SDP-181 加入的 smoke `SMOKE_BASIC_AUTH: from_secret: staging_basic_auth`；plugin 步驟的 `settings` 寫法不變。驗證：Woodpecker lint 無 error；把新舊兩份的 `from_secret` 名稱各自排序後比對（`grep -o "from_secret: [a-z_]*"`），結果相同；`grep -n "DRONE_" .woodpecker/cd-backend-staging.yaml` 無結果 — 後端 Staging workflow
- [x] 4.3 [BE] 新增 `.woodpecker/cd-frontend-staging.yaml`：做法同 4.2，對照 `cd-frontend-staging`，包含 SDP-181 的 bundle `STAGING_GATE_TOKEN: from_secret: staging_gate_token`，以及 smoke 的 `SMOKE_BASIC_AUTH`、`FRONTEND_STAGING_HOST`、預期 API `https://$$FRONTEND_STAGING_HOST`。驗證同 4.2 — 前端 Staging workflow
- [x] 4.4 [BE] 改寫 `deploy/ci/verify-commit-status.sh`（design 決策 4）：讀 GitHub combined status，篩出 context 以 `ci/woodpecker/push/` 開頭的項目，每個 context 取最新一筆；必要 context 由環境變數 `REQUIRED_CONTEXTS`（空白分隔）指定，預設為 `ci/woodpecker/push/cd-backend-staging ci/woodpecker/push/cd-frontend-staging`；缺少任一個，或任一 Woodpecker context 不是 `success`，就列出不符項目並 exit 1；API 位址可用 `GITHUB_API_URL` 覆寫（預設 `https://api.github.com`）。另新增 `deploy/ci/test-verify-commit-status.sh`：以 `python3 -m http.server` 提供 fixture JSON，驗證「全綠 → 0」「缺一個必要 context → 1」「有一個 failure → 1」「只有 drone context → 1」四種情況，最後印 `ALL PASS`。驗證：`sh deploy/ci/test-verify-commit-status.sh` 印 `ALL PASS`；暫時把「缺少必要 context」的判斷改成一律通過時，對應案例要 FAIL（不 commit）；shellcheck 無輸出 — Production 前的 push 建置檢查
- [x] 4.5 [BE] 新增 `.woodpecker/cd-backend-production.yaml`、`cd-frontend-production.yaml`：`when: event: deployment`，加 `evaluate: 'CI_PIPELINE_DEPLOY_TARGET == "production"'`；步驟同現行（verify-build 呼叫 `verify-commit-status.sh "${CI_REPO}" "${CI_COMMIT_SHA}"`、bundle、upload、deploy、smoke）；Production 部署不加 SDP-181 的 Staging secret。驗證：Woodpecker lint 無 error；`grep -n "promote\|DRONE_" .woodpecker/cd-*-production.yaml` 無結果；兩份與 `.drone.yml` 的 `from_secret` 名稱比對相同 — Production workflow
- [x] 4.6 [BE] 更新 `docs/cicd.md` 的 CD 段落與 secret 清單，以及 `docs/deployment-runbook.md` 第 2、3、5、6 節：
  - promote 全部改為 Woodpecker 的 Deploy（UI 操作與 `woodpecker-cli pipeline deploy … last production`），Target 寫錯時的行為照實記錄。
  - `verify-build` 改為檢查 `ci/woodpecker/push/*`。
  - secret 表依 design 決策 6 列出 10 個，含事件與映像限定，並附 `woodpecker-cli repo secret add` 範例；部署金鑰改為新產生的 `ci_deploy`；「目前流水線用到的 secret 只有這 8 個」更新為 10 個並改用 Woodpecker CLI。
  - 新增「從 Drone 切換」一節（design Migration Plan 的切換 7–10 步與回滾）與「GitHub branch protection」一節（必要 check、分支要最新、禁止 force push，附 `gh api` 設定指令）。

  驗證：`grep -n "drone build promote\|DRONE_RUNNER\|continuous-integration/drone" docs/cicd.md docs/deployment-runbook.md` 只剩「從 Drone 切換」一節中刻意提到的部分；`grep -n "woodpecker-cli pipeline deploy\|ci/woodpecker/push" docs/cicd.md docs/deployment-runbook.md` 有結果 — CD 文件同步

## 5. 移除 `.drone.yml` 與文字同步

- [x] 5.1 [BE] 刪除 `.drone.yml`。同步以下文字：
  - `CLAUDE.md`：「檢查指令」的說明改為 Woodpecker；「層層把關」的 PR 中 CI 改為兩個 `ci/woodpecker/pr/*` check，合併後部署改為 Woodpecker Deploy；「審查規則三分法」的 Drone CI；「高風險檔案樣式」的 `.drone.yml` 改為 `.woodpecker/**`（design 決策 7）。
  - `README.md` 的 CI/CD 段落。
  - `backend/docker-compose.prod.yml`、`backend/docker-compose.staging.yml`、`deploy/web/docker-compose.*.yml` 開頭註解的「Drone CD」改為「CD（Woodpecker）」。
  - `.gitignore` 的「Drone CD 在工作區產生的暫存檔」註解。

  驗證：`git grep -il drone -- . ':!openspec/changes/archive' ':!openspec/changes/sdp-183-woodpecker-ci' ':!openspec/specs'` 只剩 `deploy/drone/`、`deploy/edge/drone.conf.template`、`deploy/edge/render.sh`（DRONE_HOST 用法，觀察期後移除）、`deploy/ci/image-ref.sh`（fallback）、`docs/deployment-runbook.md`（切換與移除步驟），以及 `.woodpecker/` 內的 `appleboy/drone-*` 映像名 — Drone 字樣清理

## 6. 合併前的實機驗證（user 操作，Agent 只確認結果）

- [ ] 6.1 user 依 runbook「Woodpecker 初次建置」完成主機、DNS、OAuth、repo 設定與 10 個 secret。驗證：`curl -sS -o /dev/null -w '%{http_code}' https://ci.stevenshih-littleduck.com/` 為 200；`woodpecker-cli repo secret ls --repository StevenShih-0402/littleduck_website` 列出 10 個名稱，且沒有任何一個包含 `pull_request` 事件；`gh api repos/StevenShih-0402/littleduck_website/hooks --jq '.[].config.url'` 包含 `https://ci.stevenshih-littleduck.com/api/hook` — Woodpecker 上線
  - **env-blocked（2026-11-20）**：需 user 在主機、DNS、GitHub 手動操作（runbook 第 1 節），Agent 無主機與帳號權限；完成後再確認驗證指令的結果。
- [ ] 6.2 本變更的 PR 上出現 `ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend` 且都成功，沒有新的 `continuous-integration/drone/pr`；user 設好 branch protection。驗證：`gh pr checks <PR>` 的輸出；`gh api repos/StevenShih-0402/littleduck_website/branches/main/protection --jq '.required_status_checks'` 列出兩個 context 且 `strict: true` — PR CI 與合併保護
  - **env-blocked（2026-11-20）**：PR 尚未建立（由 `/issue-solver` 開 PR），且依賴 6.1 的 Woodpecker 上線；branch protection 由 user 依 runbook 第 9 節設定。

## Workflow follow-up

- **合併時**（user，依 runbook「從 Drone 切換」）：
  - 在 Drone 停用 repo，確認 GitHub webhook 只剩 Woodpecker，執行 `docker compose stop drone-runner`。
  - 合併 PR，確認 push pipeline 只有兩個 `cd-*-staging` 且都成功（含 SDP-181 的帳密冒煙）。
  - 在 Staging 人工確認後 Deploy `production`，確認兩個 `cd-*-production` 成功。
  - 再以 Target `prod` 觸發一次，確認沒有部署發生，並把 Woodpecker 的顯示方式寫進 `docs/cicd.md`。
- **合併後量測**：依 `docs/cicd.md`「各階段耗時」重新量測 PR CI、push 到 Staging、Deploy 到 Production、全程，以新表取代 Drone 的數字（舊數字留一行作為對照），另開小 PR 提交。
- **歸檔時**：直接編輯 `openspec/specs/ci-server/spec.md`、`agent-ci-gate/spec.md`、`frontend-delivery/spec.md` 的 Purpose，把 Drone 改為 Woodpecker（delta 改不到 Purpose）。
- **觀察期後（建議 7 天，另開小 PR，高風險）**：
  - repo：刪除 `deploy/drone/`、`deploy/edge/drone.conf.template`，移除 `render.sh` 用法中的 `DRONE_HOST` 與 `image-ref.sh` 的 `DRONE_COMMIT_SHA` fallback，runbook 的切換一節濃縮成歷史紀錄。
  - 主機：`docker compose down -v`、刪除 `/opt/drone`、`/opt/edge-sites/littleduck-drone.conf`（之後 `nginx -t` 並 reload）、`/var/cache/drone-trivy`；從 `authorized_keys` 移除 `drone-deploy`。
  - 外部帳號：撤銷 Drone 用的 GHCR PAT 與 OAuth App，`certbot delete --cert-name drone.stevenshih-littleduck.com`，刪除 `drone.` 的 DNS 記錄。
- 本變更命中高風險檔案樣式（`.drone.yml`、`.woodpecker/**`、`deploy/**`），PR 中要等 user 看過 diff 後回覆「合併」。
