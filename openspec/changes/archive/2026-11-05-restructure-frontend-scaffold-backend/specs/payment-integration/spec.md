# payment-integration

## ADDED Requirements

### Requirement: 金流模組骨架（綠界）
系統 SHALL 定義 `PaymentProvider` interface（建立付款、驗證 callback）並提供讀取綠界設定（MerchantID、HashKey、HashIV、環境）的 ECPay 實作骨架；綠界金鑰 MUST 來自環境變數。正式的 CheckMacValue 計算與交易流程不在本次範圍。

#### Scenario: 金流設定由環境變數載入
- **WHEN** 設定 `ECPAY_MERCHANT_ID`、`ECPAY_HASH_KEY`、`ECPAY_HASH_IV` 後啟動
- **THEN** ECPay service 以這些值初始化，程式碼中無寫死金鑰

#### Scenario: 預留端點回應未實作
- **WHEN** 呼叫 `POST /api/v1/payments` 或 `POST /api/v1/payments/ecpay/callback`
- **THEN** 回傳 501 Not Implemented 與統一錯誤格式，且端點已列於 Swagger 文件
