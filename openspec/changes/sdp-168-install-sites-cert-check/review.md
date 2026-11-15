<!--
  review-template.md 範本：複製到 openspec/changes/<change-id>/review.md（一個 change 一份，前後端的發現都記在這裡，以「端」欄區分）。
  由 /ship-issue 主對話寫入（代 reviewer、內建 /code-review 與 /security-review）：
    - PR 前（第 5 步）：reviewer 審查模式，對照規格審 diff。
    - PR 中（第 7 步）：CI、/code-review、/security-review（高風險才跑），由 reviewer 評定模式評級。
  只追加：每輪審查加一節「第 N 輪」，不改寫前面的輪次。
  「已審到」是下一輪增量審查的起點：PR 前寫審查當下的 HEAD SHA；PR 中寫該輪的收尾 HEAD（handoff.md 收尾紀錄的值）。
  review-diff.patch 是審查用的暫存 diff，已列入 .gitignore，不進 commit。
-->
# 程式碼審查：sdp-168-install-sites-cert-check

## 第 1 輪 — PR 前 — 2026-11-15

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：7530bf4a496cedb5d656c13b97c7bbad2fae3536
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：`deploy/edge/install-sites.sh`、`deploy/edge/test-install-sites.sh`（`deploy/**`）

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | BE | reviewer | 中 | 高 | `deploy/edge/install-sites.sh:25-34`；`deploy/edge/test-install-sites.sh:80-86`、`:114-122`；design.md 決策 2 | 安裝站台前在 Nginx 容器內檢查憑證 › 無法在容器內執行檢查；tasks 1.1、1.2 | 只把回傳碼 1 判為「憑證不存在」，但 daemon 端錯誤（容器未執行、`No such container`、docker.sock 無權限）時 Docker CLI 也回 1，此時仍會誤報「憑證不存在，請先依 runbook 核發」。腳本照樣非 0 中止、站台目錄不變，只是訊息錯。測試案例 3 以 PATH 包裝直接回 126，沒有重現真實失敗方式。（依 Docker CLI 行為推論，未實機執行） | PR 備註：建議改由容器內輸出 `present`／`absent` 分流，其他一律「無法檢查」；案例 3 改為先 `docker stop` 假容器再跑 |
| 2 | BE | reviewer | 中 | 高 | `deploy/edge/test-install-sites.sh`（缺案例） | 安裝後驗證失敗即還原 › 設定語法錯誤 | delta spec 新增的「`nginx -t` 失敗即還原、不 reload」Scenario 沒有測試；`EDGE_SITES_DIR` 未掛進假容器，`nginx -t` 只驗證容器預設設定。還原邏輯本次未改。「憑證不存在」案例也沒驗證不 reload | PR 備註：把暫存站台目錄 bind mount 進假容器並加語法錯誤 conf 案例 |

- **結果**：通過
