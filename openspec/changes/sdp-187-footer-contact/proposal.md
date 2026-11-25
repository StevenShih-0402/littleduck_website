# Proposal：前台頁尾加上聯絡 Email（SDP-187）

## Why

申請綠界（ECPay）金流服務時，審核方要求網站上能看到站主的聯絡方式。前台目前只有版權字樣與後台可編輯的社群連結，沒有固定的聯絡資訊，審核方可能找不到。

## What Changes

- 前台頁尾（所有前台頁面共用的 `FrontShell`）新增一個固定的聯絡 Email：`steven0210423@gmail.com`，以 `mailto:` 連結呈現，標示為「聯絡／Contact」。
- 兩種主題（Magazine＝day、Terminal＝night）共用同一個頁尾，都顯示這個 Email；顏色沿用既有的主題 token，不另寫色值。
- 頁尾的大寫樣式（`uppercase`）不套用到 Email 本身，確保顯示為小寫原文，避免審核方抄錯。
- Email 寫成站台常數（`frontend/data/site.js` 的 `SITE`），不放在後台可編輯的網站設定裡，避免資料庫內容被誤刪或後端暫時無法回應時就消失（見 design.md D1）。
- 不加功能開關：這是靜態公開資訊，不是可操作的互動功能（見 design.md D3）。
- `frontend/design-sync/manifest.json` 在對應的設計檔下記錄兩筆偏離（`data/site.js`、`components/front/front-shell.jsx`），避免日後 `/design-pull` 移植時被覆蓋回設計稿。
- 不涉及後端、API 與資料表。非破壞性變更。

## Capabilities

### New Capabilities

- `site-contact-info`：前台公開聯絡資訊的呈現規則（頁尾 Email 的位置、連結形式、在兩種主題與各前台頁面皆可見）。

### Modified Capabilities

無。既有的 `front-content-links`（社群連結的網址分類）行為不變；`design-sync` 的偏離紀錄規則已存在，本次只是照規則新增紀錄。

## Impact

預計修改的檔案（全部在 `frontend/`，可據此判斷與其他 issue 能否並行）：

| 檔案 | 變更 |
|------|------|
| `frontend/data/site.js` | `SITE` 新增聯絡 Email 常數 |
| `frontend/components/front/front-shell.jsx` | 頁尾新增聯絡 Email 區塊 |
| `frontend/design-sync/manifest.json` | `project/data.js` 與 `project/v2/front-home.jsx` 底下各加一筆 `deviations`（SDP-187） |

- 不動：`backend/**`、API、`openapi.yaml`（`paths: {}`）、功能開關登記表、後台頁面。
- 會與同樣修改 `front-shell.jsx`、`data/site.js` 或 `manifest.json` 的 issue 衝突，這類 issue 不宜並行。
- 風險分級：不命中高風險檔案樣式；動到前端執行期程式碼，屬「中」。
