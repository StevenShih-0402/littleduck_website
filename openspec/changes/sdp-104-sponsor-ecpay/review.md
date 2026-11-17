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

## 第 2 輪 — PR 中 — 2026-11-17

- **範圍**：origin/main...HEAD（PR 中第 1 輪）
- **已審到**：e74703f9016bd3b70464102201f4dc41cb9470e1（收尾 HEAD）
- **審查指令**：
  - PR 中：CI 全綠（`continuous-integration/drone/pr`，Drone 建置 #35；先前 Drone webhook 因 OAuth token 失效逾時，user 重新登入 Drone 後恢復）；`/code-review high /home/steven/side_projects/littleduck_website/wt/sdp-104-sponsor-ecpay origin/main...HEAD`；/security-review 已跑（`claude -p "/security-review"`，無高信心發現）；reviewer 評定模式（已派）
- **高風險檔案**：`backend/docker-compose.yml`、`backend/internal/controller/payment_controller.go`、`backend/internal/model/donation.go`、`backend/internal/service/payment/checkmac.go`、`backend/internal/service/payment/ecpay.go`、`backend/internal/service/payment/payment.go`（及同目錄測試檔）
- **風險等級**（只有 PR 中）：高 — 命中高風險檔案樣式（金流、資料表結構、backend/docker-compose）；無破壞性 API 變更；/security-review 無成立發現

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | FE | /code-review | — | — | `frontend/components/front/sponsor-result-page.jsx:33`、`:47` | payment-integration › 前台贊助付款流程 › 缺少或錯誤的交易編號 | 交易不存在（404）被當成查詢失敗持續輪詢，到不了「付款未完成」 | 不成立：已於第 1 輪 PR 前 #1 記錄 |
| 2 | FE | /code-review | 低 | 高 | `frontend/components/front/sponsor-result-page.jsx:47`、`:60-65`、`:74-76` | payment-integration › 前台贊助付款流程 › 通知尚未抵達 | 輪詢用完仍 pending 時 view 停在 pending，同時顯示「付款處理中…請稍候」與「請稍後重新整理本頁確認」，文案矛盾 | PR 備註（建議：exhausted 時改用 unknown 的標題） |
| 3 | FE | /code-review | 中 | 高 | `frontend/components/admin/admin-sponsor.jsx:66`、`:75` | payment-integration › 後台贊助統計只計入已付款 | 「本月」KPI 寫死 `2026 / 07`，真實贊助進來後本月統計永遠錯（既有問題，本次使其生效） | PR 備註（建議：改用台北時區當月 `yyyy / MM`） |
| 4 | BE | /code-review | 中 | 中 | `backend/internal/service/payment_service.go:115-167`、`backend/internal/router/router.go:100` | payment-integration › 建立贊助付款 | 公開建立付款端點無速率限制、pending 無清理，可被灌爆紀錄（全專案皆無限流，屬既有設計缺口） | PR 備註（建議：另開 issue 處理限流與過期 pending 清理） |
| 5 | BE | /code-review | 低 | 高 | `backend/internal/service/payment_service.go:303-312` | payment-integration › 查詢付款狀態 | 公開狀態查詢接受任意 id，舊紀錄 `d1`～`d6` 可被列舉出金額與狀態（不洩漏姓名、留言） | PR 備註（建議：只接受 `LD` 開頭 20 碼編號，其他回 404） |
| 6 | FE | /code-review | 低 | 高 | `frontend/components/admin/admin-sponsor.jsx:97-98`、`frontend/components/front/sponsor-fab.jsx:39-40`、`:89-96` | payment-integration › 建立贊助付款 › 金額不合法 | 後台檔位編輯器未檢查 10～20000，設錯的檔位前台無法使用 | PR 備註（建議：後台儲存時套用 `validateSponsorAmount`） |
| 7 | BE | /code-review | 低 | 高 | `backend/internal/service/payment_service.go:250-253`、`backend/internal/config/config.go:62-65` | payment-integration › 綠界付款結果通知 | 設定不完整（含驗章用不到的 ReturnURL／ClientBackURL）時 callback 一律回 `0|CheckMacValue Error`，誤導為簽章錯誤，進行中訂單停在 pending | PR 備註（建議：callback 只檢查 MerchantID／HashKey／HashIV，並回可辨識的原因字串） |
| 8 | BE | /code-review | — | — | `backend/internal/service/payment/ecpay.go:27`、`backend/internal/service/payment_service.go:296` | 無 | taipeiZone 與日期格式字面值重複 | 不成立：重用／風格類，沒有錯誤行為 |
| 9 | BE | /code-review | — | — | `backend/internal/service/payment_service.go:125-127`、`backend/internal/service/payment/ecpay.go:55` | 無 | 設定檢查在 service 與 provider 各做一次 | 不成立：簡化類，沒有錯誤行為 |
| 10 | FE | /code-review | — | — | `frontend/components/front/sponsor-fab.jsx:79-111` | 無 | 無作用的 Fragment 包裝 | 不成立：純風格，交給 linter |

- **結果**：通過
