# Tasks：贊助流程串接綠界 ECPay（SDP-104）

後端（`[BE]`）全部完成並定稿 `openapi.yaml` 後，前端（`[FE]`）才開工。後端驗證在 `backend/` 執行，前端驗證在 `frontend/` 執行。

## 1. 後端：設定、錯誤碼與功能開關

- [x] 1.1 [BE] `internal/config/config.go` 的 `ECPayConfig` 新增 `ReturnURL`（`ECPAY_RETURN_URL`）、`ClientBackURL`（`ECPAY_CLIENT_BACK_URL`）與 `Configured()`（MerchantID、HashKey、HashIV、ReturnURL、ClientBackURL 皆非空才為 true），不加入 `requiredVars`；`config_test.go` 補載入與 `Configured()` 的表格測試。驗證：`go test ./internal/config/...` — ECPayConfig 新增 ReturnURL、ClientBackURL 與 Configured()，補載入與表格測試
- [x] 1.2 [BE] `internal/pkg/apperror/apperror.go` 新增 `CodePaymentUnavailable = "PAYMENT_UNAVAILABLE"` 與 `PaymentUnavailable(message)`（503），補單元測試；README「錯誤碼」表加一列。驗證：`go test ./internal/pkg/...` — 新增 CodePaymentUnavailable 與 PaymentUnavailable()（503），README 錯誤碼表加一列
- [x] 1.3 [BE] `internal/service/feature_service.go` 登記表新增 `FeatureSponsorPayment = "sponsor_payment"`（名稱「贊助付款（綠界）」、說明「贊助視窗改走綠界真實付款；需與『贊助按鈕』同時啟用」、`Default: false`），補測試：空資料庫時 `sponsor_payment` 為 false、既有 `sponsor` 值不被覆寫；README「功能開關」段落列入新 key。驗證：`go test ./internal/service/...` — 登記表新增 sponsor_payment（預設停用），補預設與不覆寫測試，README 列入

## 2. 後端：綠界適配層

- [x] 2.1 [BE] 新增 `internal/service/payment/checkmac.go`：`CheckMacValue(params map[string]string, hashKey, hashIV string) string`（key 不分大小寫排序、.NET 風格 URL encode 與還原、轉小寫、SHA256、轉大寫）與 `VerifyCheckMac(...) bool`（不分大小寫、`crypto/subtle` 常數時間比較）；`checkmac_test.go` MUST 以綠界官方文件的範例參數與期望值驗證，並涵蓋空白、`- _ . ! * ( )` 等特殊字元與驗證失敗。（變異測試目標：`backend/internal/service/payment/checkmac.go`）驗證：`go test ./internal/service/payment/...` — 新增 CheckMacValue／VerifyCheckMac，以綠界官方文件範例（期望值吻合）與特殊字元、驗證失敗測試
- [x] 2.2 [BE] 改寫 `internal/service/payment/payment.go` 與 `ecpay.go`：`CreatePaymentInput` 加 `TradeDate`、`ItemName`、`ClientBackURL`；`CreatePaymentResult` 改為 `ActionURL` 與 `Params map[string]string`；`ECPayService.CreatePayment` 依 design D2 組 `params`（`ChoosePayment` 預設 `Credit`、`EncryptType=1`）並加 `CheckMacValue`，`ECPAY_ENV=production` 用正式網址、其餘用 stage 網址；`VerifyCallback` 驗 `CheckMacValue` 與 `MerchantID`，回傳可區分兩種失敗的錯誤；加 `Configured()` 委派。`ecpay_test.go` 改寫（移除「回 501」的舊測試），涵蓋 stage／production 網址、params 欄位完整、驗證成功與兩種失敗。不得在日誌輸出 HashKey、HashIV。驗證：`go test ./internal/service/payment/...` — ECPayService 產生 AIO 表單欄位與 CheckMacValue、依 ECPAY_ENV 選網址、VerifyCallback 區分檢查碼與商店代號失敗，測試改寫

## 3. 後端：資料表與 repository

- [x] 3.1 [BE] `internal/model/donation.go` 新增 `Status`（`gorm:"size:16;not null;default:paid"`，json `status`）、`PaidAt *time.Time`（`paid_at`）、`EcpayTradeNo`（`size:32`，`ecpay_trade_no`）、`PaymentType`（`size:32`，`payment_type`）與狀態常數 `DonationStatusPending/Paid/Failed`；同一項在 `internal/repository/donation_repository.go` 新增 `Create`、`FindByID`（查無回 `ErrNotFound`）、`MarkPaid(id, paidAt, ecpayTradeNo, paymentType) (bool, error)`（`WHERE id=? AND status<>'paid'`）、`MarkFailed(id) (bool, error)`（`WHERE id=? AND status='pending'`）。本專案 repository 層目前沒有單元測試、也沒有 sqlite／sqlmock 相依，本項不新增測試套件相依；條件更新的語意由 4.1 的 service 測試（mock 回傳影響列數 0／1）涵蓋，SQL 本身在 Staging 實際付款時驗證。檔頭註解移除「交易寫入屬後續 change」。驗證：`go build ./...`、`go vet ./...`，並在 handoff 註明：相容的 schema 變更，空資料庫與既有資料庫啟動由執行驗證觀察點 1 確認 — Donation 新增 status／paid_at／ecpay_trade_no／payment_type；repository 新增 Create、FindByID、MarkPaid、MarkFailed（條件更新）

## 4. 後端：付款業務邏輯

- [x] 4.1 [BE] 新增 `internal/service/payment_service.go`：`PaymentService` interface 含 `Create(in PaymentInput, asAdmin bool) (*CreatePaymentView, error)`、`HandleECPayCallback(params map[string]string) string`、`Status(tradeNo string) (*PaymentStatusView, error)`；注入 `DonationRepository`、`FeatureChecker`、`payment.PaymentProvider`、`ECPayConfig`（或必要欄位）、`now`、亂數來源。規則依 spec「建立贊助付款」「綠界付款結果通知」「查詢付款狀態」與 design D4、D5、D8、D9：金額上下限常數（10～20000，只檢查上下限、不比對預設按鈕金額）、`ChoosePayment` 固定 `Credit`、`message` ≤100、`name` ≤50（rune 計數，空白記「匿名訪客」）、非管理員時 `sponsor` 與 `sponsor_payment` 皆需啟用（讀取失敗回 500）、設定不完整回 503、交易編號 `LD`+`yyMMddHHmmss`+6 碼 `crypto/rand` 英數、`Date` 用 UTC+8 的 `2006 / 01 / 02`、`ClientBackURL` 附 `trade_no`；通知處理順序依 design「通知處理判斷」圖。`payment_service_test.go` 用 mock 覆蓋每個 scenario（含開關各組合、管理員放行、上下限邊界值、重複通知、`paid` 不被改回、`SimulatePaid=1`、金額不符、查無交易、資料庫錯誤回 `0|Internal Error`）。（變異測試目標：`backend/internal/service/payment_service.go`）驗證：`go test ./internal/service/...` — 新增 PaymentService（Create／HandleECPayCallback／Status），mock repository 測試涵蓋開關組合、邊界值、通知判斷順序與冪等

## 5. 後端：HTTP 層、文件與本機設定

- [x] 5.1 [BE] 改寫 `internal/controller/payment_controller.go`：`Create`（綁 JSON、從 `middleware.ContextKeyAdmin` 判斷管理員、回 `response.OK`）、`ECPayCallback`（`c.Request.ParseForm` 後轉 `map[string]string`，一律 `c.String(200, ...)` 純文字）、新增 `Status`（`GET /payments/:trade_no`）；更新 swag 註解（含 403、503、`text/plain`）。`internal/router/router.go`：`POST /payments` 掛 `middleware.OptionalAuth(authSvc)`，新增 `GET /payments/:trade_no`，組裝 `PaymentService`。補 controller 測試（httptest）：建立成功、403、503、callback 回純文字、狀態 404。驗證：`go test ./internal/controller/... ./internal/router/...` — PaymentController 實作 Create／ECPayCallback（純文字）／Status，router 掛 OptionalAuth 與 GET /payments/:trade_no，補 controller 與 router 測試
- [x] 5.2 [BE] 執行 `make swag` 重新產生 `docs/`；更新 `README.md`（端點表三列、環境變數表新增 `ECPAY_RETURN_URL`、`ECPAY_CLIENT_BACK_URL` 並把 `ECPAY_*` 說明改為「設定不完整時建立付款回 503」、正式值只放部署 `.env`、管理員預覽會建立真實訂單）、`.env.example`（新增兩個變數與註解）、`docker-compose.yml`（`ECPAY_*` 改為 `${ECPAY_X:-<預設>}`，預設依 design D8：綠界官方文件公開的 stage 測試特店、本機 ReturnURL 與 ClientBackURL；Staging／Production 的值不寫進 repo，README 註明由部署 `.env` 提供）。驗證：`gofmt -l .` 無輸出、`go vet ./...`、`go test ./...`、`docker compose -f docker-compose.yml config` 成功 — make swag 重新產生 docs/；README、.env.example、docker-compose（stage 測試特店預設）更新；gofmt／vet／全部測試／compose config 通過
- [x] 5.3 [BE] 依實作就地定稿 `openspec/changes/sdp-104-sponsor-ecpay/openapi.yaml`（`info.description` 改為「已由後端 Agent 依實作定稿」，與 swag 產物的路徑、狀態碼、欄位一致）。驗證：逐一比對 `docs/swagger.yaml` 的 `/api/v1/payments*` 與 `/api/v1/donations` — openapi.yaml 定稿（info.description 改為已定稿，callback 補設定不完整行為），與 swagger.yaml 逐項比對一致

## 6. 前端：純函式與資料層

- [x] 6.1 [FE] `lib/feature-flags.mjs` 的 `FEATURES` 新增 `SPONSOR_PAYMENT: "sponsor_payment"`；`lib/feature-flags.test.mjs` 補「兩個開關組合 × 是否管理員」的 `gateMode` 測試（含 `sponsor` 啟用、`sponsor_payment` 停用時一般訪客為 hide、管理員為 preview）。驗證：`npm run test:design` — FEATURES 新增 SPONSOR_PAYMENT，補雙開關 gateMode 組合測試
- [x] 6.2 [FE] 新增 `lib/sponsor.mjs` 與 `lib/sponsor.test.mjs`：`SPONSOR_AMOUNT_MIN/MAX`（與後端常數同值）、`validateSponsorAmount`、`resultView(status)`（`paid`/`pending`/`failed`/其他→`missing`）、`nextPollDelay(attempt)`（有限次數後回 `null`）。（變異測試目標：`frontend/lib/sponsor.mjs`）驗證：`npm run test:design`、`npm run typecheck` — 新增 lib/sponsor.mjs（金額驗證、resultView、nextPollDelay）與測試
- [x] 6.3 [FE] `store/lake-context.jsx` 新增 `createPayment({ amount, message })`（`apiPost("/payments", ...)`）與 `getPaymentStatus(tradeNo)`（`apiGet("/payments/" + encodeURIComponent(tradeNo))`），加入 context value。驗證：`npm run lint`、`npm run typecheck` — lake-context 新增 createPayment 與 getPaymentStatus

## 7. 前端：畫面

- [x] 7.1 [FE] `components/front/sponsor-fab.jsx`：「贊助」改為呼叫 `createPayment`，成功後以 `document.createElement("form")`（`method="post"`、`action=action_url`）逐一加入 hidden input（以屬性設值，不拼 HTML 字串）並 `submit()`；送出中按鈕停用並顯示處理中；失敗在視窗內顯示錯誤訊息且可重試；金額先以 `validateSponsorAmount` 擋下並提示上下限；移除原本計時關閉的假感謝畫面與 SDP-104 的 TODO 註解。`components/front/front-shell.jsx`：以外層 `FeatureGate flag={FEATURES.SPONSOR}`、內層 `FeatureGate flag={FEATURES.SPONSOR_PAYMENT}` 包住 `SponsorFab`，`preview` 取兩層任一為真。驗證：`npm run lint`、`npm run typecheck`、`npm run build` — SponsorFab 改呼叫 createPayment 並以隱藏表單導向綠界，front-shell 雙層 FeatureGate
- [x] 7.2 [FE] 新增 `app/(front)/sponsor/result/page.jsx`（以 `React.Suspense` 包住）與 `components/front/sponsor-result-page.jsx`：讀 `trade_no` 查詢參數，依 `resultView` 顯示——`paid`：「感謝你的蜂蜜檸檬！」與 `NT$ <amount>`；`pending`：處理中並依 `nextPollDelay` 重查，用盡後提示稍後再確認；`failed`／`missing`：「付款未完成」與回首頁連結；查詢失敗不顯示全頁錯誤。沿用前台主題 class 與 `HoneyLemonIcon`。驗證：`npm run lint`、`npm run typecheck`、`npm run build` 成功且輸出含 `/sponsor/result` — 新增 /sponsor/result 頁與 SponsorResultPage（paid／pending 輪詢／failed／missing），build 輸出含 /sponsor/result
- [x] 7.3 [FE] `components/admin/admin-sponsor.jsx`：KPI（累計、本月、平均、最近一筆）只計入 `status === "paid"`；紀錄列表新增狀態欄（沿用 `status-pill.jsx` 或同風格標示 `pending`／`paid`／`failed`）。驗證：`npm run lint`、`npm run typecheck` — 後台贊助 KPI 只計 paid，紀錄列表新增狀態標籤
- [x] 7.4 [FE] `design-sync/manifest.json`：為 `components/front/sponsor-fab.jsx`、`components/front/front-shell.jsx`、`components/admin/admin-sponsor.jsx` 各新增一筆 `issue: "SDP-104"` 的 deviation（說明真實付款流程、雙開關、只計 paid），不改各設計檔雜湊；結果頁不加入 manifest。驗證：`npm run test:design` — manifest 為 front-shell、sponsor-fab、admin-sponsor 各新增 SDP-104 deviation

## 8. 整合檢查

- [x] 8.1 [FE] 前端完整檢查：`npm run lint`、`npm run typecheck`、`npm run test:design`、`npm run build`、`npm audit --audit-level=high` 全數通過，並在 handoff 記錄結果。驗證：上述指令 — lint 0 error（9 個既有 img warning）、typecheck、test:design 38 通過、build 含 /sponsor/result、audit --audit-level=high 通過

## Workflow follow-up

- 合併後依 design「Migration Plan」在 Staging 填 `ECPAY_*`、以管理員預覽完成一次 stage 刷卡，確認紀錄轉 `paid` 後再決定正式環境啟用 `sponsor_payment`。
- 功能穩定後另開 issue 移除 `sponsor_payment` 開關。
