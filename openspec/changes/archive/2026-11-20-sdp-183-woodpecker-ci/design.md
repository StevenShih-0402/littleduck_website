# Design

## Context

動機見 proposal.md「Why」，要達成的行為見 `specs/`。以下是現況、與理想流程的逐階段比對，以及會影響做法的限制。

### 現行流程（`.drone.yml`、`docs/cicd.md`、`deploy/`）

- Drone 2.26 server 加上 docker runner 1.8.4，跑在 Hetzner CX23（4 GB）的 `/opt/drone`，經 mygo 的 Nginx 以 `drone.stevenshih-littleduck.com` 對外（`deploy/drone/`、`deploy/edge/drone.conf.template`）。runner 的 `DRONE_RUNNER_CAPACITY=1`，同一次建置的 pipeline 依序執行。
- 6 條 pipeline：
  - `ci-backend`、`ci-frontend`：觸發是 PR（目標 `main`）加上推送到 `main`。
  - `cd-*-staging`：觸發是推送到 `main`，`depends_on` 同一次建置的 `ci-*`。
  - `cd-*-production`：觸發是 `promote` 到 `production`。
- GitHub 上只有一個彙總的 check，`continuous-integration/drone/pr`。push 建置的彙總 context 是 `continuous-integration/drone/push`，`verify-commit-status.sh` 在 Production 部署前查的就是它。
- GitHub repo 的 webhook 目前只有一個，指向 `https://drone.stevenshih-littleduck.com/hook`，事件是 create、delete、deployment、pull_request、push。`main` **沒有 branch protection**（`gh api …/branches/main/protection` 回 404），所以「CI 通過才能合併」目前只靠 `/ship-issue` 的流程約束。
- agent 流程（`.claude/commands/ship-issue.md`、`issue-integrator`）用 `gh pr checks --watch` 等 CI，不寫死 check 名稱。寫死 Drone 名稱的只有 `CLAUDE.md`、`docs/`、`openspec/specs/` 的文字。
- 實測耗時（`docs/cicd.md`，2026-11-15）：PR CI 7 分 16 秒；合併後的 push 建置（兩條 CI 加兩條 Staging）8 分 15 秒；兩條 Production 各約 20 秒。

### 與理想流程的逐階段比對

| 理想階段 | 現況 | 判定 |
|----------|------|------|
| PR → CI 檢查 | `ci-backend`：gofmt → vet → test（PostgreSQL service）→ gosec → docker build → trivy。`ci-frontend`：npm ci → lint → typecheck → test:design → **npm run build** → audit → **docker build**（內含 `npm ci` 加 `next build`）→ trivy | 符合。**多餘**：前端的 `next build` 跑兩次，Dockerfile 的 build stage 已經跑過一次 |
| → 其餘檢查 | `/ship-issue`：`/code-review`、高風險時跑 `/security-review`、高風險時等 user 回覆「合併」 | 符合，不在 CI server 內，本變更不動 |
| → 合併到 main | 合併隊列：rebase 到最新 `main`，重跑 PR CI，全綠才 `--merge`。GitHub 端沒有保護 | **缺口**：手動合併或直接推送 `main` 不受任何檢查 |
| 用 main 產生 image | push 到 `main` 時**先重跑整套 `ci-*`**（同一份樹在 PR 上剛驗過；CI 還會建一次映像、掃一次），接著 `cd-*-staging` **再建一次**映像、再掃一次，才推送 | **多餘**：同一 commit 的 CI 跑兩次（PR、push），映像建三次（PR、push CI、CD） |
| → 推送到 Staging | `cd-*-staging`：推 GHCR → bundle → scp → `remote-deploy.sh`（120 秒等 healthy）→ 冒煙 → cleanup | 符合 |
| → 人工在 Staging 確認 | user 自己開 Staging 看。SDP-181 會替 Staging 加上身分驗證 | 符合。Staging 不在 CI 控制內，也不需要 |
| → 同意後推送到 Production | Drone 的 promote（UI 的 Deploy 或 `drone build promote`）→ `verify-build` 查 push 建置的彙總 status → 用**同一個** `<version>-<sha>` 部署，不重建 | 符合（已經是 promote 映像而非重建）。要對到 Woodpecker 的機制，見決策 4 |

結論：多出來的工作集中在 push 到 `main` 這一段，以及前端 CI 的重複建置。人工核准這一關本來就有，只要換成 Woodpecker 的對應機制。

### Woodpecker 的相關事實（v3.20.1，2026-10-29 發布；已對照官方文件與 v3.20.1 原始碼）

- 設定檔放在 `.woodpecker/*.yaml`，一個檔案就是一個 workflow，GitHub status context 是 `ci/woodpecker/<event>/<workflow>`（server 預設 `WOODPECKER_STATUS_CONTEXT_FORMAT`）。workflow 之間用 `depends_on`，步驟用 `when`（`event`、`branch`、`status`、`evaluate` 等）。**沒有** Drone 的 `trigger.target`，也沒有 `when.environment`：deployment 的目標要用 `evaluate: 'CI_PIPELINE_DEPLOY_TARGET == "production"'` 判斷。
- 事件值有 `push`、`pull_request`、`pull_request_closed`、`pull_request_metadata`、`tag`、`release`、`deployment`、`cron`、`manual`。
- **人工核准 → deployment 事件**：repo 設定開啟「Allow deployments」後，成功的 pipeline 頁面會出現 Deploy 按鈕，可以填 Target、Task 與參數，結果是新的 pipeline，`CI_PIPELINE_EVENT=deployment`、`CI_PIPELINE_DEPLOY_TARGET=<target>`、`CI_PIPELINE_PARENT=<原 pipeline 編號>`。CLI 是 `woodpecker-cli pipeline deploy <repo> <pipeline|last> <environment>`。按鈕只對成功的 pipeline 顯示，**但 API（`POST /repos/{id}/pipelines/{n}?event=deployment`）不檢查原 pipeline 是否成功**，所以部署前仍要自己驗證。官方也提醒，只能 deployment 拿到的 secret，任何有 push 權限的人都能觸發；本 repo 只有擁有者有 push 權限。
- 祕密用 `from_secret`，寫在 `environment` 或 plugin 的 `settings` 底下。secret 可以限定事件（`push`、`deployment`…）與 plugin 映像。預設不開放給 `pull_request`。CLI 是 `woodpecker-cli repo secret add --repository <owner/repo> --name … --value @file --event push --event deployment`。
- 變數替換同樣用 `$$` 跳脫，對應關係：`DRONE_COMMIT_SHA` 改 `CI_COMMIT_SHA`，`DRONE_BUILD_NUMBER` 改 `CI_PIPELINE_NUMBER`，`DRONE_REPO` 改 `CI_REPO`。
- 步驟掛主機路徑寫成 `volumes: ["/var/run/docker.sock:/var/run/docker.sock"]`，需要 repo 的 Trusted 設定開啟 **volumes**。Trusted 分成 network、volumes、security 三項，只有管理員能設定。Drone 那種 workflow 層級的 `volumes: temp/host` 寫法不存在，同一個 workflow 的步驟本來就共用 workspace。
- Approval 模式有 `none`、`forks`（預設）、`pull_requests`、`all_events`，另外可以指定免核准的使用者。
- GitHub forge：用 OAuth App，callback 是 `https://<host>/authorize`，環境變數為 `WOODPECKER_GITHUB=true`、`WOODPECKER_GITHUB_CLIENT`、`WOODPECKER_GITHUB_SECRET`。官方文件明講**不要用 GitHub App**。啟用 repo 時 server 會自動在 GitHub 建立 webhook，指向 `<WOODPECKER_HOST>/api/hook`；停用時會移除。`WOODPECKER_OPEN=false` 時只有 `WOODPECKER_ADMIN` 列出的帳號能第一次登入（建立帳號）。
- agent 用 `WOODPECKER_SERVER=<server>:9000` 的 gRPC 連線，`WOODPECKER_AGENT_SECRET` 與 server 共用。`WOODPECKER_MAX_WORKFLOWS` 預設 1，要掛主機的 docker.sock。server 預設用 SQLite。映像 `woodpeckerci/woodpecker-server:v3.20.1`、`woodpecker-agent:v3.20.1` 已在 Docker Hub 確認存在。官方沒有給記憶體需求；server 與 agent 都是 Go 程式，閒置時約數十 MB，實際用量由建置容器決定（建置容器不算在 agent 的上限內）。
- `appleboy/drone-scp`、`appleboy/drone-ssh` 讀的是 `PLUGIN_*` 環境變數，Woodpecker 會把 `settings` 轉成這些變數傳進去，所以可以照用。
- 啟用 repo 後，若推送的 commit 沒有 `.woodpecker/` 設定，Woodpecker 會建立一個錯誤的 pipeline，並對該 commit 回報失敗 status。這會影響切換時的時序，見 Migration Plan。

### 進行中的 SDP-181（`openspec/changes/sdp-181-staging-auth/`，尚未實作）

這個變更會改 `.drone.yml` 的 `cd-backend-staging`、`cd-frontend-staging`：smoke 加 `SMOKE_BASIC_AUTH: from_secret: staging_basic_auth`，前端 bundle 加 `STAGING_GATE_TOKEN: from_secret: staging_gate_token`，前端 smoke 的預期 API 改成 `https://$$FRONTEND_STAGING_HOST`、拿掉 `API_STAGING_HOST`。另外它會新增兩個 Drone secret，並改 `deploy/edge/*-staging.conf.template`、`install-sites.sh`、`deploy/ci/smoke-*.sh`、`frontend-bundle.sh`。它的 delta spec 修改的是 `frontend-delivery` 的「自動部署到前端 Staging 並冒煙測試」與 `backend-delivery` 的「Staging 冒煙測試」，本變更**刻意不修改這兩條需求**，避免兩邊歸檔時衝突。

## Goals / Non-Goals

**Goals:**
- 從 Drone 換到 Woodpecker v3.20.1。流水線的行為（檢查門檻、映像 tag、部署腳本、冒煙、回滾）不退化。
- 每個 commit 只跑一次 CI（在 PR 上）、只建一次上線映像（在 CD 上），Production 照舊直接用 Staging 驗過的映像。
- 人工核准這一關保留，而且不能繞過「push 建置全綠」的檢查。
- 切換期間 Production 不停機，任一階段都能退回 Drone。

**Non-Goals:**
- 不改 `deploy/ci/*.sh` 的部署邏輯（`remote-deploy.sh`、bundle、smoke 只改變數名稱）、不改 Nginx 的 Staging／Production 站台、不改容器與 `.env`。
- 不加 path filter（只改文件的 PR 也照跑完整 CI，理由見決策 3）。
- 不換掉 Trivy、gosec、npm audit 等檢查工具，也不調整門檻。
- 不處理上層 SDP-175 中其他專案（如 mygo_v2_anon 的 Jenkins）的 CI/CD。
- 不改 `.claude/commands/ship-issue.md` 與角色 skill：它們用 `gh pr checks`，與 CI 名稱無關。

## Decisions

### 1. Woodpecker 部署方式：同主機 docker compose、SQLite、新子網域 `ci.`

`deploy/woodpecker/docker-compose.yml`（project `littleduck-woodpecker`）的內容：

- `woodpecker-server`：`woodpeckerci/woodpecker-server:v3.20.1`，container_name `littleduck-woodpecker-server`，接 `default` 與 `mygo-edge` 兩個網路，資料放 named volume `woodpecker_data:/var/lib/woodpecker`（SQLite），不開埠。上限 `cpus: 0.5`、`memory: 256M`，同 Drone server。
- `woodpecker-agent`：`woodpeckerci/woodpecker-agent:v3.20.1`，container_name `littleduck-woodpecker-agent`，只接 `default` 網路，掛 docker.sock，設定 `WOODPECKER_SERVER=woodpecker-server:9000`、`WOODPECKER_MAX_WORKFLOWS=1`、`WOODPECKER_HOSTNAME=hetzner-cx23`，上限 `cpus: 0.25`、`memory: 128M`，同 Drone runner。
- `.env.example` 的欄位：`WOODPECKER_HOST=https://ci.stevenshih-littleduck.com`、`WOODPECKER_GITHUB=true`、`WOODPECKER_GITHUB_CLIENT`、`WOODPECKER_GITHUB_SECRET`、`WOODPECKER_AGENT_SECRET`（`openssl rand -hex 32`）、`WOODPECKER_OPEN=false`、`WOODPECKER_ADMIN=StevenShih-0402`、`WOODPECKER_REPO_OWNERS=StevenShih-0402`（只有這個帳號的 repo 可以啟用）。
- Nginx 站台 `deploy/edge/woodpecker.conf.template` 比照 `drone.conf.template`，樣板變數用 `${WOODPECKER_SITE_HOST}`，upstream 是 `littleduck-woodpecker-server:8000`，保留 resolver 延後解析、`proxy_buffering off`（記錄串流）與長 timeout。

理由：與 Drone 的部署模型完全相同（同主機、只經 mygo Nginx 對外、資源上限），runbook 的步驟大多可以沿用。子網域用中性的 `ci.`，不帶產品名，日後再換 CI 時不必改 DNS 與憑證。SQLite 對單一 repo、單一使用者夠用，不必多跑一個資料庫容器。

替代方案：
- **沿用 `drone.` 子網域**：切換期間兩套要同時存在，而且名稱會誤導，否決。
- **`woodpecker.` 子網域**：可行，但綁產品名，否決。這只是命名偏好，user 想改可以直接換，影響範圍只有 DNS、憑證、`.env` 與站台樣板。
- **PostgreSQL 當資料庫**：多一個常駐容器，在 4 GB 主機上沒必要，否決。
- **Woodpecker 的 GitHub App 模式**：官方文件說目前不能正常運作，否決。

### 2. 流水線檔案：6 個 workflow，名稱與 Drone 一一對應

`.woodpecker/ci-backend.yaml`、`ci-frontend.yaml`、`cd-backend-staging.yaml`、`cd-frontend-staging.yaml`、`cd-backend-production.yaml`、`cd-frontend-production.yaml`。各檔的 `when`：

| workflow | `when` | 內容 |
|----------|--------|------|
| `ci-backend` | `event: pull_request`，`branch: main`（PR 的目標分支） | 同現行，去掉 Drone 專屬寫法（見下） |
| `ci-frontend` | 同上 | 現行步驟去掉 `build`（決策 3） |
| `cd-backend-staging` | `event: push`，`branch: main` | 同現行，不再 `depends_on` CI（決策 3） |
| `cd-frontend-staging` | 同上 | 同上 |
| `cd-backend-production` | `event: deployment`，`evaluate: 'CI_PIPELINE_DEPLOY_TARGET == "production"'` | 同現行（verify → bundle → upload → deploy → smoke） |
| `cd-frontend-production` | 同上 | 同上 |

語法對應：
- `services: postgres` 照舊（service 名稱就是主機名）。`wait-postgres` 步驟保留。
- Go 模組快取：Drone 用 `temp` volume，Woodpecker 改設 `GOMODCACHE: /woodpecker/src/.gomodcache`（放在 workspace 裡，同一個 workflow 的步驟共用），不需要額外的 volume。實作時以 `$CI_WORKSPACE` 的實際路徑為準。
- docker.sock 與 Trivy 快取：步驟層級寫 `volumes: [/var/run/docker.sock:/var/run/docker.sock]`、`[/var/cache/woodpecker-trivy:/root/.cache]`。快取目錄改名，避免與 Drone 並行期間互相干擾；Drone 移除後再刪舊的 `/var/cache/drone-trivy`。
- `${DRONE_BUILD_NUMBER}` 改 `${CI_PIPELINE_NUMBER}`，`${DRONE_REPO}`、`${DRONE_COMMIT_SHA}` 改 `${CI_REPO}`、`${CI_COMMIT_SHA}`。`image-ref.sh` 改成 `sha=${CI_COMMIT_SHA:-${DRONE_COMMIT_SHA:-$(git rev-parse HEAD)}}`，保留 Drone 的 fallback，觀察期後移除 Drone 時再拿掉（Workflow follow-up）。
- `cleanup` 步驟的 `when: status: [success, failure]` 照舊。
- 每個 workflow 開頭用註解說明觸發條件，並指向 `docs/cicd.md`。

理由：名稱一一對應，`docs/cicd.md` 的階段表、runbook 與使用者的認知都可以延續。每個 workflow 在 GitHub 上各有一個 check，失敗時一眼就看得出是哪一端。

替代方案：
- **staging 與 production 合成一個 `cd-backend.yaml`，用步驟層級的 `when` 分流**：兩者的 secret 與目標目錄不同，一半的步驟要寫兩份 `when`，可讀性變差，省下的行數不多，否決。
- **把 `upload`、`deploy` 兩個 plugin 步驟合成一個 `commands` 步驟（自己 `scp` 加 `ssh`）**：可以少一個映像，但要自己處理 known_hosts 與金鑰檔權限，plugin 已經處理好了，否決。

### 3. 簡化：CI 只在 PR 跑，push 到 main 直接進 CD；前端拿掉重複的 `next build`

- `ci-*` 只在 `pull_request` 觸發。`cd-*-staging` 在 push 到 `main` 時直接開始，不 `depends_on` CI。
- `main` 的品質靠兩道關卡保證（理想流程本來就是「CI 通過才合併」）：
  1. **GitHub branch protection**（新增，user 手動設定）：`main` 要求 PR、要求 `ci/woodpecker/pr/ci-backend` 與 `ci/woodpecker/pr/ci-frontend` 通過、要求分支是最新（Require branches to be up to date）、禁止 force push。PR 分支是最新且 CI 通過時，合併後的樹就等於 CI 驗過的樹。`--merge` 產生的 merge commit 樹與 PR head 相同，所以 push 時重跑 CI 不會多驗到任何東西。
  2. `/ship-issue` 的合併隊列（既有）：rebase 後重跑 PR CI。
- CD Staging 的 `docker build` 本身就會編譯（Go 的 `go build`、前端的 `next build`），接著 Trivy 會再掃要上線的映像。所以即使有人繞過保護，壞到不能建置的程式也上不了 Staging。
- `ci-frontend` 拿掉 `build` 步驟。`docker-build` 步驟的 Dockerfile 已經跑 `npm ci && npm run build`，lint 與 typecheck 照樣排在它前面。`npm ci` 步驟保留，lint、typecheck、test:design、audit 都需要它。
- 預期效果（合併後實測，寫進 `docs/cicd.md`，見 Workflow follow-up）：
  - push 建置少掉兩條 CI。現在合併後要 8 分 15 秒，扣掉 CI 只剩兩條 CD Staging。
  - PR CI 少一次 `next build`。
  - 每個 commit 的映像建置次數從 3 次（PR CI、push CI、CD）降到 2 次（PR CI 驗 Dockerfile、CD 產出上線映像）。

替代方案：
- **push 時保留 CI（不改流程，只換 CI server）**：最保守，但正是 issue 要找出的重複步驟，否決。
- **PR CI 建出的映像直接推送，CD 只做 retag（真正的「建一次」）**：PR 映像建在 PR head 上，與合併後的 commit sha 不同。合併隊列還會 rebase，產生新的 sha，tag 對不上。而且 PR 建置拿不到 GHCR secret（不能給 PR 寫入權限），否決。
- **path filter（`when.path`，只改 `backend/` 時跳過前端 CI）**：被跳過的 workflow 不會回報 status，branch protection 的必要 check 會一直卡在等待。要另外加「永遠成功」的替身 workflow，增加複雜度，否決。之後真的需要再另開 issue。
- **CD Staging 不再掃 Trivy（PR 已掃過）**：CD 建出的是另一個映像，base image 可能已經更新，而它才是要上線的東西，所以保留。

### 4. 人工核准：Woodpecker deployment 事件取代 Drone promote

- repo 設定開啟 **Allow deployments**。user 在 Staging 確認後，到 Woodpecker 該 commit 的 push pipeline 頁按 **Deploy**，Target 填 `production`（Task 留空）。或者執行：
  ```bash
  woodpecker-cli pipeline deploy StevenShih-0402/littleduck_website <pipeline 編號|last> production
  ```
  `last` 會選 `main` 上最近一次成功的 push pipeline。
- `cd-*-production` 用 `evaluate: 'CI_PIPELINE_DEPLOY_TARGET == "production"'` 過濾，打錯目標（例如 `prod`）就不會執行任何部署 workflow。這時 deployment pipeline 沒有任何 workflow 符合，部署不會發生。Woodpecker 對這種 pipeline 的顯示方式（略過或錯誤），合併後實機測試時確認並寫進 `docs/cicd.md`（Workflow follow-up）。
- 第一個步驟 `verify-build` 保留，改成驗證 Woodpecker 的 push status，因為 API 不檢查原 pipeline 是否成功（見 Context）。`verify-commit-status.sh` 改為：
  - 取該 commit 所有 context 以 `ci/woodpecker/push/` 開頭的 status，每個 context 取最新一筆。
  - 至少要有 `ci/woodpecker/push/cd-backend-staging` 與 `ci/woodpecker/push/cd-frontend-staging` 兩個，而且全部是 `success`，否則失敗並列出不符的 context。
  - 必要 context 清單以參數或環境變數傳入，預設是上面兩個。腳本仍然用公開 API、不需要 token（repo 是 public）。
- 回滾：在 Woodpecker 對前一個良好的 push pipeline 再按一次 Deploy 到 `production`，效果等同 Drone 的「對前一個建置重新 promote」。手動改 `image.txt` 的方式也保留。

理由：這是 Woodpecker 裡與 Drone promote 對應的機制（Drone 的 promote 本來就是 deployment 事件的另一個名字）。按鈕只出現在成功的 pipeline 上，`verify-build` 再補上 API 繞過的情況，兩道一起確保「只有 Staging 驗過的 commit 能上 Production」。

替代方案：
- **用 `manual` 事件加參數**（Run pipeline 選分支）：manual 會從分支 HEAD 開新 pipeline，不是針對某個已驗證的 pipeline，容易部署到還沒上 Staging 的 commit，否決。
- **同一個 push pipeline 裡放一個「等待核准」的步驟**：Woodpecker 的 approval 只作用在整條 pipeline 開始前（例如 fork PR），沒有步驟層級的人工閘門。硬做會讓 agent 長時間佔用，而 `MAX_WORKFLOWS=1` 會擋住其他建置，否決。
- **GitHub Environments 的 required reviewers 加 GitHub Deployments API**：Woodpecker 會收 GitHub 的 deployment webhook，可以做到，但要多維護一套 GitHub 端設定與 token，對單人專案過度設計，否決。

### 5. 安全設定對齊 Drone 的三道防線

| Drone | Woodpecker |
|-------|------------|
| repo 設為 Trusted | Trusted 只開 **volumes**（network、security 關閉） |
| `--ignore-forks` | Require approval 設為 **forks**（預設值；fork PR 要擁有者核准才跑） |
| secret 不勾 Allow Pull Requests | 每個 secret 只開 `push`、`deployment` 事件（不加 `pull_request`） |
| `DRONE_USER_FILTER`、`DRONE_REGISTRATION_CLOSED` | `WOODPECKER_OPEN=false`、`WOODPECKER_ADMIN`、`WOODPECKER_REPO_OWNERS` |

另外加強一項：`deploy_ssh_key` 與 `hetzner_host` 用 `--image` 限定在 `appleboy/drone-scp`、`appleboy/drone-ssh` 兩個 plugin。限定映像的 secret 不能被 `commands` 步驟讀取，可以降低外洩面。`hetzner_host` 只在這兩個 plugin 用到，可以一起限定。其他 secret（`ghcr_*`、各環境 host、SDP-181 的兩個）有 `commands` 步驟在用，不限定映像。

### 6. 祕密搬移：重建、不匯出；部署金鑰換新

Drone 的 secret 無法讀回，只能從來源重建。來源整理成下表，寫進 runbook：

| secret | 來源 | 事件 | 限定映像 |
|--------|------|------|----------|
| `ghcr_username`、`ghcr_token` | GitHub 帳號、具 `write:packages` 的 PAT（建議另發一把給 Woodpecker，切換完撤銷 Drone 用的那把） | push | — |
| `hetzner_host` | 主機 IP | push、deployment | scp、ssh plugin |
| `deploy_ssh_key` | **新產生**的 `~/.ssh/ci_deploy`（ed25519，comment `woodpecker-deploy`），`--value @~/.ssh/ci_deploy` | push、deployment | scp、ssh plugin |
| `api_host`、`frontend_host` | 固定網域 | deployment | — |
| `api_staging_host`、`frontend_staging_host` | 固定網域 | push | — |
| `staging_basic_auth`、`staging_gate_token`（SDP-181） | 與 Drone 建立時相同的值（user 自行保管；遺失時依 SDP-181 runbook 輪替） | push | — |

- 事件依實際用途收窄：Production 的 host 只給 `deployment`，Staging 的 host 只給 `push`，所以 PR 與錯誤的事件都拿不到。
- 換新部署金鑰的理由：舊私鑰存在 Drone 的資料庫（`drone_data` volume）。換新後可以在 Drone 移除時撤銷舊的公鑰，兩套系統的權限可以分開收回。並行期間兩把公鑰同時存在 `authorized_keys`。

### 7. CLAUDE.md、文件與高風險樣式

- `CLAUDE.md`：
  - 「檢查指令」的「與 Drone CI 的 `ci-backend`、`ci-frontend` 相同」改為 Woodpecker。前端「建置」一欄維持 `npm run build`，本機檢查仍需要它，CI 是在映像建置內跑。
  - 「層層把關」的 PR 中 CI 改成等 `ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend`，「合併後的部署」改為 Woodpecker 的 Deploy。
  - 「高風險檔案樣式」把 `.drone.yml` 換成 `.woodpecker/**`（`.drone.yml` 在本 PR 就刪除，不必並列）。
  - 「審查規則三分法」的「Drone CI」改為「Woodpecker CI」。
- `docs/cicd.md` 整份改寫成 Woodpecker 版：觸發規則、各 workflow 階段表、簡化說明、Deploy 操作、secret 清單（含事件與映像限定）、本機執行（`woodpecker-cli exec`）、耗時表（新量測）。`docs/deployment-runbook.md` 的第 1 節改為「Woodpecker 初次建置」，第 2、3、5、6 節的 promote、secret、金鑰改為 Woodpecker 版，另外新增「從 Drone 切換」與「branch protection」兩節。
- `openspec/specs/` 的 Purpose 段落（`ci-server`、`agent-ci-gate`、`frontend-delivery`）提到 Drone，delta 改不到 Purpose，歸檔時直接編輯主規格（Workflow follow-up）。

### 8. 與 SDP-181 的先後：SDP-181 先在 Drone 合併，本變更再搬過去

- **順序**：SDP-181 先實作並合併（仍走 Drone），本變更的實作在 SDP-181 歸檔之後才開始（tasks 1.1 是前置檢查）。
- **理由**：
  - SDP-181 範圍小、提案已完成，能先讓 Staging 有身分驗證，正好補上理想流程中「人工在 Staging 確認」這一段的前提。
  - 本變更是整條流水線換底。如果先合併本變更，SDP-181 的 tasks 3.3（改 `.drone.yml`）就得改成改 `.woodpecker/`，而它的提案、design、驗證步驟都以 Drone 為準，要重寫。反過來，本變更只要把 SDP-181 合併後的 `.drone.yml` 照抄成 Woodpecker 語法即可，改動最少。
  - 兩者都碰 `deploy/**` 與 `docs/cicd.md`、`docs/deployment-runbook.md`。依序做，文件衝突最少。
- **搬移內容**（tasks 4.2、4.3、4.6 明列）：
  - `cd-backend-staging` 的 smoke 加 `SMOKE_BASIC_AUTH: from_secret: staging_basic_auth`。
  - `cd-frontend-staging` 的 bundle 加 `STAGING_GATE_TOKEN: from_secret: staging_gate_token`；smoke 加 `SMOKE_BASIC_AUTH` 與 `FRONTEND_STAGING_HOST`，預期 API 參數用 `https://$$FRONTEND_STAGING_HOST`。
  - 兩個 secret 都只開 `push` 事件。
  - runbook 的 secret 數量以 10 個計。
  - 實作時以**合併後 `main` 上的 `.drone.yml` 實際內容**為準，逐條 pipeline 比對，不照本段文字抄。
- **如果 user 決定讓 SDP-183 先做**（不建議）：本變更照樣以當時 `main` 上的 `.drone.yml` 為準（沒有 SDP-181 的內容），並在 SDP-181 的 `tasks.md` 註記「3.3 改為修改 `.woodpecker/cd-*-staging.yaml`、secret 改建在 Woodpecker（只開 push 事件）」。這屬於 SDP-181 的範圍，本變更不直接修改它。
- **規格**：本變更不修改 SDP-181 會改的兩條需求（見 Context），兩邊歸檔時不衝突。

### 9. 切換方式：短暫並行驗證 PR CI，合併時一次切換 CD

細節見 Migration Plan，重點如下：

- Woodpecker 先上線、啟用 repo，在本變更的 PR 上驗證 CI 的兩個 workflow，這時 Drone 仍負責 `main` 的 CD。
- 本 PR 同時刪除 `.drone.yml`（tasks 5.1）。Drone 在 PR 上找不到設定就不會建置，所以不會有兩套 CI 同時搶記憶體。
- 合併前先在 Drone 停用 repo（會移除 Drone 的 webhook），合併後的 push 只有 Woodpecker 接手 CD Staging。接著用 Deploy 驗證 Production 路徑。
- Drone 的容器保留一段觀察期（建議 7 天，期間停止 runner 以釋放記憶體），沒問題再移除（Workflow follow-up，另一個小 PR）。

替代方案：
- **長期雙跑**：兩套 CD 都會部署同一個環境，會重複部署、互相覆蓋。4 GB 主機也撐不住兩條建置同時跑，否決。
- **直接停 Drone 後再裝 Woodpecker**：中間會有一段沒有 CI/CD 的時間，而且 PR CI 沒機會先在真實環境驗證，否決。

## 執行驗證觀察點

本變更不改 `backend/`、`frontend/` 的程式碼，CLAUDE.md 的通用觀察點照跑，只確認沒有退化。本變更特有的驗證：

- 本機語法與行為（合併前，Teshima 可跑）：
  - `docker run --rm -v "$PWD:$PWD" -w "$PWD" woodpeckerci/woodpecker-cli:v3.20.1 lint .woodpecker/`：6 個檔案都沒有 error。
  - 用 `woodpecker-cli exec` 在本機跑 `ci-backend`、`ci-frontend`，都成功。CD workflow 需要 secret 與主機，不在本機跑。
  - `sh deploy/ci/verify-commit-status.sh <repo> <sha>`：對一個已知全綠的 commit 回 0；對一個沒有 Woodpecker status 的舊 commit 回非 0，訊息列出缺少的 context。
  - `grep -rn "DRONE_\|drone" .woodpecker/` 除了 plugin 映像名稱（`appleboy/drone-*`）之外沒有結果。
- 實機（由 user 依 runbook 執行，結果記在 `docs/cicd.md`）：
  - 本 PR 上出現 `ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend` 兩個 check，都成功，且沒有 `continuous-integration/drone/pr`。
  - 合併後的 push pipeline 只有兩個 `cd-*-staging` workflow，都成功，Staging 冒煙通過（含 SDP-181 的帳密檢查）。
  - 對該 pipeline Deploy `production` 後，兩個 `cd-*-production` 成功，`https://stevenshih-littleduck.com/` 回 200，`/runtime-config.js` 指向正式後端。
  - 以 Target `prod` 觸發 Deploy，沒有任何部署 workflow 執行。
  - 用一個 fork PR 或第二個帳號測試：Woodpecker 停在等待核准。這項可選，沒有第二個帳號時以 repo 設定截圖代替。
  - `gh api repos/StevenShih-0402/littleduck_website/hooks` 只剩 Woodpecker 的 `/api/hook`。

## 變異測試目標

- `deploy/ci/verify-commit-status.sh`：「缺少必要 context」與「任一 context 非 success」兩個判斷分支。

（本專案尚未設定變異測試指令，Teshima 會記為略過。）

## Risks / Trade-offs

- [push 到 `main` 不再重跑 CI，繞過 PR 的推送會帶著沒驗過的程式上 Staging] → 用 branch protection 封住直接推送與未過 CI 的合併（runbook 新增一節，屬於 user 的手動步驟）。CD 的 `docker build` 與 Trivy 仍會擋下無法建置與有高風險弱點的映像。Production 還要人工 Deploy，Staging 是最後一道人工確認。
- [branch protection 要求「分支是最新」，PR 變多時要常 rebase] → 目前一次只處理一張 issue，`/ship-issue` 的合併隊列本來就會 rebase，實際上沒有新增負擔。
- [Woodpecker 的 deployment API 不檢查原 pipeline 是否成功] → `verify-build` 檢查 GitHub 上所有 `ci/woodpecker/push/*` 都是 success（決策 4）。
- [並行期間兩套 server 與 runner 的常駐記憶體加起來約多 384M（上限）] → runbook 第 3 節的量測：Production 全滿時 `available` 約剩 1067M。並行只持續到合併當天，Drone runner 在停用 repo 後立刻 `docker compose stop drone-runner`，只留 server（256M）在觀察期內，確認時不需要它也可以一併停掉。上線前先量 `free -m`，`available` 小於 1.4 GB 就先停 Drone runner 再起 Woodpecker agent。
- [啟用 repo 後到本 PR 合併之間，其他 commit 推到 `main` 時 Woodpecker 找不到 `.woodpecker/`，會回報錯誤 status] → 這段時間不合併其他 PR（SDP-181 已先合併）。Drone 的 `verify-commit-status.sh` 只看 `continuous-integration/drone/push`，不受影響。
- [`ci/woodpecker/push/*` 的 status 名稱與數量變了，Production 的檢查漏掉某個 workflow] → 必要 context 明列在腳本裡，少一個就失敗；新增 workflow 時要同步清單（寫進 `docs/cicd.md`）。
- [`appleboy/drone-*` plugin 名稱帶 drone，日後可能不再維護] → 目前 1.8.x 仍在更新，而且只讀 `PLUGIN_*` 環境變數，與 CI server 無關。要換時是另一張 issue。
- [Woodpecker 版本更新快（約每月一個 minor），v3.20.0 就曾有意外的破壞性變更] → 映像釘在 `v3.20.1`，升級時看 release notes 後手動改 compose。
- [secret 重建時打錯值] → 合併後的第一次 push 與 Deploy 就會用到全部 secret，冒煙會失敗而不是靜默出錯。Production 失敗時依回滾步驟處理（Production 的容器在部署失敗前不會被換掉，`remote-deploy.sh` 先 pull 再 up）。
- [`.claude/` 的角色 skill 由 littleduck-skills 安裝，裡面若有 Drone 字樣，重跑 installer 會蓋回來] → 已確認 `.claude/` 內沒有寫死 Drone（只有 `gh pr checks`），不受影響。

## Migration Plan

**前置（SDP-181 合併後、本變更實作前）**：確認 SDP-181 已合併並歸檔，Drone 上的 10 個 secret 都在使用中。

**上線前（user 在主機、DNS、GitHub 手動完成，寫進 runbook 的「Woodpecker 初次建置」與「從 Drone 切換」）**：

1. 量測 `free -m`，必要時先停 Drone runner 以外的非必要容器。
2. DNS：新增 `ci.stevenshih-littleduck.com` 的 A 記錄（DNS only），用 certbot webroot 核發憑證。
3. GitHub：新建 OAuth App，Homepage `https://ci.stevenshih-littleduck.com`，callback `https://ci.stevenshih-littleduck.com/authorize`。
4. 主機：建立 `/opt/woodpecker`（擁有者 `deploy`），放 compose 與 `.env`（權限 600），`docker compose up -d`。在本機渲染 `woodpecker.conf.template`，用 `install-sites.sh` 安裝站台。
5. 登入 Woodpecker，啟用 `littleduck_website`。repo 設定：Trusted 只開 volumes（Admin → Repos）、Allow deployments 開、Require approval 為 forks、Allow pull requests 開、timeout 60 分鐘。
6. 產生新的部署金鑰 `ci_deploy`，公鑰加進 `deploy` 的 `authorized_keys`，本機確認能登入。依決策 6 的表格用 `woodpecker-cli repo secret add` 建立 10 個 secret。

**切換**：

7. 開本變更的 PR（內含 `.woodpecker/`，刪除 `.drone.yml`）。確認 PR 上的兩個 Woodpecker check 成功。
8. 設定 GitHub branch protection：要求 PR、上述兩個 check、分支是最新、禁止 force push。
9. 在 Drone 停用 repo（Settings → Disable，或 `drone repo disable StevenShih-0402/littleduck_website`），確認 GitHub webhook 只剩 Woodpecker。執行 `docker compose -f /opt/drone/docker-compose.yml stop drone-runner`。
10. 合併 PR，確認 push pipeline 的兩個 `cd-*-staging` 成功。user 在 Staging 確認後，Deploy 到 `production`，確認兩個 `cd-*-production` 成功。
11. 依 `docs/cicd.md` 的耗時表量測新流程，寫入文件（Workflow follow-up）。

**回滾**（第 10 步失敗，或觀察期內發現問題）：

- 在 GitHub revert 本 PR（`.drone.yml` 回來），在 Drone 重新啟用 repo（webhook 會重建）並啟動 runner，然後在 Woodpecker 停用 repo（webhook 移除）。branch protection 的必要 check 改回 `continuous-integration/drone/pr`。
- Production 在整個切換過程中都不會被自動動到，只有 user 按 Deploy 才會變，所以回滾不涉及 Production 的容器。如果已經用 Woodpecker 部署了有問題的版本，依 runbook 第 6 節換回前一個 `<version>-<sha>`。

**觀察期後（建議 7 天，Workflow follow-up，另開小 PR）**：

- 刪除 `deploy/drone/`、`deploy/edge/drone.conf.template`，移除 `image-ref.sh` 的 `DRONE_COMMIT_SHA` fallback 與 `CLAUDE.md` 高風險樣式中的 `.drone.yml`。
- 主機上：`docker compose -f /opt/drone/docker-compose.yml down -v`、刪除 `/opt/drone`、`/opt/edge-sites/littleduck-drone.conf`（之後 `nginx -t` 並 reload）、`/var/cache/drone-trivy`。從 `authorized_keys` 移除 `drone-deploy` 公鑰，撤銷 Drone 用的 GHCR PAT 與 GitHub OAuth App，`certbot delete --cert-name drone.stevenshih-littleduck.com`，刪除 `drone.` 的 DNS 記錄。
