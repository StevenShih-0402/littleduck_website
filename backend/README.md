<!-- backend/README.md -->
<!-- 後端專案說明：架構、環境變數、啟動方式與 Swagger 文件產生流程 -->

# LittleDuck Backend

LittleDuck 個人網站後端 RESTful API，使用 Go（Gin + GORM）+ PostgreSQL。

## 技術棧

- **Go 1.26** / **Gin**（HTTP 框架）/ **GORM**（ORM, PostgreSQL driver）
- **zap** 結構化日誌（dev 彩色 console / release JSON，相容 GCP Cloud Logging）
- **swaggo** 自動產生 Swagger 文件
- **Docker** multi-stage 建置 + docker-compose 本機環境

## 架構

分層架構，依賴方向單向：`controller → service → repository → model`，service 與 repository 皆以 interface 定義，方便單元測試 mock 替換。

```
backend/
├── cmd/server/main.go       # 進入點：config → logger → db → router → 啟動
├── internal/
│   ├── config/              # 環境變數組態（os.Getenv 集中封裝，缺必要變數 fail-fast）
│   ├── router/              # Gin engine、middleware 掛載、路由註冊、/swagger
│   ├── middleware/          # request logger、recovery、error handler、CORS
│   ├── controller/          # HTTP 層：health、auth、article、comment、setting、stat、donation、payment
│   ├── service/             # 業務邏輯（含 auth/ JWT 與 payment/ 綠界骨架）
│   ├── repository/          # GORM 資料存取
│   ├── model/               # GORM model + AutoMigrate 註冊機制
│   ├── seed/                # 啟動種子資料（各表為空時以前端既有內容初始化）
│   ├── database/            # DB 連線（重試策略）與 AutoMigrate
│   └── pkg/
│       ├── apperror/        # AppError（code、status、message、wrapped err）
│       ├── logger/          # zap 初始化
│       └── response/        # 統一 JSON envelope {code, message, data}
├── docs/                    # swag init 產物（勿手動編輯）
├── Dockerfile / docker-compose.yml / Makefile / .env.example
```

錯誤處理：controller 以 `c.Error(err)` 將錯誤交給 error-handling middleware，`AppError` 依其狀態碼回應統一格式 `{code, message}`；未知錯誤一律回 500 且不洩漏內部細節；panic 由 recovery middleware 攔截（500 + stack 寫入日誌）。

## 環境變數

組態一律來自環境變數（見 `.env.example`）：

| 變數 | 必要 | 預設 | 說明 |
|---|---|---|---|
| `PORT` | — | `8080` | HTTP 監聽埠 |
| `GIN_MODE` | — | `debug` | `debug` / `release` / `test` |
| `DB_HOST` | ✅ | — | PostgreSQL 主機 |
| `DB_PORT` | — | `5432` | PostgreSQL 埠 |
| `DB_USER` | ✅ | — | 資料庫使用者 |
| `DB_PASSWORD` | ✅ | — | 資料庫密碼 |
| `DB_NAME` | ✅ | — | 資料庫名稱 |
| `DB_SSLMODE` | — | `disable` | `disable` / `require` / `verify-full` |
| `LOG_LEVEL` | — | `info` | `debug` / `info` / `warn` / `error` |
| `ADMIN_USERNAME` | ✅ | — | 後台單一管理員帳號 |
| `ADMIN_PASSWORD` | ✅ | — | 後台單一管理員密碼 |
| `JWT_SECRET` | ✅ | — | JWT HS256 簽章秘鑰 |
| `JWT_TTL` | ✅ | — | JWT 有效期（Go duration 如 `24h`，或純數字＝秒） |
| `ECPAY_MERCHANT_ID` | — | — | 綠界商店代號（正式串接時必要） |
| `ECPAY_HASH_KEY` | — | — | 綠界 HashKey（正式串接時必要） |
| `ECPAY_HASH_IV` | — | — | 綠界 HashIV（正式串接時必要） |
| `ECPAY_ENV` | — | `stage` | 綠界環境：`stage` / `production` |

缺少必要變數時服務啟動即失敗（fail-fast）並列出全部缺漏項。

## 啟動方式

### 本機開發

```bash
cp .env.example .env        # 填入實際值
set -a && source .env && set +a
go run ./cmd/server          # 或 make run
```

### Docker Compose（api + PostgreSQL 16）

```bash
docker compose up --build
```

啟動後：

- 健康檢查：<http://localhost:8080/health>（含 DB ping 狀態）
- Swagger UI：<http://localhost:8080/swagger/index.html>
- 範例端點：`GET /api/v1/articles`

## 測試

```bash
go test ./...                # 或 make test
```

測試遵循 AAA 模式（Arrange-Act-Assert），tests 與 src 互為鏡像（`*_test.go` 與被測檔同套件同目錄）。

## Swagger 文件產生流程

端點以 [swaggo](https://github.com/swaggo/swag) 宣告式註解描述（見各 controller），修改註解後重新產生：

```bash
make swag                    # 等同 swag init -g cmd/server/main.go -o docs
```

需先安裝 swag CLI：`go install github.com/swaggo/swag/cmd/swag@latest`。
`make build` 會先執行 `swag` 再編譯，確保文件與程式碼同步。

## API 端點

端點分公開與受保護兩群；受保護端點需於 `Authorization: Bearer <jwt>` 帶有效權杖（由 `POST /auth/login` 取得）。

| Method | Path | 存取 | 說明 |
|---|---|---|---|
| GET | `/health` | 公開 | 健康檢查（含 DB 狀態） |
| POST | `/api/v1/auth/login` | 公開 | 後台登入，回 `{token, expires_at}` |
| GET | `/api/v1/articles` | 公開 | 文章列表（僅已發布） |
| GET | `/api/v1/articles/:id` | 公開 | 讀取單篇文章（不存在回 404） |
| GET | `/api/v1/comments` | 公開 | 可見留言列表（可用 `?articleId=` 過濾） |
| POST | `/api/v1/comments` | 公開（選用權杖） | 張貼留言（建立為 pending）；對應功能開關停用回 403 `FEATURE_DISABLED`。可選帶管理員 `Authorization: Bearer`，有效時略過開關（管理員預覽），無效或過期視同訪客、不回 401 |
| GET | `/api/v1/features` | 公開 | 功能開關對照表 `{key: boolean}` |
| GET | `/api/v1/settings` | 公開 | 讀取網站設定 |
| POST | `/api/v1/stats/pageview` | 公開 | 回報瀏覽事件（非阻塞，回 204） |
| POST | `/api/v1/payments` | 公開 | 建立付款（骨架，回 501） |
| POST | `/api/v1/payments/ecpay/callback` | 公開 | 綠界 callback（骨架，回 501） |
| GET | `/api/v1/admin/articles` | 🔒 | 後台文章列表（含草稿） |
| POST | `/api/v1/articles` | 🔒 | 建立文章 |
| PUT | `/api/v1/articles/:id` | 🔒 | 修改文章（含發布狀態） |
| DELETE | `/api/v1/articles/:id` | 🔒 | 刪除文章 |
| GET | `/api/v1/admin/comments` | 🔒 | 後台留言列表（全部狀態） |
| PUT | `/api/v1/comments/:id/status` | 🔒 | 變更留言狀態 |
| PUT | `/api/v1/comments/:id/reply` | 🔒 | 回覆留言 |
| PUT | `/api/v1/settings` | 🔒 | 更新網站設定 |
| GET | `/api/v1/admin/features` | 🔒 | 後台功能開關清單（含名稱、說明、狀態） |
| PUT | `/api/v1/features/:key` | 🔒 | 切換功能開關，body `{"enabled": boolean}`（未知 key 404、缺欄位 400） |
| GET | `/api/v1/stats` | 🔒 | 最近 30 天每日 views／visitors |
| GET | `/api/v1/donations` | 🔒 | 贊助紀錄清單（唯讀） |
| GET | `/swagger/*any` | 公開 | Swagger UI |

### 功能開關

開關登記表在 `internal/service/feature_service.go`（`sponsor`、`guestbook`、`article_comments`），啟動 seed 只補缺的 key、不覆寫既有值。新增功能：在登記表加一列、前端以 `FeatureGate` 包住；有公開寫入端點的功能要在 service 層比照 `comment_service.go` 做後端把關。

### 錯誤碼

| code | HTTP | 說明 |
|---|---|---|
| `BAD_REQUEST` | 400 | 請求內容錯誤 |
| `UNAUTHORIZED` | 401 | 缺少或無效的權杖 |
| `FEATURE_DISABLED` | 403 | 功能已被管理員停用（目前由 `POST /comments` 回傳） |
| `NOT_FOUND` | 404 | 資源不存在 |
| `NOT_IMPLEMENTED` | 501 | 尚未實作 |
| `INTERNAL_ERROR` | 500 | 伺服器內部錯誤 |
