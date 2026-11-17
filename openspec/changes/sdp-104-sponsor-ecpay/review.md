<!--
  review-template.md 範本：複製到 openspec/changes/<change-id>/review.md（一個 change 一份，前後端的發現都記在這裡，以「端」欄區分）。
  由 /ship-issue 主對話寫入（代 reviewer、內建 /code-review 與 /security-review）：
    - PR 前（第 5 步）：reviewer 審查模式，對照規格審 diff。
    - PR 中（第 7 步）：CI、/code-review、/security-review（高風險才跑），由 reviewer 評定模式評級。
  只追加：每輪審查加一節「第 N 輪」，不改寫前面的輪次。
  「已審到」是下一輪增量審查的起點：PR 前寫審查當下的 HEAD SHA；PR 中寫該輪的收尾 HEAD（handoff.md 收尾紀錄的值）。
  review-diff.patch 是審查用的暫存 diff，已列入 .gitignore，不進 commit。
-->
# 程式碼審查：sdp-104-sponsor-ecpay

## 第 1 輪 — PR 前 — 2026-11-17

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：f856e14bfa4d11922a2268da906817c1606f1dec
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：`backend/internal/service/payment/**`、`backend/internal/controller/payment_controller.go`、`backend/internal/model/donation.go`、`backend/docker-compose.yml`

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | reviewer | 中 | 高 | `frontend/components/front/sponsor-result-page.jsx:28-41`、`:47`；`frontend/lib/api.js:58-61` | payment-integration › 前台贊助付款流程 › 缺少或錯誤的交易編號 | `/sponsor/result?trade_no=NOPE` 後端回 404 → `request()` 的 Error 不帶狀態碼，結果頁當成查詢失敗、輪詢 8 次（約 51 秒）後顯示「還在確認付款結果」，而非 spec 要求的「付款未完成」；無測試涵蓋 404 | PR 備註（建議：`api.js` 的 Error 帶 `status`，404 時定為 missing 並停止輪詢） |
| 2 | FE | reviewer | 中 | 中 | `frontend/components/front/sponsor-fab.jsx:44`、`:75`、`:104-105` | payment-integration › 前台贊助付款流程 › 建立失敗（可重試） | 導向綠界後 `submitting` 不重設；訪客按「上一頁」且頁面由 bfcache 還原時，按鈕與遮罩都停用，全螢幕視窗關不掉，只能重新整理 | PR 備註（建議：監聽 `pageshow`，`event.persisted` 時重設 `submitting`） |
| 3 | FE | reviewer | 中 | 高 | `frontend/components/front/sponsor-fab.jsx`、`frontend/components/front/sponsor-result-page.jsx`、`frontend/components/admin/admin-sponsor.jsx:35`、`:65-77` | payment-integration › 前台贊助付款流程（導向綠界、建立失敗、付款完成返回、通知尚未抵達）；後台贊助統計只計入已付款 | 表單欄位與 `params` 一致、建立失敗後按鈕恢復、結果頁三種畫面與輪詢停止條件、KPI 只計 paid 都沒有自動化測試；專案目前沒有元件測試基礎設施 | PR 備註（建議：把 KPI 篩選與結果頁畫面判斷抽成 `lib/` 純函式並補測試） |

- **結果**：通過
