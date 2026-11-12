# Proposal：前台內容收尾：日期、連結與文章內文（SDP-105）

## Why

前台沿用設計原型時留下四個缺口（SDP-105）：日期不正確、社群連結與文章內的 Markdown 連結點了沒反應、小說與專案文章的內文仍是「整理中」佔位文字。網站已上線（Staging／Production），訪客會直接看到這些缺口。

2026-11-12 對照現況調查（細節見 design.md「Context」）：

| issue 原文 | 現況 | 本次處置 |
|------|------|------|
| `Lake.today()` 寫死 2026 / 07 / 02（`store/lake.js`） | SDP-103 已改為取當天日期；留言日期改由後端 `comment_service` 以 `time.Now()` 產生，贊助寫入尚未實作（SDP-104） | 修正兩個仍存在的日期問題：後端留言日期以 UTC 計算（容器沒設 TZ，台灣 00:00–08:00 的留言會記成前一天）；後台新增文章的年份 `y` 寫死 2026 |
| 社群連結網址是 `#`，點擊被 `preventDefault` 擋掉（`front-shell.jsx`） | 仍存在；網址來自後端 seed（`seed_data.go`），可在後台「設定」修改 | 前端移除 `preventDefault`，依網址安全分類渲染（佔位網址顯示為不可點文字）；後端驗證社群連結網址格式；真實網址由 user 上線後在後台填寫 |
| 文章內 Markdown 連結被擋（`components/ui.jsx`） | 仍存在 | 前端移除 `preventDefault`，依網址安全分類渲染 |
| 小說與專案內文是「整理中」佔位（issue 寫 `store/lake.js`） | 仍存在，但來源已改為後端 seed（`backend/internal/seed/seed_data.go`）；seed 只在資料表為空時寫入，已上線的資料庫不受 seed 修改影響 | user 決定（2026-11-12）：上線後在後台編輯器自行填寫，seed 維持現狀，本次不改程式碼 |

## What Changes

**後端**

- 留言建立時的顯示日期改以台灣時間（UTC+8）計算，不再依容器時區。
- `PUT /api/v1/settings` 新增社群連結網址驗證：只接受空字串、`#`（佔位）、站內路徑（`/` 開頭、非 `//`）、`http://`、`https://`、`mailto:`；其餘（如 `javascript:`、`data:`）回 400，設定不變。非破壞性：既有合法資料（含 `#`）照常可存。
- seed 維持現狀：真實文章內文、社群與專案網址由 user 上線後在後台填寫。

**前端**

- 新增網址安全分類 helper（`frontend/lib/safe-url.mjs`）與單元測試：站內路徑、外部網址、不可點三類。
- `components/ui.jsx` 的 Markdown 連結：站內路徑用 `next/link`；外部網址開新分頁並加 `rel="noopener noreferrer"`；不安全或空網址只顯示文字。
- `components/front/front-shell.jsx` 的頁尾社群連結：同上規則；網址為空、`#` 或不安全時仍顯示名稱，但為不可點的文字。
- `components/admin/admin-posts.jsx` 新增文章的年份改取當年。
- 設計同步：`frontend/design-sync/manifest.json` 新增「移植時要保留的偏離」紀錄（`deviations`），`design-sync.mjs` 在更新基準時保留它、在差異報告列出它；三處偏離都登記進去。

非 **BREAKING**。

## Capabilities

### New Capabilities

- `front-content-links`：前台 Markdown 內文與社群連結的可點擊行為與網址安全規則。

### Modified Capabilities

- `comments-api`：新留言的顯示日期以台灣時間計算。
- `site-settings-api`：設定更新時驗證社群連結網址格式。
- `design-sync`：同步基準可記錄單一前端檔案內「移植時要保留的偏離」，並在差異報告中列出。

## Impact

### 預計修改的檔案範圍（供判斷能否與其他 issue 並行）

| 端 | 檔案 |
|----|------|
| 後端 | `backend/internal/service/comment_service.go`、`backend/internal/service/comment_service_test.go` |
| 後端 | `backend/internal/service/setting_service.go`、`backend/internal/service/setting_service_test.go`、`backend/internal/controller/setting_controller.go`（只改 swag 註解的 `@Failure` 說明） |
| 前端 | 新增 `frontend/lib/safe-url.mjs`、`frontend/lib/safe-url.test.mjs` |
| 前端 | `frontend/components/ui.jsx`、`frontend/components/front/front-shell.jsx`、`frontend/components/admin/admin-posts.jsx` |
| 前端 | `frontend/scripts/design-sync.mjs`、`frontend/scripts/design-sync.test.mjs`、`frontend/design-sync/manifest.json`、`frontend/package.json`（`test:design` 加入 `lib/*.test.mjs`） |

- **不動** `backend/internal/model/**`、`backend/internal/database/**`、`backend/internal/seed/**`，沒有 schema 變更；不動 `.drone.yml`、部署設定。
- 與其他 issue 的衝突熱點：`frontend/design-sync/manifest.json`、`frontend/components/ui.jsx`（`/design-pull` 會改）。

### API

| 端點 | 變動 |
|------|------|
| `POST /api/v1/comments` | 回應欄位不變；`date` 的值改以台灣時間計算 |
| `PUT /api/v1/settings` | 新增 400 情境：`social[].url` 格式不合法 |

### 風險

依 `CLAUDE.md` 風險分級為**中**：動到前後端執行期程式碼，未命中高風險檔案樣式，非破壞性。
