# payment-integration Specification

## Purpose
提供綠界科技金流模組的骨架，包含設定、介面與 callback 端點預留，讓後續 change 能在此基礎上完成正式的金流串接流程。

## Requirements

### Requirement: 金流模組骨架（綠界）
系統 SHALL 定義 `PaymentProvider` interface（建立付款、驗證 callback）並提供讀取綠界設定（MerchantID、HashKey、HashIV、環境）的 ECPay 實作骨架；綠界金鑰 MUST 來自環境變數。正式的 CheckMacValue 計算與交易流程不在本次範圍。

#### Scenario: 金流設定由環境變數載入
- **WHEN** 設定 `ECPAY_MERCHANT_ID`、`ECPAY_HASH_KEY`、`ECPAY_HASH_IV` 後啟動
- **THEN** ECPay service 以這些值初始化，程式碼中無寫死金鑰

#### Scenario: 預留端點回應未實作
- **WHEN** 呼叫 `POST /api/v1/payments` 或 `POST /api/v1/payments/ecpay/callback`
- **THEN** 回傳 501 Not Implemented 與統一錯誤格式，且端點已列於 Swagger 文件

### Requirement: 後台唯讀贊助紀錄清單
系統 SHALL 提供需管理員身分的唯讀端點，列出贊助／付款紀錄（日期、金額、贊助者名稱、備註等）供後台檢視。本需求僅涵蓋讀取；正式的付款建立與綠界交易流程仍不在本範圍（屬後續 change）。

#### Scenario: 後台讀取贊助紀錄
- **WHEN** 管理員以有效權杖呼叫贊助紀錄清單端點
- **THEN** 回傳 200 與贊助紀錄清單

#### Scenario: 未授權讀取
- **WHEN** 未帶有效權杖呼叫贊助紀錄清單端點
- **THEN** 回傳 401，不回傳紀錄

#### Scenario: 付款建立仍未實作
- **WHEN** 呼叫 `POST /api/v1/payments`
- **THEN** 維持回傳 501 Not Implemented（交易流程屬後續 change）
