# Tasks: restructure-frontend-scaffold-backend

## 1. 前端 Next.js 專案初始化

- [x] 1.1 在 `frontend/` 建立 Next.js（App Router、JS）專案骨架：`package.json`、`next.config.mjs`、`jsconfig.json`、`.gitignore`，安裝 next / react / react-dom / tailwindcss
- [x] 1.2 設定 Tailwind（`tailwind.config.js`、`postcss.config.js`），將 `Personal Site v2.html` 的 inline CSS 抽出為 `app/globals.css`（theme tokens、md、inp/btn、adm skins、動畫）
- [x] 1.3 建立 `app/layout.jsx`：以 `next/font` 載入五種字型、掛 `<html>` 主題 class、metadata（title/description）
- [x] 1.4 搬移 `assets/` → `public/assets/`，`data.js` → `data/site.js`（`window.SITE` 改 export）

## 2. 前端模組化搬移

- [x] 2.1 `v2/store.js` → `store/lake.js`：移除 IIFE 與 `window.Lake`，改 ES module export；建立 `store/lake-context.jsx`（LakeCtx Provider：db、update、theme、setTheme）
- [x] 2.2 `v2/ui.jsx` → `components/ui.jsx`：拆為具名 export 的共用元件
- [x] 2.3 `v2/front-home.jsx`、`v2/front-content.jsx` → `components/front/`：拆出 FrontShell 與各頁元件，站內連結改 `next/link`
- [x] 2.4 `v2/admin-shell.jsx`、`v2/admin-manage.jsx` → `components/admin/`：AdminShell 與管理頁元件
- [x] 2.5 建立路由頁：`app/page.jsx`（首頁）、`app/blog/page.jsx`、`app/blog/[id]/page.jsx`、`app/works/page.jsx`、`app/guest/page.jsx`、`app/admin/page.jsx`（均為 client component），移除 hash 路由邏輯
- [x] 2.6 刪除原型檔（`Personal Site v2.html`、`v2/`、根層 `data.js`），確認 `npm run dev` 與 `npm run build` 通過
- [x] 2.7 手動驗證：前台各頁、文章動態路由、後台編輯寫回 localStorage、雙主題切換與持久化；frontend repo commit

## 3. 後端專案骨架

- [x] 3.1 安裝/確認 Go 1.26 環境，`backend/` 內 `go mod init`、`git init`、`.gitignore`
- [x] 3.2 建立 `internal/config`：以 `os.Getenv()` 讀取全部環境變數（PORT、GIN_MODE、DB_*、LOG_LEVEL、ECPAY_*），必要變數缺漏時 fail-fast；撰寫 config 單元測試（AAA）
- [x] 3.3 建立 `internal/pkg/logger`（zap：dev console / prod JSON、LOG_LEVEL 過濾）、`internal/pkg/apperror`（AppError：code、status、message）、`internal/pkg/response`（統一 JSON envelope）；各附單元測試
- [x] 3.4 建立 `internal/middleware`：request logger、recovery（panic → 500 + stack log）、error handler（AppError → HTTP、未知錯誤 → 500 不洩漏細節）、CORS；附單元測試
- [x] 3.5 建立 DB 初始化（GORM + postgres driver、連線重試、AutoMigrate 註冊機制）與 `internal/model` 基礎

## 4. 後端 API 與文件

- [x] 4.1 建立 `internal/router` 與 `cmd/server/main.go`：組裝 config → logger → db → middleware → routes，實作 `GET /health`（含 DB ping 狀態）
- [x] 4.2 建立 article 範例垂直切片：model / repository（interface + GORM 實作）/ service（interface + 實作）/ controller，掛載 `GET /api/v1/articles`；service 與 controller 附單元測試（mock 下層）
- [x] 4.3 建立 payment 骨架：`PaymentProvider` interface、ECPay service（讀 ECPAY_* 設定）、payment controller 預留 `POST /api/v1/payments`、`POST /api/v1/payments/ecpay/callback` 回 501；附單元測試
- [x] 4.4 導入 swaggo：全部端點加 swag 註解、`swag init` 產生 `docs/`、router 掛 `/swagger/*any`、Makefile 加 `make swag`
- [x] 4.5 建立 `.env.example`（全部變數含說明）並確認 `.env` 在 `.gitignore`

## 5. Docker 與收尾

- [ ] 5.1 撰寫 multi-stage Dockerfile 與 docker-compose.yml（api + postgres:16 + volume），驗證 `docker compose up` 後 `/health` 與 `/swagger` 可用
- [x] 5.2 執行 `go test ./...` 全綠；backend repo initial commit
- [x] 5.3 backend `README.md`：架構說明、環境變數表、啟動方式、swag 產生流程
- [x] 5.4 根目錄 `git init`、`.gitignore`（排除 `frontend/`、`backend/`、`go1.26.0.linux-amd64.tar.gz`）、根 `README.md`（三 repo 關係與專案總覽）、initial commit
