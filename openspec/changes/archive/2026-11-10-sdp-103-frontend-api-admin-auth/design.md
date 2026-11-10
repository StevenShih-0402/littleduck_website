# Design

## Context

見 proposal.md - Why。現況（已於程式碼確認）：後端為 Gin + GORM 分層架構（controller → service → repository，介面可替換；model 以 `model.Register()` 在 `init()` 註冊供 AutoMigrate），目前僅 `GET /api/v1/articles` 有實作、兩個 payment 端點回 501；前端為 Next.js（App Router），資料層 `store/lake.js` + `store/lake-context.jsx` 以 `useLake()` 同步讀寫 localStorage（key `lakeside_v2_store`），無任何 API client。設計同步基準 `frontend/design-sync/manifest.json` 以 sha256 追蹤 `project/v2/store.js ↔ store/lake.js`。

## Goals / Non-Goals

**Goals:**
- 後端補齊 auth／article CRUD／comment／setting／stat／donation(唯讀) 六組 API，沿用現有分層與統一 envelope／error 格式。
- 前端資料層由同步 localStorage 改為非同步 API client，後台改用後端 JWT 登入。
- 切換後前台內容與現有原型一致（後端 seed 現有種子資料）。
- `store/lake.js` 的設計同步基準標記為刻意分歧，`design-pull` 不覆寫。

**Non-Goals:**
- 正式付款建立與綠界 ECPay 交易流程（屬 SDP-104）；本次 payment 僅新增唯讀清單，`POST /payments` 維持 501。
- 多管理員／角色權限、密碼重設、refresh token 輪替。
- 設計專案（claude.ai）端的 store.js 改寫——分歧由基準標記處理，不回寫設計。

## Decisions

### D1：單一管理員 + env 密碼 + JWT
- 以 `golang-jwt` 簽發 HS256 存取權杖；管理員帳密與秘鑰、效期來自環境變數（`ADMIN_USERNAME`、`ADMIN_PASSWORD`、`JWT_SECRET`、`JWT_TTL`），啟動時 fail-fast 驗證存在（沿用 `backend-configuration` 慣例）。
- 密碼比對用 constant-time 比較（`subtle.ConstantTimeCompare`）。單一管理員故不建 users 表、不引入 bcrypt。
- **替代方案**：DB users 表 + bcrypt（過度設計，單站無多帳號需求）；session cookie（需 server 端 session 儲存與 cookie/CORS 設定，較重）。

### D2：auth middleware 保護後台端點
- 新增 `middleware/auth.go`，驗證 `Authorization: Bearer`，掛在後台路由群組；公開端點（文章公開讀取、公開留言讀取、留言張貼、瀏覽追蹤、設定讀取）不掛。
- 路由分兩群：`/api/v1`（公開）與 `/api/v1`（受保護，套用 auth middleware），以 Gin group 區分。

### D3：資料模型（GORM，沿用 base model 與 Register()）
- `Article` 擴充：新增 `Status`(draft/published)、`Views`、以及前台渲染欄位（分類、日期、摘要、banner/showcase/gallery、內文）——對齊 lake.js 既有欄位。
- `Comment`：`Name`、`Text`、`Status`(pending/visible/hidden)、`Reply`、`ArticleID`(可空＝留言板)。
- `SiteSetting`：單列（id=1）設定，巢狀內容（社群連結、贊助級距）以 JSON 欄位存。
- `PageView`：`Day`(date)、`VisitorKey`、`Path`——stat 端點以 `GROUP BY day` 聚合 views，distinct `visitor_key` 聚合 visitors。
- `Donation`（唯讀清單用）：`Day`、`Amount`、`Name`、`Note`；seed 既有贊助紀錄；SDP-104 再接寫入。
- **替代方案（stats）**：維護每日計數表（增量 upsert）——寫入較省，但個人站流量低，原始事件 + GROUP BY 更簡單且可重算，採後者。

### D4：visitor 去重
- `VisitorKey = sha256(client_ip + user_agent)` 截短，按日去重。不存原始 IP（隱私）。前台瀏覽事件以非阻塞方式回報（fire-and-forget fetch），不拖慢渲染。

### D5：前端 API client 與非同步資料層
- 新增 `lib/api.js`：封裝 `fetch`，base URL 來自 `NEXT_PUBLIC_API_BASE_URL`，自動附帶 token，對 401 清除 token 並導回登入。
- `lake-context.jsx` 改為：mount 時 async 載入各資料域、提供 async mutator（呼叫 API 後更新 state）、暴露 `loading`／`error`。各 admin／front 元件改用這些狀態（載入指示、錯誤提示）。
- token 存 `sessionStorage`（沿用現有後台 session 範圍行為）；主題偏好仍留 localStorage。

### D6：設計同步刻意分歧
- 於 `frontend/design-sync/manifest.json` 對 `project/v2/store.js ↔ store/lake.js` 的對應加上 `divergent: true`（或等效標記），並調整 `scripts/design-sync.mjs`：標記者從差異報告排除、不覆寫。於 manifest 記錄分歧原因與日期。

## Risks / Trade-offs

- **移除 localStorage 後需要後端在線** → 本就要成為真實服務；前端對 API 失敗顯示錯誤而非崩潰（frontend-data-layer 規格）。
- **JWT 存 sessionStorage 有 XSS 風險** → 單一低風險個人站可接受；不放長效憑證、效期短。未來要更安全可改 httpOnly cookie。
- **env 明文密碼** → 僅單一管理員、部署環境變數管控；constant-time 比較避免 timing。
- **visitor 去重以 IP+UA 雜湊，NAT 下會低估** → 可接受的近似，個人站非精準分析。
- **lake.js 刻意分歧後設計端再改 store.js 不會自動帶入** → manifest 記錄分歧原因，`design-pull` 會明確標示該檔為刻意分歧提醒人工判斷。
- **seed 與既有原型資料需一致** → 以 lake.js 現有 seed 為來源，僅在資料表為空時 seed，避免重複。

## Migration Plan

1. 後端：新增 model → `AutoMigrate` 自動建表；啟動時若對應表為空則 seed 既有內容（文章／留言／設定／贊助）。
2. 新增環境變數同步 `.env.example` 與部署設定（`backend-configuration`、部署 runbook）；前端新增 `NEXT_PUBLIC_API_BASE_URL`。
3. 前端切換資料層；後台登入改接 API。
4. 更新 `design-sync` manifest 標記分歧。
5. **Rollback**：前端資料層改動集中於 `lib/api.js` 與 `lake-context.jsx`，可整批還原回 localStorage 版本；後端新增 model／端點為增量，回滾不影響既有 `GET /articles`。

## 執行驗證觀察點

（供後續檢查用，補充 CLAUDE.md 既有觀察點）
- `POST /api/v1/auth/login` 以正確/錯誤帳密分別回 200（含 token）/401。
- 帶 token 呼叫後台端點（如文章建立、留言狀態變更、`GET` 統計）回 2xx；未帶 token 回 401。
- `GET /api/v1/articles` 公開可讀、只含已發布；單篇不存在回 404。
- 前台頁面載入後資料來自 API（Network 面板可見 API 請求，無 `lakeside_v2_store` 作為資料來源）。
- 後台統計頁不再出現「假資料示意」字樣，數字來自實際瀏覽紀錄。

## Open Questions

- 無阻擋性未決項；帳密輪替、更安全的 token 儲存策略可於日後獨立處理，不影響本次規格與任務拆解。
