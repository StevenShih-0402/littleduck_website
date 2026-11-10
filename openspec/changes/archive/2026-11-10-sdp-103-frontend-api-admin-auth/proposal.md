# Proposal

## Why

前端目前是原型狀態：後台任意密碼即可登入（`admin-shell.jsx` 只在 `sessionStorage` 設旗標），所有資料（文章、留言、贊助紀錄、網站設定、統計）只存在瀏覽器 localStorage（`store/lake.js`，key `lakeside_v2_store`），不同裝置與訪客之間不共用，後台統計更是用種子 42 產生的假資料。後端現階段只有 `GET /api/v1/articles` 與兩個回 501 的 payment 端點。要讓網站成為真正可運作的服務，必須補齊後端 API、把前端資料層改接後端，並為後台加上真實登入。

## What Changes

- **後台真實登入**：新增 `POST /api/v1/auth/login`，以單一管理員帳密（來自環境變數）驗證，成功後簽發 JWT；新增 auth middleware 保護所有後台寫入／管理端點。前端後台登入改打此 API 並以 token 呼叫受保護端點，移除「任意密碼可登入」行為。**BREAKING**（後台登入行為改變）。
- **文章 API 擴充**：在現有 `GET /api/v1/articles` 之外補齊單篇讀取與後台 CRUD／發布狀態（草稿／已發布、瀏覽數）。
- **留言 API**：訪客張貼留言（預設 pending）、前台讀取 visible 留言、後台審核（狀態 pending／visible／hidden、回覆）。
- **網站設定 API**：讀取與更新網站設定（標題、標語、社群連結、公告、贊助級距）。
- **統計 API**：新增輕量瀏覽追蹤（記錄每日 views／visitors）與後台統計聚合端點，取代前端假資料。
- **贊助紀錄唯讀清單**：後台可讀取贊助／付款紀錄清單；實際金流寫入與 ECPay 串接仍屬 SDP-104，不在本次範圍。
- **前端資料層改接 API**：新增 API client／fetch 層，`store/lake.js` 與 `lake-context.jsx` 從同步讀寫 localStorage 改為非同步呼叫後端 API，各 admin／front 元件處理載入中與錯誤狀態。
- **後端初始資料**：以 seed／migration 將前端現有的種子內容（文章／小說、留言、設定、贊助紀錄）寫入後端，確保切換後前台內容不變。
- **設計同步基準調整**：`store/lake.js` 改接 API 後會與設計專案 `project/v2/store.js`（`design-sync/manifest.json` 記錄的 sha256 `71b069086…`）分歧，需將此對應標記為刻意分歧，避免下次 `design-pull` 覆寫。

## Capabilities

### New Capabilities
- `admin-auth`: 後台身分驗證——以環境變數設定的單一管理員帳密登入、簽發與驗證 JWT、以 middleware 保護後台端點。
- `blog-content-api`: 文章內容 API——公開列表與單篇讀取，以及後台建立／修改／刪除與發布狀態管理。
- `comments-api`: 留言 API——訪客張貼、前台讀取公開留言、後台審核與回覆。
- `site-settings-api`: 網站設定 API——讀取與更新站台中繼設定（標題、標語、社群連結、公告、贊助級距）。
- `site-analytics-api`: 站台統計 API——輕量瀏覽追蹤與後台每日統計聚合，取代前端假資料。
- `frontend-data-layer`: 前端資料層——API client／fetch 封裝、非同步載入與錯誤狀態處理、後台登入 token 的保存與附帶。

### Modified Capabilities
- `frontend-project-structure`: 「保留原型功能行為」要求目前明定資料層為 localStorage（`Lake` 之 seed／load／save／reset 介面不變）；本次改為以後端 API 作為資料來源，需調整此要求（主題切換仍持久化於 localStorage）。
- `design-sync`: 同步基準新增「刻意分歧」概念——`store/lake.js` 因改接 API 而與 `project/v2/store.js` 分歧時，基準需能標記該對應為刻意分歧，使差異報告不將其視為待同步、`design-pull` 不覆寫。
- `payment-integration`: 新增後台唯讀「贊助／付款紀錄列表」端點；現有 501 預留端點與「正式交易流程不在範圍」不變（交易寫入屬 SDP-104）。

## Impact

- **後端**（`backend/`）：新增 `auth`／`comment`／`setting`／`stat` 模組（controller／service／repository／model 各層）、auth middleware、對應 GORM model 與 AutoMigrate、seed 初始資料；擴充 `article` 模組為完整 CRUD；`payment` 模組新增唯讀列表。新增環境變數（管理員帳密、JWT 秘鑰／效期）。所有端點補 swag 註解。
- **前端**（`frontend/`）：新增 API client 層；改寫 `store/lake.js`、`store/lake-context.jsx` 與 `components/admin/*`、`components/front/*` 的資料存取為非同步 API 呼叫；後台登入改接 `auth` API 並保存 token。
- **設計同步**：更新 `frontend/design-sync/manifest.json`（或其基準格式）以標記 `lake.js` 的刻意分歧。
- **相依**：後端新增 JWT 函式庫（如 `golang-jwt`）與密碼雜湊（若日後擴充）；前端沿用既有 fetch，不新增 UI 框架。
- **設定／部署**：新增必要環境變數需同步 `.env.example` 與部署設定（`backend-configuration`、部署 runbook）。
