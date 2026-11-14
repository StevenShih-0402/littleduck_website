# 程式碼審查：sdp-117-cicd-rollout

## 第 1 輪 — PR 前 — 2026-11-14

- **範圍**：origin/main...HEAD
- **已審到**：03fd8e829ce310e446698203aa2c41f4bb0988d0
- **審查指令**：
  - PR 前：reviewer 審查模式（已派）
- **高風險檔案**：`deploy/drone/.env.example`、`deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh`（命中 `deploy/**`）

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| — | — | — | — | — | — | — | 本輪無發現 | — |

- **備註**：design.md 執行驗證觀察點第 2 條在歸檔前必然命中主規格 `openspec/specs/backend-delivery/spec.md:20`，屬觀察點寫法所致、非實作錯誤；歸檔後須重跑確認為空。
- **結果**：通過

## 第 2 輪 — PR 中 — 2026-11-14

- **範圍**：origin/main...HEAD（PR 中第 1 輪）
- **已審到**：93f7559f042a8260109c3f1f2bfea4dd0f6fe114（收尾 HEAD）
- **審查指令**：
  - PR 中：CI 無 CI（PR #14 兩次查詢皆無 check，Drone 尚未上線）；`/code-review high /home/steven/side_projects/littleduck_website/wt/sdp-117-cicd-rollout origin/main...HEAD`；/security-review 已跑（無發現）；reviewer 評定模式（已派）
- **高風險檔案**：`deploy/drone/.env.example`、`deploy/ci/smoke-backend.sh`、`deploy/ci/smoke-frontend.sh`（命中 `deploy/**`）
- **風險等級**：高 — 命中高風險檔案樣式 `deploy/**`（實際只改範本值與註解）；無破壞性 API 變更；/security-review 無成立發現

| # | 端 | 來源 | 嚴重度 | 信心度 | 檔案:行號 | 對應需求 | 說明 | 處理 |
|---|----|------|--------|--------|-----------|---------|------|------|
| 1 | BE | /code-review | 中 | 高 | docs/deployment-runbook.md:42-43 | tasks 3.1、design D3 | 第 1 節第 3 步只複製 `docker-compose.yml` 到 `/opt/drone/`，接著 `cp .env.example .env` 會 No such file，預填的 `DRONE_SERVER_HOST` 到不了主機；應一併複製 `.env.example` | PR 備註 |
| 2 | BE | /code-review | 低 | 高 | docs/deployment-runbook.md:16 | tasks 1.1 | 對照表寫 Drone 主機名在 `deploy/drone/.env`，實際位置是主機的 `/opt/drone/.env` | PR 備註 |
| 3 | BE | /code-review | 低 | 中 | docs/cicd.md:132、docs/cicd.md:141 | 無 | 前端 CD 冒煙列仍寫「Staging 後端位址」「正式網域／正式後端位址」，與後端冒煙列改為實際主機名＋secret 名稱的寫法不一致 | PR 備註 |
| 4 | BE | /code-review | 低 | 高 | openspec/specs/frontend-delivery/spec.md:78-83 | proposal › frontend-delivery | delta 未改「前端可回滾到先前版本」，歸檔後情境仍寫「正式網域首頁回 200」 | PR 備註 |
| 5 | — | /code-review | — | — | design.md:172 | — | 觀察點第 2 條歸檔前必命中主規格 | 不成立：PR 前第 1 輪已記錄；docs／deploy／delta 範圍已確認無輸出，歸檔後重跑 |
| 6 | 流程 | /code-review | — | — | handoff.md:159 | 提案 Q3 | PR 以 `Fixes SDP-117` 引用，合併會轉 Done | 不成立：user 已決定合併後手動改回 In Progress |
| 7 | 流程 | /code-review | — | — | review.md:6 | — | PR 前「已審到」SHA 在重切後不存在 | 不成立：PR 中增量起點用收尾 HEAD，PR 前 SHA 不再使用 |

- **結果**：通過
