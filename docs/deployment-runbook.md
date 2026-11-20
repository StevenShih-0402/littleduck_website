# 部署操作手冊

給日常開發者的部署手冊，分成兩部分：

- 第 1–4 節：初次建置。Woodpecker、後端主機、前端主機的一次性設定，換主機或重建時才需要重做。
- 第 5–6 節：日常發布與回滾。
- 第 7 節：Staging 身分驗證。
- 第 8–9 節：從 Drone 切換到 Woodpecker、GitHub branch protection。

流水線的階段與門檻見 [cicd.md](cicd.md)。

下文直接寫實際主機名，對照如下。`api-staging`、`web-staging` 是環境名稱（目錄、容器、樣板檔名），不是主機名。

| 用途 | 主機名 | Woodpecker secret 或設定位置 |
|------|--------|--------------------------|
| 前端 Production | `stevenshih-littleduck.com` | `frontend_host` |
| 前端 Staging | `staging.stevenshih-littleduck.com` | `frontend_staging_host` |
| Woodpecker | `ci.stevenshih-littleduck.com` | `deploy/woodpecker/.env` 的 `WOODPECKER_HOST`；站台樣板變數 `WOODPECKER_SITE_HOST` |
| 後端 Production | `api.stevenshih-littleduck.com` | `api_host` |
| 後端 Staging | `api.staging.stevenshih-littleduck.com` | `api_staging_host` |

## 0. 前置條件

| 項目 | 說明 |
|------|------|
| 主機 | mygo_v2_anon 使用的 Hetzner Cloud CX23。已有 Docker、`deploy` 帳號（只能操作 docker、無 sudo）、外部網路 `mygo-edge` 與 certbot，建置方式見 mygo_v2_anon 的 `docs/deployment-runbook.md` 第 2 節 |
| mygo 的 Nginx 要能載入外部站台 | mygo_v2_anon 的 Nginx 必須把主機 `/opt/edge-sites` 掛到容器 `/etc/nginx/edge-sites`，並在 http 區塊 `include /etc/nginx/edge-sites/*.conf;`。**這是 mygo_v2_anon 的變更（SDP-114）**，完成前 littleduck 的 api、web 與 Woodpecker 都無法對外 |
| DNS | `ci.stevenshih-littleduck.com`、`api.stevenshih-littleduck.com`、`api.staging.stevenshih-littleduck.com` 三筆 A 記錄指向主機 IP，Cloudflare 代理狀態設為 DNS only（灰雲），HTTP-01 驗證與憑證續約需要直連；前端的兩筆 DNS 見第 3 節（同樣 DNS only） |

## 1. Woodpecker 初次建置

Woodpecker server 與 agent 以 docker compose 跑在同一台主機的 `/opt/woodpecker`，經 mygo 的 Nginx 以 `ci.stevenshih-littleduck.com` 對外。部署檔在 `deploy/woodpecker/`，站台樣板是 `deploy/edge/woodpecker.conf.template`。從 Drone 切換的順序見第 8 節。

1. **量測記憶體**：server（上限 256M）與 agent（上限 128M）的常駐記憶體要和 Drone 並行到切換當天。
   ```bash
   free -m
   ```
   判準：`available` 小於 1.4 GB 時，先 `docker compose -f /opt/drone/docker-compose.yml stop drone-runner` 再起 Woodpecker。runner 停掉後 Drone 不會再跑建置；切換前本來就不合併其他 PR（第 8 節），所以沒有影響。
2. **DNS 與憑證**：新增 `ci.stevenshih-littleduck.com` 的 A 記錄指向主機 IP，Cloudflare 代理狀態設為 **DNS only**（灰雲）。在主機上以 root 核發憑證，webroot 與 mygo 共用：
   ```bash
   certbot certonly --webroot -w /var/www/certbot -d ci.stevenshih-littleduck.com
   ```
   第一次核發前，Nginx 要先能回應 `ci.stevenshih-littleduck.com` 的 HTTP-01 驗證。mygo 的 Nginx 預設站台已處理 `/.well-known/acme-challenge/` 時可直接核發；否則先用 `certbot --standalone`，但要短暫停掉佔用 80 埠的 Nginx。
3. **建立 GitHub OAuth App**：在 GitHub 的 Settings → Developer settings → OAuth Apps 建立。**不要用 GitHub App**，Woodpecker 官方文件說明目前不能正常運作。
   - Homepage URL：`https://ci.stevenshih-littleduck.com`
   - Authorization callback URL：`https://ci.stevenshih-littleduck.com/authorize`
   - 記下 Client ID，並產生一組 Client Secret。
4. **放設定並啟動**：
   ```bash
   sudo mkdir -p /opt/woodpecker && sudo chown deploy: /opt/woodpecker
   # 以 deploy 身分，把 repo 的 deploy/woodpecker/docker-compose.yml 與 .env.example 複製到 /opt/woodpecker/
   cd /opt/woodpecker
   cp .env.example .env && chmod 600 .env
   # 填值：WOODPECKER_GITHUB_CLIENT、WOODPECKER_GITHUB_SECRET 來自第 3 步；
   # WOODPECKER_AGENT_SECRET 用 openssl rand -hex 32 產生；WOODPECKER_OPEN 維持 false
   docker compose up -d
   docker compose ps      # woodpecker-server、woodpecker-agent 兩個容器都是 running
   ```
5. **放 Nginx 站台**：在本機渲染後上傳，再到主機上安裝。
   ```bash
   WOODPECKER_SITE_HOST=ci.stevenshih-littleduck.com deploy/edge/render.sh /tmp/sites woodpecker
   scp -i ~/.ssh/hetzner_prod_deploy /tmp/sites/littleduck-woodpecker.conf deploy/edge/install-sites.sh deploy@<主機>:/tmp/
   ssh -i ~/.ssh/hetzner_prod_deploy deploy@<主機> 'sh /tmp/install-sites.sh /tmp'
   ```
   憑證檢查由 `install-sites.sh` 在 mygo 的 Nginx 容器內進行，`deploy` 不需要讀 `/etc/letsencrypt`；**不要**放寬 `/etc/letsencrypt/live`、`archive` 的權限（維持 certbot 預設 700）。
6. **啟用 repo 與 repo 設定**：用 `WOODPECKER_ADMIN` 列出的帳號登入 `https://ci.stevenshih-littleduck.com`，啟用 `littleduck_website`（server 會自動在 GitHub 建立指向 `https://ci.stevenshih-littleduck.com/api/hook` 的 webhook）。然後在 repo 設定中：
   - **Trusted**：只開 **volumes**，network 與 security 關閉。CI 要掛載主機的 docker.sock 與 Trivy 快取，只有管理員能設定。
   - **Allow deployments**：開啟。Production 的人工核准靠 Deploy 按鈕（deployment 事件）。
   - **Require approval**：設為 **forks**（預設值）。repo 是 public，fork 來的 PR 要擁有者核准才會在掛著 docker.sock 的 agent 上執行。
   - **Allow pull requests**：開啟。
   - **Timeout**：60 分鐘。
7. **設定 woodpecker-cli**：到 Woodpecker 的使用者設定（右上角頭像 → User settings → Personal access token）複製 token，在本機設定：
   ```bash
   export WOODPECKER_SERVER=https://ci.stevenshih-littleduck.com
   export WOODPECKER_TOKEN=<token>
   ```
   CLI 可從 [Woodpecker releases](https://github.com/woodpecker-ci/woodpecker/releases/tag/v3.20.1) 下載 `woodpecker-cli` 執行檔，或用容器版：`docker run --rm -e WOODPECKER_SERVER -e WOODPECKER_TOKEN woodpeckerci/woodpecker-cli:v3.20.1 <子指令>`（`--value @<檔案>` 讀檔時要另外把檔案掛進容器）。
8. **部署金鑰與 secret**：依第 2 節第 2 步產生新的部署金鑰 `ci_deploy`，再依 [cicd.md](cicd.md#secret-清單) 的表格建立 10 個 secret，事件與限定映像照表設定，**都不開放 `pull_request`**。
9. **驗證**：
   - `curl -sS -o /dev/null -w '%{http_code}' https://ci.stevenshih-littleduck.com/` 回 `200`。
   - 用不在 `WOODPECKER_ADMIN` 的帳號登入會被拒絕。
   - 從外部連主機任何埠都連不到 Woodpecker 容器，因為 compose 檔沒有開埠。
   - `woodpecker-cli repo secret ls --repository StevenShih-0402/littleduck_website` 列出 10 個名稱，且沒有任何一個包含 `pull_request` 事件。
   - `gh api repos/StevenShih-0402/littleduck_website/hooks --jq '.[].config.url'` 包含 `https://ci.stevenshih-littleduck.com/api/hook`。

**授權**：Woodpecker 是 Apache 2.0 授權的開源專案，沒有使用限制。

**風險**：agent 掛載 docker.sock，等同主機 root 權限。只允許 `WOODPECKER_REPO_OWNERS` 的 repo 啟用、fork PR 需核准、Trusted 只開 volumes、部署 secret 不給 PR，是四道防線。另外 `hetzner_host`、`deploy_ssh_key` 限定在 scp／ssh plugin，`commands` 步驟讀不到。只能 deployment 事件拿到的 secret，任何有 push 權限的人都能觸發；本 repo 只有擁有者有 push 權限。

**版本**：映像釘在 `v3.20.1`。Woodpecker 約每月一個 minor，曾有 minor 版帶意外的破壞性變更，升級前先看 release notes，再手動改 `deploy/woodpecker/docker-compose.yml`。

**排錯**：

| 症狀 | 原因 | 處理 |
|------|------|------|
| pipeline 一建立就 error，訊息提到 volumes 或 trusted | repo 的 Trusted 沒開 volumes | 以管理員帳號到 repo 設定開啟 Trusted → volumes，再按 Restart |
| 推送的 commit 出現失敗的 Woodpecker status，訊息是找不到設定 | 該 commit 沒有 `.woodpecker/`（切換前的 `main`） | 切換前預期會發生，不影響 Drone；切換完成後不會再出現 |
| PR 上沒有 Woodpecker check | webhook 沒建立或 OAuth token 失效 | 確認 GitHub repo 的 Settings → Webhooks 有 `/api/hook`；在 Woodpecker 登出後重新登入，再到 Webhooks 頁面 Redeliver |

容器名稱是 `littleduck-woodpecker-server`、`littleduck-woodpecker-agent`，看記錄時用這兩個名稱。

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
2. **部署金鑰**：在本機產生 CD 專用的金鑰，不沿用 Jenkins 或 Drone 的那把，各自可以分開撤銷。
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/ci_deploy -C "woodpecker-deploy" -N ""
   ```
   把 `ci_deploy.pub` 加到主機 `/home/deploy/.ssh/authorized_keys`，私鑰存成 Woodpecker secret `deploy_ssh_key`（只開 `push`、`deployment` 事件，限定 `appleboy/drone-scp`、`appleboy/drone-ssh`）：
   ```bash
   woodpecker-cli repo secret add --repository StevenShih-0402/littleduck_website --name deploy_ssh_key \
     --value @"$HOME/.ssh/ci_deploy" --event push --event deployment --image appleboy/drone-scp --image appleboy/drone-ssh
   ```
   加完後在本機確認能登入，否則 CD 的 upload 階段會以 `ssh: unable to authenticate` 失敗：
   ```bash
   ssh -i ~/.ssh/ci_deploy -o IdentitiesOnly=yes deploy@<主機> 'echo ok'
   ```
   Drone 時期的 `drone-deploy` 公鑰在切換觀察期結束後移除（第 8 節）；並行期間兩把公鑰同時存在。
3. **GHCR 讀取權限**：主機的 `docker compose pull` 要能讀 GHCR。mygo 已用 `deploy` 帳號執行過 `docker login ghcr.io`，同一個 GitHub 帳號的 package 可以共用。若 package 是 private，確認那把 PAT 有 `read:packages`。
4. **憑證**：兩個主機名各一張：
   ```bash
   sudo certbot certonly --webroot -w /var/www/certbot -d api.stevenshih-littleduck.com
   sudo certbot certonly --webroot -w /var/www/certbot -d api.staging.stevenshih-littleduck.com
   ```
   **要在第一次 CD 之前完成**，否則 `install-sites.sh` 會因找不到憑證而讓部署失敗。站台不會被放上去，所以 mygo 不受影響。憑證檢查在 Nginx 容器內進行，目錄權限維持 certbot 預設 700，不必放寬。
5. **確認目錄擁有者**：CD 以 `deploy` 身分寫入部署目錄與 `/opt/edge-sites`。目錄若是 root 擁有，upload 階段解壓會 `Permission denied`。第一次 CD 前以 root 確認：
   ```bash
   ls -ld /opt/littleduck-* /opt/edge-sites /opt/woodpecker            # 擁有者都要是 deploy
   chown -R deploy: /opt/littleduck-* /opt/edge-sites /opt/woodpecker   # 不是的話
   chmod 600 /opt/littleduck-*/.env /opt/woodpecker/.env                # .env 含密碼，只給 deploy 讀
   ```
6. **第一次部署**：合併任一 PR 到 `main`，`cd-backend-staging` 會建立 Staging。Production 等第一次 Deploy（第 5 節）。

## 3. 前端主機初次建置

前端與後端一樣部署在這台主機，每個環境一個目錄：

| 環境 | 目錄 | 容器 | 主機名 |
|------|------|------|--------|
| Staging | `/opt/littleduck-web-staging` | `littleduck-web-staging` | `staging.stevenshih-littleduck.com` |
| Production | `/opt/littleduck-web` | `littleduck-web` | `stevenshih-littleduck.com` |

1. **量測記憶體**：CX23 只有 4 GB，前端兩個容器的上限合計 448M（Staging 192M、Production 256M），`ci-frontend` 與 `cd-frontend-staging` 的 `docker build`（內含 `next build`）也在這台主機的 Woodpecker agent 上跑。
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
4. **核發憑證**：兩個主機名各一張，**要在第一次 CD 之前完成**，否則 `install-sites.sh` 找不到憑證會讓部署失敗（站台不會被放上去，mygo 不受影響）。憑證檢查在 Nginx 容器內進行，`/etc/letsencrypt/live`、`archive` 維持 700，不必放寬。
   ```bash
   sudo certbot certonly --webroot -w /var/www/certbot -d stevenshih-littleduck.com
   sudo certbot certonly --webroot -w /var/www/certbot -d staging.stevenshih-littleduck.com
   ```
5. **建目錄與 `.env`**：
   ```bash
   sudo mkdir -p /opt/littleduck-web /opt/littleduck-web-staging && sudo chown deploy: /opt/littleduck-web*
   ```
   兩個目錄各放一份 `.env`，**只有一行**：
   - Staging：`API_BASE_URL=https://staging.stevenshih-littleduck.com`（同源，經前端 Staging 站台的 `/api/` 轉送到後端 Staging，理由見第 7 節）
   - Production：`API_BASE_URL=https://api.stevenshih-littleduck.com`

   這個值在容器啟動時由 `/runtime-config.js` 交給瀏覽器，改了 `.env` 後 `docker compose up -d` 重建容器即可，不必重建映像。
6. **建 Woodpecker secret**：`frontend_staging_host`（`staging.stevenshih-littleduck.com`，只開 `push`）、`frontend_host`（`stevenshih-littleduck.com`，只開 `deployment`），都不開放 `pull_request`。`deploy_ssh_key`、`hetzner_host`、`ghcr_*` 與後端共用。
7. **第一次部署**：合併 PR 到 `main` 後 `cd-frontend-staging` 會建立 Staging，用 `smoke` 階段確認 `/runtime-config.js` 的 `apiBaseUrl` 是前端 Staging 自己。Staging 需要先完成第 7 節的帳密檔與 secret，否則部署會失敗。Production 等第一次 Deploy（第 5 節）。
8. **移除舊託管服務**：Production 部署成功、確認新站正常後，才刪除舊託管服務上的前端專案，並刪除 CI 裡殘留的四個舊 secret（token、org id、project id、preview bypass，名稱都以該服務的名字開頭）。在那之前保留舊專案，作為整體回退的後路（把 DNS 指回去）。

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
   4. **撤銷 token**：在舊服務的帳號設定 → Tokens，刪除當初給 CD 用的那把 token。只刪 CI 的 secret 的話，token 本身仍然有效。
   5. **移除 GitHub 整合**：在 GitHub → Settings → Applications → Installed GitHub Apps，找到舊服務的 App，按 Configure。只用在本 repo 的話就 Uninstall；還有其他 repo 在用的話，在 Repository access 把 `littleduck_website` 移除。沒移除的話，它可能繼續在 PR 上留言或建立預覽部署。
   6. **刪除 CI 裡的舊 secret**：在 Woodpecker 的 repo 設定 → Secrets 刪除。也可以用 CLI 先列出全部 secret，再刪掉不在下表中的名稱：
      ```bash
      woodpecker-cli repo secret ls --repository StevenShih-0402/littleduck_website
      woodpecker-cli repo secret rm --repository StevenShih-0402/littleduck_website --name <舊 secret 名稱>
      ```
      目前流水線用到的 secret 只有這 10 個，其他的都可以刪：`ghcr_username`、`ghcr_token`、`hetzner_host`、`deploy_ssh_key`、`api_host`、`api_staging_host`、`frontend_host`、`frontend_staging_host`、`staging_basic_auth`、`staging_gate_token`。刪完後執行 `git grep -n from_secret -- .woodpecker/`，確認沒有任何 workflow 引用已刪除的名稱。

## 4. 確認 CI 會回報到 PR

開一個只改文件的 PR，然後執行 `gh pr checks <PR>`。應該看到 `ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend` 且狀態都為成功；GHCR 不應出現新的映像。

## 5. 日常發布

1. **PR 合併到 `main`**：CI 已在 PR 上跑過，Woodpecker 直接跑 `cd-backend-staging` 與 `cd-frontend-staging`。
2. **檢查 Staging**：
   - 到 Woodpecker 該 commit 的 push pipeline 頁確認兩個 workflow 都是綠的。
   - 兩個 Staging 都要 Basic Auth（第 7 節），以下都帶 Staging 帳密。
   - 後端：`curl -u '<帳號>:<密碼>' https://api.staging.stevenshih-littleduck.com/health`。
   - 前端：瀏覽器開 `https://staging.stevenshih-littleduck.com`，輸入帳密後確認頁面正常，且 `/runtime-config.js` 的 `apiBaseUrl` 是 `https://staging.stevenshih-littleduck.com`；兩條 cd-*-staging 的 smoke 階段已自動檢查，也會確認不帶帳密回 401。
3. **Deploy**：在 Woodpecker 該 push pipeline 頁按 **Deploy**，Target 填 `production`，Task 留空。也可以用 CLI：
   ```bash
   woodpecker-cli pipeline deploy --branch main StevenShih-0402/littleduck_website <pipeline 編號|last> production
   ```
   `last` 是 `main` 上最近一次成功的 push pipeline。只能 Deploy 該 commit 的 `ci/woodpecker/push/*` 全部成功的 pipeline，不符合時 `verify-build` 會擋下，不會部署。Target 打錯（例如 `prod`）時不會執行任何部署 workflow，改用正確的 Target 再按一次即可。
4. **確認上線**：`cd-backend-production` 與 `cd-frontend-production` 的 smoke 階段都通過。

後端要發新版號時，改 `backend/cmd/server/main.go` 的 `@version`，再執行 `make swag` 重新產生 Swagger 文件。

## 6. 回滾

**先到 GHCR 查歷史 tag**：GitHub → 個人頁 → Packages → `littleduck-api`／`littleduck-web`，或以有 `read:packages` 的 token 執行：

```bash
gh api /users/StevenShih-0402/packages/container/littleduck-api/versions --jq '.[] | "\(.created_at) \(.metadata.container.tags|join(","))"'
```

不要只看主機上的 `docker image ls`。Woodpecker agent 與部署主機是同一台，CD 最後的 `cleanup` 會刪掉本次建出的映像，連剛部署的那個 tag 也一起刪掉（執行中的容器不受影響）。所以主機上通常只剩舊 tag，看不到完整歷史。回滾時 `docker compose up` 會從 GHCR 重新拉映像，**主機要能連到 GHCR**，`deploy` 帳號的 `docker login ghcr.io` 也要有效。

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

也可以在 Woodpecker 對前一個良好的 push pipeline 再按一次 Deploy 到 `production`，效果相同，而且會跑冒煙測試。

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

也可以在 Woodpecker 對前一個良好的 push pipeline 再按一次 Deploy 到 `production`，效果相同，而且會跑冒煙測試。前端沒有資料庫，不需評估 schema。

### 演練紀錄

| 日期 | 環境 | 內容 | 結果 |
|------|------|------|------|
| 2026-11-15 | 後端 Staging | `0.1.0-a2bb8cf` → `0.1.0-4e5d2e1` → 還原 `0.1.0-a2bb8cf` | 兩次切換都 healthy、`/health` 回 200，db 未重建；還原時從 GHCR 重新拉映像 |

## 7. Staging 身分驗證

SDP-181 起，前端 Staging（`staging.stevenshih-littleduck.com`）與後端 Staging（`api.staging.stevenshih-littleduck.com`）都要 HTTP Basic Auth，兩站共用一組帳密；Production 不受影響。機制見 [cicd.md](cicd.md#staging-身分驗證sdp-181)。

| 項目 | 位置 |
|------|------|
| 帳密雜湊檔 | 主機 `/opt/edge-sites/littleduck-staging.htpasswd`（Nginx 容器內 `/etc/nginx/edge-sites/littleduck-staging.htpasswd`） |
| 帳密明文（冒煙用） | Woodpecker secret `staging_basic_auth`，值為 `帳號:密碼` |
| 通行 cookie 值 | Woodpecker secret `staging_gate_token`，`openssl rand -hex 32` 的輸出 |

### 上線前（合併 SDP-181 的 PR 之前完成）

1. **建立帳密檔**：以 `deploy` 登入主機執行。用 Nginx 映像內的 `mkpasswd` 產生 SHA-512 crypt（`$6$`）雜湊，密碼從標準輸入讀，不留在指令列與 shell 歷史：
   ```bash
   read -r -p '帳號：' u
   read -r -s -p '密碼：' p; echo
   printf '%s:%s\n' "$u" "$(printf '%s' "$p" | docker run --rm -i nginx:1.27-alpine mkpasswd -m sha512 -P 0)" \
     > /opt/edge-sites/littleduck-staging.htpasswd
   unset p
   chmod 644 /opt/edge-sites/littleduck-staging.htpasswd   # 只含雜湊；Nginx worker 不是 deploy，需要可讀
   docker exec -u nginx "$(docker ps -q --filter label=com.docker.compose.project=mygo-v2 --filter label=com.docker.compose.service=nginx)" \
     test -r /etc/nginx/edge-sites/littleduck-staging.htpasswd && echo Nginx worker 可讀   # 以 worker 帳號檢查；root 一定讀得到，檢查無意義
   ```
   沒有這個檔時 `install-sites.sh` 會以「帳密檔 … 不存在，請先依 runbook 建立」中止部署。
2. **建 secret**：SDP-181 上線時建在 Drone；SDP-183 起改建在 Woodpecker（repo 設定 → Secrets，或 `woodpecker-cli repo secret add`），兩個都只開 `push` 事件，不開放 `pull_request`：
   - `staging_basic_auth`：`帳號:密碼`（與第 1 步相同）
   - `staging_gate_token`：在本機執行 `openssl rand -hex 32` 的輸出（必須是 32 字元以上的小寫十六進位，否則 `frontend-bundle.sh` 會失敗）
3. **改前端 Staging 的 `.env`**：把 `/opt/littleduck-web-staging/.env` 改為 `API_BASE_URL=https://staging.stevenshih-littleduck.com`。下次部署重建容器時生效，不必手動重啟。

### 上線

合併 PR 後，`cd-backend-staging` 與 `cd-frontend-staging` 依序部署並冒煙。兩條都跑完之前 Staging 會短暫不能用（後端站台先上了驗證，前端舊容器還在跨站呼叫 `api.staging`；或 `.env` 已改、站台還沒上 `/api/` 轉送），幾分鐘內自動恢復，只影響 Staging。

上線後確認：

- 兩條 cd-*-staging 的 smoke 通過。
- 瀏覽器開前端 Staging 會跳出帳密視窗；登入後台、預覽一項功能開關，各操作正常。
- `curl -s -o /dev/null -w '%{http_code}' https://stevenshih-littleduck.com/` 仍為 `200`（Production 不受影響）。

### 輪替

- **帳密**：照「上線前」第 1 步重建帳密檔（Nginx 每個請求都會讀檔，不必 reload），再更新 secret `staging_basic_auth`。瀏覽器已有的通行 cookie 仍有效，要讓所有人重新登入就一併輪替通行值。
- **通行值**：更新 secret `staging_gate_token` 後，在 Woodpecker 對最近一次 `main` 的 push pipeline 按 Restart，重跑 `cd-frontend-staging`。舊 cookie 立刻失效，使用者重新輸入一次帳密即可。

### 回滾

1. revert SDP-181 的 PR 並合併，讓兩條 Staging pipeline 重跑，站台設定回到無驗證版本。
2. 把 `/opt/littleduck-web-staging/.env` 改回 `API_BASE_URL=https://api.staging.stevenshih-littleduck.com`，再重跑一次 `cd-frontend-staging`。
3. 帳密檔與兩個 secret 留著無害，確定不再用時再刪。

### 已知限制

- **Staging Swagger 不能試打後台端點**：在 Swagger 的「Authorize」填 Bearer 權杖後，`Authorization` 標頭會被它佔用，Basic Auth 失效而回 401。後台端點改從前端 Staging 的後台操作測，或以 curl 經前端 Staging 的 `/api/` 帶通行 cookie 打：
  ```bash
  curl -H 'Cookie: ld_staging_gate=<通行值>' -H 'Authorization: Bearer <JWT>' https://staging.stevenshih-littleduck.com/api/v1/...
  ```
- **綠界付款結果通知**（`POST https://api.staging.stevenshih-littleduck.com/api/v1/payments/ecpay/callback`）不需帳密，由後端 CheckMacValue 驗證；其餘路徑（含 `/health`、`/swagger/`）都要帳密。

## 8. 從 Drone 切換

SDP-183 把 CI/CD 從 Drone 換成 Woodpecker。Woodpecker 先上線、在本變更的 PR 上驗證 CI，合併時一次切換 CD。Production 在整個過程中只有 user 按 Deploy 才會變動。

### 上線前

1. 確認 SDP-181 已合併，Drone 上的 10 個 secret 都在使用中。
2. 完成第 1 節「Woodpecker 初次建置」（含新部署金鑰與 10 個 secret）。
3. 啟用 repo 之後到本變更合併之前，推到 `main` 的 commit 沒有 `.woodpecker/`，Woodpecker 會回報失敗的 status。這段期間**不要合併其他 PR**。Drone 的 Production 前置檢查只看 `continuous-integration/drone/push`，不受影響。

### 切換

1. 開 SDP-183 的 PR（內含 `.woodpecker/`、刪除 `.drone.yml`）。確認 PR 上出現 `ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend` 且都成功，沒有新的 `continuous-integration/drone/pr`（Drone 在 PR 上找不到 `.drone.yml` 就不會建置）。
   ```bash
   gh pr checks <PR>
   ```
2. 依第 9 節設定 GitHub branch protection。
3. 在 Drone 停用 repo（repo 的 Settings → Disable，或 `drone repo disable StevenShih-0402/littleduck_website`），停用會移除 Drone 的 webhook。確認 GitHub webhook 只剩 Woodpecker，再停掉 runner：
   ```bash
   gh api repos/StevenShih-0402/littleduck_website/hooks --jq '.[].config.url'   # 只剩 https://ci.stevenshih-littleduck.com/api/hook
   docker compose -f /opt/drone/docker-compose.yml stop drone-runner
   ```
4. 合併 PR。確認 push pipeline 只有 `cd-backend-staging`、`cd-frontend-staging` 兩個 workflow 且都成功（含 SDP-181 的帳密冒煙）。
5. 在 Staging 人工確認後 Deploy `production`（第 5 節），確認兩個 `cd-*-production` 成功，`https://stevenshih-littleduck.com/` 回 200，`/runtime-config.js` 指向正式後端。
6. 再以 Target `prod` 觸發一次 Deploy，確認沒有任何部署 workflow 執行，並把 Woodpecker 的顯示方式補進 [cicd.md](cicd.md#deploy人工核准)。
7. 依 cicd.md「各階段耗時」量測新流程，以新表取代 Drone 的數字，另開小 PR 提交。

### 回滾

切換第 4、5 步失敗，或觀察期內發現問題時：

1. 在 GitHub revert SDP-183 的 PR（`.drone.yml` 回來）。
2. 在 Drone 重新啟用 repo（webhook 會重建），`docker compose -f /opt/drone/docker-compose.yml start drone-runner`。
3. 在 Woodpecker 停用 repo（webhook 移除）。
4. branch protection 的必要 check 改回 `continuous-integration/drone/pr`。

回滾不涉及 Production 的容器。如果已經用 Woodpecker 部署了有問題的版本，依第 6 節換回前一個 `<version>-<sha>`。

### 觀察期後（建議 7 天，另開小 PR，高風險）

- repo：刪除 `deploy/drone/`、`deploy/edge/drone.conf.template`，移除 `render.sh` 用法中的 `DRONE_HOST` 與 `image-ref.sh` 的 `DRONE_COMMIT_SHA` fallback，本節濃縮成歷史紀錄。
- 主機：
  ```bash
  docker compose -f /opt/drone/docker-compose.yml down -v
  sudo rm -rf /opt/drone /var/cache/drone-trivy
  rm /opt/edge-sites/littleduck-drone.conf   # 之後在 mygo 的 Nginx 容器內 nginx -t 並 reload
  ```
  從 `/home/deploy/.ssh/authorized_keys` 移除 `drone-deploy` 公鑰。
- 外部帳號：撤銷 Drone 用的 GHCR PAT 與 GitHub OAuth App，`sudo certbot delete --cert-name drone.stevenshih-littleduck.com`，刪除 `drone.` 的 DNS 記錄。

## 9. GitHub branch protection

CI 只在 PR 上跑，推送到 `main` 不重跑（理由見 [cicd.md](cicd.md#為什麼-push-不重跑-ci)），所以 `main` 要有 branch protection：

- 要求 PR 才能合併。
- 必要 check：`ci/woodpecker/pr/ci-backend`、`ci/woodpecker/pr/ci-frontend`。
- 要求分支是最新（Require branches to be up to date，`strict: true`）。
- 禁止 force push 與刪除分支。

以 `gh` 設定（要有 repo 的 admin 權限）：

```bash
gh api -X PUT repos/StevenShih-0402/littleduck_website/branches/main/protection --input - <<'JSON'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["ci/woodpecker/pr/ci-backend", "ci/woodpecker/pr/ci-frontend"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "required_linear_history": false,
  "allow_force_pushes": false,
  "allow_deletions": false
}
JSON
```

- `required_pull_request_reviews` 為 `null`：單人專案不要求 approve。`enforce_admins` 為 `false`，擁有者在緊急時仍可繞過，繞過時 CD Staging 的建置與 Trivy 仍會把關。
- 合併到 `main` 一律經 PR（`/ship-issue` 的合併隊列或 GitHub 的 Merge 按鈕），不要直接推送 `main`。

確認：

```bash
gh api repos/StevenShih-0402/littleduck_website/branches/main/protection --jq '.required_status_checks'
# 列出兩個 context 且 strict 為 true
```
