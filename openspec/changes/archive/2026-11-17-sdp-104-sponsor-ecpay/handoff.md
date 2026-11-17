<!-- handoff.md 範本：複製到 openspec/changes/<change-id>/handoff.md。「狀態」區塊就地改值、永遠反映最新進度；下方各角色紀錄只追加，不改前面的內容。 -->
# Handoff：sdp-104-sponsor-ecpay

## 狀態

<!-- 就地改值，不追加。中斷後重新派工時，開發 Agent 以這裡判斷從哪裡接續；不寫 commit SHA（以 git log --grep "wip(<ISSUE-ID>): <任務編號>" 對應）。
  階段由誰寫入：proposal（architect 建立；主對話提交提案時更新「最後完成」）→ backend（後端 Agent 開工）
  → frontend（後端 Agent 交棒，或沒有 [BE] 任務時前端 Agent 開工）→ integrating（最後一個開發 Agent 交接）→ reviewing（integrator 檢查通過；主對話做 PR 前審查）
  → pr-checking（integrator 收尾重切並開 PR 後；主對話做 PR 中檢查）
  → awaiting-approval（只有高風險：主對話等待 user 回覆「合併」）→ archived（integrator 歸檔；之後由 integrator 合併）。
  integrator 檢查、PR 前審查、CI 或 PR 中檢查退回時改回被退回那一端的 backend／frontend（退回 architect 為 proposal）。
  合併隊列失敗不改階段，停下交給 user。 -->

- **階段**：archived <!-- proposal | backend | frontend | integrating | reviewing | pr-checking | awaiting-approval | archived -->
- **最後完成**：integrator 歸檔
- **下一步**：合併
- **暫停原因**：無
- **更新**：integrator（Teshima）— 2026-11-17（歸檔）

---

## architect（Hikaru）— 2026-11-17 15:30

- **做了什麼**：
  - 建立變更目錄：proposal、specs（`payment-integration`、`feature-flags` 兩份 MODIFIED/ADDED delta）、design（資料流圖、通知處理判斷圖、狀態圖、前後端互動時序圖、執行驗證觀察點 9 項、變異測試目標 3 項）、`openapi.yaml` 草稿、tasks（`[BE]` 10 項、`[FE]` 8 項）。
  - 確認問題存在：`sponsor-fab.jsx` 的 `confirm` 只顯示假的感謝畫面（issue 說的 localStorage 寫入已在 SDP-103 移除），`POST /payments` 與 callback 仍回 501。
  - 功能開關規則適用：新增 `sponsor_payment`（預設停用），`POST /payments` 掛 `OptionalAuth`、停用回 403；前台雙開關包住贊助按鈕。
- **未完成**：無（Linear SDP-104 已補 `code: sponsor-ecpay`、回填調查結果與實作決策到 description，狀態轉 In Progress；引言兩區塊留佔位待實作後填）。實作交給開發 Agent。
- **風險**：
  - 高風險：命中金流（`service/payment/**`、`payment_controller.go`）、資料表結構（`model/donation.go`）、`backend/docker-compose.yml`。合併前需 user 審批。
  - schema 變更相容：`donations` 新增 4 欄，`status` 預設 `paid` 回填既有列。
  - 本機無法收到綠界真實通知（ReturnURL 不可達），付款成功路徑只能在 Staging 實測。
- **API 契約變動**：`POST /api/v1/payments`（501 → 200／400／403／500／503）、`POST /api/v1/payments/ecpay/callback`（501 → 200 text/plain）、新增 `GET /api/v1/payments/{trade_no}`、`GET /api/v1/donations` 新增 4 欄、`GET /api/v1/features` 新增 key。皆非破壞性。
- **退回**：無
- **待確認問題**（已於 2026-11-17 由 user 回覆，見下一節「architect（Hikaru）— 修訂」）：
  - **Q1 綠界憑證來源**：本機 `backend/docker-compose.yml` 是否以綠界官方公開的 stage 測試特店作為 `ECPAY_*` 預設值（暫定：是，讓執行驗證能測建立付款）？Staging／Production 的特店代號、HashKey、HashIV 由 user 自行放進部署 `.env`，確認已向綠界申請正式特店。
  - **Q2 贊助金額規則**：自訂金額上下限（暫定 10～20000 元）；預設三個金額按鈕沿用後台「贊助視窗設定」，是否要後端拒絕不在按鈕上的金額（暫定：不拒絕，只檢查上下限）？
  - **Q3 付款方式**：`ChoosePayment` 只開信用卡（暫定 `Credit`），或開放 `ALL`（含 ATM／超商代碼，付款完成時間會延後、結果頁會較常停在「處理中」）？
  - **Q4 功能開關取捨**：新增 `sponsor_payment`（預設停用）並讓贊助按鈕需兩個開關都啟用——部署後到你在後台啟用前，前台贊助按鈕會暫時消失（暫定採用）。或改為沿用既有 `sponsor` 一個開關（部署即對外收款，不符合「新功能預設停用」規則）？
  - **Q5 付款後導向**：付款完成後導回新頁面 `/sponsor/result?trade_no=...` 顯示感謝／處理中／未完成（暫定採用）；或改為導回首頁不顯示結果？另確認 Staging／Production 的前台與 API 對外網址，用來設定 `ECPAY_CLIENT_BACK_URL`、`ECPAY_RETURN_URL`。

## architect（Hikaru）— 修訂 — 2026-11-17 16:00

- **做了什麼**：依 user 對 Q1～Q5 的回覆（全部採用暫定做法）定案，清除待確認問題：
  - Q1：`backend/docker-compose.yml` 預設用綠界官方公開的 stage 測試特店；Staging／Production 的特店資料由 user 放進部署 `.env`（design D8、tasks 5.2）。
  - Q2：金額 10～20000 元，只檢查上下限，不比對預設按鈕金額（design D9、spec「金額不合法」「金額邊界與非預設金額」、tasks 4.1、openapi）。
  - Q3：`ChoosePayment` 固定 `Credit`（design D2、spec「只開信用卡」、tasks 4.1）。
  - Q4：新增 `sponsor_payment`（預設停用），兩個開關都啟用才顯示贊助按鈕（design D6、Risks）。
  - Q5：付款後導回 `/sponsor/result?trade_no=...`；user 沒有提供對外網址，`ECPAY_RETURN_URL`、`ECPAY_CLIENT_BACK_URL` 只以環境變數提供，部署時由 user 設定（design D8）。
  - 移除執行驗證觀察點「若 Q1 不提供預設」的分支說明。
- **未完成**：無
- **風險**：同上一節；Staging 實測前需由 user 在部署 `.env` 填好 `ECPAY_*` 六個值。
- **API 契約變動**：`openapi.yaml` 的 `amount` 說明改為確定值 10～20000，端點與欄位不變。
- **退回**：無（非退回，為 user 回覆待確認問題後的定案）

## backend（Yasumi）— 2026-11-17 21:59

- **做了什麼**：
  - 1.1～1.3：`ECPayConfig` 新增 `ReturnURL`／`ClientBackURL`／`Configured()`；`apperror.PaymentUnavailable`（503 `PAYMENT_UNAVAILABLE`）；功能開關登記 `sponsor_payment`（預設停用）。
  - 2.1～2.2：`payment/checkmac.go`（`CheckMacValue`、`VerifyCheckMac`，以綠界官方文件範例期望值驗證吻合）；`ECPayService` 產生 AIO 表單欄位與檢查碼、依 `ECPAY_ENV` 選 stage／正式網址、`VerifyCallback` 以 `ErrCheckMacMismatch`／`ErrMerchantMismatch` 區分失敗。`~` 依綠界（.NET／PHP urlencode）編為 `%7e`，Go 的 `QueryEscape` 預設不編，已特別處理。
  - 3.1：`Donation` 新增 `status`／`paid_at`／`ecpay_trade_no`／`payment_type`；repository 新增 `Create`（主鍵衝突回 `ErrDuplicate`）、`FindByID`、`MarkPaid`、`MarkFailed`（條件更新）。
  - 4.1：`service/payment_service.go`（`Create`／`HandleECPayCallback`／`Status`）與 mock 測試。
  - 5.1～5.3：`PaymentController`、router（`POST /payments` 掛 `OptionalAuth`、新增 `GET /payments/:trade_no`）、`make swag`、README、`.env.example`、`docker-compose.yml`（本機預設為綠界官方文件公開的 stage 測試特店 `3002607`）、`openapi.yaml` 定稿。
- **未完成**：無。
- **驗證**：`gofmt -l .` 無輸出、`go vet ./...`、`go test ./...` 全部通過、`docker compose -f docker-compose.yml config` 成功。另用臨時 PostgreSQL 16 容器實測（測試檔已刪除、未提交）：既有 `donations` 表（無新欄位、含 1 筆列）AutoMigrate 兩次成功，既有列 `status` 補為 `paid`；`MarkPaid`／`MarkFailed` 的條件更新（paid 不被改回、failed 可轉 paid、重複不覆蓋）與主鍵衝突偵測在真實 SQL 下行為正確。
- **schema 變更**：**相容**——`donations` 新增 4 欄（`status` 預設 `paid`），既有資料庫與空資料庫皆可啟動；舊版程式不讀新欄位，回滾安全。
- **規格同步**：專案無 `openspec/spec-sync.yaml`，不適用。
- **風險**：
  - 高風險檔案：`service/payment/**`、`payment_controller.go`、`model/donation.go`、`docker-compose.yml`，合併前需 user 審批。
  - 本機無法收到綠界真實通知（ReturnURL 不可達），付款成功路徑只能在 Staging 實測；`donation_repository.go` 的條件更新沒有單元測試（專案 repository 層無測試相依），僅上述臨時實測。
  - callback 在綠界設定不完整時一律回 `0|CheckMacValue Error`（避免以空金鑰簽章通過檢查）。
  - 管理員預覽會建立真實訂單（README 已註明）。
- **API 契約變動（openapi.yaml 定稿）**：實作與草稿的路徑、狀態碼、欄位一致，只調整文字：`info.description` 改為「已由後端 Agent 依實作定稿」；callback 描述補「設定不完整時回 `0|CheckMacValue Error`」。無欄位增減。`GET /donations` 每筆新增 `status`、`paid_at`（可為 null）、`ecpay_trade_no`、`payment_type`；`GET /features` 多 `sponsor_payment`（預設 false）。前端注意：`amount` 上下限常數為 10／20000；建立付款成功回 200（不是 201）；`params` 值全為字串，原樣放進 hidden input。
- **退回**：無

## frontend（Fujimori）

- **做了什麼**：
  - 6.1～6.3：`FEATURES.SPONSOR_PAYMENT`（含雙開關 gateMode 組合測試）；`lib/sponsor.mjs`（金額 10～20000、`resultView`、`nextPollDelay`）與測試；`lake-context` 新增 `createPayment`、`getPaymentStatus`。
  - 7.1：`sponsor-fab.jsx` 呼叫 `createPayment`，以 DOM API 建立隱藏表單 POST 到 `action_url`（`params` 值原樣放入 hidden input）；送出中停用、失敗顯示錯誤可重試、金額先驗證；`front-shell.jsx` 以 `sponsor` 外層、`sponsor_payment` 內層 `FeatureGate` 包住，`preview` 取兩層任一。
  - 7.2：`/sponsor/result` 頁（Suspense）與 `sponsor-result-page.jsx`：paid 感謝、pending 依 `nextPollDelay` 輪詢、failed／無 trade_no 顯示未完成；查詢失敗不顯示全頁錯誤。
  - 7.3：`admin-sponsor.jsx` KPI 只計 paid（無 status 的舊資料視為 paid），紀錄列表新增狀態標籤。
  - 7.4：`design-sync/manifest.json` 為 front-shell、sponsor-fab、admin-sponsor 各加 SDP-104 deviation，未動設計檔雜湊。
- **未完成**：無
- **檢查結果**（在 `frontend/`）：`npm run lint` 0 error（9 個既有 `<img>` warning，非本次造成）、`npm run typecheck` 通過、`npm run test:design` 38 項通過、`npm run build` 成功且輸出含 `/sponsor/result`、`npm audit --audit-level=high` 通過（僅 7 個 moderate）。
- **仍靠 mock 的部分**：無（直接串接已定稿的契約）。
- **風險 / 已知限制**：
  - 結果頁查無交易（後端 404）時，`apiGet` 只拋出一般錯誤、無法區分狀態碼，因此會輪詢到用盡後顯示「還在確認付款結果」，而非 spec 的「付款未完成」；`trade_no` 缺失與 `failed` 則正確顯示「付款未完成」。若要精確對應，需 `lib/api.js` 的錯誤帶 status（屬可後續小改）。
  - `admin-sponsor.jsx` 原有的「本月」仍是寫死的 `2026 / 07` 篩選，非本次引入，未更動。
  - 付款成功路徑只能在 Staging 實測（本機收不到綠界通知）。
- **契約問題**：無，無需退回後端或 architect。
- **規格同步**：專案無 `openspec/spec-sync.yaml`，不適用。

## integrator（Teshima）— 檢查通過 — 2026-11-17

- **後端**（`backend/`）：`gofmt -l .` 無輸出；`go vet ./...` 通過；`go test -count=1 ./...` 全部 ok。
- **前端**（`frontend/`）：`npm ci`；`npm run lint` 0 error（9 個既有 img warning）；`npm run typecheck` 通過；`npm run test:design` 38 通過；`npm run build` 成功且含 `/sponsor/result`；`npm audit --audit-level=high` exit 0。
- **規格**：`openspec validate sdp-104-sponsor-ecpay --strict` 有效；`openspec validate --specs` 24 通過。
- **migration**：空資料庫啟動成功（執行驗證）；既有資料庫升級依 Yasumi 記錄的臨時 PostgreSQL 實測（本次未重測）。相容的 schema 變更。
- **變異測試**：略過（CLAUDE.md 未設定變異測試）；3 個目標未跑。
- **執行驗證（專案通用）**：health `{"status":"ok","database":"up"}`；`/api/v1/articles` 200；`/blog/a1` 200。
- **執行驗證（本次專屬 1～9）**：全部符合。1 donations status 皆為 `paid`；2 `sponsor_payment` false；3 `FEATURE_DISABLED`；4 stage 網址、`100`、CheckMacValue 長度 64；5 `pending`／`50`；6 `400`；7 `0|CheckMacValue Error`；8 `404`；9 `200`。（環境無 jq，改以 python3 解析 JSON，結果等價。）無截圖觀察點。
- **核對**：tasks.md 無未勾選項；worktree 乾淨且與 origin 一致；前後端 commit 無越界；proposal.md 僅「提案」commit 動過；無 `spec-sync.yaml`，規格同步不適用；無破壞性變更。
- **規格免同步聲明**：無
- **風險**：高風險檔案（金流、model、docker-compose）合併前需 user 審批；付款成功路徑僅能在 Staging 實測；既有資料庫 migration 本次未重測；結果頁查無交易會輪詢到用盡後顯示「還在確認付款結果」而非「付款未完成」（Fujimori 已知限制）。
- **退回**：無

---

## integrator（Teshima）— 收尾

- **重切前 SHA（OLD）**：`12ff23a4294f0cd9b1373bc0c1de47b1ab4d8577`
- **重切起點（FROM）**：`3ef23b10e3541cb7c95ddb970616cd86c7ffc9ae`（merge-base）
- **重切後 commit**：
  - `17348c6` feat(payment): 新增 CheckMacValue 驗章並改寫綠界適配層
  - `d72c96a` feat(api): 贊助改走綠界付款流程與功能開關
  - `8c7015c` feat(web): 贊助視窗串接綠界付款、付款結果頁與後台已付款統計
  - `e74703f` docs(openspec): 新增 SDP-104 贊助綠界付款變更提案與審查紀錄
- **tree 比對**：`git diff OLD HEAD` 為空
- **PR**：https://github.com/StevenShih-0402/littleduck_website/pull/20（附 3 條 FE PR 備註）
- **風險**：高風險檔案（金流、model、docker-compose）合併前需 user 審批；付款成功路徑僅能在 Staging 實測。
- **收尾 HEAD**：`e74703f9016bd3b70464102201f4dc41cb9470e1`

---

## integrator（Teshima）— 歸檔

- 歸檔至 `openspec/changes/archive/2026-11-17-sdp-104-sponsor-ecpay/`，同步規格 `feature-flags`（2 modified）、`payment-integration`（7 added、2 removed）。
- `openspec validate --specs`：24 passed, 0 failed。
