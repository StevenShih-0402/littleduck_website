# Proposal: restructure-frontend-scaffold-backend

## Why

frontend 資料夾目前是從 Claude Design 擷取的原型：React 18 以 UMD `<script>` 載入、JSX 由 Babel Standalone 在瀏覽器即時編譯、元件全部掛在全域、資料存在 localStorage。這種形式無法安裝依賴、無法打包部署、也無法與後端 API 串接。同時 backend 資料夾仍是空的，需要建立正式的 RESTful API 系統架構，才能讓網站從原型走向可部署的正式版（PostgreSQL + Docker + GCP）。

## What Changes

- **BREAKING** 前端重構為標準 Next.js（App Router）+ React 18 專案結構：
  - 移除 CDN UMD React / Babel Standalone / Tailwind CDN，改為 npm 依賴與 Next.js 建置流程
  - 全域元件（`app.jsx`、`front-*.jsx`、`admin-*.jsx`、`ui.jsx`）拆分為 ES module，依 Next.js 慣例分層：`app/`（路由頁面）、`components/`、`store/`（資料層）、`data/`（種子內容）
  - hash routing（`#/home`、`#/blog`、`#/admin` …）改為 Next.js 檔案式路由
  - `Personal Site v2.html` 的 inline CSS 抽出為全域樣式與 Tailwind 設定，`assets/` 移入 `public/`
  - 保留現有行為：雙主題（雜誌/終端機）、前台各頁、後台管理、localStorage 資料層（後續換接 API 時介面不變，互動元件以 client component 實作）
- 新建 Golang backend（Gin + GORM）RESTful API 系統架構：
  - 分層架構：router / middleware / controller / service / repository / model / config
  - swaggo 產生 Swagger 文件（`/swagger/*` 端點）
  - 組態一律由環境變數讀取（`os.Getenv()` 封裝於 config 套件），不寫死資料
  - PostgreSQL（GORM driver）連線與 migration 基礎
  - 統一 Exception 處理（error middleware + 自訂 AppError）與結構化 log（請求日誌 + 應用日誌）
  - 綠界科技（ECPay）金流介接的模組骨架（介面與設定預留，正式串接後續 change 處理）
  - Dockerfile 與 docker-compose（app + PostgreSQL），供日後部署 GCP
- Git 管理調整：frontend 與 backend 各自獨立 git repo（frontend 已有），根目錄 git 初始化並以 `.gitignore` 排除兩子資料夾，只追蹤外層檔案

## Capabilities

### New Capabilities

- `frontend-project-structure`: 前端以標準 Next.js 專案結構組織（App Router、模組化元件、靜態資源、npm scripts），且維持原型的功能行為
- `backend-api-architecture`: Gin + GORM 分層 RESTful API 架構，含路由、健康檢查端點與依 domain 劃分的模組
- `backend-configuration`: 以環境變數為唯一組態來源（DB 連線、伺服器埠、ECPay 金鑰等），提供 `.env.example`
- `backend-error-logging`: 統一錯誤處理（panic recovery、AppError → HTTP 回應）與結構化日誌紀錄
- `backend-api-docs`: swaggo 自動產生並提供 Swagger UI 文件
- `backend-deployment`: Docker 映像檔打包與 docker-compose 本機環境（PostgreSQL），為 GCP 部署做準備
- `payment-integration`: 綠界科技金流模組骨架（設定、介面、callback 端點預留）

### Modified Capabilities

（無 — `openspec/specs/` 目前為空，皆為新能力）

## Impact

- `frontend/`：檔案全面搬移重組（git 已納管，重組以 git mv / 新增方式保留歷史）；新增 `package.json`、`next.config.js`、`app/` 路由樹、`components/`、`public/`
- `backend/`：從空資料夾建立完整 Go module 專案，並 `git init`
- 根目錄：`git init`、新增 `.gitignore`（排除 `frontend/`、`backend/`、`go1.26.0.linux-amd64.tar.gz`）、更新 `README.md`
- 新依賴：前端 Next.js、React、Tailwind CSS（npm）；後端 gin、gorm、gorm postgres driver、swaggo、logrus/zap 等
- 開發環境：需要 Node.js 與 Go 1.26（根目錄已有 go1.26.0 tarball 可安裝）
