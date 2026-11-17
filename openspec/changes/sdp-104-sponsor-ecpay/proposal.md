# Proposal：贊助流程串接綠界 ECPay（SDP-104）

## Why

前台贊助按鈕（`frontend/components/front/sponsor-fab.jsx`）按下「贊助」後只顯示感謝畫面，不建立任何紀錄，也沒有真的付款（issue 寫的「寫進 localStorage」在 SDP-103 後已移除，但「沒有真的付款」仍然成立）。後端的 `POST /api/v1/payments` 與 `POST /api/v1/payments/ecpay/callback` 仍是回 501 的骨架。本次把前後端串起來，完成綠界全方位金流（AIO）的正式串接。

## What Changes

- **後端：建立付款**——`POST /api/v1/payments` 由 501 改為實作：驗證金額與留言、建立一筆 `pending` 贊助紀錄（以綠界 `MerchantTradeNo` 為 id）、計算 `CheckMacValue`，回傳前端要自動送出到綠界的表單（`action_url` 與 `params`）。
- **後端：綠界付款結果通知**——`POST /api/v1/payments/ecpay/callback` 由 501 改為實作：驗證 `CheckMacValue`、商店代號、金額，依 `RtnCode` 把紀錄轉為 `paid` 或 `failed`（可重複通知、冪等），以綠界要求的純文字 `1|OK`／`0|<原因>` 回應。
- **後端：查詢付款狀態**——新增公開的 `GET /api/v1/payments/{trade_no}`，只回傳交易編號、金額、狀態，供付款完成後的前台結果頁顯示。
- **後端：贊助紀錄欄位**——`Donation` 新增 `status`、`paid_at`、`ecpay_trade_no`、`payment_type`；既有紀錄以 `paid` 補值（GORM AutoMigrate，相容的 schema 變更）。`GET /api/v1/donations` 回應多出這些欄位（新增欄位，非破壞性）。
- **後端：設定**——新增 `ECPAY_RETURN_URL`（綠界通知後端的完整網址）與 `ECPAY_CLIENT_BACK_URL`（付款後返回前台結果頁的網址）。綠界設定不完整時，`POST /api/v1/payments` 回 503 `PAYMENT_UNAVAILABLE`，服務照常啟動。
- **功能開關（SDP-149 規則適用）**——新增 key `sponsor_payment`（「贊助付款（綠界）」，預設**停用**）。`POST /api/v1/payments` 在 `sponsor` 或 `sponsor_payment` 停用時回 403 `FEATURE_DISABLED`，帶有效管理員 JWT 時放行（`middleware.OptionalAuth`）。callback 與狀態查詢不受開關影響（綠界的通知與已付款者的結果頁不能被擋）。
- **前端：贊助視窗改走真實付款**——按「贊助」改呼叫 `POST /api/v1/payments`，以回傳內容建表單自動送出到綠界；有送出中與錯誤狀態。浮動按鈕需 `sponsor` 與 `sponsor_payment` 都啟用才對一般瀏覽者顯示，任一停用時管理員預覽仍可看到並帶「停用中・預覽」標示。
- **前端：付款結果頁**——新增 `/sponsor/result?trade_no=...`，查詢付款狀態並顯示「感謝」「處理中」「付款未完成」三種畫面；處理中時短暫輪詢。
- **前端：後台贊助管理**——KPI 只計入 `paid`，紀錄列表顯示狀態。
- 非破壞性：沒有既有欄位被移除或改名；兩個骨架端點原本只回 501，沒有使用方。

## Capabilities

### New Capabilities

（無）

### Modified Capabilities

- `payment-integration`：骨架需求改為正式串接——建立付款、綠界通知驗證與狀態轉換、付款狀態查詢、設定不完整時的處理、贊助紀錄的狀態欄位。
- `feature-flags`：登記表新增 `sponsor_payment`（預設停用）；前台贊助浮動按鈕改由 `sponsor` 與 `sponsor_payment` 共同控制。

## Impact

### 預計修改的檔案範圍（供判斷能否與其他 issue 並行）

後端（`backend/`）：

- `internal/config/config.go`、`internal/config/config_test.go`
- `internal/service/payment/payment.go`、`ecpay.go`、`ecpay_test.go`，新增 `checkmac.go`、`checkmac_test.go`
- 新增 `internal/service/payment_service.go`、`payment_service_test.go`
- `internal/service/feature_service.go`（登記表）、相關測試
- `internal/model/donation.go`
- `internal/repository/donation_repository.go`
- `internal/controller/payment_controller.go`（含測試）
- `internal/router/router.go`
- `internal/pkg/apperror/apperror.go`（新增 `PAYMENT_UNAVAILABLE`）
- `docs/`（`make swag` 重新產生）、`README.md`、`.env.example`、`docker-compose.yml`

前端（`frontend/`）：

- `lib/feature-flags.mjs`、`lib/feature-flags.test.mjs`
- 新增 `lib/sponsor.mjs`、`lib/sponsor.test.mjs`
- `store/lake-context.jsx`
- `components/front/sponsor-fab.jsx`、`components/front/front-shell.jsx`
- 新增 `components/front/sponsor-result-page.jsx`、`app/(front)/sponsor/result/page.jsx`
- `components/admin/admin-sponsor.jsx`
- `design-sync/manifest.json`（deviations）

與同時期改到 `feature_service.go` 登記表、`front-shell.jsx`、`lake-context.jsx`、`router.go` 的 issue 會衝突，不宜並行。

### 風險

- 命中高風險檔案樣式：金流（`backend/internal/service/payment/**`、`payment_controller.go`）、資料表結構（`backend/internal/model/**`）、`backend/docker-compose.yml`。整體判定**高風險**，合併前需 user 審批。
- 外部相依：綠界全方位金流 AIO（`/Cashier/AioCheckOut/V5`），不新增 Go／npm 套件。
- 部署：正式環境的 backend `.env` 要補 `ECPAY_*` 六個值，綠界後台要能連到 `ECPAY_RETURN_URL`。
