# backend-api-architecture

## ADDED Requirements

### Requirement: 後端採用 Gin + GORM 分層架構
後端 SHALL 為獨立 Go module，使用 Gin 作為 HTTP 框架、GORM 作為 ORM，並依 router / middleware / controller / service / repository / model / config 分層；依賴方向 MUST 為 controller → service → repository，且 service 與 repository 以 interface 定義以利測試替換。

#### Scenario: 伺服器可啟動
- **WHEN** 設定必要環境變數後執行 `go run ./cmd/server`
- **THEN** Gin 伺服器於設定的埠啟動並輸出啟動日誌

#### Scenario: 分層目錄存在且職責分離
- **WHEN** 檢視 `backend/internal/` 目錄
- **THEN** 存在 config、router、middleware、controller、service、repository、model 套件，且 controller 不直接引用 repository

### Requirement: 提供健康檢查端點
系統 SHALL 提供 `GET /health` 端點，回報服務與資料庫連線狀態。

#### Scenario: 服務正常
- **WHEN** 呼叫 `GET /health` 且資料庫連線正常
- **THEN** 回傳 200 與 JSON 狀態（含 DB 連線狀態）

#### Scenario: 資料庫斷線
- **WHEN** 呼叫 `GET /health` 且資料庫無法 ping 通
- **THEN** 回傳非 200 狀態碼並於回應中標示 DB 異常

### Requirement: 提供範例垂直切片作為模組樣板
系統 SHALL 內含至少一個完整的範例模組（article），貫穿 controller、service、repository、model 四層並掛載於 `/api/v1` 之下，作為後續業務模組的實作樣板。

#### Scenario: 範例端點可呼叫
- **WHEN** 呼叫範例模組的列表端點（如 `GET /api/v1/articles`）
- **THEN** 回傳 200 與統一 JSON envelope 格式的資料

### Requirement: PostgreSQL 連線與自動遷移
系統 SHALL 以 GORM PostgreSQL driver 連線資料庫，啟動時對已註冊的 model 執行 AutoMigrate；連線失敗時 MUST 依重試策略重試後才失敗退出。

#### Scenario: 啟動時自動建表
- **WHEN** 伺服器以可連線的空白資料庫啟動
- **THEN** 已註冊 model 的資料表自動建立

#### Scenario: 資料庫暫時不可用
- **WHEN** 啟動時資料庫尚未就緒（如 docker-compose 順序）
- **THEN** 系統依設定間隔重試連線，重試耗盡才以錯誤結束
