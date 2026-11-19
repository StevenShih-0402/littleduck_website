# Tasks

本次沒有 `[FE]` 任務，也不改 `backend/` 的 Go 程式碼；以下 `[BE]` 任務改的是 `deploy/`、`.drone.yml`、`docs/`（與 SDP-168 相同，後端 Agent 的寫入範圍需暫時放寬到這些路徑）。不需跑 `go test`／`gofmt`，各項驗證寫在任務內，指令在 worktree 根目錄執行。shellcheck 指令：`docker run --rm -v "$PWD:/mnt" -w /mnt koalaman/shellcheck:v0.10.0 -s sh -S warning <檔案…>`。

## 1. 站台安裝：帳密檔檢查

- [x] 1.1 [BE] 改 `deploy/edge/install-sites.sh`（design 決策 5）：在憑證檢查迴圈之後、複製設定之前，對每個 `littleduck-*.conf` 以 `grep` 取出 `auth_basic_user_file` 的路徑（去掉結尾 `;`），在 Nginx 容器內以 `sh -c 'test -f "$1" || exit 3'` 檢查；回 3 → stderr `<conf 檔名>：帳密檔 <路徑> 不存在，請先依 runbook 建立` 並 exit 1；其他非 0 → 含 conf 檔名、路徑與「無法在 Nginx 容器內檢查」的訊息並 exit 1。沒有 `auth_basic_user_file` 的 conf 不做檢查。開頭註解補上帳密檔前置。驗證：`sh -n deploy/edge/install-sites.sh` 無輸出，shellcheck 無輸出 — 安裝前檢查帳密檔
- [x] 1.2 [BE] 擴充 `deploy/edge/test-install-sites.sh`：新增「帳密檔存在」（帳密檔只建在容器內的 `/etc/nginx/edge-sites/littleduck-staging.htpasswd`，conf 註解含 `auth_basic_user_file` 行）→ exit 0；「帳密檔不存在」→ exit 非 0、stderr 含「請先依 runbook 建立」、站台目錄不變。「無法檢查」案例維持放最後。驗證：`sh deploy/edge/test-install-sites.sh` 最後一行 `ALL PASS`、結束碼 0；暫時把 1.1 的 `exit 3` 分流改成一律通過時「帳密檔不存在」案例要 FAIL（確認測試有效，不 commit 暫時修改） — 帳密檔檢查有本機測試

## 2. Nginx 站台樣板

- [x] 2.1 [BE] 改 `deploy/edge/web-staging.conf.template`（design 決策 2、3）：http 層加 `map $cookie_ld_staging_gate $ld_web_staging_auth { "${STAGING_GATE_TOKEN}" off; default "littleduck staging"; }`；443 server 加 `auth_basic $ld_web_staging_auth;`、`auth_basic_user_file /etc/nginx/edge-sites/littleduck-staging.htpasswd;`、`add_header Set-Cookie "ld_staging_gate=${STAGING_GATE_TOKEN}; Path=/; Secure; HttpOnly; SameSite=Strict";`（不加 `always`）；新增 `location /api/` 以 `set $api_upstream littleduck-api-staging:8080;` 轉送，proxy 標頭與 `client_max_body_size 5m` 同 `api-staging.conf.template`。80 埠 server 不變。驗證：`FRONTEND_STAGING_HOST=staging.example STAGING_GATE_TOKEN=$(printf '%064d' 0 | tr 0 a) sh deploy/edge/render.sh /tmp/ld-r web-staging` 成功，且產出檔不含 `${` — 前端 Staging 樣板加驗證與同源轉送
  - 實作註記：map 鍵改為錨定正規式 `"~^${STAGING_GATE_TOKEN}$" off;`。64 字元的字串鍵超過預設 `map_hash_bucket_size`（64），`nginx -t` 會失敗（2.3 實測抓到）；調大該值會影響共用 http 區塊的其他站台。語意仍是「等於通行值才關閉 Basic Auth」，通行值已由 3.1 限定為十六進位。
- [x] 2.2 [BE] 改 `deploy/edge/api-staging.conf.template`（design 決策 4）：443 server 加 `auth_basic "littleduck staging";` 與同一個 `auth_basic_user_file`；新增 `location = /api/v1/payments/ecpay/callback { auth_basic off; … }`，轉送設定同 `location /`。驗證：`API_STAGING_HOST=api.staging.example sh deploy/edge/render.sh /tmp/ld-r api-staging` 成功 — 後端 Staging 樣板加驗證並放行綠界通知
- [x] 2.3 [BE] 新增 `deploy/edge/test-staging-auth.sh`（design「執行驗證觀察點」）：建立暫用 docker network，起兩個假 upstream 容器（名稱 `littleduck-web-staging`、`littleduck-api-staging`，回應中能辨識是哪一個，例：`nginx:1.27-alpine` 回固定字串），以 `render.sh` 渲染兩個 Staging 樣板、在 `nginx:1.27-alpine` 容器內產生自簽憑證與 `$6$` 測試帳密檔後載入，`nginx -t` 通過，再以 `curl --resolve` 逐項驗證 design 列出的前端 5 項、後端 4 項觀察點；印 `PASS:`／`FAIL:`，全過印 `ALL PASS`。開頭若已有同名容器就中止，`trap` 清除容器、network 與暫存目錄。驗證：`sh deploy/edge/test-staging-auth.sh` 印 `ALL PASS`、結束碼 0，結束後 `docker ps -a --filter name=littleduck- -q` 無新增容器；shellcheck 無輸出 — 站台驗證行為有本機測試

## 3. 部署腳本與流水線

- [x] 3.1 [BE] 改 `deploy/ci/frontend-bundle.sh`：`staging` 時檢查 `STAGING_GATE_TOKEN` 符合 `^[0-9a-f]{32,}$`，不符以「STAGING_GATE_TOKEN 未設定或格式錯誤（需 32 字元以上十六進位）」失敗；`production` 不檢查。驗證：在暫存目錄放假 `.image-ref`，`STAGING_GATE_TOKEN=abc` 與未設定時 `sh deploy/ci/frontend-bundle.sh staging` 皆失敗，給 64 字元十六進位值時成功產出 `bundle/littleduck-web-staging.conf`；`FRONTEND_HOST=x sh deploy/ci/frontend-bundle.sh production` 不受影響 — 通行值格式檢查
  - 實作註記：改以整串比對（`case` 排除非十六進位字元＋長度檢查），`grep -x` 逐行比對會放過多行值。
- [x] 3.2 [BE] 改 `deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh`（design 決策 6）：讀選用 `SMOKE_BASIC_AUTH`，有值時每個請求加 `-u`，並另做一次不帶帳密請求（後端 `/api/v1/articles`、前端 `/`），非 401 即失敗；`smoke-frontend.sh` 在預期 API 等於前端 base 時加測 `$base/api/v1/articles` 回 2xx。未設 `SMOKE_BASIC_AUTH` 時行為與現在相同。驗證：以 2.3 的測試環境（或本機 `python3 -m http.server` 等假服務）手動跑兩支腳本：有帳密且站台有驗證 → exit 0；站台沒驗證但設了 `SMOKE_BASIC_AUTH` → exit 1 並印「驗證未生效」類訊息；shellcheck 無輸出 — 冒煙帶帳密並驗證 401
- [x] 3.3 [BE] 改 `.drone.yml`：`cd-backend-staging` 的 smoke 加 `SMOKE_BASIC_AUTH: from_secret: staging_basic_auth`；`cd-frontend-staging` 的 bundle 加 `STAGING_GATE_TOKEN: from_secret: staging_gate_token`，smoke 加 `SMOKE_BASIC_AUTH`、`FRONTEND_STAGING_HOST`，預期 API 參數改為 `https://$$FRONTEND_STAGING_HOST`（移除不再需要的 `API_STAGING_HOST`）。Production pipeline 不動。驗證：`docker run --rm -v "$PWD:/mnt" -w /mnt drone/cli:1.8 lint .drone.yml`（或 `python3 -c 'import yaml,sys;list(yaml.safe_load_all(open(".drone.yml")))'`）通過，`git diff .drone.yml` 只動上述兩條 Staging pipeline — 流水線帶入新 secret
  - 實作註記：`drone lint` 需加 `--trusted`（既有 pipeline 掛載 host volume，`main` 不加也同樣失敗，非本次造成）；加上後通過，`yaml.safe_load_all` 也通過。

## 4. 文件

- [x] 4.1 [BE] 更新 `docs/cicd.md`：secret 表新增 `staging_basic_auth`、`staging_gate_token`（不勾 Allow Pull Requests）；`cd-backend-staging`／`cd-frontend-staging` 的 smoke 列改寫為帶帳密並驗證 401、前端預期 `apiBaseUrl` 為前端 Staging 自己；「Nginx 站台」一節補 Staging 驗證、同源 `/api/` 轉送、通行 cookie 與帳密檔檢查，以及「Staging 同源、Production 跨站，日後收斂 CORS 要在 Production 前另行驗證」。驗證：`grep -n "staging_basic_auth\|staging_gate_token\|ld_staging_gate" docs/cicd.md` 命中上述段落 — cicd.md 同步
- [x] 4.2 [BE] 更新 `docs/deployment-runbook.md`：新增「Staging 身分驗證」一節，寫 design Migration Plan 的三個上線前步驟（以 `deploy` 用 docker 產生 `$6$` 雜湊建立 `/opt/edge-sites/littleduck-staging.htpasswd` 的指令、建 Drone secret、改前端 Staging `.env`）、上線過程的短暫不可用、輪替帳密與通行值的方法、回滾步驟、Swagger 試打後台端點的限制；前端主機初次建置第 4 步的 Staging `API_BASE_URL` 改為 `https://staging.stevenshih-littleduck.com`，第 8 節「檢查 Staging」改為帶帳密；secret 數量（目前寫 8 個）同步更新為 10 個。驗證：`grep -n "littleduck-staging.htpasswd\|API_BASE_URL=https://staging" docs/deployment-runbook.md` 命中，且文件中沒有 Staging `API_BASE_URL` 指向 `api.staging` 的殘留步驟 — runbook 同步

## Workflow follow-up

- 合併前由 user 完成 runbook「Staging 身分驗證」的三個上線前步驟（帳密檔、兩個 Drone secret、前端 Staging `.env`）；未完成時 `install-sites.sh` 或 bundle 會讓 CD 失敗。
- 合併後確認兩條 cd-*-staging 的 smoke 通過，並以瀏覽器實測登入後台與功能開關預覽；Production 不帶帳密仍可存取。
- 本變更命中高風險檔案樣式（`.drone.yml`、`deploy/**`），PR 中需等 user 看過 diff 回覆「合併」。
