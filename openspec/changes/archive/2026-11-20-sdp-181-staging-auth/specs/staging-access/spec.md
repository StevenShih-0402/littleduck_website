# Spec Delta

## Purpose

規範 Staging 站台（前端 `staging.stevenshih-littleduck.com`、後端 `api.staging.stevenshih-littleduck.com`）的對外存取控制：沒有 Staging 帳密的人看不到尚未上線的內容，同時不影響 Staging 上的正常測試流程與 Production。

## ADDED Requirements

### Requirement: 前端 Staging 需 Basic Auth
前端 Staging 站台 SHALL 對所有 HTTPS 請求要求 HTTP Basic Auth；沒有帳密或帳密錯誤 MUST 回 401 並帶 `WWW-Authenticate: Basic` 標頭，不回傳任何頁面內容。HTTP（80 埠）的憑證驗證路徑與轉址行為不變。

#### Scenario: 未帶帳密
- **WHEN** 不帶帳密請求 `https://staging.stevenshih-littleduck.com/`
- **THEN** 回 401，回應含 `WWW-Authenticate: Basic`，內容不含網站頁面

#### Scenario: 帳密錯誤
- **WHEN** 以錯誤密碼請求 `https://staging.stevenshih-littleduck.com/blog`
- **THEN** 回 401

#### Scenario: 帳密正確
- **WHEN** 以正確帳密請求 `https://staging.stevenshih-littleduck.com/blog`
- **THEN** 回 2xx，內容與未加驗證前相同

#### Scenario: 憑證續約不受影響
- **WHEN** certbot 以 HTTP 請求 `http://staging.stevenshih-littleduck.com/.well-known/acme-challenge/<token>`
- **THEN** 不要求帳密，行為與加驗證前相同

### Requirement: 前端 Staging 同源轉送後端
前端 Staging 站台 SHALL 把 `/api/` 開頭的請求轉送到後端 Staging 服務，並沿用與頁面相同的存取控制；前端 Staging 的執行期設定 `apiBaseUrl` MUST 是 `https://staging.stevenshih-littleduck.com`，讓瀏覽器以同源方式呼叫後端。

#### Scenario: 經同源讀取文章
- **WHEN** 以正確帳密請求 `https://staging.stevenshih-littleduck.com/api/v1/articles`
- **THEN** 回 2xx，內容與後端 Staging 的 `/api/v1/articles` 相同

#### Scenario: 未帶帳密呼叫同源 API
- **WHEN** 不帶帳密也不帶通行 cookie 請求 `https://staging.stevenshih-littleduck.com/api/v1/articles`
- **THEN** 回 401

### Requirement: 通過驗證後發放通行 cookie
前端 Staging 站台 SHALL 在通過 Basic Auth 的回應中設定通行 cookie（HttpOnly、Secure、SameSite=Strict、僅限前端 Staging 主機）；之後帶有效通行 cookie 的請求 MUST 免再驗 Basic Auth，即使 `Authorization` 標頭是 Bearer 權杖。cookie 值錯誤時 MUST 回到 Basic Auth 驗證。

#### Scenario: 管理員以 JWT 呼叫後台 API
- **WHEN** 已通過 Basic Auth 的瀏覽器帶通行 cookie 與 `Authorization: Bearer <有效管理員 JWT>` 請求 `https://staging.stevenshih-littleduck.com/api/v1/admin/articles`
- **THEN** 請求送達後端並依 JWT 判定權限（有效則 2xx），不被 Basic Auth 擋下

#### Scenario: 偽造的通行 cookie
- **WHEN** 帶錯誤值的通行 cookie、不帶 Basic Auth 請求前端 Staging 任一路徑
- **THEN** 回 401

#### Scenario: Bearer 權杖不能單獨通過
- **WHEN** 只帶 `Authorization: Bearer <任意值>`、沒有有效通行 cookie 請求前端 Staging 任一路徑
- **THEN** 回 401

### Requirement: 後端 Staging 需 Basic Auth
後端 Staging 站台 SHALL 對所有 HTTPS 請求要求與前端 Staging 相同的 Basic Auth 帳密（含 `/health`、`/swagger/`），未通過 MUST 回 401；唯一例外是綠界付款結果通知 `POST /api/v1/payments/ecpay/callback`，該路徑 MUST 不要求帳密，由後端既有的 CheckMacValue 驗證把關。

#### Scenario: 直接讀取後端 Staging
- **WHEN** 不帶帳密請求 `https://api.staging.stevenshih-littleduck.com/api/v1/articles` 或 `/swagger/index.html`
- **THEN** 回 401

#### Scenario: 帶帳密讀取後端 Staging
- **WHEN** 以正確帳密請求 `https://api.staging.stevenshih-littleduck.com/health`
- **THEN** 回 2xx

#### Scenario: 綠界付款結果通知
- **WHEN** 綠界測試環境不帶帳密 `POST https://api.staging.stevenshih-littleduck.com/api/v1/payments/ecpay/callback`
- **THEN** 請求送達後端，由後端驗證 CheckMacValue 並回應，不被 Basic Auth 擋下

### Requirement: Staging 帳密不進版控
Staging 的 Basic Auth 帳密 SHALL 只以雜湊檔存在主機上、以明文只存在 Drone secret；版控中的站台設定與腳本 MUST NOT 含帳密或通行 cookie 的實際值。

#### Scenario: 檢查版控
- **WHEN** 在 repo 搜尋 Staging 帳密或通行 cookie 值
- **THEN** 找不到；站台樣板只含變數佔位與帳密檔路徑

### Requirement: Production 不受影響
Production 的前端與後端站台 MUST NOT 要求 Basic Auth，行為與本變更前相同。

#### Scenario: 存取 Production
- **WHEN** 不帶帳密請求 `https://stevenshih-littleduck.com/` 與 `https://api.stevenshih-littleduck.com/health`
- **THEN** 兩者都回 2xx
