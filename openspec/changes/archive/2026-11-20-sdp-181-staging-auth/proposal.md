# Proposal

來源：Linear SDP-181「littleduck_website：Staging 環境加上身分驗證」（上層 SDP-174「在 Staging 環境加上身分驗證」的本專案範圍）。

## Why

前端 Staging（`staging.stevenshih-littleduck.com`）與後端 Staging（`api.staging.stevenshih-littleduck.com`）目前任何人連線都能看到，裡面有尚未上線、待測試的畫面與資料（含 Swagger 列出的新端點）。issue 要求加上 Basic Auth 或 OTP，讓沒有帳密的人看不到 Staging。

## What Changes

- 兩個 Staging 站台在 Nginx（mygo 共用的 edge Nginx）加上 HTTP Basic Auth，共用一組帳密；帳密的雜湊檔只放在主機上，不進版控。Production 站台不變。
- 前端 Staging 改為**同源**呼叫後端：Staging 站台新增 `/api/` 轉送到後端 Staging 容器，主機 `.env` 的 `API_BASE_URL` 改為 `https://staging.stevenshih-littleduck.com`。原因：瀏覽器跨站呼叫 `api.staging.…` 時，CORS 預檢不帶帳密、管理員 JWT 也佔用同一個 `Authorization` 標頭，兩者都會被 Basic Auth 擋下。
- 通過 Basic Auth 後，Staging 站台發一個 HttpOnly 的通行 cookie；帶有效 cookie 的請求免再驗 Basic Auth，讓帶 `Authorization: Bearer` 的管理員請求能通過。cookie 值由新的 Drone secret 帶入站台設定。
- 後端 Staging 站台只對綠界付款結果通知（`POST /api/v1/payments/ecpay/callback`，伺服器對伺服器、以 CheckMacValue 驗證）免驗證，其餘全部要 Basic Auth。
- `install-sites.sh` 安裝前檢查站台引用的帳密檔在 Nginx 容器內存在；缺檔就中止（Nginx 對缺檔只會在請求時回 500，`nginx -t` 抓不到）。
- Staging 冒煙測試改為帶帳密，並新增「不帶帳密回 401」的檢查，確認驗證確實生效；前端冒煙改為確認 `apiBaseUrl` 指向前端 Staging 自己、且經 `/api/` 轉送讀得到文章列表。Production 冒煙不變。
- 新增 Drone secret：`staging_basic_auth`（冒煙用的 `帳號:密碼`）、`staging_gate_token`（通行 cookie 值）；更新 `docs/cicd.md` 與 `docs/deployment-runbook.md`（建立帳密檔、改 `.env`、上線順序）。
- **BREAKING（僅 Staging）**：不帶帳密存取兩個 Staging 網域會得到 401；前端 Staging 的後端位址改為同源。

## Capabilities

### New Capabilities

- `staging-access`：Staging 站台的對外存取控制——哪些請求要驗證、以什麼方式通過、哪些路徑例外，以及 Production 不受影響。

### Modified Capabilities

- `edge-sites`：新增「安裝前檢查帳密檔存在」需求。
- `frontend-delivery`：「自動部署到前端 Staging 並冒煙測試」改為帶帳密冒煙、驗證未帶帳密回 401、`apiBaseUrl` 指向前端 Staging 同源並經 `/api/` 讀到文章列表。
- `backend-delivery`：「Staging 冒煙測試」改為帶帳密，並驗證未帶帳密回 401。

## Impact

- `deploy/edge/web-staging.conf.template`、`deploy/edge/api-staging.conf.template`、`deploy/edge/install-sites.sh`、`deploy/edge/test-install-sites.sh`
- `deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh`、`deploy/ci/frontend-bundle.sh`（渲染需要新變數）
- `.drone.yml`：`cd-backend-staging`、`cd-frontend-staging` 的 bundle／smoke 步驟加入新 secret（高風險檔案，合併前需 user 看過 diff）
- `docs/cicd.md`、`docs/deployment-runbook.md`
- 主機手動步驟：建立帳密檔、改前端 Staging `.env`、在 Drone 新增兩個 secret
- 不改 `backend/`、`frontend/` 的程式碼；Production 的站台、`.env`、pipeline 都不變
