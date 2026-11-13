# Proposal：功能開關（SDP-149）

## Why

SDP-149 要求仿照 Anthropic 新功能發布的做法：每個功能都有一個功能開關，後台啟用後一般瀏覽者才看得到、用得到對應的操作區塊。目前網站沒有任何開關機制，前台的贊助按鈕、留言表單都是寫死顯示，要暫停某個功能只能改程式碼重新部署。

## What Changes

- 後端新增「功能開關」資料表（GORM AutoMigrate）與程式碼內的開關登記表（registry）；啟動時把登記表中尚未存在的開關補進資料表，已存在的開關值不覆寫。
- 新增 3 個端點（皆為新增，**非破壞性**）：
  - `GET /api/v1/features`（公開）：回傳各開關是否啟用的對照表。
  - `GET /api/v1/admin/features`（需管理員 JWT）：回傳開關清單（含名稱、說明、更新時間）。
  - `PUT /api/v1/features/{key}`（需管理員 JWT）：切換單一開關。
- 第一批套用開關的功能（user 已確認）：
  | key | 功能 | 前台受影響的操作區塊 |
  |-----|------|---------------------|
  | `sponsor` | 贊助 | 右下角贊助浮動按鈕與贊助視窗 |
  | `guestbook` | 留言板投書 | `/guest` 的投書表單（已通過的留言列表照常顯示） |
  | `article_comments` | 文章留言 | 文章頁留言區的留言表單（已通過的留言照常顯示） |
- 只藏操作區塊（表單、按鈕）：導覽列「留言板」連結、首頁「簽到留言」按鈕、已公開留言列表照常顯示。
- 既有功能的開關預設為**啟用**，上線後前台外觀不變；之後新增的功能預設為**停用**。
- 後端同步把關：開關停用時 `POST /api/v1/comments` 回 403 `FEATURE_DISABLED`（留言板看 `guestbook`、文章留言看 `article_comments`），避免只藏 UI 卻仍能直接呼叫 API。贊助金流端點本次不動（留給 SDP-104）。
- **管理員預覽**：管理員在後台登入後（同一分頁的有效 JWT），前台仍顯示停用中的操作區塊並可試用，區塊加上「停用中・預覽」標示；帶有效 JWT 的 `POST /api/v1/comments` 不受開關限制（後端以選用式驗證判斷，權杖無效或過期時視同訪客、仍受開關限制）。
- 前端：資料層載入開關、提供 `FeatureGate` 元件包住受控的操作區塊（啟用→顯示、停用且為管理員預覽→加標示顯示、其他→隱藏）；開關讀取失敗時一律視為停用（fail-closed），頁面不崩潰。
- 後台新增「功能開關」分頁，列出全部開關並可逐一切換，切換後前台立即反映。
- 「新功能都要掛功能開關」寫入 repo 根目錄 `CLAUDE.md` 作為後續開發原則——**不在本提案的 Agent 寫入範圍**，見下方「CLAUDE.md 開發原則」。

## Capabilities

### New Capabilities
- `feature-flags`：功能開關的登記、讀取、後台切換，以及前台依開關顯示或隱藏操作區塊的行為。

### Modified Capabilities
- `comments-api`：「訪客張貼留言」增加開關停用時拒絕張貼的情境。

## Impact

- **API**：新增 3 個端點；`POST /api/v1/comments` 新增 403 回應（僅在開關被停用且非管理員時發生，預設啟用下行為不變），並開始讀取選用的 `Authorization` 標頭（無標頭或權杖無效時行為與現在相同，不回 401）。無破壞性變更。
- **資料表**：新增 `feature_flags` 表（相容的 schema 新增，空資料庫與既有資料庫皆可啟動）。
- **風險等級**：命中高風險檔案樣式 `backend/internal/model/**`（新增 model）與 `backend/internal/middleware/**`（新增選用式驗證），整體判為**高風險**。
- **預計修改的檔案範圍**（供判斷能否與其他 issue 並行）：
  - 後端：
    - 新增 `backend/internal/model/feature_flag.go`
    - 新增 `backend/internal/repository/feature_repository.go`
    - 新增 `backend/internal/service/feature_service.go`（含開關登記表）與 `_test.go`
    - 新增 `backend/internal/controller/feature_controller.go` 與 `_test.go`
    - 修改 `backend/internal/service/comment_service.go`、`comment_service_test.go`（注入開關檢查、管理員預覽放行）
    - 修改 `backend/internal/middleware/auth.go`、`auth_test.go`（新增 `OptionalAuth`）
    - 修改 `backend/internal/controller/comment_controller.go`、`comment_controller_test.go`（讀取管理員身分傳給 service）
    - 修改 `backend/internal/pkg/apperror/apperror.go`（新增 403 `FEATURE_DISABLED`）
    - 修改 `backend/internal/router/router.go`、`router_test.go`
    - 修改 `backend/internal/seed/seed.go`（補齊開關）、`backend/cmd/server/main.go`（若 seed 需傳入新 repository）
    - 重新產生 `backend/docs/`（swag）；修改 `backend/README.md` 端點表
  - 前端：
    - 新增 `frontend/lib/feature-flags.mjs` 與 `frontend/lib/feature-flags.test.mjs`
    - 新增 `frontend/components/feature-gate.jsx`、`frontend/components/admin/admin-features.jsx`
    - 修改 `frontend/store/lake-context.jsx`
    - 修改 `frontend/components/front/front-shell.jsx`、`guest-page.jsx`、`article-page.jsx`、`sponsor-fab.jsx`（預覽標示）
    - 修改 `frontend/components/admin/admin-shell.jsx`
    - 修改 `frontend/design-sync/manifest.json`（登記上述設計稿對應檔的 deviations）
  - 與其他 issue 的衝突熱點：`router.go`、`lake-context.jsx`、`admin-shell.jsx`、`front-shell.jsx`、`manifest.json`。

## CLAUDE.md 開發原則（交主對話處理）

repo 根目錄 `CLAUDE.md` 不在 architect（`openspec/`）、後端（`backend/`）、前端（`frontend/`）任何一方的寫入範圍，本提案**不排 [BE]／[FE] 任務**。user 已同意由主對話在後端、前端完成後，於同一分支另一個 commit 寫入以下段落（放在「## 審查規則三分法」之前），並依實際檔名與做法調整文字：

```markdown
## 功能開關（SDP-149）

- 前台新增一般瀏覽者可操作的功能（表單、按鈕、互動區塊）時，MUST 同時：
  1. 在後端功能開關登記表（`backend/internal/service/feature_service.go`）新增一個 key，新功能預設**停用**；
  2. 前端以 `FeatureGate`（`frontend/components/feature-gate.jsx`）包住該操作區塊：啟用時顯示；停用時一般瀏覽者看不到；
  3. 若該功能有公開寫入端點，後端在開關停用時回 403 `FEATURE_DISABLED`，但帶有效管理員 JWT 的請求放行（管理員預覽）。
- 只藏操作區塊，不藏導覽與既有公開內容。
- 管理員預覽：管理員在後台登入後，同一分頁的前台會顯示停用中的區塊並標示「停用中・預覽」，可實際試用；確認無誤後到後台「功能開關」分頁啟用才對外開放。
- 開關讀取失敗一律視為停用（fail-closed），頁面其他內容照常顯示。
- 功能穩定、不再需要開關時，另開 issue 移除開關與相關判斷，不留永久開關。
```
