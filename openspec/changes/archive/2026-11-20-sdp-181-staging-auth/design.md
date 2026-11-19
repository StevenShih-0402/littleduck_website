# Design

## Context

動機見 proposal.md「Why」；要達成的行為見 `specs/staging-access/spec.md`。現況與限制：

- **對外入口只有 mygo 共用的 Nginx**。`deploy/edge/*-staging.conf.template` 由 `render.sh` 以 envsubst 渲染（只替換 `${大寫變數}`，`$host` 等 Nginx 變數保留），CD 上傳後由 `install-sites.sh` 放到主機 `/opt/edge-sites/`（掛進容器 `/etc/nginx/edge-sites/`，http 區塊 `include *.conf`），`nginx -t` 通過才 reload。Staging 容器不開埠，只接 `mygo-edge`，Nginx 以容器名轉送。
- **前端所有 API 呼叫都在瀏覽器端**：`frontend/lib/api.js` 以 `fetch(apiBaseUrl + "/api/v1" + path)` 呼叫，`apiBaseUrl` 來自執行期 `/runtime-config.js`（主機 `.env` 的 `API_BASE_URL`）。目前 Staging 是跨站呼叫 `https://api.staging.stevenshih-littleduck.com`；fetch 未設 `credentials`（預設 same-origin）。沒有 Next.js 伺服器端呼叫後端的路徑。
- **管理員 JWT 走 `Authorization: Bearer`**（`api.js` 第 37–38 行），與 Basic Auth 用的是同一個標頭。
- **後端 CORS 全開**（`backend/internal/middleware/middleware.go` 的 `CORS()`：`Allow-Origin: *`、OPTIONS 直接 204），不允許帶 credentials。
- **綠界付款結果通知是伺服器對伺服器**：`POST /api/v1/payments/ecpay/callback`（`ECPAY_RETURN_URL`，Staging 指向 `api.staging`），由後端 CheckMacValue 驗證；綠界不會帶 Basic Auth。`ECPAY_CLIENT_BACK_URL` 是瀏覽器導回前台，瀏覽器已有帳密，不受影響。
- **既有的自動檢查會打到 Staging**：
  - 容器健康檢查（compose 的 `wget http://127.0.0.1:…`）在容器內部，不經 Nginx，不受影響。
  - `cd-backend-staging` 的 `smoke-backend.sh` 打 `/health`、`/api/v1/articles`；`cd-frontend-staging` 的 `smoke-frontend.sh` 打 `/`、`/blog`、`/runtime-config.js`。加上驗證後都會得到 401，必須改。
  - 沒有其他 webhook 或排程打 Staging（Drone 的 GitHub webhook 打的是 Drone 自己的站台）。
- Production 與 Staging 用同一個映像；Production 的站台樣板、`.env`、pipeline 都不在本變更範圍。
- 主機是 4 GB 的 CX23，已跑 mygo、littleduck 兩套環境與 Drone runner，記憶體吃緊（runbook 第 3 節的量測）。

## Goals / Non-Goals

**Goals:**
- 沒有 Staging 帳密的人看不到前端與後端 Staging 的任何內容（含 Swagger）。
- Staging 上的既有測試流程照常運作：一般瀏覽、管理員登入與後台、管理員預覽、綠界測試付款、CD 冒煙。
- 不改 `backend/`、`frontend/` 的程式碼，不改 Production 的任何行為。

**Non-Goals:**
- 不做 OTP／SSO／個別帳號；全體共用一組 Staging 帳密。
- 不收斂 Production 的 CORS 白名單（程式碼註解中的 TODO 另開 issue）。
- 不處理上層 SDP-174 中其他專案（如 mygo_v2_anon）的 Staging。
- 不讓 Staging Swagger 的「Authorize」帶 Bearer 試打後台端點（見 Risks）。

## Decisions

### 1. 選 Basic Auth，不選 OTP

在 Nginx 層以 `auth_basic` 擋。

- **理由**：issue 允許二擇一。Basic Auth 是 Nginx 內建功能，不需新服務；只有一位開發者與少數測試者，共用帳密足夠；CD 冒煙用 `curl -u` 就能通過。
- **替代方案**：
  - **OTP／SSO（oauth2-proxy、Authelia 等前置驗證服務）**：要多跑一個常駐容器與 session 儲存，在 4 GB 主機上與建置尖峰搶記憶體；CD 冒煙也要另外處理繞過機制。以目前規模是過度設計，被否決。
  - **Cloudflare Access**：DNS 必須改為代理（橘雲），但 runbook 規定 DNS only 以免 HTTP-01 驗證與憑證續約失敗，被否決。
  - **IP 白名單**：開發者與測試者的 IP 不固定，被否決。
  - **應用層驗證（後端 middleware、Next.js middleware）**：要動 `backend/`、`frontend/` 程式碼並用環境變數區分 Staging／Production，同一映像在兩個環境行為不同，增加 Production 出錯的風險，被否決。

### 2. 前端 Staging 改為同源呼叫後端（`/api/` 轉送）

前端 Staging 站台新增 `location /api/`，轉送到 `littleduck-api-staging:8080`（同 `api-staging.conf.template` 的 resolver 與 proxy 標頭、`client_max_body_size 5m`）；Staging 主機 `.env` 的 `API_BASE_URL` 改為 `https://staging.stevenshih-littleduck.com`。`proxy_pass` 用不帶 URI 的變數形式，原始路徑 `/api/v1/...` 原樣送到後端。

- **理由**：如果後端 Staging 加 Basic Auth 而前端繼續跨站呼叫：(a) CORS 預檢 OPTIONS 依規範不帶帳密，會被 401 擋下；(b) 跨站 fetch 預設不帶 HTTP 帳密；(c) 管理員請求的 `Authorization: Bearer` 會蓋掉 Basic Auth。同源後 (a)(b) 消失，(c) 由決策 3 解決；不需改程式碼，只改設定與主機 `.env`。
- **替代方案**：
  - **後端 Staging 不加驗證**：程式碼不必動，但任何人都能直接讀 Staging API 與 Swagger 上未上線的端點，不符 issue「避免任何人連線就能看到」，被否決。
  - **維持跨站，改用 cookie＋`credentials: "include"`**：要改 `frontend/lib/api.js`，並把後端 CORS 改成回傳明確 origin 與 `Allow-Credentials`；這會連帶改到 Production 的 CORS 行為，被否決。
  - **以 `Origin` 標頭放行**：`Origin` 可被 curl 任意偽造，不算驗證，被否決。
- **取捨**：Staging 的呼叫路徑（同源）與 Production（跨站到 `api.`）不同，Staging 測不到 CORS 問題。目前 CORS 全開，風險低；記在 Risks。

### 3. 通過 Basic Auth 後發通行 cookie，帶 cookie 的請求免驗

前端 Staging 站台在 http 區塊加一個 `map`，以 `$cookie_ld_staging_gate` 對照 `auth_basic` 的值：等於通行值 → `off`，其他 → realm 字串；`auth_basic` 使用該變數。server 區塊以 `add_header Set-Cookie "ld_staging_gate=<通行值>; Path=/; Secure; HttpOnly; SameSite=Strict"`（不加 `always`，Nginx 只在 2xx／3xx 加上，401 不會發 cookie）。

- **理由**：解決決策 2 的 (c)：管理員請求帶 Bearer 時瀏覽器不會再送 Basic 帳密，但同源請求會帶 cookie，因此照樣通過。cookie 只在通過驗證的回應裡發，HttpOnly 讓前端 JS 讀不到。
- **通行值**由新 Drone secret `staging_gate_token` 經 `render.sh` 帶入（樣板變數 `${STAGING_GATE_TOKEN}`）。為避免值破壞 Nginx 語法，`frontend-bundle.sh` 在渲染前檢查它是 32 字元以上的十六進位字串，不符就失敗。輪替：改 secret 後重跑前端 Staging 部署，舊 cookie 立即失效、使用者重新輸入一次 Basic 帳密即可。
- **map 變數名稱要加站台前綴**（例：`$ld_web_staging_auth`）：所有 `littleduck-*.conf` 與 mygo 設定在同一個 http 區塊，同名 map 會讓 `nginx -t` 失敗。只有前端 Staging 需要這個 map，後端 Staging 站台不用（見決策 4）。
- **注意**：`add_header` 寫在 server 層時，任何 location 若自己有 `add_header` 會整個覆蓋 server 層的設定；目前 location 都沒有，實作時保持如此，或在該 location 重複宣告。
- **替代方案**：只靠 Basic Auth、遇到 Bearer 就放行 → 任何人帶 `Authorization: Bearer x` 就能繞過，被否決。

### 4. 後端 Staging 站台：全站 Basic Auth，只放行綠界通知

`api-staging.conf.template` 的 server 加 `auth_basic`（固定 realm，不需 cookie map），另加 `location = /api/v1/payments/ecpay/callback { auth_basic off; … }` 轉送到同一 upstream。`/health`、`/swagger/` 都在驗證範圍內。

- **理由**：直接存取 `api.staging` 的只剩開發者（curl、Swagger）、CD 冒煙與綠界；前兩者帶帳密即可，綠界無法帶帳密但端點本身有 CheckMacValue 驗證，且只接受 POST。
- **替代方案**：放行 `/health` 給外部監控 → 目前沒有外部監控打 Staging，冒煙可帶帳密，不放行以縮小面。

### 5. 帳密檔放主機站台目錄，安裝前檢查存在

兩個 Staging 站台共用 `auth_basic_user_file /etc/nginx/edge-sites/littleduck-staging.htpasswd`（主機 `/opt/edge-sites/littleduck-staging.htpasswd`）。檔名不是 `.conf`，不會被 `include *.conf` 載入，也不會被 `install-sites.sh` 的備份／還原（只處理 `littleduck-*.conf`）動到。由 `deploy` 依 runbook 手動建立，雜湊用 SHA-512 crypt（`$6$`，glibc 與 musl 的 crypt 都支援）。

- `install-sites.sh` 比照憑證檢查（SDP-168 的作法）：以 `grep` 取出每個 conf 的 `auth_basic_user_file` 路徑，在 Nginx 容器內 `test -f`，容器內回 3 → 「帳密檔不存在」，其他非 0 → 「無法在 Nginx 容器內檢查」。理由：缺檔時 `nginx -t` 照樣通過，但每個請求都回 500，必須在安裝前擋下。
- `test-install-sites.sh` 補「帳密檔存在」「帳密檔不存在」兩個案例。
- **替代方案**：帳密檔放部署目錄、由 CD 從 secret 產生 → CD 需要 htpasswd 工具且雜湊會出現在建置環境；手動建立一次即可，被否決。

### 6. 冒煙改為帶帳密，並驗證驗證確實生效

`smoke-backend.sh`、`smoke-frontend.sh` 讀選用環境變數 `SMOKE_BASIC_AUTH`（`帳號:密碼`）：

- 有設定時，所有請求加 `curl -u "$SMOKE_BASIC_AUTH"`，另外不帶帳密請求一次（後端 `/api/v1/articles`、前端 `/`），結果不是 401 就失敗。
- 沒設定時（Production）行為與現在完全相同。
- `smoke-frontend.sh` 在「預期 API 位址等於前端 base」時（即同源），額外檢查 `$base/api/v1/articles` 回 2xx；Production 傳入的預期位址是 `api.` 網域，不會觸發。
- `.drone.yml`：`cd-backend-staging`、`cd-frontend-staging` 的 smoke 步驟加 `SMOKE_BASIC_AUTH: from_secret: staging_basic_auth`；`cd-frontend-staging` 的 smoke 預期 API 改為 `https://$$FRONTEND_STAGING_HOST`、bundle 步驟加 `STAGING_GATE_TOKEN: from_secret: staging_gate_token`。兩個 secret 都不勾 Allow Pull Requests。Production pipeline 不動。

## 執行驗證觀察點

本變更只改 Nginx 設定、部署腳本與文件，本機的 `backend/`、`frontend/` 執行驗證不會經過這些設定，CLAUDE.md 的通用觀察點照跑但只確認沒有退化。本變更特有的驗證由新增的 `deploy/edge/test-staging-auth.sh` 執行（tasks 2.3），它在本機 Docker 上以 `nginx:1.27-alpine` 載入渲染後的兩個 Staging 站台（自簽憑證、測試帳密檔、兩個名為 `littleduck-web-staging`／`littleduck-api-staging` 的假 upstream 容器，同一個 user-defined network）。觀察點：

- 前端 Staging：不帶帳密 `/` → 401；帶帳密 `/` → 200 且回應有 `Set-Cookie: ld_staging_gate=`；只帶正確 cookie＋`Authorization: Bearer x` 請求 `/api/v1/articles` → 200 且請求送達假 api upstream；帶錯誤 cookie → 401；只帶 Bearer → 401。
- 後端 Staging：不帶帳密 `/api/v1/articles`、`/swagger/index.html` → 401；帶帳密 `/health` → 200；不帶帳密 `POST /api/v1/payments/ecpay/callback` → 送達 upstream（非 401）。
- `sh deploy/edge/test-install-sites.sh` → `ALL PASS`。
- 實機（合併後，由 user 或 CD）：兩個 cd-*-staging 的 smoke 通過；瀏覽器開 Staging 會跳帳密視窗，登入後台、預覽一項功能開關、各操作正常；`curl -s -o /dev/null -w '%{http_code}' https://stevenshih-littleduck.com/` 仍為 200。

## 變異測試目標

- `deploy/edge/install-sites.sh` 的帳密檔檢查分流（存在／不存在／無法檢查）。
- `deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh` 的「不帶帳密須回 401」判斷。

（本專案尚未設定變異測試指令，Teshima 會記為略過。）

## Risks / Trade-offs

- [合併後到兩條 Staging pipeline 都跑完之間，Staging 短暫不能用：後端站台先上了驗證，前端舊容器還在跨站呼叫 `api.staging`] → 只影響 Staging、幾分鐘內自動恢復；runbook 註明。
- [`.env` 先改、站台設定還沒上時，前端 Staging 呼叫同源 `/api/` 會打到 Next.js 而 404] → 同上，屬上線過程中的短暫狀態；runbook 的上線順序讓兩者在同一次部署內收斂。
- [Staging 與 Production 呼叫後端的路徑不同（同源 vs 跨站），Staging 驗不到 CORS 問題] → 目前 CORS 全開；日後收斂 CORS 白名單時要在 Production 前另行驗證（記在 `docs/cicd.md`）。
- [Staging Swagger 的「Authorize」改用 Bearer 後會蓋掉 Basic，無法在 Swagger 試打後台端點] → 後台端點改從前端 Staging 測，或用 curl 經前端 Staging 的 `/api/` 加 cookie；記在 runbook。
- [帳密雜湊格式不被 mygo Nginx 的 crypt 支援，導致帳密正確仍 401] → 選 `$6$`；萬一發生，CD 冒煙會失敗而非靜默放行。
- [通行值外洩等同帳密外洩] → HttpOnly、Secure、SameSite=Strict、只限前端 Staging 主機；可隨時輪替 secret。
- [`.drone.yml` 與 `deploy/**` 命中高風險檔案樣式] → 依 CLAUDE.md，合併前停下等 user 看 diff。

## Migration Plan

上線前（user 在主機與 Drone 手動完成，寫進 runbook）：

1. 以 `deploy` 建立 `/opt/edge-sites/littleduck-staging.htpasswd`（SHA-512 crypt），權限讓 Nginx 容器可讀。
2. Drone 新增 secret `staging_basic_auth`（`帳號:密碼` 明文）與 `staging_gate_token`（`openssl rand -hex 32`），都不勾 Allow Pull Requests。
3. 把 `/opt/littleduck-web-staging/.env` 的 `API_BASE_URL` 改為 `https://staging.stevenshih-littleduck.com`（下次部署重建容器時生效）。

上線：合併 PR，`cd-backend-staging`、`cd-frontend-staging` 依序部署並冒煙。

回滾：revert PR 並重跑兩條 Staging pipeline（站台設定回到無驗證版本），再把 `.env` 的 `API_BASE_URL` 改回 `https://api.staging.stevenshih-littleduck.com` 後重跑前端 Staging；帳密檔留在主機無害。
