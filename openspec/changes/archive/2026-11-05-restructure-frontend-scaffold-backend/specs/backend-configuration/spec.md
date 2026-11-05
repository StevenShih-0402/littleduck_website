# backend-configuration

## ADDED Requirements

### Requirement: 組態一律來自環境變數
所有可變組態（伺服器埠、Gin 模式、資料庫連線資訊、日誌等級、綠界金鑰）SHALL 由 `internal/config` 套件以 `os.Getenv()` 讀取並集中為 config struct 注入各層；程式碼中 MUST NOT 寫死連線字串、金鑰或密碼。

#### Scenario: 由環境變數組成 DB 連線
- **WHEN** 設定 `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME` 後啟動
- **THEN** 系統以這些值組成連線字串連上 PostgreSQL

#### Scenario: 缺少必要變數 fail-fast
- **WHEN** 必要環境變數（如 `DB_HOST`）未設定即啟動
- **THEN** 系統於啟動階段輸出明確錯誤並終止，不進入服務狀態

### Requirement: 提供環境變數範本
專案 SHALL 提供 `.env.example` 列出全部環境變數與說明，且真實秘密值 MUST NOT 進版控（`.env` 列入 `.gitignore`）。

#### Scenario: 新環境依範本設定
- **WHEN** 開發者複製 `.env.example` 為 `.env` 並填入值
- **THEN** 依該檔匯入環境變數後服務可正常啟動
