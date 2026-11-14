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

## 第 2 輪 — PR 中 — 2026-11-14

- **範圍**：origin/main...HEAD（PR 中第 1 輪）
- **已審到**：b9a32437ca7d7663a5f21dc96f89bb8f8c7c7a8e（收尾 HEAD）
- **審查指令**：
  - PR 中：CI 無 CI（PR #12 間隔 30 秒兩次查詢 statusCheckRollup 皆為空，Drone 尚未接上，見 SDP-117）；`/code-review high /home/steven/side_projects/littleduck_website/wt/sdp-157-frontend-self-host origin/main...HEAD`；/security-review 已跑（`claude -p "/security-review"`，無發現）；reviewer 評定模式（已派）
- **高風險檔案**：`.drone.yml`、`deploy/ci/frontend-bundle.sh`、`deploy/ci/frontend-version.sh`、`deploy/ci/image-ref.sh`、`deploy/ci/remote-deploy.sh`、`deploy/ci/smoke-frontend.sh`、`deploy/ci/vercel-deploy.sh`（刪除）、`deploy/edge/render.sh`、`deploy/edge/web-staging.conf.template`、`deploy/edge/web.conf.template`、`deploy/web/docker-compose.prod.yml`、`deploy/web/docker-compose.staging.yml`、`frontend/Dockerfile`
- **風險等級**（只有 PR 中）：高 — 命中高風險檔案樣式（CI／CD 與部署：`.drone.yml`、`deploy/**`、`**/Dockerfile`）；無破壞性 API 變更；/security-review 無成立發現

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | BE | /code-review | — | — | `deploy/ci/remote-deploy.sh:31`、`deploy/edge/install-sites.sh:28-38` | design.md:10 | 前後端 pipeline 平行呼叫 install-sites.sh 互相還原設定 | 不成立：`deploy/drone/docker-compose.yml:47` 設 `DRONE_RUNNER_CAPACITY: "1"` 且只有一個 runner，pipeline 依序執行 |
| 2 | FE | /code-review | 中 | 高 | `frontend/Dockerfile:5,10,17`；`.drone.yml:135-164` | proposal.md:11、design D1 | runtime 映像用 `node:20-alpine`，Node 20 已於 2026-04-30 EOL，無上游安全修補；版本為提案指定 | PR 備註：另開 issue 把前端映像與 CI 升到 Node 22/24 LTS |
| 3 | FE、BE | /code-review | 中 | 高 | `frontend/app/runtime-config.js/route.js:9`、`frontend/lib/runtime-config.mjs:29-31`；`deploy/web/docker-compose.prod.yml:20-26`、`docker-compose.staging.yml` healthcheck；`deploy/ci/remote-deploy.sh:29-31` | frontend-data-layer › 後端位址於執行期決定 › 未設定；design.md:194 | 主機 `.env` 缺少或設錯 `API_BASE_URL` → `apiBaseUrl:null`，healthcheck 只看 200 仍判 healthy、Nginx 照切；smoke 失敗但不回滾，壞容器已接流量 | PR 備註：healthcheck 改為斷言 `apiBaseUrl` 非 null（例：`wget -qO- … \| grep -q '"apiBaseUrl":"http'`） |
| 4 | FE | /code-review | 中 | 高 | `frontend/app/layout.jsx:6,48-49`；`frontend/design-sync/manifest.json:14-19,74-95` | CLAUDE.md「設計同步」 | `layout.jsx` 是設計同步目標，新加的 `<Script src="/runtime-config.js" beforeInteractive>` 未記為 deviation，之後 `/design-pull` 可能覆蓋，瀏覽器退回 `localhost:8080`，smoke 抓不到 | PR 備註：在 manifest 對應項目補 `issue: SDP-157` 的 deviation |
| 5 | FE | /code-review | — | — | `frontend/app/runtime-config.js/route.js:13` | frontend-data-layer › 不被快取 | no-store 使每次載入多一個阻塞請求 | 不成立：spec 要求 MUST NOT 被快取，design.md:195 已評估代價 |
| 6 | BE | /code-review | 低 | 高 | `.drone.yml:158-164,172-178,190-196,398-406` | 無 | 每次 push `main` 前端建置三次，CI 映像未重用 | PR 備註 |
| 7 | BE | /code-review | — | — | `deploy/ci/frontend-bundle.sh:1` | 無 | 與 `backend-bundle.sh` 幾乎逐行重複 | 不成立：重用類意見，兩支腳本各自正確 |
| 8 | BE | /code-review | — | — | `.drone.yml:406,484` | design.md:10 | 兩次快速合併時共用的可變 tag 互相干擾 | 不成立：runner 容量 1，pipeline 依序執行；與後端既有模式相同 |
| 9 | FE | /code-review | — | — | `frontend/next.config.mjs:7`；CLAUDE.md「執行驗證」 | design.md「執行驗證觀察點」 | 通用執行驗證用 `next start`，未驗到出貨產物 | 不成立：本次專屬觀察點以 3101 埠跑真實 standalone 映像並驗證 runtime-config 與 smoke，handoff 已記錄實跑通過 |

- **結果**：通過
