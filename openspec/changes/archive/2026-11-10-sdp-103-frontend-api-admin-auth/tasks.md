# Tasks

## 1. 後端：設定與認證基礎

- [x] 1.1 [BE] 於 `internal/config` 新增 `ADMIN_USERNAME`、`ADMIN_PASSWORD`、`JWT_SECRET`、`JWT_TTL` 讀取與啟動 fail-fast 檢查；更新 `.env.example`。驗證：缺變數時啟動輸出明確錯誤並退出，齊備時正常啟動。
- [x] 1.2 [BE] 新增 JWT 相依（`golang-jwt`）與 `internal/service/auth`：簽發 HS256 權杖、constant-time 密碼比對。驗證：單元測試涵蓋正確/錯誤帳密與過期權杖解析。
- [x] 1.3 [BE] 新增 `POST /api/v1/auth/login` controller 與路由（含 swag 註解）。驗證：正確帳密回 200+token、錯誤回 401（統一錯誤格式）。
- [x] 1.4 [BE] 新增 `middleware/auth.go` 驗證 `Authorization: Bearer`，建立受保護路由群組。驗證：中介層單元測試——有效權杖放行、缺失/過期/格式錯誤回 401。

## 2. 後端：文章 CRUD 擴充

- [x] 2.1 [BE] 擴充 `Article` model（status、views 與前台欄位），確認 `model.Register()` 註冊、AutoMigrate 建表。驗證：啟動後資料表含新欄位。
- [x] 2.2 [BE] repository/service 新增單篇讀取、建立、修改、刪除、發布狀態切換。驗證：service 單元測試（含 mock repository）涵蓋 CRUD 與發布篩選。
- [x] 2.3 [BE] controller/路由：公開 `GET /articles`（只回已發布）、`GET /articles/:id`、後台（受保護）建立/修改/刪除（含 swag 註解）。驗證：公開列表排除草稿、單篇不存在回 404、未授權管理操作回 401。

## 3. 後端：留言 API

- [x] 3.1 [BE] 新增 `Comment` model（name、text、status、reply、articleId）+ Register + AutoMigrate。驗證：資料表建立。
- [x] 3.2 [BE] repository/service：張貼（預設 pending）、公開讀取（只 visible）、後台列全部、狀態變更、回覆。驗證：service 單元測試涵蓋狀態轉換與公開篩選。
- [x] 3.3 [BE] controller/路由：公開張貼與讀取、後台（受保護）審核/回覆（含 swag 註解）。驗證：張貼建立 pending、缺內容回 400、公開讀取不含 pending/hidden、未授權審核回 401。

## 4. 後端：網站設定 API

- [x] 4.1 [BE] 新增 `SiteSetting` model（單列 + 巢狀 JSON 欄位）+ Register + AutoMigrate。驗證：資料表建立。
- [x] 4.2 [BE] service/controller/路由：公開讀取、後台（受保護）更新（含 swag 註解與基本驗證）。驗證：更新後讀取反映新值、無效內容回 400、未授權更新回 401。

## 5. 後端：統計與瀏覽追蹤 API

- [x] 5.1 [BE] 新增 `PageView` model（day、visitor_key、path）+ Register + AutoMigrate。驗證：資料表建立。
- [x] 5.2 [BE] service：記錄瀏覽事件（visitor_key = sha256(ip+ua) 截短、不存原始 IP）、以 GROUP BY day 聚合 views 與 distinct visitors。驗證：單元測試——多筆事件聚合出正確每日 views/visitors、同 visitor 當日去重。
- [x] 5.3 [BE] controller/路由：公開回報瀏覽（非阻塞語意）、後台（受保護）讀取每日統計（含 swag 註解）。驗證：回報後統計增加、無紀錄回零序列、未授權讀統計回 401。

## 6. 後端：贊助紀錄唯讀清單與種子資料

- [x] 6.1 [BE] 新增 `Donation` model（day、amount、name、note）+ Register + AutoMigrate；後台（受保護）唯讀列表端點（含 swag 註解）。驗證：帶權杖回 200 清單、未授權回 401、`POST /payments` 仍回 501。
- [x] 6.2 [BE] 新增啟動 seed：各資料表為空時以 `store/lake.js` 既有種子內容（文章、留言、設定、贊助）寫入。驗證：空庫啟動後 `GET /articles` 等回傳與原型一致的內容；非空庫不重複 seed。

## 7. 前端：API client 與認證

- [x] 7.1 [FE] 新增 `lib/api.js`：fetch 封裝、base URL 取自 `NEXT_PUBLIC_API_BASE_URL`、自動附帶 token、401 清除 token。更新 `.env`/`.env.example`（前端）與型別。驗證：`npm run build` 通過；對 mock 端點的單元測試涵蓋附帶 token 與 401 處理。
- [x] 7.2 [FE] 後台登入改接 `POST /auth/login`：成功存 token（sessionStorage）進後台、失敗顯示錯誤、移除「任意密碼可登入」。驗證：正確帳密進後台、錯誤帳密不進、token 失效導回登入。

## 8. 前端：資料層改接 API

- [x] 8.1 [FE] 改寫 `store/lake-context.jsx`（與 `store/lake.js`）：mount 時 async 載入各資料域、提供 async mutator、暴露 `loading`/`error`；移除以 localStorage 作為業務資料來源（主題偏好仍留 localStorage）。驗證：前台載入時資料來自 API（無 `lakeside_v2_store` 作資料源）。
- [x] 8.2 [FE] 更新前台元件（文章/留言/設定）改用 API 資料與載入/錯誤狀態。驗證：載入中顯示指示、API 失敗顯示錯誤不崩潰、不以假資料填補。
- [x] 8.3 [FE] 更新後台元件（文章、留言審核、設定、統計、贊助清單）改打後端 API；統計改用真實資料並移除「假資料示意」。驗證：後台編輯經 API 寫入、重整後一致；統計數字來自後端。

## 9. 設計同步基準

- [x] 9.1 [FE] 於 `frontend/design-sync/manifest.json` 將 `project/v2/store.js ↔ store/lake.js` 對應標記為刻意分歧（含原因與日期），並調整 `scripts/design-sync.mjs` 使標記者從差異報告排除、不覆寫。驗證：`npm run design:check` 不把該檔列為待同步；`npm run test:design` 通過。

## 10. 整合驗證

- [x] 10.1 [BE] `make swag`（或 `swag init`）重新產生文件並編譯通過；Swagger UI 列出新端點。驗證：`/swagger/index.html` 可見 auth/comment/setting/stat/donation 端點。
- [x] 10.2 全端整合：依 CLAUDE.md「執行驗證」啟動前後端，跑 design.md「執行驗證觀察點」。驗證：登入流程、受保護端點 401/2xx、公開讀取、前台資料來自 API、後台統計為真實資料皆符合預期。已於 docker compose（API :18080）+ `next start`（:3100）驗證：health=ok、公開 /articles 只回已發布（17 篇，隱藏 3 篇草稿）、單篇 404、登入正確回 token／錯誤 401、受保護端點無 token 皆 401、帶 token 皆 2xx、pageview 回報後 /stats 由 0 變為 views=3/visitors=2（visitor 去重正確、非假資料）、donations 合計 NT$480 與種子一致、POST /payments 仍 501、前台 6 條路由皆 200。附帶：為讓容器啟動補上 `docker-compose.yml` 的 ADMIN_*／JWT_* 環境變數（dev 預設值）。
