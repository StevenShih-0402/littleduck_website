# agent-ci-gate Specification

## Purpose
讓 agent-team 的 `/ship-issue` 在開 PR 後等待 Drone CI 結果，CI 失敗就退回開發 Agent，CI 全綠才進入後續 PR 中檢查與合併。

## Requirements

### Requirement: 開 PR 後等待 CI
`/ship-issue` 在開 PR 後 SHALL 先等 PR 上的 CI check，等待上限取 team-hq `CLAUDE.md` 設定的值；所有 check 成功才進行後續的 PR 中檢查。

#### Scenario: CI 全綠
- **WHEN** PR 上的前後端 Drone check 都成功
- **THEN** 流程往下進行 `/code-review` 等 PR 中檢查

#### Scenario: CI 逾時
- **WHEN** 超過等待上限 check 仍未結束
- **THEN** 流程停下並回報 user，不合併

### Requirement: CI 失敗時退回對應的開發 Agent
任一 CI check 失敗時，`/ship-issue` SHALL 在 `handoff.md` 寫下失敗的 check 與連結，退回負責該端的開發 Agent；退回次數 MUST 與其他檢查的退回合併計算上限。

#### Scenario: 後端 check 失敗
- **WHEN** 後端 Drone check 失敗
- **THEN** `handoff.md` 多一筆退回紀錄，寫明 check 名稱與 Drone 建置連結，並退回後端開發 Agent

### Requirement: 合併前重跑 CI
模擬合併隊列 rebase 到最新 `main` 後，SHALL 再等一次 CI；CI 不是全綠 MUST NOT 合併。

#### Scenario: rebase 後 CI 失敗
- **WHEN** rebase 後重新觸發的 Drone check 失敗
- **THEN** 不合併 PR，流程停下交給 user

### Requirement: CI 設定檔視為高風險
變更命中 `.drone.yml`、`deploy/**` 或 `.claude/hooks/**`、`.claude/agents/**` 時，`/ship-issue` SHALL 判定為高風險，MUST 等 user 看過 diff 回覆「合併」才合併。

#### Scenario: 修改流水線
- **WHEN** PR 改到 `.drone.yml`
- **THEN** 流程在合併前停下，等 user 回覆「合併」
