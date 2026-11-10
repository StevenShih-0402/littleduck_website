# Proposal

## Why

littleduck_website 與 code_city 兩個專案的多 Agent 團隊目前沿用 agent-team 範本的同一組名字（hyuse／chika／yuma／osamu／shiori），而且只有 shiori 設了顏色。兩個專案同時開著時，分不出是哪個專案的哪個 Agent 在跑。SDP-88 要讓兩個專案各有一組自己的名字，並為每個 Agent 指定顯示顏色。

## What Changes

**littleduck_website**（角色不變，只改名字與顏色）：

| 角色 | 舊名 | 新名 | 顏色 |
|------|------|------|------|
| architect | hyuse | hikaru | purple |
| backend | chika | yasumi | pink |
| frontend | yuma | fujimori | green |
| integrator | osamu | teshima | cyan |
| reviewer | shiori | akafuku | yellow |

**code_city**（沒有 backend 角色）：

| 角色 | 舊名 | 新名 | 顏色 |
|------|------|------|------|
| architect | hyuse | gintoki | cyan |
| frontend | yuma | kagura | red |
| integrator | osamu | shinpachi | blue |
| reviewer | shiori | sadaharu | pink |

- **BREAKING**：Agent 定義檔改名（例：`.claude/agents/hyuse.md` → `hikaru.md`），frontmatter 的 `name` 隨之改變，`subagent_type: "hyuse"` 等舊名呼叫方式失效；`/ship-issue` 改用新名派工。
- 每個 Agent 定義檔的 frontmatter 加上（或改成）指定的 `color`。
- `guard-write-scope.sh` 的角色參數改成新名（舊名不再接受），Agent 定義檔掛 hook 時傳入新名。
- 兩個專案中提到舊名的骨架檔與已安裝的 skills 一律改成新名：`CLAUDE.md`、`.claude/commands/ship-issue.md`、`docs/multi-agent-workflow.md`、`openspec/review-template.md`、`spec-sync/README.md`、`.gitignore` 註解、`.claude/skills/**`（含 hooks 註解；code_city 另有 `gen-story`）。
- code_city 的主規格 `openspec/specs/agent-ci-gate/spec.md` 直接改用新名（code_city 不在本 repo 的歸檔同步範圍）。
- **不改**：`openspec/changes/archive/**` 的歷史紀錄；littleduck-skills 的 skill 來源。

## Capabilities

### New Capabilities
- `agent-roster`：本專案多 Agent 團隊的名冊。規定每個角色的 Agent 名稱、顯示顏色、定義檔路徑，以及寫入限制 hook 認得的角色參數。

### Modified Capabilities
（無。`agent-ci-gate` 的需求文字沒有提到 Agent 名稱，行為不變。）

## Impact

- **littleduck_website**：`.claude/agents/*`（5 檔改名＋改內容）、`.claude/hooks/guard-write-scope.sh`、`.claude/commands/ship-issue.md`、`.claude/skills/{issue-architect,backend-processor,frontend-processor,issue-integrator,custom-review}/**`、`CLAUDE.md`、`docs/multi-agent-workflow.md`、`openspec/review-template.md`、`spec-sync/README.md`、`.gitignore`。命中高風險樣式 `.claude/hooks/**`、`.claude/agents/**`，`/ship-issue` 會判定為**高風險**。
- **code_city**（跨 repo）：同類檔案，外加 `gen-story` skill 與 `openspec/specs/agent-ci-gate/spec.md`。要在 code_city 另開分支並開自己的 PR。
- **已知風險**：之後在任一專案重跑 littleduck-skills 的 `install.sh agent-team .`／`install.sh backend-processor .`，skills 會被覆寫回舊名；骨架檔不受影響。
- 執行期程式碼（`backend/`、`frontend/`、code_city 的 `src/`）、API 都不受影響。
