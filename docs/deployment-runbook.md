# 部署操作手冊

給日常開發者的部署手冊，分成兩部分：

- 第 1–4 節：初次建置。Drone、後端主機、前端主機的一次性設定，換主機或重建時才需要重做。
- 第 5–6 節：日常發布與回滾。

流水線的階段與門檻見 [cicd.md](cicd.md)。

下文直接寫實際主機名，對照如下。`api-staging`、`web-staging` 是環境名稱（目錄、容器、樣板檔名），不是主機名。

| 用途 | 主機名 | Drone secret 或設定位置 |
|------|--------|--------------------------|
| 前端 Production | `stevenshih-littleduck.com` | `frontend_host` |
| 前端 Staging | `staging.stevenshih-littleduck.com` | `frontend_staging_host` |
| Drone | `drone.stevenshih-littleduck.com` | `deploy/drone/.env` 的 `DRONE_SERVER_HOST` |
| 後端 Production | `api.stevenshih-littleduck.com` | `api_host` |
| 後端 Staging | `api.staging.stevenshih-littleduck.com` | `api_staging_host` |

## 0. 前置條件

| 項目 | 說明 |
|------|------|
| 主機 | mygo_v2_anon 使用的 Hetzner Cloud CX23。已有 Docker、`deploy` 帳號（只能操作 docker、無 sudo）、外部網路 `mygo-edge` 與 certbot，建置方式見 mygo_v2_anon 的 `docs/deployment-runbook.md` 第 2 節 |
| mygo 的 Nginx 要能載入外部站台 | mygo_v2_anon 的 Nginx 必須把主機 `/opt/edge-sites` 掛到容器 `/etc/nginx/edge-sites`，並在 http 區塊 `include /etc/nginx/edge-sites/*.conf;`。**這是 mygo_v2_anon 的變更（SDP-114）**，完成前 littleduck 的 api、web 與 Drone 都無法對外 |
| DNS | `drone.stevenshih-littleduck.com`、`api.stevenshih-littleduck.com`、`api.staging.stevenshih-littleduck.com` 三筆 A 記錄指向主機 IP，Cloudflare 代理狀態設為 DNS only（灰雲），HTTP-01 驗證與憑證續約需要直連；前端的兩筆 DNS 見第 3 節（同樣 DNS only） |

## 1. Drone 初次建置

1. **建立 GitHub OAuth App**：在 GitHub 的 Settings → Developer settings → OAuth Apps 建立。
   - Homepage URL：`https://drone.stevenshih-littleduck.com`
   - Authorization callback URL：`https://drone.stevenshih-littleduck.com/login`
   - 記下 Client ID，並產生一組 Client Secret。
2. **核發憑證**：在主機上以 root 執行，webroot 與 mygo 共用。
   ```bash
   certbot certonly --webroot -w /var/www/certbot -d drone.stevenshih-littleduck.com
   ```
   第一次核發前，Nginx 要先能回應 `drone.stevenshih-littleduck.com` 的 HTTP-01 驗證。若尚未放站台設定，可先用 `certbot --standalone`，但要短暫停掉佔用 80 埠的 Nginx。
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
   DRONE_HOST=drone.stevenshih-littleduck.com deploy/edge/render.sh /tmp/sites drone
   scp -i ~/.ssh/hetzner_prod_deploy /tmp/sites/littleduck-drone.conf deploy/edge/install-sites.sh deploy@<主機>:/tmp/
   ssh -i ~/.ssh/hetzner_prod_deploy deploy@<主機> 'sh /tmp/install-sites.sh /tmp'
   ```
5. **啟用 repo**：用 `DRONE_USER_FILTER` 列出的帳號登入 `https://drone.stevenshih-littleduck.com`，啟用 `littleduck_website`，然後在 repo 設定中：
   - 勾選 **Trusted**。CI 要掛載主機的 docker.sock 與快取目錄，只有管理員能設定。
   - 用 drone CLI 執行 `drone repo update --ignore-forks StevenShih-0402/littleduck_website`。repo 是 public，fork 來的 PR 不應在掛著 docker.sock 的 runner 上執行。
   - 依 [cicd.md](cicd.md#secret-清單) 建立 secret，全部不勾 Allow Pull Requests。
6. **驗證**：
   - 瀏覽器開 `https://drone.stevenshih-littleduck.com` 看到登入頁。
   - 用不在 `DRONE_USER_FILTER` 的帳號登入會被拒絕。
   - 從外部連主機任何埠都連不到 Drone 容器，因為 compose 檔沒有開埠。

**授權**：`drone/drone` 官方映像是企業版建置，個人或年營收低於一百萬美元的組織可免費使用。條件改變時，改用以 `-tags oss` 自行編譯的開源版。

**風險**：runner 掛載 docker.sock，等同主機 root 權限。只啟用自己的 repo、開啟 ignore forks、部署 secret 不給 PR，是三道防線。

**排錯**（第一次上線時遇過）：

| 症狀 | 原因 | 處理 |
|------|------|------|
| PR 上沒有 Drone check；GitHub repo 的 Settings → Webhooks 顯示 push、pull_request 送出逾時，但 ping、create 成功 | Drone 存的 GitHub OAuth token 失效。`docker logs littleduck-drone-server` 會有 `trigger: cannot find yaml ... error="Bad credentials"` | 在 Drone 登出後重新登入，換一個新 token；再推一個 commit，或在 Webhooks 頁面 Redeliver |
| 建置一建立就 error，頁面顯示 `linter: untrusted repositories cannot mount host volumes` | repo 沒有設成 Trusted | 在 repo 設定勾選 Trusted 後，到建置頁按 Restart。看不到 Trusted 選項代表帳號不是管理員，要在 `.env` 設 `DRONE_USER_CREATE=username:<帳號>,admin:true` 後重建 server |
| 建置一建立就 error，頁面顯示 `yaml: missing kind attribute` | `.drone.yml` 有一份沒有 `kind` 的文件，例如開頭註解後多一行 `---` | 每份以 `---` 分隔的文件都要有 `kind: pipeline` |

容器名稱是 `littleduck-drone-server`、`littleduck-drone-runner`，看記錄時用這兩個名稱。

## 2. 後端主機初次建置

每個環境一個目錄，兩個環境完全獨立：

| 環境 | 目錄 | 容器 | 網域 |
|------|------|------|------|
| Staging | `/opt/littleduck-api-staging` | `littleduck-api-staging`、`littleduck-db-staging` | `api.staging.stevenshih-littleduck.com` |
| Production | `/opt/littleduck-api` | `littleduck-api`、`littleduck-db` | `api.stevenshih-littleduck.com` |

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
   把 `drone_deploy.pub` 加到主機 `/home/deploy/.ssh/authorized_keys`，私鑰內容存成 Drone secret `deploy_ssh_key`。加完後在本機確認能登入，否則 CD 的 upload 階段會以 `ssh: unable to authenticate` 失敗：
   ```bash
   ssh -i ~/.ssh/drone_deploy -o IdentitiesOnly=yes deploy@<主機> 'echo ok'
   ```
3. **GHCR 讀取權限**：主機的 `docker compose pull` 要能讀 GHCR。mygo 已用 `deploy` 帳號執行過 `docker login ghcr.io`，同一個 GitHub 帳號的 package 可以共用。若 package 是 private，確認那把 PAT 有 `read:packages`。
4. **憑證**：兩個主機名各一張：
   ```bash
   sudo certbot certonly --webroot -w /var/www/certbot -d api.stevenshih-littleduck.com
   sudo certbot certonly --webroot -w /var/www/certbot -d api.staging.stevenshih-littleduck.com
   ```
   **要在第一次 CD 之前完成**，否則 `install-sites.sh` 會因找不到憑證而讓部署失敗。站台不會被放上去，所以 mygo 不受影響。
5. **確認目錄擁有者**：CD 以 `deploy` 身分寫入部署目錄與 `/opt/edge-sites`。目錄若是 root 擁有，upload 階段解壓會 `Permission denied`。第一次 CD 前以 root 確認：
   ```bash
   ls -ld /opt/littleduck-* /opt/edge-sites /opt/drone            # 擁有者都要是 deploy
   chown -R deploy: /opt/littleduck-* /opt/edge-sites /opt/drone   # 不是的話
   chmod 600 /opt/littleduck-*/.env /opt/drone/.env                # .env 含密碼，只給 deploy 讀
   ```
6. **第一次部署**：合併任一 PR 到 `main`，`cd-backend-staging` 會建立 Staging。Production 等第一次 promote。

## 3. 前端主機初次建置

前端與後端一樣部署在這台主機，每個環境一個目錄：

| 環境 | 目錄 | 容器 | 主機名 |
|------|------|------|--------|
| Staging | `/opt/littleduck-web-staging` | `littleduck-web-staging` | `staging.stevenshih-littleduck.com` |
| Production | `/opt/littleduck-web` | `littleduck-web` | `stevenshih-littleduck.com` |

1. **量測記憶體**：CX23 只有 4 GB，前端兩個容器的上限合計 448M（Staging 192M、Production 256M），`ci-frontend` 與 `cd-frontend-staging` 的 `next build`／`docker build` 也在這台 runner 上跑。
   ```bash
   free -m
   docker stats --no-stream
   ```
   判準：`available` 減去 448M 後仍 >= 1024M（保留給建置尖峰）。不足時先調低 `deploy/web/docker-compose.staging.yml` 的記憶體上限與 `NODE_OPTIONS`，或加 swap，再往下做。量測結果記在這裡：

   | 日期 | total | used | available | `docker stats` 合計 | 判定 |
   |------|-------|------|-----------|---------------------|------|
   | 2026-11-15 | 3819M | 1727M | 2091M | 約 975MiB（15 個容器） | 通過：2091 − 448 = 1643M ≥ 1024M |

   - 量測時 Staging（api、db、web）已在執行，Production 尚未部署，所以數字已包含 `littleduck-web-staging`，結果偏保守。
   - Production 三個容器的上限合計 1024M（api 256M、db 512M、web 256M）。全部用到上限時，`available` 約剩 1067M，剛好高於 1024M，餘裕不多。量測時 swap 已用 243M。
   - 磁碟：`/` 共 38G，已用 21G（57%），剩 16G。每次部署都會留下舊映像，要定期 `docker image prune`。

2. **DNS**：`stevenshih-littleduck.com`（apex）與 `staging.stevenshih-littleduck.com` 都改成 **A 記錄**指向主機 IP；原本指向舊託管服務（切換前前端所在的平台）的記錄與轉址規則一併移除。Cloudflare 的代理狀態設為 **DNS only**（灰雲），否則 HTTP-01 驗證與後續憑證續約會失敗。
3. **清理 mygo 舊的 `staging.` 記錄**：`staging.stevenshih-littleduck.com` 是 mygo 舊 Staging 的主機名。先確認 mygo 已不使用，再移除舊的 A 記錄與 certbot 續約設定：
   ```bash
   sudo certbot certificates                                    # 找出舊憑證
   sudo certbot delete --cert-name staging.stevenshih-littleduck.com
   ```
   並確認 mygo 的 Nginx 設定沒有引用該主機名。
4. **核發憑證**：兩個主機名各一張，**要在第一次 CD 之前完成**，否則 `install-sites.sh` 找不到憑證會讓部署失敗（站台不會被放上去，mygo 不受影響）。
   ```bash
   sudo certbot certonly --webroot -w /var/www/certbot -d stevenshih-littleduck.com
   sudo certbot certonly --webroot -w /var/www/certbot -d staging.stevenshih-littleduck.com
   ```
5. **建目錄與 `.env`**：
   ```bash
   sudo mkdir -p /opt/littleduck-web /opt/littleduck-web-staging && sudo chown deploy: /opt/littleduck-web*
   ```
   兩個目錄各放一份 `.env`，**只有一行**：
   - Staging：`API_BASE_URL=https://api.staging.stevenshih-littleduck.com`
   - Production：`API_BASE_URL=https://api.stevenshih-littleduck.com`

   這個值在容器啟動時由 `/runtime-config.js` 交給瀏覽器，改了 `.env` 後 `docker compose up -d` 重建容器即可，不必重建映像。
6. **建 Drone secret**：新增 `frontend_staging_host`（`staging.stevenshih-littleduck.com`），`frontend_host` 改成 `stevenshih-littleduck.com`；不勾 Allow Pull Requests。`deploy_ssh_key`、`hetzner_host`、`ghcr_*` 與後端共用。
7. **第一次部署**：合併 PR 到 `main` 後 `cd-frontend-staging` 會建立 Staging，用 `smoke` 階段確認 `/runtime-config.js` 含 Staging 後端位址。Production 等第一次 promote。
8. **移除舊託管服務**：Production promote 成功、確認新站正常後，才刪除舊託管服務上的前端專案，並刪除 Drone 裡殘留的四個舊 secret（token、org id、project id、preview bypass，名稱都以該服務的名字開頭）。在那之前保留舊專案，作為整體回退的後路（把 DNS 指回去）。

   刪除後就無法用「把 DNS 指回去」回退，請依序做：

   1. **確認新站已接手**：
      ```bash
      dig +short A stevenshih-littleduck.com            # 只有主機 IP
      dig +short A staging.stevenshih-littleduck.com    # 只有主機 IP
      dig +short CNAME stevenshih-littleduck.com        # 沒有輸出
      curl -fsS https://stevenshih-littleduck.com/runtime-config.js        # 指向 https://api.stevenshih-littleduck.com
      ```
      再到 Cloudflare DNS 確認沒有其他記錄（例如 `www`）還指向舊服務。有的話，先改指主機或刪掉。
   2. **解除網域綁定**：登入**舊服務的 Dashboard**（不是 Cloudflare），進入原本的前端專案 → Settings → Domains，移除 `stevenshih-littleduck.com`、`staging.stevenshih-littleduck.com` 與其他自訂網域。這一步只清掉舊服務那邊的設定。**Cloudflare 的 A 記錄不要動**：它們指向主機，新站靠它們運作。先解除綁定，以免之後別人把同一個網域加到他們的專案時出現衝突。
   3. **刪除專案**：在專案設定最底下的刪除區塊，輸入專案名稱確認刪除。部署紀錄、環境變數與預覽網址會一起刪除，無法復原。
   4. **撤銷 token**：在舊服務的帳號設定 → Tokens，刪除當初給 CD 用的那把 token。只刪 Drone secret 的話，token 本身仍然有效。
   5. **移除 GitHub 整合**：在 GitHub → Settings → Applications → Installed GitHub Apps，找到舊服務的 App，按 Configure。只用在本 repo 的話就 Uninstall；還有其他 repo 在用的話，在 Repository access 把 `littleduck_website` 移除。沒移除的話，它可能繼續在 PR 上留言或建立預覽部署。
   6. **刪除 Drone 裡的舊 secret**：在 repo 的 Settings → Secrets 刪除。也可以用 CLI 先列出全部 secret，再刪掉不在下表中的那四個：
      ```bash
      drone secret ls StevenShih-0402/littleduck_website
      drone secret rm --name <舊 secret 名稱> StevenShih-0402/littleduck_website
      ```
      目前流水線用到的 secret 只有這 8 個，其他的都可以刪：`ghcr_username`、`ghcr_token`、`hetzner_host`、`deploy_ssh_key`、`api_host`、`api_staging_host`、`frontend_host`、`frontend_staging_host`。刪完後在 `.drone.yml` 執行 `git grep -n from_secret`，確認沒有任何 pipeline 引用已刪除的名稱。

## 4. 確認 CI 會回報到 PR

開一個只改文件的 PR，然後執行 `gh pr checks <PR>`。應該看到 `continuous-integration/drone/pr` 且狀態為成功；GHCR 不應出現新的映像。

## 5. 日常發布

1. **PR 合併到 `main`**：Drone 自動跑 CI，接著跑 `cd-backend-staging` 與 `cd-frontend-staging`。
2. **檢查 Staging**：
   - 到 Drone 的建置頁確認整個建置是綠的。
   - 後端打 `https://api.staging.stevenshih-littleduck.com/health`。
   - 前端開 `https://staging.stevenshih-littleduck.com`，並確認 `https://staging.stevenshih-littleduck.com/runtime-config.js` 的 `apiBaseUrl` 是 Staging 後端；`cd-frontend-staging` 的 smoke 階段已自動檢查。
3. **promote**：在 Drone 建置頁右上角按 Deploy，Target 填 `production`。也可以用 CLI：
   ```bash
   drone build promote StevenShih-0402/littleduck_website <build 編號> production
   ```
   只能 promote 整個建置都通過的 commit。不符合時，`verify-build` 會擋下，不會部署。
4. **確認上線**：`cd-backend-production` 與 `cd-frontend-production` 的 smoke 階段都通過。

後端要發新版號時，改 `backend/cmd/server/main.go` 的 `@version`，再執行 `make swag` 重新產生 Swagger 文件。

## 6. 回滾

**先到 GHCR 查歷史 tag**：GitHub → 個人頁 → Packages → `littleduck-api`／`littleduck-web`，或以有 `read:packages` 的 token 執行：

```bash
gh api /users/StevenShih-0402/packages/container/littleduck-api/versions --jq '.[] | "\(.created_at) \(.metadata.container.tags|join(","))"'
```

不要只看主機上的 `docker image ls`。runner 與部署主機是同一台，CD 最後的 `cleanup` 會刪掉本次建出的映像，連剛部署的那個 tag 也一起刪掉（執行中的容器不受影響）。所以主機上通常只剩舊 tag，看不到完整歷史。回滾時 `docker compose up` 會從 GHCR 重新拉映像，**主機要能連到 GHCR**，`deploy` 帳號的 `docker login ghcr.io` 也要有效。

各環境的對照：

| 環境 | 目錄 | compose 檔 | 容器 |
|------|------|------------|------|
| 後端 Production | `/opt/littleduck-api` | `docker-compose.prod.yml` | `littleduck-api` |
| 後端 Staging | `/opt/littleduck-api-staging` | `docker-compose.staging.yml` | `littleduck-api-staging` |
| 前端 Production | `/opt/littleduck-web` | `docker-compose.prod.yml` | `littleduck-web` |
| 前端 Staging | `/opt/littleduck-web-staging` | `docker-compose.staging.yml` | `littleduck-web-staging` |

### 後端

SSH 到主機，改用前一個已知良好的不可變 tag：

```bash
ssh deploy@<主機>
cd /opt/littleduck-api
cat image.txt                                   # 目前的版本，回滾後要還原時用得到
echo ghcr.io/stevenshih-0402/littleduck-api:<前一版>-<sha> > image.txt
LITTLEDUCK_API_IMAGE=$(cat image.txt) docker compose -f docker-compose.prod.yml up -d
docker inspect -f '{{.Config.Image}} {{.State.Health.Status}}' littleduck-api   # 確認 tag 正確且為 healthy，剛啟動時會先顯示 starting
curl -fsS https://api.stevenshih-littleduck.com/health
```

也可以在 Drone 對前一個良好的建置重新 promote 到 `production`，效果相同，而且會跑冒煙測試。

回滾只重建 api 容器，資料庫容器維持運作，不會被重建。

**資料庫 migration 不會自動回退。** AutoMigrate 只會新增欄位與資料表，舊版程式通常能在新 schema 上運作。若新版做了不相容的 schema 變更，回滾前要先人工評估。

### 前端

與後端相同，改用前一個不可變 tag（Staging 的目錄、compose 檔、容器名稱見上方對照表）：

```bash
ssh deploy@<主機>
cd /opt/littleduck-web
cat image.txt                                   # 目前的版本，回滾後要還原時用得到
echo ghcr.io/stevenshih-0402/littleduck-web:<前一版>-<sha> > image.txt
LITTLEDUCK_WEB_IMAGE=$(cat image.txt) docker compose -f docker-compose.prod.yml up -d
docker inspect -f '{{.Config.Image}} {{.State.Health.Status}}' littleduck-web   # 確認 tag 正確且為 healthy，剛啟動時會先顯示 starting
curl -fsS https://stevenshih-littleduck.com/runtime-config.js
```

也可以在 Drone 對前一個良好的建置重新 promote 到 `production`，效果相同，而且會跑冒煙測試。前端沒有資料庫，不需評估 schema。

### 演練紀錄

| 日期 | 環境 | 內容 | 結果 |
|------|------|------|------|
| 2026-11-15 | 後端 Staging | `0.1.0-a2bb8cf` → `0.1.0-4e5d2e1` → 還原 `0.1.0-a2bb8cf` | 兩次切換都 healthy、`/health` 回 200，db 未重建；還原時從 GHCR 重新拉映像 |
