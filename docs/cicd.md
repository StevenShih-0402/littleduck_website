# CI/CD 流水線

本 repo 以自架的 [Woodpecker CI](https://woodpecker-ci.org/)（v3.20.1）跑 CI 與 CD，設定在根目錄的 `.woodpecker/`，一個檔案就是一個 workflow。CI 與 CD 是不同的 workflow：CI 只在 PR 上做檢查，不推映像也不部署；CD 在 `main` 收到推送時部署 Staging，在 user 按 Deploy 時部署 Production。流程骨架參考 mygo_v2_anon 的 Jenkins 流水線。

SDP-183 起由 Drone 換成 Woodpecker，同時簡化流程：

```
PR → CI（ci-backend、ci-frontend）→ /ship-issue 的其餘檢查 → 合併到 main
main 推送 → CD Staging（建映像 → Trivy → 推 GHCR → 部署 → 冒煙）→ 人工在 Staging 確認 → Deploy production → CD Production
```

主機與帳號的初次設定、日常發布與回滾的操作步驟見 [deployment-runbook.md](deployment-runbook.md)。

## 觸發規則

| workflow | 觸發 | 前置 | GitHub status context |
|----------|------|------|-----------------------|
| `ci-backend` | PR 開啟／更新（目標 `main`） | — | `ci/woodpecker/pr/ci-backend` |
| `ci-frontend` | 同上 | — | `ci/woodpecker/pr/ci-frontend` |
| `cd-backend-staging` | 推送到 `main` | 無（不等 CI，見下方「為什麼 push 不重跑 CI」） | `ci/woodpecker/push/cd-backend-staging` |
| `cd-frontend-staging` | 推送到 `main` | 同上 | `ci/woodpecker/push/cd-frontend-staging` |
| `cd-backend-production` | Deploy，Target 為 `production` | 該 commit 的 `ci/woodpecker/push/*` 全部成功 | `ci/woodpecker/deployment/cd-backend-production` |
| `cd-frontend-production` | 同上 | 同上 | `ci/woodpecker/deployment/cd-frontend-production` |

- 每個 workflow 各自以 GitHub commit status 回報。PR 上會看到 `ci/woodpecker/pr/ci-backend` 與 `ci/woodpecker/pr/ci-frontend` 兩個 check，`gh pr checks` 也讀得到；agent-team 的 `/ship-issue` 在「PR 中」第一步就是等這兩個 check。GitHub branch protection 把這兩個設為必要 check（runbook「GitHub branch protection」）。
- agent 的 `WOODPECKER_MAX_WORKFLOWS` 是 1，同一個 pipeline 的 workflow 依序執行。
- 本 repo 在 Woodpecker 的 Trusted 只開 **volumes**（步驟掛主機 docker.sock 與 Trivy 快取），Require approval 為 **forks**（fork 來的 PR 要擁有者核准才跑）。部署用 secret 只開 `push`、`deployment` 事件，PR 拿不到。
- YAML 內的 shell 變數寫成 `$$VAR`，避免被 Woodpecker 的變數替換吃掉。

### 為什麼 push 不重跑 CI

`main` 的品質由兩道關卡保證，所以推送到 `main` 時直接進 CD，不再對同一份程式重跑整套 CI：

1. **GitHub branch protection**：`main` 要求 PR、要求 `ci/woodpecker/pr/ci-backend` 與 `ci/woodpecker/pr/ci-frontend` 通過、要求分支是最新（Require branches to be up to date）、禁止 force push。分支是最新且 CI 通過時，`--merge` 產生的 merge commit 樹與 PR head 相同，push 時重跑 CI 不會多驗到任何東西。
2. **`/ship-issue` 的合併隊列**：rebase 到最新 `main` 後重跑 PR CI，全綠才合併。

即使有人繞過保護，CD Staging 的 `docker build` 本身會編譯（Go 的 `go build`、前端的 `next build`），接著 Trivy 會再掃要上線的映像，無法建置或有高風險弱點的程式上不了 Staging。Production 還要人工 Deploy。

這樣每個 commit 的映像只建 2 次（PR CI 驗 Dockerfile、CD 產出上線映像），Drone 時期是 3 次（PR CI、push CI、CD）。

## CI

### 後端（`ci-backend`）

| 階段 | 指令 | 失敗時 |
|------|------|--------|
| gofmt | `gofmt -l .` 有輸出即失敗，並列出未格式化的檔案 | 中止 |
| go-vet | `go vet ./...` | 中止 |
| wait-postgres | 等臨時 PostgreSQL 16 service 就緒（30 秒上限） | 中止 |
| go-test | `go test ./...`，`DB_*` 指向臨時 PostgreSQL | 任一測試失敗即中止 |
| gosec | `gosec -quiet -exclude-generated ./...`（對應 mygo 的 Bandit） | 有任何 finding 即中止 |
| docker-build | `docker build -t littleduck-api:ci-<pipeline 編號> backend` | 中止 |
| trivy | `trivy image --severity CRITICAL,HIGH --exit-code 1 --ignorefile backend/.trivyignore` | 有未登記的 CRITICAL／HIGH 即中止 |
| cleanup | 刪除本次建出的映像 | 不論成敗都執行 |

- 臨時 PostgreSQL 是 Woodpecker 的 `services`，workflow 結束後自動移除。
- go-vet、go-test、gosec 以 `GOMODCACHE=$CI_WORKSPACE/.gomodcache` 共用模組快取（同一個 workflow 的步驟共用 workspace）。

### 前端（`ci-frontend`）

| 階段 | 指令 | 失敗時 |
|------|------|--------|
| install | `npm ci` | 中止 |
| lint | `npm run lint`（ESLint，`eslint-config-next/core-web-vitals`） | 有 error 等級的違規即中止；warning 不擋 |
| typecheck | `npm run typecheck`（`tsc --noEmit`，以 `jsconfig.json` 的 `checkJs` 檢查 JS） | 有型別錯誤即中止 |
| test-design | `npm run test:design` | 中止 |
| audit | `npm audit --audit-level=high` | 有 high 以上弱點即中止 |
| docker-build | `docker build -t littleduck-web:ci-<pipeline 編號> frontend`（Dockerfile 內執行 `npm ci` 與 `next build`） | 中止 |
| trivy | `trivy image --severity CRITICAL,HIGH --exit-code 1 --ignorefile frontend/.trivyignore` | 有未登記的 CRITICAL／HIGH 即中止 |
| cleanup | 刪除本次建出的映像 | 不論成敗都執行 |

沒有獨立的 `npm run build` 階段：正式建置在 docker-build 內的 `next build` 進行，原本兩者各跑一次 `next build`。lint 與 typecheck 仍排在建置之前。本機檢查（CLAUDE.md「檢查指令」）仍要跑 `npm run build`。

### 弱點例外登記

- **Trivy**：後端登記在 `backend/.trivyignore`、前端登記在 `frontend/.trivyignore`，一行一個 CVE。上一行用註解寫明理由、確認日期與預計移除條件。只有在人工確認是誤判或可接受風險後才登記。
- **npm audit**：沒有例外清單。遇到無法即時升級的弱點，另開 issue 處理，不要降低門檻。

### 在本機跑 CI

用 woodpecker-cli 的容器版即可，不必安裝 CLI。repo 要用相同的絕對路徑掛進去，因為建置容器是由主機的 docker 建立的；CLI 容器不是 root，要加入 docker.sock 的群組：

```bash
docker run --rm --group-add "$(stat -c %g /var/run/docker.sock)" \
  -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD:$PWD" -w "$PWD" \
  woodpeckerci/woodpecker-cli:v3.20.1 exec --repo-trusted-volumes --pipeline-event pull_request .woodpecker/ci-backend.yaml
```

檔名換成 `.woodpecker/ci-frontend.yaml` 即可跑前端。語法檢查用 `docker run --rm -v "$PWD:$PWD" -w "$PWD" woodpeckerci/woodpecker-cli:v3.20.1 lint .woodpecker/`。注意：

- 建置容器以 root 身分寫入工作目錄，跑完後要把擁有者改回來（`.gomodcache` 已列入 `.gitignore`）：
  `docker run --rm -v "$PWD:/w" alpine:3.21 chown -R "$(id -u):$(id -g)" /w/.gomodcache /w/frontend/node_modules`
  不再需要時以 `docker run --rm -v "$PWD:/w" alpine:3.21 rm -rf /w/.gomodcache` 刪除模組快取。
- Trivy 快取會建立在本機的 `/var/cache/woodpecker-trivy`（root 擁有），不需要時以 `docker run --rm -v /var/cache:/c alpine:3.21 rm -rf /c/woodpecker-trivy` 刪除。
- CD workflow 需要 secret 與主機，不能在本機 exec。

## CD：後端

### `cd-backend-staging`（推送到 `main` 時）

| 階段 | 內容 | 失敗時 |
|------|------|--------|
| build | 以三個 tag 建置映像（見下方「映像 tag」） | 中止 |
| trivy | 對要推送的映像再掃一次，門檻同 CI | 中止，不推送 |
| push-ghcr | 推送三個 tag 到 GHCR，結束後 `docker logout` | 中止 |
| bundle | 組部署包：compose 檔、渲染好的 Nginx 站台、主機端腳本、`image.txt` | 中止 |
| upload | `scp` 部署包到 `/opt/littleduck-api-staging` | 中止 |
| deploy | 主機上執行 `remote-deploy.sh`：`pull` → `up -d` → 120 秒內等 api 變成 healthy → 更新 Nginx 站台 | 逾時即失敗，印出 `docker compose ps` 與 api 記錄 |
| smoke | 以 Staging 帳密（secret `staging_basic_auth`）對 `https://api.staging.stevenshih-littleduck.com/health`、`/api/v1/articles` 各發一次請求（主機名來自 secret `api_staging_host`），再不帶帳密請求 `/api/v1/articles` | 帶帳密任一非 2xx、或不帶帳密不是 401（驗證未生效）即失敗 |
| cleanup | 刪除 agent 主機上本次建出的映像 | 不論成敗都執行 |

CD Staging 的 Trivy 不省略：CD 建出的是另一個映像，base image 可能已經更新，而它才是要上線的東西。

### `cd-backend-production`（Deploy 到 `production` 時）

| 階段 | 內容 | 失敗時 |
|------|------|--------|
| verify-build | `verify-commit-status.sh`：該 commit 所有 `ci/woodpecker/push/*` status 都是 `success`，且至少有 `cd-backend-staging` 與 `cd-frontend-staging` 兩個 | 中止，不部署，列出不符的 context |
| bundle／upload／deploy | 同 Staging，目錄為 `/opt/littleduck-api`，映像是同一個 `<version>-<sha>` | 同上 |
| smoke | 對 `https://api.stevenshih-littleduck.com/health`、`/api/v1/articles`（主機名來自 secret `api_host`） | 任一非 2xx 即失敗，需依 runbook 回滾 |

Production 不重新建置映像，直接部署 Staging 驗證過的那個 tag。

### Deploy（人工核准）

Staging 確認後，user 以 Woodpecker 的 deployment 事件核准 Production，取代 Drone 的 promote：

- **UI**：到 Woodpecker 該 commit 的 push pipeline 頁，按 **Deploy**，Target 填 `production`，Task 留空。按鈕只出現在成功的 pipeline 上，repo 設定要開啟 Allow deployments。
- **CLI**：
  ```bash
  woodpecker-cli pipeline deploy --branch main StevenShih-0402/littleduck_website <pipeline 編號|last> production
  ```
  `last` 會選 `main` 上最近一次成功的 push pipeline。

兩者都會建立一個新的 pipeline（事件 `deployment`，`CI_PIPELINE_DEPLOY_TARGET=production`），跑 `cd-backend-production` 與 `cd-frontend-production`。

- **Target 打錯**（例如 `prod`）：兩個 Production workflow 都以 `evaluate: 'CI_PIPELINE_DEPLOY_TARGET == "production"'` 過濾，不會執行，部署不會發生。Woodpecker 對這種沒有符合任何 workflow 的 pipeline 怎麼顯示（略過或錯誤），合併後實機確認時補在這裡。
- **為什麼還要 verify-build**：Woodpecker 的 Deploy API（`POST /repos/{id}/pipelines/{n}?event=deployment`）不檢查原 pipeline 是否成功，所以由 `verify-build` 補上「只有 Staging 驗過的 commit 能上 Production」。必要的 context 清單在 `deploy/ci/verify-commit-status.sh` 的 `REQUIRED_CONTEXTS` 預設值；**新增 push 觸發的 workflow 時要同步這份清單**。本機測試：`sh deploy/ci/test-verify-commit-status.sh`（需要 python3、curl、jq）。

### 映像 tag

| Tag | 範例 | 用途 |
|-----|------|------|
| `<version>-<git 短 sha>` | `0.1.0-a1b2c3d` | 不可變，部署與回滾都用它 |
| `<version>` | `0.1.0` | 語意版本，給人看 |
| `latest` | `latest` | 移動標籤 |

- 映像位置：`ghcr.io/stevenshih-0402/littleduck-api`
- `<version>` 取自 `backend/cmd/server/main.go` 的 `@version` 註解，與 Swagger 顯示的版本一致。
- 部署腳本在 `deploy/ci/`：`image-ref.sh` 產生不可變 tag（commit sha 讀 `CI_COMMIT_SHA`），`remote-deploy.sh` 在主機上執行。

### Nginx 站台

Hetzner 主機的 80／443 由 mygo_v2_anon 的 Nginx 佔用。本 repo 的站台樣板在 `deploy/edge/*.conf.template`，CD 會做兩件事：

1. 用 `deploy/edge/render.sh` 只帶入網域變數（前端 Staging 另帶入通行值 `STAGING_GATE_TOKEN`），產生 `littleduck-*.conf`。
2. 在主機上執行 `install-sites.sh`，把設定放進 `/opt/edge-sites/`。

`install-sites.sh` 放設定前會先檢查憑證存在，放完跑 `nginx -t`，失敗就還原，所以不會把 mygo 的 Nginx 弄壞。憑證檢查是在 mygo 的 Nginx 容器內以 `docker exec … test -f` 進行（前置：Nginx 以相同路徑唯讀掛載 `/etc/letsencrypt`），檢查的就是 Nginx 實際讀得到的檔案，所以 `deploy` 不需要讀主機 `/etc/letsencrypt/live`、`archive`，也不需要放寬它們的權限。站台以 `auth_basic_user_file` 引用的帳密檔也用同樣方式在容器內檢查：缺檔時 `nginx -t` 照樣通過、但每個請求都回 500，所以缺檔就中止、不放設定。本機驗證可執行 `sh deploy/edge/test-install-sites.sh`（需要 Docker）。

#### Staging 身分驗證（SDP-181）

兩個 Staging 站台在 Nginx 層以 HTTP Basic Auth 擋住沒有帳密的人，Production 站台不受影響。建立帳密檔、輪替與回滾見 runbook「Staging 身分驗證」。

- **帳密檔**：兩站共用主機 `/opt/edge-sites/littleduck-staging.htpasswd`（容器內 `/etc/nginx/edge-sites/littleduck-staging.htpasswd`），SHA-512 crypt 雜湊，手動建立、不進版控。檔名不是 `.conf`，不會被 `include *.conf` 載入，也不會被 `install-sites.sh` 的備份／還原動到。
- **後端 Staging**（`api-staging.conf.template`）：全站要帳密，含 `/health`、`/swagger/`；唯一例外是綠界付款結果通知 `POST /api/v1/payments/ecpay/callback`（伺服器對伺服器，由後端 CheckMacValue 驗證）。
- **前端 Staging 同源轉送**（`web-staging.conf.template`）：`location /api/` 轉送到 `littleduck-api-staging:8080`，主機 `.env` 的 `API_BASE_URL` 是前端 Staging 自己。理由：跨站呼叫時 CORS 預檢不帶帳密，會被 401 擋下。
- **通行 cookie `ld_staging_gate`**：通過 Basic Auth 的回應會帶 `Set-Cookie: ld_staging_gate=<通行值>`（HttpOnly、Secure、SameSite=Strict）；之後帶正確 cookie 的請求免再驗 Basic Auth，讓帶 `Authorization: Bearer` 的管理員請求能通過（Bearer 與 Basic 共用同一個標頭）。通行值來自 secret `staging_gate_token`，`frontend-bundle.sh` 會檢查它是 32 字元以上的十六進位字串。map 用錨定正規式比對，因為 64 字元的字串鍵會超過預設 `map_hash_bucket_size`。
- **Staging 同源、Production 跨站**：Staging 驗不到 CORS 問題。目前後端 CORS 全開所以無妨；日後收斂 CORS 白名單時，要在上 Production 前另行驗證跨站呼叫。
- 本機驗證：`sh deploy/edge/test-staging-auth.sh`（需要 Docker 與網路），以真實樣板與假 upstream 驗證上述行為。

## CD：前端

前端與後端同一套模型：Woodpecker 建置一個 Next.js standalone 映像，推到 GHCR，部署到 Hetzner 主機；Staging 驗證後，Production 直接部署同一個映像，不重新建置。API 位址不寫進映像，由主機 `.env` 的 `API_BASE_URL` 在執行期經 `/runtime-config.js` 交給瀏覽器，所以同一個映像在 Staging 打 Staging 後端、在 Production 打正式後端。

### `cd-frontend-staging`（推送到 `main` 時）

| 階段 | 內容 | 失敗時 |
|------|------|--------|
| build | 以三個 tag 建置映像（見下方「映像 tag」） | 中止 |
| trivy | 對要推送的映像再掃一次，門檻同 CI | 中止，不推送 |
| push-ghcr | 推送三個 tag 到 GHCR，結束後 `docker logout` | 中止 |
| bundle | 組部署包：`deploy/web/docker-compose.staging.yml`、渲染好的 Nginx 站台、主機端腳本、`image.txt` | 中止 |
| upload | `scp` 部署包到 `/opt/littleduck-web-staging` | 中止 |
| deploy | 主機上執行 `remote-deploy.sh … web`：`pull` → `up -d` → 120 秒內等 web 變成 healthy → 更新 Nginx 站台 | 逾時即失敗，印出 `docker compose ps` 與 web 記錄 |
| smoke | `smoke-frontend.sh`，以 Staging 帳密（secret `staging_basic_auth`）請求：`/`、`/blog`、同源的 `/api/v1/articles` 回 2xx，且 `/runtime-config.js` 的 `apiBaseUrl` 是前端 Staging 自己（`https://staging.stevenshih-littleduck.com`）；再不帶帳密請求 `/` | 任一不符、或不帶帳密不是 401（驗證未生效）即失敗 |
| cleanup | 刪除 agent 主機上本次建出的映像 | 不論成敗都執行 |

### `cd-frontend-production`（Deploy 到 `production` 時）

| 階段 | 內容 | 失敗時 |
|------|------|--------|
| verify-build | 同後端：該 commit 所有 `ci/woodpecker/push/*` 都是 `success`，且兩個 `cd-*-staging` 都在 | 中止，不部署 |
| bundle／upload／deploy | 同 Staging，目錄為 `/opt/littleduck-web`，映像是同一個 `<version>-<sha>` | 同上 |
| smoke | 對正式網域檢查 `/`、`/blog` 回 2xx，`/runtime-config.js` 須含正式後端位址（不帶帳密、不檢查 401） | 任一不符即失敗，需依 runbook 回滾 |

### 映像與站台

- 映像位置：`ghcr.io/stevenshih-0402/littleduck-web`；tag 規則同後端（`<version>-<git 短 sha>`、`<version>`、`latest`），`<version>` 取自 `frontend/package.json` 的 `version`。
- 部署腳本在 `deploy/ci/`：`image-ref.sh frontend`、`frontend-version.sh`、`frontend-bundle.sh`、`smoke-frontend.sh`；主機端沿用 `remote-deploy.sh`（第 4 個參數指定服務 `web`）。
- compose 檔在 `deploy/web/`（容器 `littleduck-web`、`littleduck-web-staging`，不開埠，只接 `mygo-edge`）。
- Nginx 站台樣板是 `deploy/edge/web.conf.template`、`web-staging.conf.template`，機制同上方「Nginx 站台」。

### 回滾

與後端相同：在 Woodpecker 對前一個良好的 push pipeline 再按一次 Deploy 到 `production`，或在主機上把部署目錄的 `image.txt` 換成前一個 `<version>-<sha>` 後重新 `docker compose up -d`。步驟見 runbook。

## Secret 清單

在 Woodpecker 的 repo 設定 → Secrets 建立，或用 `woodpecker-cli repo secret add`。**不要**開放 `pull_request` 事件，這樣 PR 建置拿不到。事件依實際用途收窄：Production 的 host 只給 `deployment`，Staging 的 host 只給 `push`。`hetzner_host`、`deploy_ssh_key` 另限定在 `appleboy/drone-scp`、`appleboy/drone-ssh` 兩個 plugin，`commands` 步驟讀不到。

| secret | 用途 | 值 | 事件 | 限定映像 |
|--------|------|----|------|----------|
| `ghcr_username`、`ghcr_token` | 推送 GHCR | GitHub 帳號與具 `write:packages` 的 PAT（建議另發一把給 Woodpecker） | push | — |
| `hetzner_host` | Hetzner 主機 IP | 主機 IP，不寫進 repo | push、deployment | `appleboy/drone-scp`、`appleboy/drone-ssh` |
| `deploy_ssh_key` | `deploy` 帳號的 `ci_deploy` 私鑰 | runbook 第 2 節產生的私鑰內容 | push、deployment | `appleboy/drone-scp`、`appleboy/drone-ssh` |
| `api_host` | 後端 Production 網域 | `api.stevenshih-littleduck.com` | deployment | — |
| `api_staging_host` | 後端 Staging 網域 | `api.staging.stevenshih-littleduck.com` | push | — |
| `frontend_host` | 前端 Production 網域 | `stevenshih-littleduck.com` | deployment | — |
| `frontend_staging_host` | 前端 Staging 網域 | `staging.stevenshih-littleduck.com` | push | — |
| `staging_basic_auth` | Staging 冒煙用的 Basic Auth 帳密（`cd-backend-staging`、`cd-frontend-staging` 的 smoke） | `帳號:密碼` 明文，與主機帳密檔一致 | push | — |
| `staging_gate_token` | 前端 Staging 通行 cookie `ld_staging_gate` 的值（`cd-frontend-staging` 的 bundle） | `openssl rand -hex 32` 的輸出 | push | — |

共 10 個。建立範例（`WOODPECKER_SERVER`、`WOODPECKER_TOKEN` 的設定見 runbook 第 1 節）：

```bash
R=StevenShih-0402/littleduck_website
woodpecker-cli repo secret add --repository $R --name api_staging_host --value api.staging.stevenshih-littleduck.com --event push
woodpecker-cli repo secret add --repository $R --name deploy_ssh_key --value @"$HOME/.ssh/ci_deploy" \
  --event push --event deployment --image appleboy/drone-scp --image appleboy/drone-ssh
woodpecker-cli repo secret ls --repository $R
```

`--value @<檔案>` 從檔案讀值，密碼類的值也用這個方式，避免留在 shell 歷史。

## 各階段耗時

由 user 實機量測。Woodpecker 的數字在合併 SDP-183 後重新量測（Workflow follow-up），量完以新表取代下表。

| 階段 | 起點→終點 | 耗時 | 量測日期 |
|------|-----------|------|----------|
| PR CI | PR 開啟 → 兩個 `ci/woodpecker/pr/*` 成功 | 待量測 | — |
| 後端＋前端 Staging | 合併 → 兩個 `cd-*-staging` 成功 | 待量測 | — |
| Production | Deploy → 兩個 `cd-*-production` 完成 | 待量測 | — |
| 全程 | PR 開啟 → Production 冒煙通過 | 待量測 | — |

對照（Drone，2026-11-15，PR #16）：PR CI 7 分 16 秒；合併後的 push 建置（兩條 CI 加兩條 Staging）8 分 15 秒；兩條 Production 各約 20 秒；全程 22 分（機器執行約 16 分 10 秒）。Woodpecker 的 push 建置少了兩條 CI，PR CI 少一次 `next build`。
