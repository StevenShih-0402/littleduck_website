# backend-deployment

## ADDED Requirements

### Requirement: 後端可打包為 Docker 映像檔
後端 SHALL 提供 multi-stage Dockerfile：build 階段編譯 Go binary，run 階段以最小映像執行；映像 MUST 僅透過環境變數取得組態。

#### Scenario: 映像可建置並執行
- **WHEN** 執行 `docker build` 後以必要環境變數 `docker run`
- **THEN** 容器啟動並可回應 `GET /health`

### Requirement: docker-compose 本機環境
專案 SHALL 提供 docker-compose 定義 api 與 PostgreSQL 服務：資料庫資料 MUST 以 volume 持久化，api 服務由 compose 注入環境變數。

#### Scenario: 一鍵啟動完整環境
- **WHEN** 執行 `docker compose up`
- **THEN** PostgreSQL 與 api 服務啟動，api 完成 DB 連線與 AutoMigrate 並可服務請求

#### Scenario: 資料持久化
- **WHEN** `docker compose down` 後再次 `up`（未刪 volume）
- **THEN** 資料庫既有資料仍存在
