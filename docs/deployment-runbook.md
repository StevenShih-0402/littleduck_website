# 部署操作手冊

給日常開發者的部署手冊，分成兩部分：

- 第 1–4 節：初次建置。Drone、後端主機、Vercel 的一次性設定，換主機或重建時才需要重做。
- 第 5–6 節：日常發布與回滾。

流水線的階段與門檻見 [cicd.md](cicd.md)。

> 下文的 `<網域>` 是正式網域，目前尚未決定。子網域固定為 `drone.`、`api.`、`api-staging.`。

## 0. 前置條件

| 項目 | 說明 |
|------|------|
| 主機 | mygo_v2_anon 使用的 Hetzner Cloud CX23。已有 Docker、`deploy` 帳號（只能操作 docker、無 sudo）、外部網路 `mygo-edge` 與 certbot，建置方式見 mygo_v2_anon 的 `docs/deployment-runbook.md` 第 2 節 |
| mygo 的 Nginx 要能載入外部站台 | mygo_v2_anon 的 Nginx 必須把主機 `/opt/edge-sites` 掛到容器 `/etc/nginx/edge-sites`，並在 http 區塊 `include /etc/nginx/edge-sites/*.conf;`。**這是 mygo_v2_anon 的變更（SDP-114）**，完成前 littleduck 的 api 與 Drone 都無法對外 |
| DNS | `drone.<網域>`、`api.<網域>`、`api-staging.<網域>` 的 A 記錄指向主機 IP |

## 1. Drone 初次建置

1. **建立 GitHub OAuth App**：在 GitHub 的 Settings → Developer settings → OAuth Apps 建立。
   - Homepage URL：`https://drone.<網域>`
   - Authorization callback URL：`https://drone.<網域>/login`
   - 記下 Client ID，並產生一組 Client Secret。
2. **核發憑證**：在主機上以 root 執行，webroot 與 mygo 共用。
   ```bash
   certbot certonly --webroot -w /var/www/certbot -d drone.<網域>
   ```
   第一次核發前，Nginx 要先能回應 `drone.<網域>` 的 HTTP-01 驗證。若尚未放站台設定，可先用 `certbot --standalone`，但要短暫停掉佔用 80 埠的 Nginx。
3. **放設定並啟動**：
   ```bash
   sudo mkdir -p /opt/drone /opt/edge-sites && sudo chown deploy: /opt/drone /opt/edge-sites
   # 以 deploy 身分，把 repo 的 deploy/drone/docker-compose.yml 複製到 /opt/drone/
   cp .env.example .env   # 依 deploy/drone/.env.example 填值；DRONE_RPC_SECRET 用 openssl rand -hex 16
   docker compose up -d
   docker compose ps      # drone-server、drone-runner 兩個容器都是 running
   ```
4. **放 Nginx 站台**：在本機渲染後上傳，再到主機上安裝。
   ```bash
   DRONE_HOST=drone.<網域> deploy/edge/render.sh /tmp/sites drone
   scp /tmp/sites/littleduck-drone.conf deploy/edge/install-sites.sh deploy@<主機>:/tmp/
   ssh deploy@<主機> 'sh /tmp/install-sites.sh /tmp'
   ```
5. **啟用 repo**：用 `DRONE_USER_FILTER` 列出的帳號登入 `https://drone.<網域>`，啟用 `littleduck_website`，然後在 repo 設定中：
   - 勾選 **Trusted**。CI 要掛載主機的 docker.sock 與快取目錄，只有管理員能設定。
   - 用 drone CLI 執行 `drone repo update --ignore-forks StevenShih-0402/littleduck_website`。repo 是 public，fork 來的 PR 不應在掛著 docker.sock 的 runner 上執行。
   - 依 [cicd.md](cicd.md#secret-清單) 建立 secret，全部不勾 Allow Pull Requests。
6. **驗證**：
   - 瀏覽器開 `https://drone.<網域>` 看到登入頁。
   - 用不在 `DRONE_USER_FILTER` 的帳號登入會被拒絕。
   - 從外部連主機任何埠都連不到 Drone 容器，因為 compose 檔沒有開埠。

**授權**：`drone/drone` 官方映像是企業版建置，個人或年營收低於一百萬美元的組織可免費使用。條件改變時，改用以 `-tags oss` 自行編譯的開源版。

**風險**：runner 掛載 docker.sock，等同主機 root 權限。只啟用自己的 repo、開啟 ignore forks、部署 secret 不給 PR，是三道防線。

## 2. 後端主機初次建置

每個環境一個目錄，兩個環境完全獨立：

| 環境 | 目錄 | 容器 | 網域 |
|------|------|------|------|
| Staging | `/opt/littleduck-api-staging` | `littleduck-api-staging`、`littleduck-db-staging` | `api-staging.<網域>` |
| Production | `/opt/littleduck-api` | `littleduck-api`、`littleduck-db` | `api.<網域>` |

1. **建目錄與 `.env`**：
   ```bash
   sudo mkdir -p /opt/littleduck-api /opt/littleduck-api-staging && sudo chown deploy: /opt/littleduck-api*
   ```
   兩個目錄各放一份 `.env`，欄位見 `backend/.env.example`：
   - `GIN_MODE=release`
   - `DB_USER`、`DB_PASSWORD`、`DB_NAME` 兩環境各自產生，`DB_PASSWORD` 不可共用。
   - `DB_HOST` 由 compose 指定，不用填。
   - 綠界金流的值在 Staging 用測試商店，Production 用正式商店。
2. **部署金鑰**：在本機產生 CD 專用的金鑰，不沿用 Jenkins 的那把，兩者可以分開撤銷。
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/drone_deploy -C "drone-deploy" -N ""
   ```
   把 `drone_deploy.pub` 加到主機 `/home/deploy/.ssh/authorized_keys`，私鑰內容存成 Drone secret `deploy_ssh_key`。
3. **GHCR 讀取權限**：主機的 `docker compose pull` 要能讀 GHCR。mygo 已用 `deploy` 帳號執行過 `docker login ghcr.io`，同一個 GitHub 帳號的 package 可以共用。若 package 是 private，確認那把 PAT 有 `read:packages`。
4. **憑證**：分別對 `api.<網域>`、`api-staging.<網域>` 執行 `certbot certonly --webroot -w /var/www/certbot -d <網域>`。**要在第一次 CD 之前完成**，否則 `install-sites.sh` 會因找不到憑證而讓部署失敗。站台不會被放上去，所以 mygo 不受影響。
5. **第一次部署**：合併任一 PR 到 `main`，`cd-backend-staging` 會建立 Staging。Production 等第一次 promote。

## 3. Vercel 初次建置

1. **建立專案**：在 Vercel 建立專案並連到本 repo。
   - Root Directory 設為 `frontend`。
   - Framework Preset 為 Next.js。
   - Node.js 版本選 20.x 以上（Next 16 需要 ≥ 20.9）。
2. **關閉 Git 自動部署**：`frontend/vercel.json` 已設 `git.deploymentEnabled: false`。推一個非 `main` 分支後，在 Deployments 頁確認沒有新部署。
3. **Token 與 ID**：
   - 在 Account Settings → Tokens 建立 token，存成 Drone secret `vercel_token`。
   - 在本機 **repo 根目錄**執行 `vercel link`，從產生的 `.vercel/project.json` 取 `orgId`、`projectId`，存成 `vercel_org_id`、`vercel_project_id`。專案的 Root Directory 已是 `frontend`，CLI 要從 repo 根目錄執行；CD 腳本也是在根目錄執行。`.vercel/` 已列入 `.gitignore`。
4. **Preview 冒煙用的 bypass**：在 Project Settings → Deployment Protection 建立 Protection Bypass for Automation，存成 `vercel_bypass_secret`。
5. **正式網域**：在 Project Settings → Domains 綁定網域，同一個值存成 `frontend_host`。

## 4. 確認 CI 會回報到 PR

開一個只改文件的 PR，然後執行 `gh pr checks <PR>`。應該看到 `continuous-integration/drone/pr` 且狀態為成功；GHCR 與 Vercel 都不應出現新的映像或部署。

## 5. 日常發布

1. **PR 合併到 `main`**：Drone 自動跑 CI，接著跑 `cd-backend-staging` 與 `cd-frontend-preview`。
2. **檢查 Staging**：
   - 到 Drone 的建置頁確認整個建置是綠的。
   - 後端打 `https://api-staging.<網域>/health`。
   - 前端打記錄中印出的 Preview 網址。
3. **promote**：在 Drone 建置頁右上角按 Deploy，Target 填 `production`。也可以用 CLI：
   ```bash
   drone build promote StevenShih-0402/littleduck_website <build 編號> production
   ```
   只能 promote 整個建置都通過的 commit。不符合時，`verify-build` 會擋下，不會部署。
4. **確認上線**：`cd-backend-production` 與 `cd-frontend-production` 的 smoke 階段都通過。

後端要發新版號時，改 `backend/cmd/server/main.go` 的 `@version`，再執行 `make swag` 重新產生 Swagger 文件。

## 6. 回滾

### 後端

SSH 到主機，改用前一個已知良好的不可變 tag：

```bash
ssh deploy@<主機>
cd /opt/littleduck-api
docker image ls ghcr.io/stevenshih-0402/littleduck-api   # 或到 GitHub → Packages 查歷史 tag
echo ghcr.io/stevenshih-0402/littleduck-api:<前一版>-<sha> > image.txt
LITTLEDUCK_API_IMAGE=$(cat image.txt) docker compose -f docker-compose.prod.yml up -d
docker inspect -f '{{.State.Health.Status}}' littleduck-api   # 等到 healthy
curl -fsS https://api.<網域>/health
```

也可以在 Drone 對前一個良好的建置重新 promote 到 `production`，效果相同，而且會跑冒煙測試。

**資料庫 migration 不會自動回退。** AutoMigrate 只會新增欄位與資料表，舊版程式通常能在新 schema 上運作。若新版做了不相容的 schema 變更，回滾前要先人工評估。

### 前端

```bash
vercel rollback <前一個部署網址> --token <token>
```

也可以在 Vercel 的 Deployments 頁，對前一個部署選 Promote to Production（Instant Rollback）。
