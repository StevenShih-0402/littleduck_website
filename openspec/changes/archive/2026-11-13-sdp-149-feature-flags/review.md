<!--
  review-template.md 範本：複製到 openspec/changes/<change-id>/review.md（一個 change 一份，前後端的發現都記在這裡，以「端」欄區分）。
  由 /ship-issue 主對話寫入（代 reviewer、內建 /code-review 與 /security-review）：
    - PR 前（第 5 步）：reviewer 審查模式，對照規格審 diff。
    - PR 中（第 7 步）：CI、/code-review、/security-review（高風險才跑），由 reviewer 評定模式評級。
  只追加：每輪審查加一節「第 N 輪」，不改寫前面的輪次。
  「已審到」是下一輪增量審查的起點：PR 前寫審查當下的 HEAD SHA；PR 中寫該輪的收尾 HEAD（handoff.md 收尾紀錄的值）。
  review-diff.patch 是審查用的暫存 diff，已列入 .gitignore，不進 commit。
-->
# 程式碼審查：sdp-149-feature-flags

## 第 1 輪 — PR 前 — 2026-11-13

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：8812e37ff54de56b12bc24cafe3834315208d5bc
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：`backend/internal/model/feature_flag.go`、`backend/internal/middleware/auth.go`（含測試）；PR 中 /code-review 用 high 並跑 /security-review

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | reviewer | 中 | 高 | `frontend/store/lake-context.jsx:45-71, 74-79, 197-203, 216`；`frontend/components/feature-gate.jsx:58-74`；`frontend/components/admin/admin-features.jsx:14-24` | feature-flags › 前台依開關顯示操作區塊（開關讀取失敗）、管理員預覽（權杖失效退出預覽）、後台功能開關管理介面（切換失敗、切換後前台反映） | 狀態層與元件行為沒有自動化測試（`/features` 失敗設 `{}`、`adminPreview` 判斷、401／logout 清 `featureList`、`setFeatureEnabled` 同步兩份狀態、切換失敗狀態不變），只測了 `lib/feature-flags.test.mjs` 的純函式；design.md 已說明改由審查涵蓋，逐一讀過邏輯與規格一致 | PR 備註 |
| 2 | FE | reviewer | 低 | 高 | `frontend/store/lake-context.jsx:77-79, 119-127`；`frontend/lib/api.js:41-44` | feature-flags › 管理員預覽 › 權杖失效退出預覽 | 管理員停在前台不重新整理、JWT 過期後，前台不再發後台請求，預覽標示與停用中表單持續顯示；此時送出留言，`OptionalAuth` 視為訪客回 403 `FEATURE_DISABLED`（非 401），前端不清權杖，畫面顯示「功能已停用」。不產生錯誤資料，重新整理即恢復 | PR 備註 |
| 3 | FE | reviewer | 低 | 高 | `frontend/store/lake-context.jsx:77-79`；`frontend/components/admin/admin-features.jsx:578` | feature-flags › 後台功能開關管理介面 | `GET /admin/features` 回 500 或網路失敗時 `featureList` 留在 `null` 且不寫錯誤，「功能開關」分頁一直顯示「載入中…」，無錯誤訊息與重試；規格未定義載入失敗 | PR 備註 |

- **結果**：通過

## 第 1 輪 — PR 中 — 2026-11-13

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：74faff8251a83f2bb98e244cc1bd2e29052004b5（收尾 HEAD）
- **審查指令**：
  - PR 中：CI 無 CI（PR #11 間隔 30 秒兩次查詢 statusCheckRollup 皆為空）；`/code-review high /home/steven/side_projects/littleduck_website/wt/sdp-149-feature-flags origin/main...HEAD`；/security-review 以 /code-review high 代替：`claude -p "/security-review"` 執行失敗（本機無 `origin/HEAD` ref，`git diff origin/HEAD...` 報 unknown revision）；reviewer 評定模式（已派）
- **高風險檔案**：`backend/internal/middleware/auth.go`、`backend/internal/middleware/auth_test.go`、`backend/internal/model/feature_flag.go`
- **風險等級**（只有 PR 中）：高 — 命中高風險檔案樣式「資料表結構」（`backend/internal/model/**`）與「認證與權限」（`backend/internal/middleware/**`）；非破壞性 API 變更；無成立的安全發現

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | BE | /code-review | — | — | `backend/internal/service/comment_service.go:73-81, 98-102` | comments-api › 訪客張貼留言 | guestbook 停用、article_comments 啟用時，帶不存在的 articleId 可通過 article_comments 檢查而寫入 | 不成立：依規格 Scenario 以 articleId 是否為空決定檢查哪個開關；留言板只顯示 `!c.articleId` 的留言（`guest-page.jsx:20`），帶假 articleId 的留言不會出現在留言板，且為 pending；不檢查文章存在是既有行為 |
| 2 | FE | /code-review | — | — | `frontend/store/lake-context.jsx:216` | feature-flags › 管理員預覽 › 權杖失效退出預覽 | JWT 過期後預覽仍顯示，送出得 403 而非 401 | 不成立：重複，已記錄（PR 前第 1 輪 #2） |
| 3 | FE | /code-review | 低 | 中 | `frontend/store/lake-context.jsx:77-79, 102-108` | feature-flags › 管理員預覽 | `/admin/features` 請求不可取消，登出後晚到的回應填回 `featureList`；同分頁重新登入時新請求回來前短暫顯示舊清單（登出時 `authed=false`，不影響訪客） | PR 備註 |
| 4 | FE | /code-review | — | — | `frontend/components/admin/admin-features.jsx:34` | feature-flags › 後台功能開關管理介面 | 載入失敗時一直顯示「載入中…」 | 不成立：重複，已記錄（PR 前第 1 輪 #3） |
| 5 | BE | /code-review | — | — | `backend/internal/service/comment_service.go:74-81` | comments-api › 留言內容缺漏／開關停用 | 開關停用時送不完整 body 得 400 而非 403 | 不成立：規格未規定 400／403 優先順序，兩者皆不建立留言 |
| 6 | BE | /code-review | 低 | 中 | `backend/internal/service/feature_service.go:97-101` | feature-flags › 後台功能開關管理介面 | 沒有資料列的 key 回零值 `updated_at`（0001-01-01T00:00:00Z）；啟動時 `EnsureDefaults` 會補齊，只有執行期間手動刪列才會觸發 | PR 備註 |
| 7 | FE | /code-review | 低 | 高 | `frontend/store/lake-context.jsx:48-59` | feature-flags › 前台依開關顯示操作區塊 | `loadFeatures` 放在同一個 `Promise.all`，`/features` 慢時拖住整頁 `loading`，與註解「另外請求、不影響其他資料」不完全一致 | PR 備註 |
| 8 | FE | /code-review | 低 | 高 | `frontend/components/feature-gate.jsx:20-29` | 無 | `inline` prop 為死碼，三個呼叫端都未傳入，JSDoc 易誤導 | PR 備註 |
| 9 | BE | /code-review | — | — | `backend/cmd/server/main.go:50` | 無 | main.go 與 router.New 各建一份 FeatureService／Repository | 不成立：兩者無狀態，與既有 Article／Comment／Setting／Donation 寫法一致，屬簡化建議 |
| 10 | FE | /code-review | — | — | `frontend/store/lake-context.jsx:198` | 無 | `setFeatureEnabled` 組 URL 未 encodeURIComponent | 不成立：key 只來自後端登記表，僅含 `[a-z_]`，無觸發條件 |

- **結果**：通過
