# backend-error-logging

## ADDED Requirements

### Requirement: 統一錯誤處理
系統 SHALL 提供自訂 `AppError`（含錯誤碼、HTTP 狀態、訊息）與 error-handling middleware：controller 回傳的錯誤 MUST 由 middleware 統一轉為 JSON 錯誤回應；未定義的錯誤一律回 500 且 MUST NOT 洩漏內部錯誤細節。

#### Scenario: 已定義的業務錯誤
- **WHEN** service 回傳 not-found 類的 AppError
- **THEN** 回應 404 與統一格式 `{code, message}`，不含 stack trace

#### Scenario: 未預期錯誤
- **WHEN** 處理過程發生未包裝的 error
- **THEN** 回應 500 與一般化訊息，完整錯誤只寫入日誌

### Requirement: panic 復原
系統 SHALL 以 recovery middleware 攔截 handler 內的 panic，回應 500 並將 stack trace 記入日誌，伺服器 MUST 繼續服務後續請求。

#### Scenario: handler 發生 panic
- **WHEN** 某請求處理中 panic
- **THEN** 該請求得到 500 回應、stack 寫入日誌，下一個請求正常處理

### Requirement: 結構化日誌
系統 SHALL 使用結構化 logger（zap）：每個 HTTP 請求 MUST 記錄 method、path、status、latency、client IP；日誌等級由環境變數控制；production 模式輸出 JSON 格式以相容 GCP Cloud Logging。

#### Scenario: 請求日誌
- **WHEN** 任一 API 請求完成
- **THEN** 輸出一筆含 method、path、status、latency 的結構化日誌

#### Scenario: 日誌等級過濾
- **WHEN** `LOG_LEVEL=warn` 時發生 info 等級事件
- **THEN** 該事件不輸出，warn 以上照常輸出
