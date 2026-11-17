# payment-integration Specification

## Purpose
提供綠界科技金流模組的骨架，包含設定、介面與 callback 端點預留，讓後續 change 能在此基礎上完成正式的金流串接流程。

## Requirements

### Requirement: 綠界金流設定
系統 SHALL 從環境變數讀取綠界全方位金流（AIO）設定：MerchantID、HashKey、HashIV、環境、付款通知網址 ReturnURL、付款後返回網址 ClientBackURL；程式碼中不得寫死正式環境金鑰。環境為 `production` 時 MUST 送往綠界正式環境，其餘一律送往測試環境（stage）。設定不完整時服務 MUST 照常啟動，但建立付款回傳 503 `PAYMENT_UNAVAILABLE`。

#### Scenario: 金流設定由環境變數載入
- **WHEN** 設定 `ECPAY_MERCHANT_ID`、`ECPAY_HASH_KEY`、`ECPAY_HASH_IV`、`ECPAY_RETURN_URL`、`ECPAY_CLIENT_BACK_URL` 後啟動
- **THEN** ECPay service 以這些值初始化，程式碼中無寫死金鑰

#### Scenario: 設定不完整
- **WHEN** 上述任一必要值為空，且 `sponsor`、`sponsor_payment` 皆啟用時呼叫 `POST /api/v1/payments`
- **THEN** 回傳 503 與統一錯誤格式 `PAYMENT_UNAVAILABLE`，不建立任何贊助紀錄；其他端點不受影響

#### Scenario: 依環境選擇綠界網址
- **WHEN** `ECPAY_ENV` 為 `production`
- **THEN** 建立付款回傳的 `action_url` 為綠界正式環境的 AIO 網址；`ECPAY_ENV` 為其他值或未設定時為測試環境網址

### Requirement: 後台贊助紀錄清單
系統 SHALL 提供需管理員身分的唯讀端點，列出贊助紀錄（日期、金額、贊助者名稱、備註、付款狀態 `status`、付款完成時間 `paid_at`、綠界交易編號 `ecpay_trade_no`、付款方式 `payment_type`）。`status` 為 `pending`、`paid`、`failed` 之一；本變更之前就存在的紀錄 MUST 視為 `paid`。

#### Scenario: 後台讀取贊助紀錄
- **WHEN** 管理員以有效權杖呼叫贊助紀錄清單端點
- **THEN** 回傳 200 與贊助紀錄清單，每筆含 `status`

#### Scenario: 未授權讀取
- **WHEN** 未帶有效權杖呼叫贊助紀錄清單端點
- **THEN** 回傳 401，不回傳紀錄

#### Scenario: 既有紀錄升級後的狀態
- **WHEN** 以本變更之前已有贊助紀錄的資料庫啟動服務
- **THEN** 服務啟動成功，既有紀錄的 `status` 為 `paid`

### Requirement: 建立贊助付款
系統 SHALL 提供公開的 `POST /api/v1/payments`，以 JSON `{amount, message?, name?}` 建立一筆 `pending` 贊助紀錄，並回傳 200 與 `{trade_no, action_url, params}`。`trade_no` 是本系統產生、全站唯一、最長 20 碼英數字的綠界 `MerchantTradeNo`，同時是贊助紀錄的 id；`params` 是前端以表單 POST 到 `action_url` 的全部欄位（含 `CheckMacValue`）。非管理員時 `sponsor` 與 `sponsor_payment` MUST 皆啟用。

#### Scenario: 建立付款成功
- **WHEN** `sponsor` 與 `sponsor_payment` 皆啟用、綠界設定完整，訪客送出 `{"amount": 100, "message": "加油"}`
- **THEN** 回傳 200，`data.params.TotalAmount` 為 `"100"`、`data.params.MerchantTradeNo` 等於 `data.trade_no`，並新增一筆 `status` 為 `pending`、`amt` 為 100、`note` 為「加油」的贊助紀錄

#### Scenario: 通知與返回網址
- **WHEN** 建立付款成功
- **THEN** `params.ReturnURL` 等於 `ECPAY_RETURN_URL`，`params.ClientBackURL` 為 `ECPAY_CLIENT_BACK_URL` 加上 `trade_no` 查詢參數

#### Scenario: 未提供名稱
- **WHEN** 請求沒有 `name` 或只有空白
- **THEN** 贊助紀錄的名稱為「匿名訪客」

#### Scenario: 金額不合法
- **WHEN** `amount` 缺少、非整數、小於 10 或大於 20000
- **THEN** 回傳 400 `BAD_REQUEST`，不建立紀錄

#### Scenario: 金額邊界與非預設金額
- **WHEN** `amount` 為 10、20000，或任何不等於後台預設按鈕的介於其間整數
- **THEN** 照常建立付款

#### Scenario: 只開信用卡
- **WHEN** 建立付款成功
- **THEN** `params.ChoosePayment` 為 `Credit`

#### Scenario: 文字過長
- **WHEN** `message` 超過 100 字或 `name` 超過 50 字
- **THEN** 回傳 400 `BAD_REQUEST`，不建立紀錄

#### Scenario: 功能停用
- **WHEN** `sponsor` 或 `sponsor_payment` 任一停用，未帶有效管理員權杖呼叫
- **THEN** 回傳 403 `FEATURE_DISABLED`，不建立紀錄

#### Scenario: 管理員預覽
- **WHEN** `sponsor_payment` 停用，請求帶有效管理員 `Authorization: Bearer` 權杖
- **THEN** 略過功能開關檢查，照常建立付款

#### Scenario: 權杖無效視同訪客
- **WHEN** 請求帶無效或過期權杖，且 `sponsor_payment` 停用
- **THEN** 回傳 403 `FEATURE_DISABLED`，不回 401

#### Scenario: CheckMacValue 正確
- **WHEN** 以綠界官方文件的 CheckMacValue 範例參數與 HashKey／HashIV 計算
- **THEN** 結果與官方文件的範例值一致（SHA256、大寫）

### Requirement: 綠界付款結果通知
系統 SHALL 在 `POST /api/v1/payments/ecpay/callback` 接收綠界以 `application/x-www-form-urlencoded` 送來的付款結果，一律以 HTTP 200、`text/plain` 回應：處理成功（含重複通知）回 `1|OK`，驗證失敗回 `0|<原因>`。此端點 MUST NOT 受功能開關影響；只有 CheckMacValue、MerchantID、金額都相符的通知能改變紀錄狀態，且 `paid` 的紀錄 MUST NOT 被改回其他狀態。

#### Scenario: 付款成功通知
- **WHEN** 綠界送來 CheckMacValue 正確、`RtnCode=1`、`TradeAmt` 與紀錄相符的通知
- **THEN** 回應 `1|OK`，該紀錄 `status` 為 `paid`，`paid_at`、`ecpay_trade_no`、`payment_type` 有值

#### Scenario: 重複通知
- **WHEN** 同一筆成功通知再送一次
- **THEN** 回應 `1|OK`，紀錄維持 `paid`，`paid_at` 不變

#### Scenario: 檢查碼錯誤
- **WHEN** 通知的 `CheckMacValue` 與重新計算的值不符
- **THEN** 回應 `0|CheckMacValue Error`，紀錄狀態不變

#### Scenario: 商店代號不符
- **WHEN** 通知 CheckMacValue 正確但 `MerchantID` 與設定不同
- **THEN** 回應 `0|MerchantID Error`，紀錄狀態不變

#### Scenario: 查無交易
- **WHEN** 通知的 `MerchantTradeNo` 沒有對應的贊助紀錄
- **THEN** 回應 `0|Order Not Found`

#### Scenario: 金額不符
- **WHEN** 通知的 `TradeAmt` 與紀錄的 `amt` 不同
- **THEN** 回應 `0|Amount Mismatch`，紀錄狀態不變

#### Scenario: 付款失敗通知
- **WHEN** 綠界對 `pending` 紀錄送來驗證通過、`RtnCode` 不為 1 的通知
- **THEN** 回應 `1|OK`，紀錄 `status` 為 `failed`

#### Scenario: 已付款不被覆寫為失敗
- **WHEN** 對已是 `paid` 的紀錄送來 `RtnCode` 不為 1 的通知
- **THEN** 回應 `1|OK`，紀錄維持 `paid`

#### Scenario: 模擬付款
- **WHEN** 通知帶 `SimulatePaid=1`
- **THEN** 回應 `1|OK`，紀錄狀態不變

#### Scenario: 功能停用時仍處理通知
- **WHEN** `sponsor_payment` 已停用，綠界送來合法的付款成功通知
- **THEN** 照常處理並回應 `1|OK`

### Requirement: 查詢付款狀態
系統 SHALL 提供公開的 `GET /api/v1/payments/{trade_no}`，回傳 `{trade_no, amount, status}`；MUST NOT 回傳贊助者名稱、留言或綠界交易細節。此端點不受功能開關影響。

#### Scenario: 查詢存在的交易
- **WHEN** 以剛建立的 `trade_no` 查詢
- **THEN** 回傳 200，`data.status` 為 `pending`，`data.amount` 為建立時的金額

#### Scenario: 查無交易
- **WHEN** 以不存在的 `trade_no` 查詢
- **THEN** 回傳 404 `NOT_FOUND`

### Requirement: 前台贊助付款流程
前台贊助視窗按下「贊助」後 SHALL 呼叫建立付款端點，成功時以回傳的 `action_url` 與 `params` 建立表單並自動送出，把瀏覽器導向綠界付款頁；送出期間按鈕 MUST 不可重複點擊，失敗時 MUST 在視窗內顯示錯誤並可重試。前台 SHALL 提供 `/sponsor/result` 結果頁，依 `trade_no` 查詢並顯示付款結果。

#### Scenario: 導向綠界
- **WHEN** 訪客在贊助視窗選擇金額並按「贊助」，建立付款成功
- **THEN** 瀏覽器以 POST 送出表單到 `action_url`，欄位與 `params` 完全一致

#### Scenario: 建立失敗
- **WHEN** 建立付款回傳 400、403、503 或網路錯誤
- **THEN** 視窗顯示錯誤訊息、按鈕恢復可點，不導向綠界、不顯示感謝畫面

#### Scenario: 付款完成返回
- **WHEN** 訪客從綠界返回 `/sponsor/result?trade_no=<已付款的交易>`
- **THEN** 頁面顯示「感謝你的蜂蜜檸檬！」與 NT$ 金額

#### Scenario: 通知尚未抵達
- **WHEN** 返回結果頁時交易仍為 `pending`
- **THEN** 頁面顯示處理中，並在有限次數內重新查詢；仍未完成時提示稍後再確認

#### Scenario: 缺少或錯誤的交易編號
- **WHEN** 開啟 `/sponsor/result` 未帶 `trade_no`，或交易不存在或為 `failed`
- **THEN** 頁面顯示付款未完成與回首頁連結，不顯示全頁錯誤

### Requirement: 後台贊助統計只計入已付款
後台贊助管理的累計、本月、平均、最近一筆等統計 SHALL 只計入 `status` 為 `paid` 的紀錄；紀錄列表 SHALL 顯示每筆的付款狀態。

#### Scenario: 待付款不列入統計
- **WHEN** 資料庫有一筆 `pending` 與一筆 `paid` 紀錄
- **THEN** 累計贊助只計入 `paid` 那筆，列表兩筆都顯示並標示各自狀態
