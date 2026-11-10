# Design

## Context

改名範圍與名稱對照見 `proposal.md`，littleduck_website 的行為要求見 `specs/agent-roster/spec.md`。目前狀況：

- 兩個專案的 Agent 名稱寫死在定義檔檔名、frontmatter `name`、`guard-write-scope.sh` 的 `case`、`/ship-issue` 的 `subagent_type`，以及大量文件與 skill 的說明文字裡（littleduck_website 約 20 個檔、code_city 約 20 個檔，排除 archive）。
- 顏色只有 `shiori.md` 設了 `color: red`，其餘沒設。
- code_city 是另一個 repo（`/home/steven/side_projects/code_city`），目前停在 `feature/SDP-113-drone-cicd` 分支，工作樹乾淨。它沒有 backend 角色，`CLAUDE.md` 明寫「不使用 Chika」。
- skills 由 littleduck-skills 的 `install.sh` 安裝，重跑會覆寫；骨架檔（Agent 定義、`ship-issue.md`、hook、範本）不會被覆寫。

## Goals / Non-Goals

**Goals:**
- 兩個專案的現行檔案（排除 archive）中不再有對應的舊名稱，Agent 可用新名稱派工並顯示指定顏色。
- 寫入限制的行為（各角色可寫的目錄）完全不變，只換名稱。

**Non-Goals:**
- 不改 littleduck-skills 來源（user 決定）。
- 不改 `openspec/changes/archive/**` 的歷史紀錄。
- 不調整任何角色的職責、工具、模型或流程步驟。

## Decisions

1. **定義檔用 `git mv` 改名**，再改內容。保留 git 歷史，diff 也會顯示為 rename。
2. **文字替換分大小寫兩種形式**：小寫（`hyuse` → `hikaru`，用於 `name`、`subagent_type`、hook 參數、CSS class）與首字大寫（`Hyuse` → `Hikaru`，用於說明文字）。逐檔 `sed` 後再人工看 diff，避免誤換（例：其他單字裡剛好含 `yuma`）。替換完以 `grep -riE` 確認沒有殘留。
3. **code_city 的 `chika`／`Chika` 保持不變**：issue 沒有給 code_city 的 backend 新名，且 code_city 本來就不使用該角色；那些字樣只出現在通用說明（`docs/multi-agent-workflow.md`、`CLAUDE.md` 的「不使用 Chika」、`backend-processor` 等 skill、`gen-story` 的 `.chika` 樣式）。若改成 littleduck_website 的 `yasumi`，反而會把兩個專案的名冊混在一起。
4. **code_city 的 `gen-story` 樣式 class 一併改名**（`.hyuse` → `.gintoki` 等），`SKILL.md` 中列出的角色清單同步；`.chika` 保留。
5. **code_city 的主規格直接改**：`openspec/specs/agent-ci-gate/spec.md` 中的 `Yuma` 改為 `Kagura`。本 change 的歸檔只會同步 littleduck_website 的 `openspec/specs/`，碰不到 code_city，因此不寫 delta，直接在 code_city 的分支上改並在該 PR 說明。
6. **code_city 另開分支與 PR**：從 code_city 的 `origin/main` 開 `improvement/SDP-88-rename-agents`，只放 code_city 的改動；littleduck_website 的 PR 內文用連結交叉引用。兩個 PR 互不依賴，合併順序不限。
7. **顏色值**直接寫在 frontmatter 的 `color` 欄位，所有指定值都在 Claude Code 支援的顏色清單內（red、blue、green、yellow、purple、orange、pink、cyan）。

## Risks / Trade-offs

- [重跑 littleduck-skills 的 `install.sh` 會把 skills 覆寫回舊名] → 在兩個專案的 `CLAUDE.md`「Agent 與指令」段補一行提醒：重跑 install.sh 後要重新套用改名；Linear 回填也註明此風險。
- [正在進行中的舊 change 的 `handoff.md`／`review.md` 用舊名] → 目前兩個專案 `openspec/changes/` 只有 archive，沒有進行中的 change，不受影響。
- [Claude Code 已載入的 Agent 清單要重開 session 才更新] → 實作後重開 session，用 `/agents` 確認名稱與顏色。
- [命中高風險樣式（`.claude/agents/**`、`.claude/hooks/**`）] → 依 CLAUDE.md 規則停下等 user 看 diff 後回覆「合併」。

## Migration Plan

1. littleduck_website：在分支上改完並驗證（grep 無殘留、hook 手動測試、`/agents` 顯示），開 PR。
2. code_city：從 `origin/main` 開分支改完同樣驗證，開 PR。
3. 回復：revert 各自的 PR 即可，沒有資料或執行期狀態需要處理。
