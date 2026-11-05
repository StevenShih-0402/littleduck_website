# backend-api-docs Specification

## Purpose
由 swaggo 從程式碼註解自動產生 API 文件，並透過 Swagger UI 提供瀏覽與呼叫，讓 API 使用方能查閱與測試後端提供的端點。

## Requirements

### Requirement: Swagger 文件自動產生
系統 SHALL 以 swaggo 註解描述 API 端點並產生 OpenAPI 文件，透過 `/swagger/*any` 提供 Swagger UI；所有對外 API 端點 MUST 具備 swag 註解（摘要、參數、回應格式）。

#### Scenario: Swagger UI 可瀏覽
- **WHEN** 伺服器啟動後開啟 `/swagger/index.html`
- **THEN** 顯示 Swagger UI，列出 health 與範例模組端點

#### Scenario: 文件可重新產生
- **WHEN** 執行 `make swag`（或 `swag init`）
- **THEN** `docs/` 目錄依最新註解重新產生且編譯通過
