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

## 第 2 輪 — PR 中 — 2026-11-15

- **範圍**：origin/main...HEAD（PR 中第 1 輪）
- **已審到**：941b98d7d2e9622012689a17c51de0b7b51d5218（收尾 HEAD）
- **審查指令**：
  - PR 中：CI 全綠（`continuous-integration/drone/pr` build 21）；`/code-review high /home/steven/side_projects/littleduck_website/wt/sdp-168-install-sites-cert-check origin/main...HEAD`；/security-review 已跑（`claude -p "/security-review"`，無發現）；reviewer 評定模式（已派）
- **高風險檔案**：`deploy/edge/install-sites.sh`、`deploy/edge/test-install-sites.sh`（`deploy/**`）
- **風險等級**（只有 PR 中）：高 — 命中高風險檔案樣式 `deploy/**`

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | BE | /code-review | 中 | 高 | `deploy/edge/install-sites.sh:25-35`；`deploy/edge/test-install-sites.sh:80-86`、`:114-122` | 安裝站台前在 Nginx 容器內檢查憑證 › 無法在容器內執行檢查；tasks 1.1、1.2；design 決策 2 | 同第 1 輪 #1，**已實機重現**：容器不存在、已停止、ID 多行時 `docker exec` 都回 1，與 `test -f` 不存在相同，會誤報「憑證不存在，請先依 runbook 核發」，該 Scenario 實際未達成；案例 3 回 126 未重現真實失敗。仍以非 0 中止、不 reload，只有訊息錯 | PR 備註：建議 `docker exec "$c" sh -c 'test -f "$1" \|\| exit 3' _ "$cert"`，只有 3 判為不存在；案例 3 改先 `docker stop` 假容器 |
| 2 | BE | /code-review | — | — | `deploy/edge/test-install-sites.sh:46` | — | 案例 3 假 docker 回 126 | 不成立：同第 1 輪 #1，併入本輪 #1 |
| 3 | BE | /code-review | — | — | `deploy/edge/install-sites.sh:16-18` | — | `docker ps -q` 多行 ID | 不成立：併入 #1；第 16 行未改動，中止結果與改動前相同 |
| 4 | BE | /code-review | 低 | 高 | `deploy/edge/install-sites.sh:23-24`、`:40` | 無 | `src_dir` 無 `littleduck-*.conf` 時 glob 字面傳給 grep，憑證檢查靜默跳過，最後在 `cp` 以費解訊息中止 | PR 備註：既有問題，建議另開 issue |
| 5 | BE | /code-review | 低 | 中 | `deploy/edge/test-install-sites.sh:23` | tasks 1.2 | dash 收到 SIGINT／SIGTERM 不跑 EXIT trap，Ctrl-C 會留下 mygo-v2 label 的假容器（下次執行會被開頭檢查擋下） | PR 備註：建議加 `trap 'exit 130' INT TERM` |
| 6 | BE | /code-review | — | — | `.drone.yml` | design Non-Goals | 測試腳本未接進 CI | 不成立：design 明訂不在範圍內 |
| 7 | BE | /code-review | — | — | `deploy/edge/install-sites.sh:22` | — | 事先檢查與 `nginx -t` 重疊 | 不成立：設計取捨，非錯誤行為 |
| 8 | BE | /code-review | — | — | `deploy/edge/install-sites.sh:25` | — | 每個憑證一次 docker exec | 不成立：效率類，design Risks 已接受 |
| 9 | BE | /code-review | — | — | `deploy/edge/install-sites.sh:25-29` | — | if/else 取 rc 寫法可簡化 | 不成立：簡化類，現寫法正確 |

- **結果**：通過

## 第 3 輪 — PR 前 — 2026-11-16

- **範圍**：941b98d7d2e9622012689a17c51de0b7b51d5218..HEAD（第 1 輪已審到的 7530bf4 已被收尾重切，內容等同收尾 HEAD 941b98d，故自 941b98d 起算）
- **已審到**：e41eb4e448c67ebc78355cde1ede92931e384dff
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：`deploy/edge/install-sites.sh`、`deploy/edge/test-install-sites.sh`（`deploy/**`）

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| — | — | — | — | — | — | — | 無新發現。第 2 輪 #1 已修正：容器內 `test -f "$1" \|\| exit 3`，只有 3 判為不存在；案例 3 以真實停止的容器重現 docker exec 失敗，斷言可擋住 `rc -eq 1` 舊寫法 | — |

- **結果**：通過
