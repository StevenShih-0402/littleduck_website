# CI/CD 流水線

本 repo 以自架的 [Drone](https://docs.drone.io/) 跑 CI 與 CD，設定在根目錄的 `.drone.yml`。CI 與 CD 是不同的 pipeline：CI 只做檢查，不推映像也不部署；CD 只在 `main` 收到推送或人工 promote 時執行。流程骨架參考 mygo_v2_anon 的 Jenkins 流水線。

主機與帳號的初次設定、日常發布與回滾的操作步驟見 [deployment-runbook.md](deployment-runbook.md)。

## 觸發規則

| pipeline | 觸發 | 前置 |
|----------|------|------|
| `ci-backend` | PR 開啟／更新（目標 `main`）、推送到 `main` | — |
| `ci-frontend` | 同上 | — |
| `cd-backend-staging` | 推送到 `main` | 同一次建置的 `ci-backend` 通過 |
| `cd-frontend-preview` | 推送到 `main` | 同一次建置的 `ci-frontend` 通過 |
| `cd-backend-production` | promote 到 `production` | 該 commit 的 push 建置全部通過 |
| `cd-frontend-production` | promote 到 `production` | 同上 |

每個 pipeline 的結果以 GitHub commit status 回報。PR 上會看到 `continuous-integration/drone/pr`，`gh pr checks` 也讀得到；agent-team 的 `/ship-issue` 在「PR 中」第一步就是等這個 check。

## CI

### 後端（`ci-backend`）

| 階段 | 指令 | 失敗時 |
|------|------|--------|
| gofmt | `gofmt -l .` 有輸出即失敗，並列出未格式化的檔案 | 中止 |
| go-vet | `go vet ./...` | 中止 |
| wait-postgres | 等臨時 PostgreSQL 16 service 就緒（30 秒上限） | 中止 |
| go-test | `go test ./...`，`DB_*` 指向臨時 PostgreSQL | 任一測試失敗即中止 |
| gosec | `gosec -quiet -exclude-generated ./...`（對應 mygo 的 Bandit） | 有任何 finding 即中止 |
| docker-build | `docker build -t littleduck-api:ci-<build> backend` | 中止 |
| trivy | `trivy image --severity CRITICAL,HIGH --exit-code 1 --ignorefile backend/.trivyignore` | 有未登記的 CRITICAL／HIGH 即中止 |
| cleanup | 刪除本次建出的映像 | 不論成敗都執行 |

臨時 PostgreSQL 是 Drone 的 `services`，建置結束後自動移除。

### 前端（`ci-frontend`）

| 階段 | 指令 | 失敗時 |
|------|------|--------|
| install | `npm ci` | 中止 |
| test-design | `npm run test:design` | 中止 |
| build | `npm run build` | 中止 |
| audit | `npm audit --audit-level=high` | 有 high 以上弱點即中止 |

### 弱點例外登記

- **Trivy**：登記在 `backend/.trivyignore`，一行一個 CVE。上一行用註解寫明理由、確認日期與預計移除條件。只有在人工確認是誤判或可接受風險後才登記。
- **npm audit**：沒有例外清單。遇到無法即時升級的弱點，另開 issue 處理，不要降低門檻。

### 在本機跑 CI

用 drone CLI 的容器版即可，不必安裝 CLI。repo 要用相同的絕對路徑掛進去，因為建置容器是由主機的 docker 建立的：

```bash
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD:$PWD" -w "$PWD" \
  drone/cli:1.8.0 exec --trusted --pipeline ci-backend --branch main --event push .drone.yml
```

`--pipeline` 換成 `ci-frontend` 即可跑前端。注意：

- 建置容器以 root 身分寫入工作目錄，跑完前端後要把擁有者改回來：
  `docker run --rm -v "$PWD:/w" alpine:3.21 chown -R "$(id -u):$(id -g)" /w/frontend/node_modules /w/frontend/.next`
- Trivy 快取會建立在本機的 `/var/cache/drone-trivy`（root 擁有），不需要時以 `docker run --rm -v /var/cache:/c alpine:3.21 rm -rf /c/drone-trivy` 刪除。
- CD pipeline 需要 secret 與主機，不能在本機 exec。

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
| smoke | 對 `https://<staging>/health`、`/api/v1/articles` 各發一次請求 | 任一非 2xx 即失敗 |
| cleanup | 刪除 runner 主機上本次建出的映像 | 不論成敗都執行 |

### `cd-backend-production`（promote 到 `production` 時）

| 階段 | 內容 | 失敗時 |
|------|------|--------|
| verify-build | 查該 commit 的 `continuous-integration/drone/push` 狀態，必須是 `success` | 中止，不部署 |
| bundle／upload／deploy | 同 Staging，目錄為 `/opt/littleduck-api`，映像是同一個 `<version>-<sha>` | 同上 |
| smoke | 對 `https://<production>/health`、`/api/v1/articles` | 任一非 2xx 即失敗，需依 runbook 回滾 |

Production 不重新建置映像，直接部署 Staging 驗證過的那個 tag。

### 映像 tag

| Tag | 範例 | 用途 |
|-----|------|------|
| `<version>-<git 短 sha>` | `0.1.0-a1b2c3d` | 不可變，部署與回滾都用它 |
| `<version>` | `0.1.0` | 語意版本，給人看 |
| `latest` | `latest` | 移動標籤 |

- 映像位置：`ghcr.io/stevenshih-0402/littleduck-api`
- `<version>` 取自 `backend/cmd/server/main.go` 的 `@version` 註解，與 Swagger 顯示的版本一致。
- 部署腳本在 `deploy/ci/`：`image-ref.sh` 產生不可變 tag，`remote-deploy.sh` 在主機上執行。

### Nginx 站台

Hetzner 主機的 80／443 由 mygo_v2_anon 的 Nginx 佔用。本 repo 的站台樣板在 `deploy/edge/*.conf.template`，CD 會做兩件事：

1. 用 `deploy/edge/render.sh` 只帶入網域變數，產生 `littleduck-*.conf`。
2. 在主機上執行 `install-sites.sh`，把設定放進 `/opt/edge-sites/`。

`install-sites.sh` 放設定前會先檢查憑證存在，放完跑 `nginx -t`，失敗就還原，所以不會把 mygo 的 Nginx 弄壞。

## CD：前端

前端部署在 Vercel，只由 Drone 觸發。`frontend/vercel.json` 已關閉 Vercel 的 Git 自動部署，推送分支不會產生部署。部署腳本是 `deploy/ci/vercel-deploy.sh`。

| pipeline | 流程 | 冒煙測試 |
|----------|------|----------|
| `cd-frontend-preview` | `vercel pull --environment=preview` → `vercel build` → `vercel deploy --prebuilt`，記錄中印出 Preview 網址 | 對 Preview 首頁發請求，帶 `x-vercel-protection-bypass` 通過 Vercel Authentication |
| `cd-frontend-production` | verify-build → `vercel pull --environment=production` → `vercel build --prod` → `vercel deploy --prebuilt --prod` | 對正式網域首頁發請求 |

回滾：`vercel rollback <前一個部署網址> --token <token>`，或在 Vercel 專案的 Deployments 頁把前一個部署設為 Production（Instant Rollback）。

## Secret 清單

在 Drone 的 repo 設定 → Secrets 建立。全部**不要勾** Allow Pull Requests，這樣 PR 建置拿不到。

| secret | 用途 |
|--------|------|
| `ghcr_username`、`ghcr_token` | 推送 GHCR（PAT 需 `write:packages`） |
| `hetzner_host` | Hetzner 主機 IP |
| `deploy_ssh_key` | `deploy` 帳號的 `drone-deploy` 私鑰 |
| `api_host`、`api_staging_host` | 後端 Production／Staging 網域 |
| `vercel_token`、`vercel_org_id`、`vercel_project_id` | Vercel CLI |
| `vercel_bypass_secret` | Preview 冒煙測試用的 Protection Bypass for Automation |
| `frontend_host` | 前端正式網域 |
