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
