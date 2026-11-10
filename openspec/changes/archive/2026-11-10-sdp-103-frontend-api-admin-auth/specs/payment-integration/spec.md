# Spec Delta

## ADDED Requirements

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
