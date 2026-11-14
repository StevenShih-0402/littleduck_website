<!--
  review-template.md 範本：複製到 openspec/changes/<change-id>/review.md（一個 change 一份，前後端的發現都記在這裡，以「端」欄區分）。
  由 /ship-issue 主對話寫入（代 reviewer、內建 /code-review 與 /security-review）：
    - PR 前（第 5 步）：reviewer 審查模式，對照規格審 diff。
    - PR 中（第 7 步）：CI、/code-review、/security-review（高風險才跑），由 reviewer 評定模式評級。
  只追加：每輪審查加一節「第 N 輪」，不改寫前面的輪次。
  「已審到」是下一輪增量審查的起點：PR 前寫審查當下的 HEAD SHA；PR 中寫該輪的收尾 HEAD（handoff.md 收尾紀錄的值）。
  review-diff.patch 是審查用的暫存 diff，已列入 .gitignore，不進 commit。
-->
# 程式碼審查：sdp-157-frontend-self-host

## 第 1 輪 — PR 前 — 2026-11-14

- **範圍**：origin/main...HEAD（第 1 輪）
- **已審到**：24b8559
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：`.drone.yml`、`deploy/**`（`deploy/ci/*`、`deploy/edge/web*.conf.template`、`deploy/web/docker-compose.*.yml`）、`frontend/Dockerfile`

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | BE | reviewer | 中 | 高 | `docs/deployment-runbook.md` 第 3 節步驟 2（DNS）、4（憑證）、7（第一次部署） | tasks 3.3；frontend-delivery › Production 需人工核准 › 未核准 | runbook 在步驟 2 就把 apex `stevenshih-littleduck.com` 的 A 記錄指到 Hetzner，並移除舊託管的記錄。但正式站 Nginx 設定要到步驟 7 合併、第一次 promote 後才放上主機，這段期間正式網域會打到 mygo 的 Nginx，憑證也不符，正式站停擺。 | PR 備註：建議先只切 `staging.`，apex 的 DNS 切換、核發憑證、第一次 promote 排成同一段連續操作並寫明預期停機；或改用 DNS-01 先核發憑證 |
| 2 | 需求 | reviewer | 中 | 高 | `openspec/changes/sdp-157-frontend-self-host/design.md:205`；`deploy/ci/verify-commit-status.sh:11-20`、`.drone.yml` 的 `cd-backend-production` verify-build | design D5；frontend-delivery › Production 需人工核准 › 建置未全綠 | 依 design 的 Migration Plan 先合併、後設定主機時，主機設定完成前每次 push `main` 都會在 `cd-frontend-staging` 失敗，`continuous-integration/drone/push` 變 failure，後端 Production 無法 promote。design 寫「不影響後端 CD」，與 runbook 的順序（步驟 1–6 完成後才在步驟 7 合併）矛盾，runbook 也沒有寫出這個警告。 | PR 備註：修正 design 的 Migration Plan，並在 runbook 步驟 7 前補警告 |
| 3 | FE、BE | reviewer | 中 | 高 | `frontend/app/runtime-config.js/route.js:8-15`、`frontend/lib/api.js:7-11, 45`；`.drone.yml:180-188`、`frontend/Dockerfile:17-28` | frontend-data-layer › 後端位址於執行期決定 › 不被快取／執行期有設定；ci-pipeline › 前端 CI 門檻 › 映像有高風險弱點；tasks 4.2、4.3 | route 的 `no-store`、Content-Type、`force-dynamic`，以及 `api.js` 讀 `window.__LD_RUNTIME_CONFIG__` 的串接都沒有自動化測試，只手動 curl 驗證過；日後退回模組層級常數或拿掉 `no-store` 時 `test:design` 照樣通過。Trivy 掃描從未實跑過，要到第一次 PR CI 才知道結果。 | PR 備註：可把 `getBaseUrl` 抽到 `runtime-config.mjs` 加測試，route 加呼叫 `GET()` 檢查標頭的 node:test；開 PR 後確認 ci-frontend 的 trivy 結果 |

- **結果**：通過
