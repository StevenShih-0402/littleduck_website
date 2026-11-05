# Design: restructure-frontend-scaffold-backend

## Context

前端現況是 Claude Design 匯出的單頁原型：`Personal Site v2.html` 以 CDN 載入 React 18 UMD、Babel Standalone、Tailwind CDN，七個 JS/JSX 檔透過全域變數互相引用（`window.SITE`、`window.Lake`、全域元件函式），路由用 hash（`#/home`、`#/blog/:id`、`#/admin/...`），資料層 `Lake` 以 localStorage 持久化並內建種子資料。功能完整：前台（首頁、部落格、文章、作品、留言板）＋ 後台（總覽、文章管理、留言、贊助、設定）＋ 雙主題（day=雜誌 / night=終端機）。

backend 為空資料夾。目標是 Gin + GORM 的 RESTful API，未來接 PostgreSQL、包 Docker 上 GCP，金流串綠界。frontend 已有獨立 git repo（1 commit），根目錄尚未 git init。

## Goals / Non-Goals

**Goals:**

- 前端轉為標準 Next.js 專案，`npm run dev` 可跑、`npm run build` 可建置，行為與原型一致
- 後端建立完整可執行的分層骨架：`go run ./cmd/server` 起得來、`/health` 可打、Swagger UI 可看
- 組態全部走環境變數；錯誤處理與日誌從第一天就是統一機制
- 三個 git 範圍各自乾淨：frontend repo、backend repo、根 repo（只管外層）

**Non-Goals:**

- 不實作實際業務 API（文章、留言、贊助的 CRUD 端點）— 只建骨架與範例端點，業務 API 留待後續 change
- 不做前端串接後端 API（localStorage 資料層原樣保留，僅模組化）
- 不做綠界正式串接（只留 service 介面、設定與 callback 路由骨架）
- 不做 GCP 實際部署（CI/CD、Cloud Run 設定留待部署 change）
- 不引入 TypeScript（原型是 JSX，維持 JS 降低搬移風險；後續可另案導入）

## Decisions

### D1. 前端框架：Next.js（App Router）

使用者指定 Next.js。採 App Router（`app/` 目錄），Node 版 `create-next-app` 慣例：

```
frontend/
├── app/
│   ├── layout.jsx          # 根 layout：字型、主題 class、Provider
│   ├── page.jsx            # 首頁
│   ├── blog/page.jsx
│   ├── blog/[id]/page.jsx  # 文章頁（原 #/article/:id）
│   ├── works/page.jsx
│   ├── guest/page.jsx      # 留言板
│   ├── admin/page.jsx      # 後台（原 #/admin/*，內部 tab 導覽先保留 state 切換）
│   └── globals.css         # 原 inline CSS（theme tokens、md、inp/btn、adm skins）
├── components/
│   ├── ui.jsx              # 原 v2/ui.jsx 共用小元件
│   ├── front/              # 原 front-home.jsx、front-content.jsx 拆出的元件
│   └── admin/              # 原 admin-shell.jsx、admin-manage.jsx 拆出的元件
├── store/
│   ├── lake.js             # 原 v2/store.js（localStorage 資料層）
│   └── lake-context.jsx    # LakeCtx Provider + hooks（theme、db、update）
├── data/site.js            # 原 data.js（export 取代 window.SITE）
├── public/assets/          # 原 assets/ 圖片
├── package.json / next.config.mjs / tailwind.config.js / postcss.config.js
```

- 互動頁面均為 client component（`"use client"`）：資料層依賴 localStorage，原型本質是 CSA。SSR/SEO 最佳化不是本次目標。
- hash 路由 → 檔案式路由；原 `#/article/:id` 對應 `/blog/[id]`。站內連結改 `next/link`。
- Tailwind CDN → 正式 Tailwind 依賴；原型只用 utility class，遷移成本低。
- Google Fonts 改用 `next/font` 載入，避免 CDN link。
- 替代方案：維持 hash 路由包在單一 client 頁面 — 否決，失去 Next.js 路由與日後 SSR 的意義。

### D2. 後端分層：標準 Go 專案佈局 + controller 命名

```
backend/
├── cmd/server/main.go          # 進入點：載 config、init logger/db、啟動 router
├── internal/
│   ├── config/config.go        # os.Getenv() 統一封裝，啟動時驗證必要變數
│   ├── router/router.go        # Gin engine、路由註冊、middleware 掛載、/swagger
│   ├── middleware/             # logger、recovery、error handler、CORS
│   ├── controller/             # HTTP 層：綁定/驗證請求、呼叫 service、統一回應
│   ├── service/                # 業務邏輯（含 ecpay service 介面骨架）
│   ├── repository/             # GORM 資料存取
│   ├── model/                  # GORM model + AutoMigrate 清單
│   └── pkg/
│       ├── apperror/           # AppError（code、status、message、wrapped err）
│       ├── logger/             # 結構化 logger 初始化
│       └── response/           # 統一 JSON envelope {code, message, data}
├── docs/                       # swag init 產物
├── Dockerfile / docker-compose.yml / .env.example / Makefile
```

- 依賴方向單向：controller → service → repository → model，符合 SOLID 依賴反轉（service、repository 以 interface 注入，方便測試 mock）。
- 範例垂直切片：`/health`（不進 DB）與 `article` 模組一條完整 controller/service/repository 鏈，作為後續模組的樣板。
- 替代方案：扁平單套件 — 否決，與 SOLID／後續多模組擴充不合。

### D3. 組態：env-only，config 套件集中

所有設定（`PORT`、`GIN_MODE`、`DB_HOST/PORT/USER/PASSWORD/NAME/SSLMODE`、`LOG_LEVEL`、`ECPAY_MERCHANT_ID/HASH_KEY/HASH_IV/ENV`）由 `internal/config` 以 `os.Getenv()` 讀取一次、填入 struct 注入各層；缺必要變數時啟動即 fail-fast。提供 `.env.example` 文件化全部變數；本機由 docker-compose / shell 匯入，不引入 viper（範圍小，標準庫即可，也符合「用 os.Getenv()」的要求）。

### D4. 錯誤處理與日誌

- 日誌用 **zap**（結構化、效能佳、GCP Cloud Logging 可直接吃 JSON 輸出）；開發模式輸出彩色 console，production 輸出 JSON。
- Gin middleware：request log（method、path、status、latency、client IP）+ recovery（panic → 500 + stack log）。
- `apperror.AppError` 實作 `error`，controller 把錯誤丟給 error-handling middleware 統一轉 HTTP 回應；未知錯誤一律 500 且不洩漏內部訊息。
- 替代方案：logrus — 已進維護模式，zap 是現行主流。

### D5. Swagger：swaggo/swag

controller 以宣告式註解描述端點，`swag init` 產 `docs/`，`gin-swagger` 掛 `/swagger/*any`。Makefile 提供 `make swag` 重新產生。替代方案：手寫 OpenAPI YAML — 否決，易與程式碼脫節。

### D6. 資料庫與 Docker

- GORM + `gorm.io/driver/postgres`；schema 以 `AutoMigrate` 起步（正式 migration 工具留待資料模型定案後導入）。
- DB 連線失敗採重試（docker-compose 啟動順序），健康檢查 `/health` 回報 DB ping 狀態。
- Dockerfile 採 multi-stage（golang:1.26 build → distroless/alpine run），docker-compose 拉起 `api` + `postgres:16`，volume 持久化。

### D7. 綠界（ECPay）骨架

`internal/service/payment/` 定義 `PaymentProvider` interface（建立訂單、驗證 callback），`ecpay.go` 空實作讀取 config 中的 MerchantID/HashKey/HashIV；`controller/payment_controller.go` 預留 `POST /api/v1/payments` 與 `POST /api/v1/payments/ecpay/callback` 路由回 501。正式串接（CheckMacValue、測試環境）另開 change。

### D8. Git 佈局

frontend、backend 各自 `git init`（frontend 已有）；根目錄 `git init`，`.gitignore` 排除 `frontend/`、`backend/`、`go1.26.0.linux-amd64.tar.gz`，根 repo 只追蹤 `CLAUDE.md`、`openspec/`、`README.md` 等外層檔案。不用 git submodule — 使用者明確要求「根只記錄兩資料夾外的檔案」，submodule 會把子 repo 的 commit 記進根 repo。

## Risks / Trade-offs

- [原型全域元件互相耦合，拆模組可能漏 import 或漏全域相依] → 逐檔搬移後以 `npm run build` + 手動走過前台/後台每頁驗證；localStorage key 與 seed version 不變，資料相容。
- [hash 路由改檔案式路由會改變 URL 結構（`#/article/x` → `/blog/x`）] → 原型無外部流量，無需轉址；admin 內部 tab 仍以 state 切換，行為不變。
- [`html { zoom: 1.263 }` 為非標準屬性，Next.js 下仍沿用] → 先原樣保留確保視覺一致，後續可另案改 rem 方案。
- [AutoMigrate 不適合正式環境 schema 演進] → 骨架階段可接受；資料模型定案後導入 migration 工具（如 golang-migrate）。
- [swag 註解與實作可能不同步] → Makefile 將 `swag init` 納入建置流程，README 記載規範。
- [根 repo 用 .gitignore 排除子 repo，三 repo 版本無法互相對應] → 接受此 trade-off（使用者指定）；README 記錄三 repo 關係。

## Migration Plan

1. frontend：在既有 git repo 內重組（保留原檔於 git 歷史），完成後單一 commit「Restructure to Next.js」
2. backend：建骨架 → `git init` → initial commit
3. 根目錄：`git init` + `.gitignore` + README → initial commit
4. 回滾：各 repo 皆可 `git revert`；frontend 原型檔案在歷史中隨時可取回

## Open Questions

- 後端業務 API 的資料模型（articles/comments/donations/settings/stats）細節 — 留待下一個 change 依前端 `store/lake.js` 的 seed 結構定案
- GCP 部署形式（Cloud Run vs GCE）與前端部署位置（Vercel vs GCP）— 部署 change 再決定
